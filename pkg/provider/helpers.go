// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
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

	if !strings.HasPrefix(providerID, prefix) {
		return "", "", fmt.Errorf("ProviderID must start with 'stackit://'")
	}

	// Remove prefix and split by '/'
	remainder := strings.TrimPrefix(providerID, prefix)
	parts := strings.Split(remainder, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("ProviderID must have format 'stackit://<projectId>/<serverId>'")
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("projectId and serverId cannot be empty")
	}

	return parts[0], parts[1], nil
}

// ========== SDK Conversion Helpers ==========

// ptr returns a pointer to the given value
// This helper is needed because the STACKIT SDK uses pointers for optional fields
func ptr[T any](v T) *T {
	return &v
}

// convertLabelsToSDK converts map[string]string to *map[string]interface{} for SDK
func convertLabelsToSDK(labels map[string]string) *map[string]interface{} {
	if labels == nil {
		return nil
	}

	result := make(map[string]interface{}, len(labels))
	for k, v := range labels {
		result[k] = v
	}
	return &result
}

// convertLabelsFromSDK converts *map[string]interface{} from SDK to map[string]string
func convertLabelsFromSDK(labels *map[string]interface{}) map[string]string {
	if labels == nil {
		return nil
	}

	result := make(map[string]string, len(*labels))
	for k, v := range *labels {
		if strVal, ok := v.(string); ok {
			result[k] = strVal
		}
	}
	return result
}

// convertStringSliceToSDK converts []string to *[]string for SDK
func convertStringSliceToSDK(slice []string) *[]string {
	if slice == nil {
		return nil
	}
	return &slice
}

// convertMetadataToSDK converts map[string]interface{} to *map[string]interface{} for SDK
func convertMetadataToSDK(metadata map[string]interface{}) *map[string]interface{} {
	if metadata == nil {
		return nil
	}
	return &metadata
}
