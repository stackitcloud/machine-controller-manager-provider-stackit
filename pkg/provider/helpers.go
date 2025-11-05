// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/base64"
)

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

// convertUserDataToSDK converts base64-encoded string to *string for SDK
// The STACKIT API expects userData as a base64-encoded string
func convertUserDataToSDK(userData string) *string {
	if userData == "" {
		return nil
	}

	// Check if already base64-encoded
	if _, err := base64.StdEncoding.DecodeString(userData); err == nil {
		// Already base64, use as-is
		return ptr(userData)
	}

	// Not base64, encode it
	encoded := base64.StdEncoding.EncodeToString([]byte(userData))
	return ptr(encoded)
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
