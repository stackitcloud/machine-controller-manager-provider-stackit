package provider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis/validation"
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
	prefix := fmt.Sprintf("%s://", StackitProviderName)

	if !strings.HasPrefix(providerID, prefix) {
		return "", "", fmt.Errorf("ProviderID must start with '%s://'", StackitProviderName)
	}

	// Remove prefix and split by '/'
	remainder := strings.TrimPrefix(providerID, prefix)
	parts := strings.Split(remainder, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("ProviderID must have format '%s://<projectId>/<serverId>'", StackitProviderName)
	}

	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("projectId and serverId cannot be empty")
	}

	return parts[0], parts[1], nil
}

func extractSecretCredentials(secretData map[string][]byte) (projectID, serviceAccountKey string) {
	projectID = string(secretData[validation.StackitProjectIDSecretKey])
	serviceAccountKey = string(secretData[validation.StackitServiceAccountKey])
	return projectID, serviceAccountKey
}
