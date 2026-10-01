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
//   - DeadlineExceeded: Failed waiting for server or volume to be deleted
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

	// Error is ignored to have compatibility with non-migrated OpenStack Machines that have no annotations.
	migrated, _ := strconv.ParseBool(req.Machine.Annotations[migratedMachineAnnotation])
	projectID, serverIDs, err := p.serverIDsForMachine(ctx, req, projectIDFromSecret, providerSpec.Region, migrated)
	if err != nil {
		return nil, err
	}
	if err := p.deleteServers(ctx, projectID, providerSpec.Region, req.Machine.Name, serverIDs); err != nil {
		return nil, err
	}

	// Migrated servers are created with the openstack MCM.
	// The openstack MCM creates a NIC and Volumes separately and attaches it to a server.
	// The STACKIT API unifies this in one call which results in IaaS deleting the NIC and Volume on server deletion aswell.
	// This does not happen when NICs and Volumes are created independently and were attached to the server.
	// Therefore, NICs and Volumes get cleaned up by name for migrated machines.
	if migrated {
		if providerSpec.Networking != nil && providerSpec.Networking.NetworkID != "" {
			if err := p.deleteMachineNICs(ctx, projectID, providerSpec.Region, providerSpec.Networking.NetworkID, req.Machine.Name); err != nil {
				return nil, err
			}
		}
		if err := p.deleteMachineVolumes(ctx, projectID, providerSpec.Region, req.Machine.Name); err != nil {
			return nil, err
		}
	}
	klog.V(2).Infof("Successfully deleted server for machine %q", req.Machine.Name)

	return &driver.DeleteMachineResponse{}, nil
}

// serverIDsForMachine gets the projectID and serverID from the providerID. In case it could not find a ServerID in the providerID it lists all servers based in labels to find the serverIDs.
// serverIDsForMachine retruns a list if IDs servernames are not uniq on infrastructure side.
// In case of a migrated machine with the stackit.cloud/migrated-machine annotation the deletion needs to get all servers and filters internally.
// We can not relay on labels as servers that are migrated during the creation (without a providerID) does not have the new labels.
func (p *Provider) serverIDsForMachine(ctx context.Context, req *driver.DeleteMachineRequest, projectIDFromSecret, region string, migrated bool) (projectID string, serverIDs []string, err error) {
	if providerID := req.Machine.Spec.ProviderID; providerID != "" {
		if !strings.HasPrefix(providerID, StackitProviderName+"://") {
			return "", nil, status.Error(codes.InvalidArgument, "providerID is not empty and does not start with stackit://")
		}

		var serverID string
		projectID, serverID, err = parseProviderID(providerID)
		if err != nil {
			klog.V(2).Infof("invalid ProviderID format: %v", err)
		}
		if serverID != "" {
			serverIDs = append(serverIDs, serverID)
		}
	}
	if projectID == "" {
		projectID = projectIDFromSecret
	}
	if len(serverIDs) != 0 {
		return projectID, serverIDs, nil
	}

	var selector map[string]string
	if !migrated {
		selector = map[string]string{StackitMachineLabel: req.Machine.Name}
	}

	servers, err := p.getServersByLabelSelector(ctx, projectID, region, selector)
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

func (p *Provider) deleteServers(ctx context.Context, projectID, region, machineName string, serverIDs []string) error {
	var allErrors error
	var deletedServerIDs []string

	for _, serverID := range serverIDs {
		if err := p.client.DeleteServer(ctx, projectID, region, serverID); err != nil {
			if errors.Is(err, client.ErrServerNotFound) {
				klog.V(2).Infof("Server %q already deleted for machine %q (idempotent)", serverID, machineName)
				continue
			}

			klog.Errorf("Failed to delete server %q for machine %q: %v", serverID, machineName, err)
			allErrors = errors.Join(allErrors, fmt.Errorf("failed to delete server %q: %w", serverID, err))
			continue
		}
		deletedServerIDs = append(deletedServerIDs, serverID)
	}

	if allErrors != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to delete servers: %v", allErrors))
	}

	for _, serverID := range deletedServerIDs {
		if err := p.WaitUntilServerDeleted(ctx, projectID, region, serverID); err != nil {
			klog.Errorf("Failed waiting for server %q to be deleted for machine %q: %v", serverID, machineName, err)
			allErrors = errors.Join(allErrors, fmt.Errorf("failed waiting for server %q to be deleted: %w", serverID, err))
		}
	}

	if allErrors != nil {
		return status.Error(codes.DeadlineExceeded, fmt.Sprintf("failed waiting for server to be deleted: %v", allErrors))
	}

	return nil
}

func (p *Provider) deleteMachineNICs(ctx context.Context, projectID, region, networkID, machineName string) error {
	nics, err := p.client.ListNICs(ctx, projectID, region, networkID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to list NICs: %v", err))
	}

	var allErrors error

	for _, nic := range nics {
		if nic.Name != machineName {
			continue
		}

		if err = p.client.DeleteNIC(ctx, projectID, region, networkID, nic.ID); err != nil {
			if errors.Is(err, client.ErrNicNotFound) {
				klog.V(2).Infof("NIC %q already deleted for machine %q (idempotent)", nic.ID, machineName)
				continue
			}

			klog.Errorf("Failed to delete NIC %q for machine %q: %v", nic.ID, machineName, err)
			allErrors = errors.Join(allErrors, fmt.Errorf("failed to delete NIC %q: %w", nic.ID, err))
		}
	}

	if allErrors != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to delete NICs: %v", allErrors))
	}

	return nil
}

func (p *Provider) deleteMachineVolumes(ctx context.Context, projectID, region, machineName string) error {
	volumes, err := p.client.ListVolumes(ctx, projectID, region)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to list volumes: %v", err))
	}

	var allErrors error
	var deletedVolumeIDs []string

	for _, volume := range volumes {
		if volume.Name != machineName {
			continue
		}

		if err = p.client.DeleteVolume(ctx, projectID, region, volume.ID); err != nil {
			if errors.Is(err, client.ErrVolumeNotFound) {
				klog.V(2).Infof("Volume %q already deleted for machine %q (idempotent)", volume.ID, machineName)
				continue
			}

			klog.Errorf("Failed to delete volume %q for machine %q: %v", volume.ID, machineName, err)
			allErrors = errors.Join(allErrors, fmt.Errorf("failed to delete volume %q: %w", volume.ID, err))
			continue
		}
		deletedVolumeIDs = append(deletedVolumeIDs, volume.ID)
	}

	if allErrors != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to delete volumes: %v", allErrors))
	}

	for _, volumeID := range deletedVolumeIDs {
		if err := p.WaitUntilVolumeDeleted(ctx, projectID, region, volumeID); err != nil {
			klog.Errorf("Failed waiting for volume %q to be deleted for machine %q: %v", volumeID, machineName, err)
			allErrors = errors.Join(allErrors, fmt.Errorf("failed waiting for volume %q to be deleted: %w", volumeID, err))
		}
	}

	if allErrors != nil {
		return status.Error(codes.DeadlineExceeded, fmt.Sprintf("failed waiting for volume to be deleted: %v", allErrors))
	}

	return nil
}

func (p *Provider) WaitUntilVolumeDeleted(ctx context.Context, projectID, region, volumeID string) error {
	return wait.PollUntilContextTimeout(ctx, p.pollingInterval, p.pollingTimeout, true, func(ctx context.Context) (bool, error) {
		_, err := p.client.GetVolume(ctx, projectID, region, volumeID)
		if err != nil {
			// Volume is deleted if we get a not found error
			if errors.Is(err, client.ErrVolumeNotFound) {
				klog.V(2).Infof("Volume %q has been deleted", volumeID)
				return true, nil
			}
		}

		return false, err
	})
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
