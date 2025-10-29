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
}

// CreateServerRequest represents the request to create a server
type CreateServerRequest struct {
	Name        string `json:"name"`
	MachineType string `json:"machineType"`
	ImageID     string `json:"imageId"`
}

// Server represents a STACKIT server response
type Server struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
