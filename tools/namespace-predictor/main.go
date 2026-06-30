package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

// PredictRuntimeNamespace implements the deterministic OpenChoreo data-plane namespace naming.
// This is a pure function — byte-for-byte equivalent to the logic in OpenChoreo (as validated by the Option C sub-agent).
// Reference implementation from openchoreo/internal/dataplane/kubernetes/name.go
func PredictRuntimeNamespace(controlPlaneNs, projectName, environmentName string) string {
	const maxLen = 63
	input := fmt.Sprintf("%s-%s-%s", controlPlaneNs, projectName, environmentName)
	hash := sha256.Sum256([]byte(input))
	short := hex.EncodeToString(hash[:])[:8]

	name := fmt.Sprintf("dp-%s-%s-%s-%s", controlPlaneNs, projectName, environmentName, short)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")

	if len(name) > maxLen {
		name = name[:maxLen]
	}
	return name
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <controlPlaneNs> <projectName> <environmentName>\n", os.Args[0])
		os.Exit(1)
	}
	fmt.Println(PredictRuntimeNamespace(os.Args[1], os.Args[2], os.Args[3]))
}