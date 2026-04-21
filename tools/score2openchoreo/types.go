// Package main: score2openchoreo shared types. Score input and OpenChoreo
// Component output structs. No I/O.
package main

type ScoreDocument struct {
	APIVersion string                    `yaml:"apiVersion"`
	Metadata   ScoreMetadata             `yaml:"metadata"`
	Containers map[string]ScoreContainer `yaml:"containers"`
	Resources  map[string]ScoreResource  `yaml:"resources,omitempty"`
	Service    *ScoreService             `yaml:"service,omitempty"`
}

type ScoreMetadata struct {
	Name        string            `yaml:"name"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type ScoreContainer struct {
	Image     string                   `yaml:"image"`
	Command   []string                 `yaml:"command,omitempty"`
	Args      []string                 `yaml:"args,omitempty"`
	Variables map[string]string        `yaml:"variables,omitempty"`
	Resources *ScoreContainerResources `yaml:"resources,omitempty"`
}

type ScoreContainerResources struct {
	Requests *ScoreResourceQuantities `yaml:"requests,omitempty"`
	Limits   *ScoreResourceQuantities `yaml:"limits,omitempty"`
}

type ScoreResourceQuantities struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type ScoreResource struct {
	Type     string            `yaml:"type"`
	Class    string            `yaml:"class,omitempty"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
	Params   map[string]any    `yaml:"params,omitempty"`
}

type ScoreService struct {
	Ports map[string]ScorePort `yaml:"ports"`
}

type ScorePort struct {
	Port       int    `yaml:"port"`
	TargetPort int    `yaml:"targetPort,omitempty"`
	Protocol   string `yaml:"protocol,omitempty"`
}

type OpenChoreoComponent struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   ComponentMetadata `yaml:"metadata"`
	Spec       ComponentSpec     `yaml:"spec"`
}

type ComponentMetadata struct {
	Name      string            `yaml:"name"`
	Namespace string            `yaml:"namespace"`
	Labels    map[string]string `yaml:"labels,omitempty"`
}

type ComponentSpec struct {
	WorkloadTemplate WorkloadTemplate `yaml:"workloadTemplate"`
	Environment      string           `yaml:"environment"`
	Owner            ComponentOwner   `yaml:"owner"`
}

type WorkloadTemplate struct {
	Containers []ContainerSpec `yaml:"containers"`
	Ports      []PortSpec      `yaml:"ports,omitempty"`
}

type ContainerSpec struct {
	Name      string                  `yaml:"name"`
	Image     string                  `yaml:"image"`
	Command   []string                `yaml:"command,omitempty"`
	Args      []string                `yaml:"args,omitempty"`
	Env       []EnvVarSpec            `yaml:"env,omitempty"`
	Resources *ContainerResourcesSpec `yaml:"resources,omitempty"`
}

type EnvVarSpec struct {
	Name      string            `yaml:"name"`
	Value     string            `yaml:"value,omitempty"`
	ValueFrom *EnvVarSourceSpec `yaml:"valueFrom,omitempty"`
}

type EnvVarSourceSpec struct {
	SecretKeyRef *SecretKeySelectorSpec `yaml:"secretKeyRef,omitempty"`
}

type SecretKeySelectorSpec struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type ContainerResourcesSpec struct {
	Requests *ResourceQuantitiesSpec `yaml:"requests,omitempty"`
	Limits   *ResourceQuantitiesSpec `yaml:"limits,omitempty"`
}

type ResourceQuantitiesSpec struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type PortSpec struct {
	Name       string `yaml:"name"`
	Port       int    `yaml:"port"`
	TargetPort int    `yaml:"targetPort,omitempty"`
	Protocol   string `yaml:"protocol,omitempty"`
}

type ComponentOwner struct {
	Project string `yaml:"project"`
}
