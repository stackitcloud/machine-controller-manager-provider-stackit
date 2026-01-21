package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/klog/v2"
)

// DeleteMachine handles a machine deletion request by deleting the STACKIT server
//
// This method deletes the server identified by the ProviderID from STACKIT infrastructure.
// It is idempotent - if the server is already deleted (404), it returns success.
//
// Error codes:
//   - InvalidArgument: Missing or invalid ProviderID
//   - Internal: Failed to delete server or communicate with STACKIT API
func (p *Provider) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)

	// Extract credentials from Secret
	_, serviceAccountKey := decodeCloudProviderSecret(req.Secret)

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	var projectID, serverID string
	var err error
	if req.Machine.Spec.ProviderID != "" {
		if !strings.HasPrefix(req.Machine.Spec.ProviderID, StackitProviderName) {
			return nil, status.Error(codes.InvalidArgument, "providerID is not empty and does not start with stackit://")
		}

		// Parse ProviderID to extract projectID and serverID
		projectID, serverID, err = parseProviderID(req.Machine.Spec.ProviderID)
		if err != nil {
			klog.V(2).Infof("invalid ProviderID format: %v", err)
		}
	}

	if projectID == "" {
		projectID, _ = decodeCloudProviderSecret(req.Secret)
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if serverID == "" {
		server, err := p.getServerByName(ctx, projectID, providerSpec.Region, req.Machine.Name)
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find server by name: %v", err))
		}

		if server != nil {
			serverID = server.ID
		}
	}

	if serverID == "" {
		klog.V(2).Infof("Server is already deleted for machine %q", req.Machine.Name)
		return &driver.DeleteMachineResponse{}, nil
	}

	// Call STACKIT API to delete server
	err = p.client.DeleteServer(ctx, projectID, providerSpec.Region, serverID)
	if err != nil {
		// Check if server was not found (404) - this is OK for idempotency
		if errors.Is(err, ErrServerNotFound) {
			klog.V(2).Infof("Server %q already deleted for machine %q (idempotent)", serverID, req.Machine.Name)
			return &driver.DeleteMachineResponse{}, nil
		}
		// All other errors are internal errors
		klog.Errorf("Failed to delete server for machine %q: %v", req.Machine.Name, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete server: %v", err))
	}

	klog.V(2).Infof("Successfully deleted server %q for machine %q", serverID, req.Machine.Name)

	return &driver.DeleteMachineResponse{}, nil
}
