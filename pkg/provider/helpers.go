// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/json"
	"fmt"

	api "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
)

// decodeProviderSpec decodes the ProviderSpec from a MachineClass
func decodeProviderSpec(machineClass *v1alpha1.MachineClass) (*api.ProviderSpec, error) {
	if machineClass == nil {
		return nil, fmt.Errorf("machineClass is nil")
	}

	var providerSpec *api.ProviderSpec
	if err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec); err != nil {
		return nil, fmt.Errorf("failed to decode ProviderSpec: %w", err)
	}

	return providerSpec, nil
}

// encodeProviderSpecForResponse encodes a ProviderSpec to JSON bytes
func encodeProviderSpecForResponse(spec *api.ProviderSpec) ([]byte, error) {
	return json.Marshal(spec)
}

// parseProviderID parses a STACKIT ProviderID and extracts the projectID and serverID
// Expected format: stackit://<projectId>/<serverId>
func parseProviderID(providerID string) (projectID, serverID string, err error) {
	const prefix = "stackit://"
	if len(providerID) < len(prefix) {
		return "", "", fmt.Errorf("ProviderID too short")
	}
	if providerID[:len(prefix)] != prefix {
		return "", "", fmt.Errorf("ProviderID must start with 'stackit://'")
	}

	// Remove prefix
	remainder := providerID[len(prefix):]
	if remainder == "" {
		return "", "", fmt.Errorf("ProviderID missing project and server IDs")
	}

	// Split by '/' - find exactly one separator
	slashIdx := -1
	for i, c := range remainder {
		if c == '/' {
			if slashIdx >= 0 {
				// Multiple slashes found
				return "", "", fmt.Errorf("ProviderID must have format 'stackit://<projectId>/<serverId>'")
			}
			slashIdx = i
		}
	}

	if slashIdx < 0 {
		// No slash found
		return "", "", fmt.Errorf("ProviderID must have format 'stackit://<projectId>/<serverId>'")
	}

	projectID = remainder[:slashIdx]
	serverID = remainder[slashIdx+1:]

	if projectID == "" {
		return "", "", fmt.Errorf("projectId cannot be empty")
	}
	if serverID == "" {
		return "", "", fmt.Errorf("serverId cannot be empty")
	}

	return projectID, serverID, nil
}
