// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package validation - validation is used to validate cloud specific ProviderSpec
package validation

import (
	"fmt"
	"regexp"

	api "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
)

// uuidRegex is a regex pattern for validating UUID format
var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

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

	// Validate Networking
	if spec.Networking != nil {
		networkingErrors := validateNetworking(spec.Networking)
		errors = append(errors, networkingErrors...)
	}

	// Validate SecurityGroups
	if len(spec.SecurityGroups) > 0 {
		for i, sg := range spec.SecurityGroups {
			if sg == "" {
				errors = append(errors, fmt.Errorf("providerSpec.securityGroups[%d] cannot be empty", i))
			}
		}
	}

	return errors
}

// validateNetworking validates the NetworkingSpec
func validateNetworking(networking *api.NetworkingSpec) []error {
	var errors []error

	hasNetworkID := networking.NetworkID != ""
	hasNICIDs := len(networking.NICIDs) > 0

	// Either NetworkID or NICIDs must be set, but not both
	if !hasNetworkID && !hasNICIDs {
		errors = append(errors, fmt.Errorf("providerSpec.networking must specify either networkId or nicIds"))
		return errors
	}

	if hasNetworkID && hasNICIDs {
		errors = append(errors, fmt.Errorf("providerSpec.networking cannot specify both networkId and nicIds (mutually exclusive)"))
		return errors
	}

	// Validate NetworkID format if specified
	if hasNetworkID {
		if !isValidUUID(networking.NetworkID) {
			errors = append(errors, fmt.Errorf("providerSpec.networking.networkId must be a valid UUID"))
		}
	}

	// Validate NICIDs if specified
	if hasNICIDs {
		for i, nicID := range networking.NICIDs {
			if nicID == "" {
				errors = append(errors, fmt.Errorf("providerSpec.networking.nicIds[%d] cannot be empty", i))
			} else if !isValidUUID(nicID) {
				errors = append(errors, fmt.Errorf("providerSpec.networking.nicIds[%d] must be a valid UUID", i))
			}
		}
	}

	return errors
}

// isValidUUID checks if a string is a valid UUID
func isValidUUID(s string) bool {
	return uuidRegex.MatchString(s)
}
