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

	// UserData is cloud-init script or user data for VM bootstrapping
	// Optional field. Can be used to override Secret.userData for this MachineClass.
	// If specified, takes precedence over Secret.userData.
	// Note: Secret.userData is typically required by MCM for node bootstrapping.
	UserData string `json:"userData,omitempty"`

	// BootVolume defines detailed boot disk configuration
	// Optional field. If not specified, a boot volume will be created from ImageID with default settings.
	// If specified, provides fine-grained control over boot disk size, performance, and lifecycle.
	BootVolume *BootVolumeSpec `json:"bootVolume,omitempty"`

	// Volumes are UUIDs of existing volumes to attach to the server
	// Optional field. Allows attaching additional data volumes beyond the boot disk.
	Volumes []string `json:"volumes,omitempty"`
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

// BootVolumeSpec defines the boot disk configuration for a server
// Provides detailed control over boot volume size, performance, and lifecycle
type BootVolumeSpec struct {
	// DeleteOnTermination controls whether the boot volume is deleted when the server is terminated
	// Optional field. Defaults to true (volume deleted with server).
	DeleteOnTermination *bool `json:"deleteOnTermination,omitempty"`

	// PerformanceClass defines the performance tier for the boot volume
	// Optional field. Examples: "standard", "premium", "fast" (depends on STACKIT offerings)
	PerformanceClass string `json:"performanceClass,omitempty"`

	// Size is the boot volume size in GB
	// Optional field. If not specified, size is determined from the image.
	// Must be >= image size if specified.
	Size int `json:"size,omitempty"`

	// Source defines where to create the boot volume from
	// Optional field. If not specified, uses ImageID from ProviderSpec.
	// Allows creating boot volume from snapshots or existing volumes.
	Source *BootVolumeSourceSpec `json:"source,omitempty"`
}

// BootVolumeSourceSpec defines the source for creating a boot volume
// Can be an image, snapshot, or existing volume
type BootVolumeSourceSpec struct {
	// Type is the source type: "image", "snapshot", or "volume"
	// Required field when Source is specified.
	Type string `json:"type"`

	// ID is the UUID of the source (image/snapshot/volume)
	// Required field when Source is specified.
	ID string `json:"id"`
}
