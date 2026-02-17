package api

// ProviderSpec is the spec to be used while parsing the calls.
type ProviderSpec struct {
	// Region is the STACKIT region (e.g., "eu01", "eu02")
	// Required field for creating a server.
	Region string `json:"region"`

	// MachineType is the STACKIT server type (e.g., "c2i.2", "m2i.8")
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

	// AllowedAddresses are the IP address ranges (CIDRs) allowed to originate traffic from the server's network interface.
	// Optional field. If specified, these ranges are configured as AllowedAddresses on the network interface of the server to bypass anti-spoofing rules.
	AllowedAddresses []string `json:"allowedAddresses,omitempty"`

	// SecurityGroups are the UUIDs of security groups to attach to the server
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

	// KeypairName is the name of the SSH keypair for server access
	// Optional field. If specified, the public key will be injected into the server for SSH access.
	// The keypair must already exist in the STACKIT project.
	KeypairName string `json:"keypairName,omitempty"`

	// AvailabilityZone is the availability zone where the server will be created
	// Optional field. If not specified:
	// - If an existing volume is used as boot volume, the server will be created in the same AZ as the volume
	// - For requests with no volumes, it will be set to the metro availability zone
	// Example values: "eu01-1", "eu01-2"
	AvailabilityZone string `json:"availabilityZone,omitempty"`

	// AffinityGroup is the UUID of the affinity group to associate with the server
	// Optional field. Affinity groups control server placement for performance or availability requirements
	// The affinity group must already exist in the STACKIT project
	// Example: "880e8400-e29b-41d4-a716-446655440000"
	AffinityGroup string `json:"affinityGroup,omitempty"`

	// ServiceAccountMails are email addresses of service accounts to associate with the server
	// Optional field. Service accounts provide identity and access management for the server
	// Service accounts must already exist in the STACKIT project
	// Note: STACKIT API currently limits this to a maximum of 1 service account per server
	// Example: ["my-service@sa.stackit.cloud"]
	ServiceAccountMails []string `json:"serviceAccountMails,omitempty"`

	// Agent configures the STACKIT agent on the server
	// Optional field. The STACKIT agent provides monitoring and management capabilities
	// If not specified, defaults to the STACKIT platform default behavior
	Agent *AgentSpec `json:"agent,omitempty"`

	// Metadata is a generic JSON object for storing arbitrary key-value pairs
	// Optional field. Can be used to store custom metadata that doesn't fit into other fields
	// Example: {"environment": "production", "cost-center": "12345"}
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AgentSpec defines the STACKIT agent configuration for a server
type AgentSpec struct {
	// Provisioned controls whether the STACKIT agent is installed on the server
	// Optional field. Set to true to install the agent, false to skip installation
	Provisioned *bool `json:"provisioned,omitempty"`
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
