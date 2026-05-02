package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestConvertMinimal(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{
			"web": {Image: "registry/hello:1"},
		},
	}
	got, err := Convert(in, ConvertOptions{
		Environment: "dev",
		Namespace:   "openchoreo-data-plane",
		Project:     "openchoreo",
	})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	want := OpenChoreoComponent{
		APIVersion: "openchoreo.dev/v1alpha1",
		Kind:       "Component",
		Metadata:   ComponentMetadata{Name: "hello", Namespace: "openchoreo-data-plane"},
		Spec: ComponentSpec{
			WorkloadTemplate: WorkloadTemplate{
				Containers: []ContainerSpec{{Name: "web", Image: "registry/hello:1"}},
			},
			Environment: "dev",
			Owner:       ComponentOwner{Project: "openchoreo"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("\n got: %+v\nwant: %+v", got, want)
	}
}

func TestConvertImageOverride(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{"web": {Image: "registry/hello:latest"}},
	}
	got, err := Convert(in, ConvertOptions{
		Environment: "dev", Namespace: "ns", Project: "p", ImageRef: "registry/hello:abc123",
	})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if got.Spec.WorkloadTemplate.Containers[0].Image != "registry/hello:abc123" {
		t.Fatalf("image=%q want abc123 override", got.Spec.WorkloadTemplate.Containers[0].Image)
	}
}

func TestConvertVariablesBecomeEnv(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{
			"web": {Image: "i", Variables: map[string]string{"FOO": "bar"}},
		},
	}
	got, _ := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
	env := got.Spec.WorkloadTemplate.Containers[0].Env
	if len(env) != 1 || env[0].Name != "FOO" || env[0].Value != "bar" {
		t.Fatalf("env=%+v", env)
	}
}

func TestConvertSecretResourceBecomesSecretKeyRef(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{
			"web": {
				Image:     "i",
				Variables: map[string]string{"TOKEN": "${resources.example.password}"},
			},
		},
		Resources: map[string]ScoreResource{
			"example": {Type: "secret", Metadata: map[string]string{"name": "example-secret"}},
		},
	}
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	env := got.Spec.WorkloadTemplate.Containers[0].Env[0]
	if env.Name != "TOKEN" {
		t.Fatalf("env name=%q", env.Name)
	}
	if env.ValueFrom == nil || env.ValueFrom.SecretKeyRef == nil {
		t.Fatalf("want secretKeyRef, got %+v", env)
	}
	if env.ValueFrom.SecretKeyRef.Name != "example-secret" {
		t.Fatalf("secret name=%q", env.ValueFrom.SecretKeyRef.Name)
	}
	if env.ValueFrom.SecretKeyRef.Key != "password" {
		t.Fatalf("secret key=%q", env.ValueFrom.SecretKeyRef.Key)
	}
}

func TestConvertPortsMapToList(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{"web": {Image: "i"}},
		Service: &ScoreService{Ports: map[string]ScorePort{
			"http": {Port: 8080},
		}},
	}
	got, _ := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
	if len(got.Spec.WorkloadTemplate.Ports) != 1 {
		t.Fatalf("ports=%+v", got.Spec.WorkloadTemplate.Ports)
	}
	p := got.Spec.WorkloadTemplate.Ports[0]
	if p.Name != "http" || p.Port != 8080 {
		t.Fatalf("bad port: %+v", p)
	}
}

func TestConvertUnsupportedResourceTypeErrors(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "x"},
		Containers: map[string]ScoreContainer{"web": {Image: "i"}},
		Resources:  map[string]ScoreResource{"pg": {Type: "postgres"}},
	}
	if _, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"}); err == nil {
		t.Fatal("expected unsupported-type error")
	}
}

func TestConvertMultipleContainersSorted(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "multi"},
		Containers: map[string]ScoreContainer{
			"zzz": {Image: "zzz:1"},
			"aaa": {Image: "aaa:1"},
			"mmm": {Image: "mmm:1"},
		},
	}
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	names := make([]string, 0, len(got.Spec.WorkloadTemplate.Containers))
	for _, c := range got.Spec.WorkloadTemplate.Containers {
		names = append(names, c.Name)
	}
	want := []string{"aaa", "mmm", "zzz"}
	for i, n := range want {
		if names[i] != n {
			t.Fatalf("container[%d]=%q want %q (full order: %v)", i, names[i], n, names)
		}
	}
}

func TestConvertMultipleVariablesSorted(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "envs"},
		Containers: map[string]ScoreContainer{
			"web": {Image: "i", Variables: map[string]string{
				"ZULU": "3", "ALPHA": "1", "MIKE": "2",
			}},
		},
	}
	got, _ := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
	env := got.Spec.WorkloadTemplate.Containers[0].Env
	want := []string{"ALPHA", "MIKE", "ZULU"}
	for i, n := range want {
		if env[i].Name != n {
			t.Fatalf("env[%d]=%q want %q", i, env[i].Name, n)
		}
	}
}

func TestConvertEnvironmentResource(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "x"},
		Containers: map[string]ScoreContainer{
			"web": {Image: "i", Variables: map[string]string{
				"ENV_NAME": "${resources.env.value}",
			}},
		},
		Resources: map[string]ScoreResource{
			"env": {Type: "environment"},
		},
	}
	got, err := Convert(in, ConvertOptions{Environment: "staging", Namespace: "ns", Project: "p"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	env := got.Spec.WorkloadTemplate.Containers[0].Env[0]
	if env.Name != "ENV_NAME" || env.Value != "staging" {
		t.Fatalf("env=%+v want Name=ENV_NAME Value=staging", env)
	}
	if env.ValueFrom != nil {
		t.Fatalf("env.ValueFrom should be nil for environment resource, got %+v", env.ValueFrom)
	}
}

func TestConvertMissingResourceReferenceErrors(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "x"},
		Containers: map[string]ScoreContainer{
			"web": {Image: "i", Variables: map[string]string{
				"TOKEN": "${resources.nonexistent.key}",
			}},
		},
		Resources: map[string]ScoreResource{
			// nonexistent is NOT here
		},
	}
	_, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
	if err == nil {
		t.Fatal("expected error for missing resource reference")
	}
}

func TestConvertInlineResourceRefErrors(t *testing.T) {
	// Score-1: a value that embeds ${resources.X.Y} as a substring (rather
	// than being the whole value) is not supported. The converter must
	// reject it explicitly instead of silently passing the literal
	// "prefix-${resources.db.password}" through as an env var value.
	cases := []struct {
		name, value string
	}{
		{"prefix before ref", "prefix-${resources.db.password}"},
		{"suffix after ref", "${resources.db.password}-suffix"},
		{"ref in middle", "prefix-${resources.db.password}-suffix"},
		{"two refs", "${resources.db.user}:${resources.db.password}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := ScoreDocument{
				APIVersion: "score.dev/v1b1",
				Metadata:   ScoreMetadata{Name: "x"},
				Containers: map[string]ScoreContainer{
					"web": {Image: "i", Variables: map[string]string{"VAR": tc.value}},
				},
				Resources: map[string]ScoreResource{
					"db": {Type: "secret"},
				},
			}
			_, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
			if err == nil {
				t.Fatalf("expected error for inline resource ref in %q", tc.value)
			}
			if !strings.Contains(err.Error(), "VAR") {
				t.Errorf("error should name the variable VAR, got: %v", err)
			}
		})
	}
}

func TestConvertAnnotationsBecomeLabels(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata: ScoreMetadata{
			Name:        "annotated",
			Annotations: map[string]string{"team": "platform", "tier": "dev"},
		},
		Containers: map[string]ScoreContainer{"web": {Image: "i"}},
	}
	got, _ := Convert(in, ConvertOptions{Environment: "dev", Namespace: "ns", Project: "p"})
	if got.Metadata.Labels["team"] != "platform" || got.Metadata.Labels["tier"] != "dev" {
		t.Fatalf("labels=%+v", got.Metadata.Labels)
	}
}
