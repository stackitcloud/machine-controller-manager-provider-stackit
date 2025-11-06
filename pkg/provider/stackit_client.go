// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
)

// StackitClient is an interface for interacting with STACKIT IAAS API
// This allows us to mock the client in unit tests
//
// Note: region parameter is required by STACKIT SDK v1.0.0+
// It must be extracted from the Secret (e.g., "eu01-1", "eu01-2")
type StackitClient interface {
	// CreateServer creates a new server in STACKIT
	CreateServer(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error)
	// GetServer retrieves a server by ID from STACKIT
	GetServer(ctx context.Context, token, projectID, region, serverID string) (*Server, error)
	// DeleteServer deletes a server by ID from STACKIT
	DeleteServer(ctx context.Context, token, projectID, region, serverID string) error
	// ListServers lists all servers in a project
	ListServers(ctx context.Context, token, projectID, region string) ([]*Server, error)
}

// CreateServerRequest represents the request to create a server
type CreateServerRequest struct {
	Name                string                   `json:"name"`
	MachineType         string                   `json:"machineType"`
	ImageID             string                   `json:"imageId,omitempty"`
	Labels              map[string]string        `json:"labels,omitempty"`
	Networking          *ServerNetworkingRequest `json:"networking"` // Required in v2 API, no omitempty
	SecurityGroups      []string                 `json:"securityGroups,omitempty"`
	UserData            string                   `json:"userData,omitempty"`
	BootVolume          *BootVolumeRequest       `json:"bootVolume,omitempty"`
	Volumes             []string                 `json:"volumes,omitempty"`
	KeypairName         string                   `json:"keypairName,omitempty"`
	AvailabilityZone    string                   `json:"availabilityZone,omitempty"`
	AffinityGroup       string                   `json:"affinityGroup,omitempty"`
	ServiceAccountMails []string                 `json:"serviceAccountMails,omitempty"`
	Agent               *AgentRequest            `json:"agent,omitempty"`
	Metadata            map[string]interface{}   `json:"metadata,omitempty"`
}

// ServerNetworkingRequest represents the networking configuration for a server
//
// Union type - use one of the following (mutually exclusive):
//   - NetworkID: Auto-create a NIC in the specified network (takes precedence)
//   - NICIDs: Attach pre-existing NICs to the server
//
// If both are specified, NetworkID takes precedence and NICIDs is ignored.
type ServerNetworkingRequest struct {
	NetworkID string   `json:"networkId,omitempty"`
	NICIDs    []string `json:"nicIds,omitempty"`
}

// BootVolumeRequest represents the boot volume configuration for a server
type BootVolumeRequest struct {
	DeleteOnTermination *bool                    `json:"deleteOnTermination,omitempty"`
	PerformanceClass    string                   `json:"performanceClass,omitempty"`
	Size                int                      `json:"size,omitempty"`
	Source              *BootVolumeSourceRequest `json:"source,omitempty"`
}

// BootVolumeSourceRequest represents the source for creating a boot volume
type BootVolumeSourceRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// AgentRequest represents the STACKIT agent configuration for a server
type AgentRequest struct {
	Provisioned *bool `json:"provisioned,omitempty"`
}

// Server represents a STACKIT server response
type Server struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Labels map[string]string `json:"labels,omitempty"`
}
