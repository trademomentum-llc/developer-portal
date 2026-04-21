package main

import (
	"reflect"
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
		APIVersion: "core.choreo.dev/v1alpha1",
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
