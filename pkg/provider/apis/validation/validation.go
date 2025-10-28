// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package validation - validation is used to validate cloud specific ProviderSpec
package validation

import (
	"fmt"

	api "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
)

// ValidateProviderSpecNSecret validates provider spec and secret to check if all fields are present and valid
func ValidateProviderSpecNSecret(spec *api.ProviderSpec, secrets *corev1.Secret) []error {
	var errors []error

	// Validate Secret
	if secrets == nil {
		errors = append(errors, fmt.Errorf("secret is required"))
		return errors // Return early if secret is nil
	}

	projectID, ok := secrets.Data["projectId"]
	if !ok {
		errors = append(errors, fmt.Errorf("secret must contain 'projectId' field"))
	} else if len(projectID) == 0 {
		errors = append(errors, fmt.Errorf("secret 'projectId' cannot be empty"))
	}

	// Validate ProviderSpec
	if spec.MachineType == "" {
		errors = append(errors, fmt.Errorf("providerSpec.machineType is required"))
	}

	if spec.ImageID == "" {
		errors = append(errors, fmt.Errorf("providerSpec.imageId is required"))
	}

	return errors
}
