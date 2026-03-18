package client

// convertLabelsToSDK converts map[string]string to *map[string]any for SDK
func convertLabelsToSDK(labels map[string]string) map[string]any {
	if labels == nil {
		return nil
	}

	result := make(map[string]any, len(labels))
	for k, v := range labels {
		result[k] = v
	}
	return result
}

// convertLabelsFromSDK converts *map[string]any from SDK to map[string]string
//
//nolint:gocritic // SDK requires *map
func convertLabelsFromSDK(labels *map[string]any) map[string]string {
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
