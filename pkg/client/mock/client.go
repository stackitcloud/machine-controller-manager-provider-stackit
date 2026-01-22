package mock

import (
	"context"
	"encoding/json"

	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/client"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
)

// StackitClient is a mock implementation of StackitClient for testing
// Note: Single-tenant design - each client is bound to one set of credentials
type StackitClient struct {
	CreateServerFunc func(ctx context.Context, projectID, region string, req *client.CreateServerRequest) (*client.Server, error)
	GetServerFunc    func(ctx context.Context, projectID, region, serverID string) (*client.Server, error)
	DeleteServerFunc func(ctx context.Context, projectID, region, serverID string) error
	ListServersFunc  func(ctx context.Context, projectID, region string, labelSelector map[string]string) ([]*client.Server, error)
	GetNICsFunc      func(ctx context.Context, projectID, region, serverID string) ([]*client.NIC, error)
	UpdateNICFunc    func(ctx context.Context, projectID, region, networkID, nicID string, allowedAddresses []string) (*client.NIC, error)
}

func (m *StackitClient) CreateServer(ctx context.Context, projectID, region string, req *client.CreateServerRequest) (*client.Server, error) {
	if m.CreateServerFunc != nil {
		return m.CreateServerFunc(ctx, projectID, region, req)
	}
	return &client.Server{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   req.Name,
		Status: "CREATING",
	}, nil
}

func (m *StackitClient) GetServer(ctx context.Context, projectID, region, serverID string) (*client.Server, error) {
	if m.GetServerFunc != nil {
		return m.GetServerFunc(ctx, projectID, region, serverID)
	}
	return &client.Server{
		ID:     serverID,
		Name:   "test-machine",
		Status: "ACTIVE",
	}, nil
}

func (m *StackitClient) DeleteServer(ctx context.Context, projectID, region, serverID string) error {
	if m.DeleteServerFunc != nil {
		return m.DeleteServerFunc(ctx, projectID, region, serverID)
	}
	return nil
}

func (m *StackitClient) ListServers(ctx context.Context, projectID, region string, labelSelector map[string]string) ([]*client.Server, error) {
	if m.ListServersFunc != nil {
		return m.ListServersFunc(ctx, projectID, region, labelSelector)
	}
	return []*client.Server{}, nil
}

func (m *StackitClient) GetNICsForServer(ctx context.Context, projectID, region, serverID string) ([]*client.NIC, error) {
	if m.GetNICsFunc != nil {
		return m.GetNICsFunc(ctx, projectID, region, serverID)
	}
	return []*client.NIC{}, nil
}

func (m *StackitClient) UpdateNIC(ctx context.Context, projectID, region, networkID, nicID string, allowedAddresses []string) (*client.NIC, error) {
	if m.UpdateNICFunc != nil {
		return m.UpdateNICFunc(ctx, projectID, region, networkID, nicID, allowedAddresses)
	}
	return &client.NIC{}, nil
}

// UpdateNIC updates a network interface

// encodeProviderSpec is a helper function to encode ProviderSpec for tests
func EncodeProviderSpec(spec *api.ProviderSpec) ([]byte, error) {
	return json.Marshal(spec)
}
