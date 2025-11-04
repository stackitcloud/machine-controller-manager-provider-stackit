// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
)

// StackitClient is an interface for interacting with STACKIT IAAS API
// This allows us to mock the client in unit tests
type StackitClient interface {
	// CreateServer creates a new server in STACKIT
	CreateServer(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error)
	// GetServer retrieves a server by ID from STACKIT
	GetServer(ctx context.Context, projectID, serverID string) (*Server, error)
	// DeleteServer deletes a server by ID from STACKIT
	DeleteServer(ctx context.Context, projectID, serverID string) error
	// ListServers lists all servers in a project
	ListServers(ctx context.Context, projectID string) ([]*Server, error)
}

// CreateServerRequest represents the request to create a server
type CreateServerRequest struct {
	Name                string                   `json:"name"`
	MachineType         string                   `json:"machineType"`
	ImageID             string                   `json:"imageId,omitempty"`
	Labels              map[string]string        `json:"labels,omitempty"`
	Networking          *ServerNetworkingRequest `json:"networking,omitempty"`
	SecurityGroups      []string                 `json:"securityGroups,omitempty"`
	UserData            string                   `json:"userData,omitempty"`
	BootVolume          *BootVolumeRequest       `json:"bootVolume,omitempty"`
	Volumes             []string                 `json:"volumes,omitempty"`
	KeypairName         string                   `json:"keypairName,omitempty"`
	AvailabilityZone    string                   `json:"availabilityZone,omitempty"`
	AffinityGroup       string                   `json:"affinityGroup,omitempty"`
	ServiceAccountMails []string                 `json:"serviceAccountMails,omitempty"`
}

// ServerNetworkingRequest represents the networking configuration for a server
// Use either NetworkID or NICIDs (mutually exclusive)
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

// Server represents a STACKIT server response
type Server struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Labels map[string]string `json:"labels,omitempty"`
}
