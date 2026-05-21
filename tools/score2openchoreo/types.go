// Package main: score2openchoreo shared types. Score input and OpenChoreo
// output structs. No I/O.
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

type OpenChoreoResource any

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
	Owner         ComponentOwner   `yaml:"owner"`
	AutoDeploy    bool             `yaml:"autoDeploy,omitempty"`
	ComponentType ComponentTypeRef `yaml:"componentType"`
}

type ComponentOwner struct {
	ProjectName string `yaml:"projectName"`
}

type ComponentTypeRef struct {
	Kind string `yaml:"kind"`
	Name string `yaml:"name"`
}

type OpenChoreoWorkload struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   ComponentMetadata `yaml:"metadata"`
	Spec       WorkloadSpec      `yaml:"spec"`
}

type WorkloadSpec struct {
	Owner     WorkloadOwner               `yaml:"owner"`
	Endpoints map[string]WorkloadEndpoint `yaml:"endpoints,omitempty"`
	Container WorkloadContainer           `yaml:"container"`
}

type WorkloadOwner struct {
	ComponentName string `yaml:"componentName"`
	ProjectName   string `yaml:"projectName"`
}

type WorkloadContainer struct {
	Image   string       `yaml:"image"`
	Command []string     `yaml:"command,omitempty"`
	Args    []string     `yaml:"args,omitempty"`
	Env     []EnvVarSpec `yaml:"env,omitempty"`
}

type EnvVarSpec struct {
	Key       string            `yaml:"key"`
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

type WorkloadEndpoint struct {
	Type       string   `yaml:"type"`
	Port       int      `yaml:"port"`
	TargetPort int      `yaml:"targetPort,omitempty"`
	Visibility []string `yaml:"visibility,omitempty"`
}

type OpenChoreoProject struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   ComponentMetadata `yaml:"metadata"`
	Spec       ProjectSpec       `yaml:"spec"`
}

type ProjectSpec struct {
	DeploymentPipelineRef DeploymentPipelineRef `yaml:"deploymentPipelineRef"`
}

type DeploymentPipelineRef struct {
	Name string `yaml:"name"`
}

type OpenChoreoSecretReference struct {
	APIVersion string              `yaml:"apiVersion"`
	Kind       string              `yaml:"kind"`
	Metadata   ComponentMetadata   `yaml:"metadata"`
	Spec       SecretReferenceSpec `yaml:"spec"`
}

type SecretReferenceSpec struct {
	Template        SecretReferenceTemplate `yaml:"template"`
	Data            []SecretReferenceData   `yaml:"data"`
	RefreshInterval string                  `yaml:"refreshInterval,omitempty"`
}

type SecretReferenceTemplate struct {
	Type     string             `yaml:"type"`
	Metadata SecretTemplateMeta `yaml:"metadata,omitempty"`
}

type SecretTemplateMeta struct {
	Labels map[string]string `yaml:"labels,omitempty"`
}

type SecretReferenceData struct {
	SecretKey string          `yaml:"secretKey"`
	RemoteRef SecretRemoteRef `yaml:"remoteRef"`
}

type SecretRemoteRef struct {
	Key string `yaml:"key"`
}
