package client

// ptr returns a pointer to the given value
// This helper is needed because the STACKIT SDK uses pointers for optional fields
func ptr[T any](v T) *T {
	return &v
}

// convertLabelsToSDK converts map[string]string to *map[string]interface{} for SDK
//
//nolint:gocritic // SDK requires *map
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
//
//nolint:gocritic // SDK requires *map
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
//
//nolint:gocritic // SDK requires *map
func convertMetadataToSDK(metadata map[string]interface{}) *map[string]interface{} {
	if metadata == nil {
		return nil
	}
	return &metadata
}
