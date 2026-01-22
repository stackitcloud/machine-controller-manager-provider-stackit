package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/client"
	"k8s.io/klog/v2"
)

// GetMachineStatus retrieves the current status of a STACKIT server
//
// This method queries STACKIT API to get the current state of the server identified
// by the Machine's ProviderID. If the ProviderID is empty (machine not created yet)
// or the server doesn't exist, it returns NotFound error.
//
// Returns:
//   - ProviderID: The machine's ProviderID
//   - NodeName: Name that the VM registered with in Kubernetes
//
// Error codes:
//   - NotFound: Machine has no ProviderID yet, or server not found in STACKIT
//   - InvalidArgument: Invalid ProviderID format
//   - Internal: Failed to get server status or communicate with STACKIT API
func (p *Provider) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)

	// When ProviderID is empty, the machine doesn't exist yet
	// Return NotFound so MCM knows to call CreateMachine
	if req.Machine.Spec.ProviderID == "" {
		klog.V(2).Infof("Machine %q has no ProviderID, returning NotFound", req.Machine.Name)
		return nil, status.Error(codes.NotFound, "machine does not have a ProviderID yet")
	}

	// Extract credentials from Secret
	serviceAccountKey := string(req.Secret.Data["serviceaccount.json"])

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	// Parse ProviderID to extract projectID and serverID
	// Expected format: stackit://<projectId>/<serverId>
	projectID, serverID, err := parseProviderID(req.Machine.Spec.ProviderID)
	if projectID == "" {
		projectID = string(req.Secret.Data["project-id"])
	}
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid ProviderID format: %v", err))
	}

	// Decode ProviderSpec from MachineClass
	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Call STACKIT API to get server status
	server, err := p.client.GetServer(ctx, projectID, providerSpec.Region, serverID)
	if err != nil {
		// Check if server was not found (404)
		if errors.Is(err, client.ErrServerNotFound) {
			klog.V(2).Infof("Server %q not found for machine %q", serverID, req.Machine.Name)
			return nil, status.Error(codes.NotFound, fmt.Sprintf("server %q not found", serverID))
		}
		// All other errors are internal errors
		klog.Errorf("Failed to get server status for machine %q: %v", req.Machine.Name, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get server status: %v", err))
	}

	klog.V(2).Infof("Retrieved server status for machine %q: status=%s", req.Machine.Name, server.Status)

	return &driver.GetMachineStatusResponse{
		ProviderID: req.Machine.Spec.ProviderID,
		NodeName:   req.Machine.Name,
	}, nil
}
