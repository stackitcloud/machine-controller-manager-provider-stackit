// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	api "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis"
)

// mockStackitClient is a mock implementation of StackitClient for testing
type mockStackitClient struct {
	createServerFunc func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error)
	getServerFunc    func(ctx context.Context, projectID, serverID string) (*Server, error)
	deleteServerFunc func(ctx context.Context, projectID, serverID string) error
	listServersFunc  func(ctx context.Context, projectID string) ([]*Server, error)
}

func (m *mockStackitClient) CreateServer(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
	if m.createServerFunc != nil {
		return m.createServerFunc(ctx, projectID, req)
	}
	return &Server{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   req.Name,
		Status: "CREATING",
	}, nil
}

func (m *mockStackitClient) GetServer(ctx context.Context, projectID, serverID string) (*Server, error) {
	if m.getServerFunc != nil {
		return m.getServerFunc(ctx, projectID, serverID)
	}
	return &Server{
		ID:     serverID,
		Name:   "test-machine",
		Status: "RUNNING",
	}, nil
}

func (m *mockStackitClient) DeleteServer(ctx context.Context, projectID, serverID string) error {
	if m.deleteServerFunc != nil {
		return m.deleteServerFunc(ctx, projectID, serverID)
	}
	return nil
}

func (m *mockStackitClient) ListServers(ctx context.Context, projectID string) ([]*Server, error) {
	if m.listServersFunc != nil {
		return m.listServersFunc(ctx, projectID)
	}
	return []*Server{}, nil
}

// encodeProviderSpec is a helper function to encode ProviderSpec for tests
func encodeProviderSpec(spec *api.ProviderSpec) ([]byte, error) {
	// For now, use a simple JSON encoding
	// In real implementation, this would use proper encoding
	return []byte(fmt.Sprintf(`{"machineType":"%s","imageId":"%s"}`, spec.MachineType, spec.ImageID)), nil
}
