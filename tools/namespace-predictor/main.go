package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	maxNameLength = 253
	hashLength    = 8
	separator     = "-"
)

// sanitizeName mirrors openchoreo/internal/dataplane/kubernetes.name.go.
func sanitizeName(name string) string {
	name = strings.ToLower(name)
	var sanitized []rune
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '.' {
			sanitized = append(sanitized, r)
		} else {
			sanitized = append(sanitized, '-')
		}
	}
	return strings.Trim(string(sanitized), "-.")
}

// ensureDNSSubdomainCompliance mirrors the OpenChoreo helper.
func ensureDNSSubdomainCompliance(name string) string {
	name = strings.TrimLeftFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	name = strings.TrimRightFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return name
}

// generateK8sNameWithLengthLimit is a byte-for-byte replica of OpenChoreo's
// GenerateK8sNameWithLengthLimit, used to predict the runtime data-plane
// namespace name before OpenChoreo creates it.
func generateK8sNameWithLengthLimit(limit int, names ...string) string {
	cleanedNames := make([]string, 0, len(names))
	for _, name := range names {
		cleanedNames = append(cleanedNames, sanitizeName(name))
	}

	fullName := strings.Join(names, separator)
	hashBytes := sha256.Sum256([]byte(fullName))
	hashString := hex.EncodeToString(hashBytes[:])[:hashLength]

	numberOfNames := len(cleanedNames)
	numberOfSeparatorsInBaseName := numberOfNames - 1
	totalSeparatorLength := len(separator) * numberOfSeparatorsInBaseName
	totalSeparatorLength += len(separator)

	maxBaseNameLength := limit - hashLength - totalSeparatorLength
	maxPartLength := maxBaseNameLength / numberOfNames
	extraChars := maxBaseNameLength % numberOfNames

	truncatedNames := make([]string, numberOfNames)
	for i, name := range cleanedNames {
		allocatedLength := maxPartLength
		if i < extraChars {
			allocatedLength++
		}
		if len(name) > allocatedLength {
			truncatedNames[i] = name[:allocatedLength]
		} else {
			truncatedNames[i] = name
		}
	}

	baseName := strings.Join(truncatedNames, separator)
	finalName := fmt.Sprintf("%s%s%s", baseName, separator, hashString)
	return ensureDNSSubdomainCompliance(finalName)
}

// PredictRuntimeNamespace predicts the OpenChoreo data-plane namespace name.
// It uses the same algorithm and inputs as
// internal/controller/project/integrations/kubernetes/namespace_handler.go:
//   GenerateK8sNameWithLengthLimit(63, "dp", controlPlaneNs, projectName, environmentName)
func PredictRuntimeNamespace(controlPlaneNs, projectName, environmentName string) string {
	return generateK8sNameWithLengthLimit(63, "dp", controlPlaneNs, projectName, environmentName)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <controlPlaneNs> <projectName> <environmentName>\n", os.Args[0])
		os.Exit(1)
	}
	fmt.Println(PredictRuntimeNamespace(os.Args[1], os.Args[2], os.Args[3]))
}
