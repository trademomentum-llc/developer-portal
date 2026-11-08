package main

import (
	"regexp"
	"strings"
	"testing"
)

// Canonical predictor vectors, mirroring scripts/smoke-m3.sh section 1
// (FR-37 / OQ-30). Expected values were produced by this implementation and
// cross-checked against the TS port; vector 0 is additionally verified against
// the live k3d-openchoreo cluster (the hello-m2 ReleaseBinding namespace).
var canonicalVectors = []struct {
	name           string
	controlPlaneNs string
	project        string
	environment    string
	expected       string
}{
	{
		name:           "canonical (live-verified, must never change)",
		controlPlaneNs: "default",
		project:        "default",
		environment:    "development",
		expected:       "dp-default-default-development-f8e58905",
	},
	{
		name:           "hello-m2 development",
		controlPlaneNs: "default",
		project:        "hello-m2",
		environment:    "development",
		expected:       "dp-default-hello-m2-development-bd0274a8",
	},
	{
		name:           "part truncation",
		controlPlaneNs: "openchoreo-control",
		project:        "prod-api",
		environment:    "production",
		expected:       "dp-openchoreo-co-prod-api-production-bf865e69",
	},
	{
		name:           "underscore normalization",
		controlPlaneNs: "underscore_ns",
		project:        "my_project",
		environment:    "prod_env",
		expected:       "dp-underscore-ns-my-project-prod-env-f1cc0757",
	},
	{
		name:           "63-char length limit with long parts",
		controlPlaneNs: "long-control-ns",
		project:        "very-long-project-name-that-keeps-going",
		environment:    "development",
		expected:       "dp-long-control--very-long-pro-development-121ccff5",
	},
}

// dnsSubdomain matches RFC 1123 DNS subdomain names (the Kubernetes
// namespace constraint).
var dnsSubdomain = regexp.MustCompile(`^[a-z0-9]([a-z0-9\-.]*[a-z0-9])?$`)

func TestPredictRuntimeNamespace_CanonicalVectors(t *testing.T) {
	for _, v := range canonicalVectors {
		t.Run(v.name, func(t *testing.T) {
			got := PredictRuntimeNamespace(v.controlPlaneNs, v.project, v.environment)
			if got != v.expected {
				t.Fatalf("PredictRuntimeNamespace(%q, %q, %q) = %q, want %q",
					v.controlPlaneNs, v.project, v.environment, got, v.expected)
			}
		})
	}
}

func TestPredictRuntimeNamespace_Invariants(t *testing.T) {
	for _, v := range canonicalVectors {
		t.Run(v.name, func(t *testing.T) {
			got := PredictRuntimeNamespace(v.controlPlaneNs, v.project, v.environment)

			if len(got) > 63 {
				t.Fatalf("namespace %q exceeds the 63-char limit (len %d)", got, len(got))
			}
			if !strings.HasPrefix(got, "dp-") {
				t.Fatalf("namespace %q lost the dp- prefix", got)
			}
			if !dnsSubdomain.MatchString(got) {
				t.Fatalf("namespace %q is not DNS-1123 subdomain compliant", got)
			}
			if strings.Contains(got, "_") {
				t.Fatalf("namespace %q still contains underscores (normalization broken)", got)
			}
			// Determinism contract: identical inputs must give identical output.
			again := PredictRuntimeNamespace(v.controlPlaneNs, v.project, v.environment)
			if again != got {
				t.Fatalf("non-deterministic output: %q vs %q", got, again)
			}
		})
	}
}
