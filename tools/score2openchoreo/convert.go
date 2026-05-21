// Package main: score2openchoreo pure conversion from Score to OpenChoreo.
package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type ConvertOptions struct {
	Environment string
	Namespace   string
	ImageRef    string
	Project     string
}

const componentTypeAnnotation = "pipeline.m2/component-type"

var resourceRefPattern = regexp.MustCompile(`^\$\{resources\.([a-zA-Z0-9_-]+)\.([a-zA-Z0-9_-]+)\}$`)

func Convert(in ScoreDocument, opts ConvertOptions) ([]OpenChoreoResource, error) {
	if opts.Environment == "" {
		return nil, fmt.Errorf("environment required")
	}
	if opts.Namespace == "" {
		return nil, fmt.Errorf("namespace required")
	}
	if opts.Project == "" {
		return nil, fmt.Errorf("project required")
	}
	if in.Metadata.Name == "" {
		return nil, fmt.Errorf("metadata.name required")
	}
	if len(in.Containers) == 0 {
		return nil, fmt.Errorf("one container required")
	}
	if len(in.Containers) > 1 {
		return nil, fmt.Errorf("multiple containers are not supported by OpenChoreo Workload")
	}

	for key, res := range in.Resources {
		if res.Type != "secret" && res.Type != "environment" {
			return nil, fmt.Errorf("unsupported resource type %q for resource %q", res.Type, key)
		}
	}

	scoreContainer := singleContainer(in.Containers)
	container := WorkloadContainer{
		Image:   scoreContainer.Image,
		Command: scoreContainer.Command,
		Args:    scoreContainer.Args,
	}
	if opts.ImageRef != "" {
		container.Image = opts.ImageRef
	}

	env, err := convertEnv(in, scoreContainer, opts.Environment)
	if err != nil {
		return nil, err
	}
	container.Env = env

	component := OpenChoreoComponent{
		APIVersion: "openchoreo.dev/v1alpha1",
		Kind:       "Component",
		Metadata: ComponentMetadata{
			Name:      in.Metadata.Name,
			Namespace: opts.Namespace,
			Labels:    componentLabels(in.Metadata.Annotations),
		},
		Spec: ComponentSpec{
			Owner:      ComponentOwner{ProjectName: opts.Project},
			AutoDeploy: true,
			ComponentType: ComponentTypeRef{
				Kind: "ClusterComponentType",
				Name: inferComponentType(in),
			},
		},
	}

	workload := OpenChoreoWorkload{
		APIVersion: "openchoreo.dev/v1alpha1",
		Kind:       "Workload",
		Metadata: ComponentMetadata{
			Name:      in.Metadata.Name + "-workload",
			Namespace: opts.Namespace,
		},
		Spec: WorkloadSpec{
			Owner: WorkloadOwner{
				ComponentName: in.Metadata.Name,
				ProjectName:   opts.Project,
			},
			Endpoints: convertEndpoints(in.Service),
			Container: container,
		},
	}

	return []OpenChoreoResource{component, workload}, nil
}

func singleContainer(containers map[string]ScoreContainer) ScoreContainer {
	names := make([]string, 0, len(containers))
	for name := range containers {
		names = append(names, name)
	}
	sort.Strings(names)
	return containers[names[0]]
}

func inferComponentType(in ScoreDocument) string {
	if t := strings.TrimSpace(in.Metadata.Annotations[componentTypeAnnotation]); t != "" {
		return t
	}
	if in.Service != nil && len(in.Service.Ports) > 0 {
		return "deployment/service"
	}
	return "deployment/worker"
}

func componentLabels(annotations map[string]string) map[string]string {
	if len(annotations) == 0 {
		return nil
	}
	labels := make(map[string]string, len(annotations))
	for key, value := range annotations {
		if key == componentTypeAnnotation {
			continue
		}
		labels[key] = value
	}
	if len(labels) == 0 {
		return nil
	}
	return labels
}

func convertEnv(in ScoreDocument, c ScoreContainer, environment string) ([]EnvVarSpec, error) {
	varNames := make([]string, 0, len(c.Variables))
	for k := range c.Variables {
		varNames = append(varNames, k)
	}
	sort.Strings(varNames)

	env := make([]EnvVarSpec, 0, len(varNames))
	for _, k := range varNames {
		v := c.Variables[k]
		if m := resourceRefPattern.FindStringSubmatch(v); m != nil {
			resKey, field := m[1], m[2]
			res, ok := in.Resources[resKey]
			if !ok {
				return nil, fmt.Errorf("variable %s refers to missing resource %s", k, resKey)
			}
			switch res.Type {
			case "secret":
				secretName := resKey + "-secret"
				if n := res.Metadata["name"]; n != "" {
					secretName = n
				}
				env = append(env, EnvVarSpec{
					Key: k,
					ValueFrom: &EnvVarSourceSpec{
						SecretKeyRef: &SecretKeySelectorSpec{Name: secretName, Key: field},
					},
				})
			case "environment":
				env = append(env, EnvVarSpec{Key: k, Value: environment})
			}
		} else if strings.Contains(v, "${resources.") {
			return nil, fmt.Errorf("variable %s has an inline resource reference in %q; the whole value must be a single ${resources.X.Y} expression (inline substitution is not supported)", k, v)
		} else {
			env = append(env, EnvVarSpec{Key: k, Value: v})
		}
	}
	return env, nil
}

func convertEndpoints(service *ScoreService) map[string]WorkloadEndpoint {
	if service == nil || len(service.Ports) == 0 {
		return nil
	}
	portNames := make([]string, 0, len(service.Ports))
	for k := range service.Ports {
		portNames = append(portNames, k)
	}
	sort.Strings(portNames)

	endpoints := make(map[string]WorkloadEndpoint, len(portNames))
	for _, name := range portNames {
		p := service.Ports[name]
		tp := p.TargetPort
		if tp == 0 {
			tp = p.Port
		}
		endpoint := WorkloadEndpoint{
			Type:       endpointType(p.Protocol),
			Port:       p.Port,
			Visibility: []string{"external"},
		}
		if tp != p.Port {
			endpoint.TargetPort = tp
		}
		endpoints[name] = endpoint
	}
	return endpoints
}

func endpointType(protocol string) string {
	switch strings.ToUpper(protocol) {
	case "", "HTTP", "TCP":
		if strings.EqualFold(protocol, "TCP") {
			return "TCP"
		}
		return "HTTP"
	case "UDP":
		return "UDP"
	case "GRPC":
		return "gRPC"
	case "GRAPHQL":
		return "GraphQL"
	case "WEBSOCKET":
		return "Websocket"
	default:
		return strings.ToUpper(protocol)
	}
}
