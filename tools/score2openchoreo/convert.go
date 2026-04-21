// Package main: score2openchoreo pure conversion from Score to OpenChoreo Component.
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

var resourceRefPattern = regexp.MustCompile(`^\$\{resources\.([a-zA-Z0-9_-]+)\.([a-zA-Z0-9_-]+)\}$`)

func Convert(in ScoreDocument, opts ConvertOptions) (OpenChoreoComponent, error) {
	if opts.Environment == "" {
		return OpenChoreoComponent{}, fmt.Errorf("environment required")
	}
	if opts.Namespace == "" {
		return OpenChoreoComponent{}, fmt.Errorf("namespace required")
	}
	if opts.Project == "" {
		return OpenChoreoComponent{}, fmt.Errorf("project required")
	}

	// Reject unsupported resource types early
	for key, res := range in.Resources {
		if res.Type != "secret" && res.Type != "environment" {
			return OpenChoreoComponent{}, fmt.Errorf("unsupported resource type %q for resource %q", res.Type, key)
		}
	}

	comp := OpenChoreoComponent{
		APIVersion: "core.choreo.dev/v1alpha1",
		Kind:       "Component",
		Metadata: ComponentMetadata{
			Name:      in.Metadata.Name,
			Namespace: opts.Namespace,
			Labels:    in.Metadata.Annotations,
		},
		Spec: ComponentSpec{
			Environment: opts.Environment,
			Owner:       ComponentOwner{Project: opts.Project},
		},
	}

	// Deterministic iteration order over container names
	containerNames := make([]string, 0, len(in.Containers))
	for k := range in.Containers {
		containerNames = append(containerNames, k)
	}
	sort.Strings(containerNames)

	for _, name := range containerNames {
		c := in.Containers[name]
		cs := ContainerSpec{
			Name:    name,
			Image:   c.Image,
			Command: c.Command,
			Args:    c.Args,
		}
		if opts.ImageRef != "" {
			cs.Image = opts.ImageRef
		}
		if c.Resources != nil {
			cs.Resources = &ContainerResourcesSpec{}
			if c.Resources.Requests != nil {
				cs.Resources.Requests = &ResourceQuantitiesSpec{
					CPU: c.Resources.Requests.CPU, Memory: c.Resources.Requests.Memory,
				}
			}
			if c.Resources.Limits != nil {
				cs.Resources.Limits = &ResourceQuantitiesSpec{
					CPU: c.Resources.Limits.CPU, Memory: c.Resources.Limits.Memory,
				}
			}
		}
		varNames := make([]string, 0, len(c.Variables))
		for k := range c.Variables {
			varNames = append(varNames, k)
		}
		sort.Strings(varNames)
		for _, k := range varNames {
			v := c.Variables[k]
			if m := resourceRefPattern.FindStringSubmatch(v); m != nil {
				resKey, field := m[1], m[2]
				res, ok := in.Resources[resKey]
				if !ok {
					return OpenChoreoComponent{}, fmt.Errorf("variable %s refers to missing resource %s", k, resKey)
				}
				switch res.Type {
				case "secret":
					secretName := resKey + "-secret"
					if n := res.Metadata["name"]; n != "" {
						secretName = n
					}
					cs.Env = append(cs.Env, EnvVarSpec{
						Name: k,
						ValueFrom: &EnvVarSourceSpec{
							SecretKeyRef: &SecretKeySelectorSpec{Name: secretName, Key: field},
						},
					})
				case "environment":
					cs.Env = append(cs.Env, EnvVarSpec{Name: k, Value: opts.Environment})
				}
			} else {
				cs.Env = append(cs.Env, EnvVarSpec{Name: k, Value: v})
			}
		}
		comp.Spec.WorkloadTemplate.Containers = append(comp.Spec.WorkloadTemplate.Containers, cs)
	}

	if in.Service != nil {
		portNames := make([]string, 0, len(in.Service.Ports))
		for k := range in.Service.Ports {
			portNames = append(portNames, k)
		}
		sort.Strings(portNames)
		for _, name := range portNames {
			p := in.Service.Ports[name]
			tp := p.TargetPort
			if tp == 0 {
				tp = p.Port
			}
			proto := p.Protocol
			if proto == "" {
				proto = "TCP"
			}
			comp.Spec.WorkloadTemplate.Ports = append(comp.Spec.WorkloadTemplate.Ports, PortSpec{
				Name: name, Port: p.Port, TargetPort: tp, Protocol: strings.ToUpper(proto),
			})
		}
	}

	return comp, nil
}
