// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package api

// ProviderSpec is the spec to be used while parsing the calls.
type ProviderSpec struct {
	// MachineType is the STACKIT server type (e.g., "c1.2", "m1.4")
	// Required field for creating a server.
	MachineType string `json:"machineType"`

	// ImageID is the UUID of the OS image to use for the server
	// Required field for creating a server.
	ImageID string `json:"imageId"`

	// Labels are key-value pairs used to tag and identify servers
	// Used by MCM for mapping servers to MachineClasses and orphan VM detection
	// Optional field. MCM will automatically add standard labels.
	Labels map[string]string `json:"labels,omitempty"`

	// Networking configuration for the server
	// Specify either a NetworkID (simple) or NICIDs (advanced)
	// Optional field. If not specified, the server may use default networking or require manual configuration.
	Networking *NetworkingSpec `json:"networking,omitempty"`

	// SecurityGroups are the names of security groups to attach to the server
	// Optional field. If not specified, the project's default security group will be used.
	SecurityGroups []string `json:"securityGroups,omitempty"`
}

// NetworkingSpec defines the network configuration for a server
// Use either NetworkID for simple single-network attachment,
// or NICIDs for advanced multi-NIC configuration (not both)
type NetworkingSpec struct {
	// NetworkID is the UUID of the network to attach the server to
	// Simple variant: Server will be attached to this network with auto-configured NIC
	// Mutually exclusive with NICIDs
	NetworkID string `json:"networkId,omitempty"`

	// NICIDs are the UUIDs of pre-created Network Interface Cards to attach
	// Advanced variant: Allows fine-grained control over NICs, IPs, and security groups
	// Mutually exclusive with NetworkID
	NICIDs []string `json:"nicIds,omitempty"`
}
