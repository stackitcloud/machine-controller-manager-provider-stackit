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
