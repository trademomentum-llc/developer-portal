package main

import (
	"strings"
	"testing"
)

func TestConvertMinimalServiceReturnsComponentAndWorkload(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{
			"web": {Image: "registry/hello:1"},
		},
		Service: &ScoreService{Ports: map[string]ScorePort{
			"http": {Port: 8080},
		}},
	}

	got, err := Convert(in, ConvertOptions{
		Environment: "dev",
		Namespace:   "default",
		Project:     "default",
	})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("resources=%d want 2", len(got))
	}

	component, ok := got[0].(OpenChoreoComponent)
	if !ok {
		t.Fatalf("resource[0]=%T want OpenChoreoComponent", got[0])
	}
	if component.APIVersion != "openchoreo.dev/v1alpha1" || component.Kind != "Component" {
		t.Fatalf("bad component type: %+v", component)
	}
	if component.Metadata.Name != "hello" || component.Metadata.Namespace != "default" {
		t.Fatalf("bad component metadata: %+v", component.Metadata)
	}
	if component.Spec.Owner.ProjectName != "default" {
		t.Fatalf("projectName=%q", component.Spec.Owner.ProjectName)
	}
	if !component.Spec.AutoDeploy {
		t.Fatal("component should enable autoDeploy")
	}
	if component.Spec.ComponentType.Kind != "ClusterComponentType" {
		t.Fatalf("componentType.kind=%q", component.Spec.ComponentType.Kind)
	}
	if component.Spec.ComponentType.Name != "deployment/service" {
		t.Fatalf("componentType.name=%q", component.Spec.ComponentType.Name)
	}

	workload, ok := got[1].(OpenChoreoWorkload)
	if !ok {
		t.Fatalf("resource[1]=%T want OpenChoreoWorkload", got[1])
	}
	if workload.Metadata.Name != "hello-workload" || workload.Metadata.Namespace != "default" {
		t.Fatalf("bad workload metadata: %+v", workload.Metadata)
	}
	if workload.Spec.Owner.ProjectName != "default" || workload.Spec.Owner.ComponentName != "hello" {
		t.Fatalf("bad workload owner: %+v", workload.Spec.Owner)
	}
	if workload.Spec.Container.Image != "registry/hello:1" {
		t.Fatalf("image=%q", workload.Spec.Container.Image)
	}
	ep, ok := workload.Spec.Endpoints["http"]
	if !ok {
		t.Fatalf("endpoints=%+v", workload.Spec.Endpoints)
	}
	if ep.Type != "HTTP" || ep.Port != 8080 || ep.TargetPort != 0 {
		t.Fatalf("bad endpoint: %+v", ep)
	}
	if len(ep.Visibility) != 1 || ep.Visibility[0] != "external" {
		t.Fatalf("visibility=%+v", ep.Visibility)
	}
}

func TestConvertEndpointTargetPortOnlyWhenDifferent(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{"web": {Image: "registry/hello:1"}},
		Service: &ScoreService{Ports: map[string]ScorePort{
			"http": {Port: 80, TargetPort: 8080},
		}},
	}
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	workload := got[1].(OpenChoreoWorkload)
	ep := workload.Spec.Endpoints["http"]
	if ep.Port != 80 || ep.TargetPort != 8080 {
		t.Fatalf("bad endpoint: %+v", ep)
	}
}

func TestConvertWorkerInferenceWithoutServicePorts(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "worker"},
		Containers: map[string]ScoreContainer{"worker": {Image: "registry/worker:1"}},
	}
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	component := got[0].(OpenChoreoComponent)
	if component.Spec.ComponentType.Name != "deployment/worker" {
		t.Fatalf("componentType.name=%q", component.Spec.ComponentType.Name)
	}
	workload := got[1].(OpenChoreoWorkload)
	if len(workload.Spec.Endpoints) != 0 {
		t.Fatalf("worker endpoints=%+v want none", workload.Spec.Endpoints)
	}
}

func TestConvertComponentTypeAnnotationOverridesInference(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata: ScoreMetadata{
			Name: "webapp",
			Annotations: map[string]string{
				"pipeline.m2/component-type": "deployment/web-application",
			},
		},
		Containers: map[string]ScoreContainer{"web": {Image: "registry/web:1"}},
		Service: &ScoreService{Ports: map[string]ScorePort{
			"http": {Port: 8080},
		}},
	}
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	component := got[0].(OpenChoreoComponent)
	if component.Spec.ComponentType.Name != "deployment/web-application" {
		t.Fatalf("componentType.name=%q", component.Spec.ComponentType.Name)
	}
}

func TestConvertImageOverride(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{"web": {Image: "registry/hello:latest"}},
	}
	got, err := Convert(in, ConvertOptions{
		Environment: "dev", Namespace: "default", Project: "default", ImageRef: "registry/hello:abc123",
	})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	workload := got[1].(OpenChoreoWorkload)
	if workload.Spec.Container.Image != "registry/hello:abc123" {
		t.Fatalf("image=%q want abc123 override", workload.Spec.Container.Image)
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
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	env := got[1].(OpenChoreoWorkload).Spec.Container.Env
	if len(env) != 1 || env[0].Key != "FOO" || env[0].Value != "bar" {
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
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("resources=%d want 3", len(got))
	}
	secretReference, ok := got[1].(OpenChoreoSecretReference)
	if !ok {
		t.Fatalf("resource[1]=%T want OpenChoreoSecretReference", got[1])
	}
	if secretReference.Metadata.Name != "example-secret" || secretReference.Metadata.Namespace != "default" {
		t.Fatalf("bad secret reference metadata: %+v", secretReference.Metadata)
	}
	if secretReference.Spec.Template.Type != "Opaque" || secretReference.Spec.RefreshInterval != "1h" {
		t.Fatalf("bad secret reference spec: %+v", secretReference.Spec)
	}
	if len(secretReference.Spec.Data) != 1 {
		t.Fatalf("secret reference data=%+v", secretReference.Spec.Data)
	}
	if secretReference.Spec.Data[0].SecretKey != "password" {
		t.Fatalf("secretKey=%q", secretReference.Spec.Data[0].SecretKey)
	}
	if secretReference.Spec.Data[0].RemoteRef.Key != "apps/hello/dev/example-secret" {
		t.Fatalf("remote key=%q", secretReference.Spec.Data[0].RemoteRef.Key)
	}
	if secretReference.Spec.Data[0].RemoteRef.Property != "password" {
		t.Fatalf("remote property=%q", secretReference.Spec.Data[0].RemoteRef.Property)
	}
	env := got[2].(OpenChoreoWorkload).Spec.Container.Env[0]
	if env.Key != "TOKEN" {
		t.Fatalf("env key=%q", env.Key)
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

func TestConvertSecretReferenceMetadataOverrides(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "hello"},
		Containers: map[string]ScoreContainer{
			"web": {
				Image: "i",
				Variables: map[string]string{
					"PASSWORD": "${resources.db.password}",
					"USER":     "${resources.db.username}",
				},
			},
		},
		Resources: map[string]ScoreResource{
			"db": {
				Type: "secret",
				Metadata: map[string]string{
					"name":              "database-secret",
					"remoteRef.key":     "shared/database",
					"remoteRef.version": "v2",
				},
			},
		},
	}
	got, err := Convert(in, ConvertOptions{Environment: "staging", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	secretReference := got[1].(OpenChoreoSecretReference)
	if secretReference.Metadata.Name != "database-secret" {
		t.Fatalf("name=%q", secretReference.Metadata.Name)
	}
	if len(secretReference.Spec.Data) != 2 {
		t.Fatalf("data=%+v", secretReference.Spec.Data)
	}
	if secretReference.Spec.Data[0].SecretKey != "password" || secretReference.Spec.Data[1].SecretKey != "username" {
		t.Fatalf("data order=%+v", secretReference.Spec.Data)
	}
	for _, data := range secretReference.Spec.Data {
		if data.RemoteRef.Key != "shared/database" || data.RemoteRef.Version != "v2" {
			t.Fatalf("remote ref=%+v", data.RemoteRef)
		}
		if data.RemoteRef.Property != data.SecretKey {
			t.Fatalf("remote property=%q secretKey=%q", data.RemoteRef.Property, data.SecretKey)
		}
	}
}

func TestConvertMultipleContainersErrors(t *testing.T) {
	in := ScoreDocument{
		APIVersion: "score.dev/v1b1",
		Metadata:   ScoreMetadata{Name: "multi"},
		Containers: map[string]ScoreContainer{
			"aaa": {Image: "aaa:1"},
			"zzz": {Image: "zzz:1"},
		},
	}
	_, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err == nil {
		t.Fatal("expected multiple-container error")
	}
	if !strings.Contains(err.Error(), "multiple containers") {
		t.Fatalf("error=%v", err)
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
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	env := got[1].(OpenChoreoWorkload).Spec.Container.Env
	want := []string{"ALPHA", "MIKE", "ZULU"}
	for i, n := range want {
		if env[i].Key != n {
			t.Fatalf("env[%d]=%q want %q", i, env[i].Key, n)
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
	got, err := Convert(in, ConvertOptions{Environment: "staging", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	env := got[1].(OpenChoreoWorkload).Spec.Container.Env[0]
	if env.Key != "ENV_NAME" || env.Value != "staging" {
		t.Fatalf("env=%+v want Key=ENV_NAME Value=staging", env)
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
		Resources: map[string]ScoreResource{},
	}
	_, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err == nil {
		t.Fatal("expected error for missing resource reference")
	}
}

func TestConvertInlineResourceRefErrors(t *testing.T) {
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
			_, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
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
	got, err := Convert(in, ConvertOptions{Environment: "dev", Namespace: "default", Project: "default"})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	component := got[0].(OpenChoreoComponent)
	if component.Metadata.Labels["team"] != "platform" || component.Metadata.Labels["tier"] != "dev" {
		t.Fatalf("labels=%+v", component.Metadata.Labels)
	}
}
