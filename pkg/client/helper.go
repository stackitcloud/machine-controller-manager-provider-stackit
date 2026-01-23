package client

// convertLabelsToSDK converts map[string]string to *map[string]interface{} for SDK
func convertLabelsToSDK(labels map[string]string) map[string]interface{} {
	if labels == nil {
		return nil
	}

	result := make(map[string]interface{}, len(labels))
	for k, v := range labels {
		result[k] = v
	}
	return result
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
