package provider

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/client"
	"k8s.io/apimachinery/pkg/util/wait"
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
	projectIDFromSecret, serviceAccountKey := extractSecretCredentials(req.Secret.Data)

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	migrated, _ := strconv.ParseBool(req.Machine.Annotations[migratedMachineAnnotation])
	projectID, serverIDs, err := p.serverIDsForMachine(ctx, req, projectIDFromSecret, providerSpec.Region, migrated)
	if err != nil {
		return nil, err
	}
	serverAlreadyDeleted, err := p.deleteServers(ctx, projectID, providerSpec.Region, req.Machine.Name, serverIDs)
	if err != nil {
		return nil, err
	}
	if serverAlreadyDeleted {
		return &driver.DeleteMachineResponse{}, nil
	}
	if migrated {
		nicAlreadyDeleted, err := p.deleteMachineNICs(ctx, projectID, providerSpec.Region, providerSpec.Networking.NetworkID, req.Machine.Name)
		if err != nil {
			return nil, err
		}
		if nicAlreadyDeleted {
			return &driver.DeleteMachineResponse{}, nil
		}
	}
	klog.V(2).Infof("Successfully deleted server for machine %q", req.Machine.Name)

	return &driver.DeleteMachineResponse{}, nil
}

func (p *Provider) serverIDsForMachine(ctx context.Context, req *driver.DeleteMachineRequest, secretProjectID, region string, migrated bool) (projectID string, serverIDs []string, err error) {
	projectID, serverIDs = "", nil
	if providerID := req.Machine.Spec.ProviderID; providerID != "" {
		if !strings.HasPrefix(providerID, StackitProviderName) {
			return "", nil, status.Error(codes.InvalidArgument, "providerID is not empty and does not start with stackit://")
		}

		var serverID string
		projectID, serverID, err = parseProviderID(providerID)
		if err != nil {
			klog.V(2).Infof("invalid ProviderID format: %v", err)
		}
		serverIDs = append(serverIDs, serverID)
	}
	if projectID == "" {
		projectID = secretProjectID
	}
	if len(serverIDs) != 0 {
		return projectID, serverIDs, nil
	}

	selector := map[string]string{StackitMachineLabel: req.Machine.Name}
	if migrated {
		selector = nil
	}
	servers, err := p.getServersByName(ctx, projectID, region, selector)
	if err != nil {
		return "", nil, status.Error(codes.Internal, fmt.Sprintf("failed to find server by name: %v", err))
	}
	for _, server := range servers {
		if server.Name == req.Machine.Name {
			serverIDs = append(serverIDs, server.ID)
		}
	}
	return projectID, serverIDs, nil
}

func (p *Provider) deleteServers(ctx context.Context, projectID, region, machineName string, serverIDs []string) (bool, error) {
	for _, serverID := range serverIDs {
		if err := p.client.DeleteServer(ctx, projectID, region, serverID); err != nil {
			if errors.Is(err, client.ErrServerNotFound) {
				klog.V(2).Infof("Server %q already deleted for machine %q (idempotent)", serverID, machineName)
				return true, nil
			}
			klog.Errorf("Failed to delete server for machine %q: %v", machineName, err)
			return false, status.Error(codes.Internal, fmt.Sprintf("failed to delete server: %v", err))
		}
	}
	return false, nil
}

func (p *Provider) deleteMachineNICs(ctx context.Context, projectID, region, networkID, machineName string) (bool, error) {
	nics, err := p.client.ListNICs(ctx, projectID, region, networkID)
	if err != nil {
		return false, err
	}
	for _, nic := range nics {
		if nic.Name != machineName {
			continue
		}
		if err := p.client.DeleteNIC(ctx, projectID, region, nic.NetworkID, nic.ID); err != nil {
			if errors.Is(err, client.ErrNicNotFound) {
				klog.V(2).Infof("Nic %q already deleted for machine %q (idempotent)", nic.ID, machineName)
				return true, nil
			}
			klog.Errorf("Failed to delete nic for machine %q: %v", machineName, err)
			return false, status.Error(codes.Internal, fmt.Sprintf("failed to delete nic: %v", err))
		}
	}
	return false, nil
}

func (p *Provider) WaitUntilServerDeleted(ctx context.Context, projectID, region, serverID string) error {
	return wait.PollUntilContextTimeout(ctx, p.pollingInterval, p.pollingTimeout, true, func(ctx context.Context) (bool, error) {
		_, err := p.client.GetServer(ctx, projectID, region, serverID)
		if err != nil {
			// Server is deleted if we get a not found error
			if errors.Is(err, client.ErrServerNotFound) {
				klog.V(2).Infof("Server %q has been deleted", serverID)
				return true, nil
			}
		}

		return false, err
	})
}
