// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package provider contains the cloud provider specific implementations to manage machines
package provider

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis/validation"
	"k8s.io/klog"
)

// CreateMachine handles a machine creation request by creating a STACKIT server
//
// This method creates a new server in STACKIT infrastructure based on the ProviderSpec
// configuration in the MachineClass. It assigns MCM-specific labels to the server for
// tracking and orphan VM detection.
//
// Returns:
//   - ProviderID: Unique identifier in format "stackit://<projectId>/<serverId>"
//   - NodeName: Name that the VM will register with in Kubernetes (matches Machine name)
//
// Error codes:
//   - InvalidArgument: Invalid ProviderSpec or missing required fields
//   - Internal: Failed to create server or communicate with STACKIT API
func (p *Provider) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been received for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine creation request has been processed for %q", req.Machine.Name)

	// Decode ProviderSpec from MachineClass
	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Validate ProviderSpec and Secret
	validationErrs := validation.ValidateProviderSpecNSecret(providerSpec, req.Secret)
	if len(validationErrs) > 0 {
		return nil, status.Error(codes.InvalidArgument, validationErrs[0].Error())
	}

	// Extract credentials from Secret
	projectID := string(req.Secret.Data["projectId"])
	serviceAccountKey := string(req.Secret.Data["serviceAccountKey"])
	region := string(req.Secret.Data["region"])

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	// Build labels: merge ProviderSpec labels with MCM-specific labels
	labels := make(map[string]string)
	// Start with user-provided labels from ProviderSpec
	if providerSpec.Labels != nil {
		for k, v := range providerSpec.Labels {
			labels[k] = v
		}
	}
	// Add MCM-specific labels for server identification and orphan VM detection
	labels["mcm.gardener.cloud/machine"] = req.Machine.Name
	labels["mcm.gardener.cloud/machineclass"] = req.MachineClass.Name
	labels["mcm.gardener.cloud/role"] = "node"

	// Create server request
	createReq := &CreateServerRequest{
		Name:        req.Machine.Name,
		MachineType: providerSpec.MachineType,
		ImageID:     providerSpec.ImageID,
		Labels:      labels,
	}

	// Add networking configuration (required in v2 API)
	// If not specified in ProviderSpec, try to use networkId from Secret, or use empty
	if providerSpec.Networking != nil {
		createReq.Networking = &ServerNetworkingRequest{
			NetworkID: providerSpec.Networking.NetworkID,
			NICIDs:    providerSpec.Networking.NICIDs,
		}
	} else {
		// v2 API requires networking field - use networkId from Secret if available
		// This allows tests/deployments to specify a default network without modifying each MachineClass
		networkID := string(req.Secret.Data["networkId"])
		createReq.Networking = &ServerNetworkingRequest{
			NetworkID: networkID, // Can be empty string if not in Secret
		}
	}

	// Add security groups if specified
	if len(providerSpec.SecurityGroups) > 0 {
		createReq.SecurityGroups = providerSpec.SecurityGroups
	}

	// Add userData for VM bootstrapping
	// Priority: ProviderSpec.UserData > Secret.userData
	// Note: IAAS API requires base64-encoded userData (OpenAPI spec: format=byte)
	var userDataPlain string
	if providerSpec.UserData != "" {
		userDataPlain = providerSpec.UserData
	} else if userData, ok := req.Secret.Data["userData"]; ok && len(userData) > 0 {
		userDataPlain = string(userData)
	}

	if userDataPlain != "" {
		createReq.UserData = base64.StdEncoding.EncodeToString([]byte(userDataPlain))
	}

	// Add boot volume configuration if specified
	if providerSpec.BootVolume != nil {
		createReq.BootVolume = &BootVolumeRequest{
			DeleteOnTermination: providerSpec.BootVolume.DeleteOnTermination,
			PerformanceClass:    providerSpec.BootVolume.PerformanceClass,
			Size:                providerSpec.BootVolume.Size,
		}

		// Add boot volume source if specified
		if providerSpec.BootVolume.Source != nil {
			createReq.BootVolume.Source = &BootVolumeSourceRequest{
				Type: providerSpec.BootVolume.Source.Type,
				ID:   providerSpec.BootVolume.Source.ID,
			}
		}
	}

	// Add additional volumes if specified
	if len(providerSpec.Volumes) > 0 {
		createReq.Volumes = providerSpec.Volumes
	}

	// Add keypair name if specified
	if providerSpec.KeypairName != "" {
		createReq.KeypairName = providerSpec.KeypairName
	}

	// Add availability zone if specified
	if providerSpec.AvailabilityZone != "" {
		createReq.AvailabilityZone = providerSpec.AvailabilityZone
	}

	// Add affinity group if specified
	if providerSpec.AffinityGroup != "" {
		createReq.AffinityGroup = providerSpec.AffinityGroup
	}

	// Add service account mails if specified
	if len(providerSpec.ServiceAccountMails) > 0 {
		createReq.ServiceAccountMails = providerSpec.ServiceAccountMails
	}

	// Add agent configuration if specified
	if providerSpec.Agent != nil {
		createReq.Agent = &AgentRequest{
			Provisioned: providerSpec.Agent.Provisioned,
		}
	}

	// Add metadata if specified
	if len(providerSpec.Metadata) > 0 {
		createReq.Metadata = providerSpec.Metadata
	}

	// Call STACKIT API to create server
	server, err := p.client.CreateServer(ctx, projectID, region, createReq)
	if err != nil {
		klog.Errorf("Failed to create server for machine %q: %v", req.Machine.Name, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create server: %v", err))
	}

	// Generate ProviderID in format: stackit://<projectId>/<serverId>
	providerID := fmt.Sprintf("stackit://%s/%s", projectID, server.ID)

	// NodeName is the machine name (will register with this name in Kubernetes)
	nodeName := req.Machine.Name

	klog.V(2).Infof("Successfully created server %q with ID %q for machine %q", server.Name, server.ID, req.Machine.Name)

	return &driver.CreateMachineResponse{
		ProviderID: providerID,
		NodeName:   nodeName,
	}, nil
}

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

	// Validate ProviderID exists
	if req.Machine.Spec.ProviderID == "" {
		return nil, status.Error(codes.InvalidArgument, "ProviderID is required")
	}

	// Extract credentials from Secret
	serviceAccountKey := string(req.Secret.Data["serviceAccountKey"])
	region := string(req.Secret.Data["region"])

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	// Parse ProviderID to extract projectID and serverID
	projectID, serverID, err := parseProviderID(req.Machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid ProviderID format: %v", err))
	}

	// Call STACKIT API to delete server
	err = p.client.DeleteServer(ctx, projectID, region, serverID)
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
	serviceAccountKey := string(req.Secret.Data["serviceAccountKey"])
	region := string(req.Secret.Data["region"])

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	// Parse ProviderID to extract projectID and serverID
	// Expected format: stackit://<projectId>/<serverId>
	projectID, serverID, err := parseProviderID(req.Machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid ProviderID format: %v", err))
	}

	// Call STACKIT API to get server status
	server, err := p.client.GetServer(ctx, projectID, region, serverID)
	if err != nil {
		// Check if server was not found (404)
		if errors.Is(err, ErrServerNotFound) {
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

// ListMachines lists all STACKIT servers that belong to the specified MachineClass
//
// This method retrieves all servers in the STACKIT project and filters them based on
// the "mcm.gardener.cloud/machineclass" label. This enables the MCM safety controller
// to detect and clean up orphan VMs that are not backed by Machine CRs.
//
// Returns:
//   - MachineList: Map of ProviderID to MachineName for all servers matching the MachineClass
//
// Error codes:
//   - Internal: Failed to list servers or communicate with STACKIT API
func (p *Provider) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been received for %q", req.MachineClass.Name)
	defer klog.V(2).Infof("List machines request has been processed for %q", req.MachineClass.Name)

	// Extract credentials from Secret
	projectID := string(req.Secret.Data["projectId"])
	serviceAccountKey := string(req.Secret.Data["serviceAccountKey"])
	region := string(req.Secret.Data["region"])

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	// Call STACKIT API to list all servers
	servers, err := p.client.ListServers(ctx, projectID, region)
	if err != nil {
		klog.Errorf("Failed to list servers for MachineClass %q: %v", req.MachineClass.Name, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list servers: %v", err))
	}

	// Filter servers by MachineClass label
	// We use the "mcm.gardener.cloud/machineclass" label to identify which servers belong to this MachineClass
	machineList := make(map[string]string)
	for _, server := range servers {
		// Skip servers without labels
		if server.Labels == nil {
			continue
		}

		// Check if server has the matching MachineClass label
		if machineClassLabel, ok := server.Labels["mcm.gardener.cloud/machineclass"]; ok {
			if machineClassLabel == req.MachineClass.Name {
				// Generate ProviderID in format: stackit://<projectId>/<serverId>
				providerID := fmt.Sprintf("stackit://%s/%s", projectID, server.ID)

				// Get machine name from labels (fallback to server name if not found)
				machineName := server.Name
				if machineLabel, ok := server.Labels["mcm.gardener.cloud/machine"]; ok {
					machineName = machineLabel
				}

				machineList[providerID] = machineName
			}
		}
	}

	klog.V(2).Infof("Found %d machines for MachineClass %q", len(machineList), req.MachineClass.Name)

	return &driver.ListMachinesResponse{
		MachineList: machineList,
	}, nil
}

// GetVolumeIDs extracts volume IDs from PersistentVolume specs
//
// This method is used by MCM to get volume IDs for persistent volumes.
// Currently unimplemented for STACKIT provider - volumes are managed directly
// through the ProviderSpec (bootVolume and volumes fields).
//
// Returns:
//   - Unimplemented: This functionality is not required for STACKIT provider
func (p *Provider) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been received for %q", req.PVSpecs)
	defer klog.V(2).Infof("GetVolumeIDs request has been processed successfully for %q", req.PVSpecs)

	return &driver.GetVolumeIDsResponse{}, status.Error(codes.Unimplemented, "")
}

// GenerateMachineClassForMigration generates a MachineClass for migration purposes
//
// This method is used to migrate from provider-specific MachineClass CRDs
// (e.g., AWSMachineClass) to the generic MachineClass format.
//
// STACKIT provider does not have a legacy provider-specific MachineClass format,
// so this method is not needed and returns Unimplemented.
//
// Returns:
//   - Unimplemented: No migration required for STACKIT provider
func (p *Provider) GenerateMachineClassForMigration(ctx context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been received for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	return &driver.GenerateMachineClassForMigrationResponse{}, status.Error(codes.Unimplemented, "")
}
