// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
)

// mockStackitClient is a mock implementation of StackitClient for testing
// Note: Single-tenant design - each client is bound to one set of credentials
type mockStackitClient struct {
	createServerFunc func(ctx context.Context, projectID, region string, req *CreateServerRequest) (*Server, error)
	getServerFunc    func(ctx context.Context, projectID, region, serverID string) (*Server, error)
	deleteServerFunc func(ctx context.Context, projectID, region, serverID string) error
	listServersFunc  func(ctx context.Context, projectID, region string, labelSelector map[string]string) ([]*Server, error)
	getNICsFunc      func(ctx context.Context, projectID, region, serverID string) ([]*NIC, error)
	updateNICFunc    func(ctx context.Context, projectID, region, networkID, nicID string, allowedAddresses []string) (*NIC, error)
}

func (m *mockStackitClient) CreateServer(ctx context.Context, projectID, region string, req *CreateServerRequest) (*Server, error) {
	if m.createServerFunc != nil {
		return m.createServerFunc(ctx, projectID, region, req)
	}
	return &Server{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   req.Name,
		Status: "CREATING",
	}, nil
}

func (m *mockStackitClient) GetServer(ctx context.Context, projectID, region, serverID string) (*Server, error) {
	if m.getServerFunc != nil {
		return m.getServerFunc(ctx, projectID, region, serverID)
	}
	return &Server{
		ID:     serverID,
		Name:   "test-machine",
		Status: "ACTIVE",
	}, nil
}

func (m *mockStackitClient) DeleteServer(ctx context.Context, projectID, region, serverID string) error {
	if m.deleteServerFunc != nil {
		return m.deleteServerFunc(ctx, projectID, region, serverID)
	}
	return nil
}

func (m *mockStackitClient) ListServers(ctx context.Context, projectID, region string, labelSelector map[string]string) ([]*Server, error) {
	if m.listServersFunc != nil {
		return m.listServersFunc(ctx, projectID, region, labelSelector)
	}
	return []*Server{}, nil
}

func (m *mockStackitClient) GetNICsForServer(ctx context.Context, projectID, region, serverID string) ([]*NIC, error) {
	if m.getNICsFunc != nil {
		return m.getNICsFunc(ctx, projectID, region, serverID)
	}
	return []*NIC{}, nil
}

func (m *mockStackitClient) UpdateNIC(ctx context.Context, projectID, region, networkID, nicID string, allowedAddresses []string) (*NIC, error) {
	if m.updateNICFunc != nil {
		return m.updateNICFunc(ctx, projectID, region, networkID, nicID, allowedAddresses)
	}
	return &NIC{}, nil
}

// UpdateNIC updates a network interface

// encodeProviderSpec is a helper function to encode ProviderSpec for tests
func encodeProviderSpec(spec *api.ProviderSpec) ([]byte, error) {
	return encodeProviderSpecForResponse(spec)
}
