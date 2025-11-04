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

	"github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis/validation"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/klog"
)

// NOTE
//
// The basic working of the controller will work with just implementing the CreateMachine() & DeleteMachine() methods.
// You can first implement these two methods and check the working of the controller.
// Leaving the other methods to NOT_IMPLEMENTED error status.
// Once this works you can implement the rest of the methods.
//
// Also make sure each method return appropriate errors mentioned in `https://github.com/gardener/machine-controller-manager/blob/master/docs/development/machine_error_codes.md`

// CreateMachine handles a machine creation request
// REQUIRED METHOD
//
// REQUEST PARAMETERS (driver.CreateMachineRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM is to be created
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.CreateMachineResponse)
// ProviderID            string                   Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//
//	ProviderID typically matches with the node.Spec.ProviderID on the node object.
//	Eg: gce://project-name/region/vm-ProviderID
//
// NodeName              string                   Returns the name of the node-object that the VM register's with Kubernetes.
//
//	This could be different from req.MachineName as well
//
// LastKnownState        string                   (Optional) Last known state of VM during the current operation.
//
//	Could be helpful to continue operations in future requests.
//
// OPTIONAL IMPLEMENTATION LOGIC
// It is optionally expected by the safety controller to use an identification mechanisms to map the VM Created by a providerSpec.
// These could be done using tag(s)/resource-groups etc.
// This logic is used by safety controller to delete orphan VMs which are not backed by any machine CRD
func (p *Provider) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been recieved for %q", req.Machine.Name)
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

	// Extract projectId from Secret
	projectID := string(req.Secret.Data["projectId"])

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

	// Add networking configuration if specified
	if providerSpec.Networking != nil {
		createReq.Networking = &ServerNetworkingRequest{
			NetworkID: providerSpec.Networking.NetworkID,
			NICIDs:    providerSpec.Networking.NICIDs,
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

	// Call STACKIT API to create server
	server, err := p.client.CreateServer(ctx, projectID, createReq)
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

// DeleteMachine handles a machine deletion request
//
// REQUEST PARAMETERS (driver.DeleteMachineRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM is to be deleted
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.DeleteMachineResponse)
// LastKnownState        bytes(blob)              (Optional) Last known state of VM during the current operation.
//
//	Could be helpful to continue operations in future requests.
func (p *Provider) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)

	// Validate ProviderID exists
	if req.Machine.Spec.ProviderID == "" {
		return nil, status.Error(codes.InvalidArgument, "ProviderID is required")
	}

	// Parse ProviderID to extract projectID and serverID
	projectID, serverID, err := parseProviderID(req.Machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid ProviderID format: %v", err))
	}

	// Call STACKIT API to delete server
	err = p.client.DeleteServer(ctx, projectID, serverID)
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

// GetMachineStatus handles a machine get status request
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.GetMachineStatusRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM status needs to be returned
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.GetMachineStatueResponse)
// ProviderID            string                   Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//
//	ProviderID typically matches with the node.Spec.ProviderID on the node object.
//	Eg: gce://project-name/region/vm-ProviderID
//
// NodeName             string                    Returns the name of the node-object that the VM register's with Kubernetes.
//
//	This could be different from req.MachineName as well
//
// The request should return a NOT_FOUND (5) status error code if the machine is not existing
func (p *Provider) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been recieved for %q", req.Machine.Name)
	defer klog.V(2).Infof("Machine get request has been processed successfully for %q", req.Machine.Name)

	// When ProviderID is empty, the machine doesn't exist yet
	// Return NotFound so MCM knows to call CreateMachine
	if req.Machine.Spec.ProviderID == "" {
		klog.V(2).Infof("Machine %q has no ProviderID, returning NotFound", req.Machine.Name)
		return nil, status.Error(codes.NotFound, "machine does not have a ProviderID yet")
	}

	// Parse ProviderID to extract projectID and serverID
	// Expected format: stackit://<projectId>/<serverId>
	projectID, serverID, err := parseProviderID(req.Machine.Spec.ProviderID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid ProviderID format: %v", err))
	}

	// Call STACKIT API to get server status
	server, err := p.client.GetServer(ctx, projectID, serverID)
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

// ListMachines lists all the machines possibilly created by a providerSpec
// Identifying machines created by a given providerSpec depends on the OPTIONAL IMPLEMENTATION LOGIC
// you have used to identify machines created by a providerSpec. It could be tags/resource-groups etc
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.ListMachinesRequest)
// MachineClass          *v1alpha1.MachineClass   MachineClass based on which VMs created have to be listed
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.ListMachinesResponse)
// MachineList           map<string,string>  A map containing the keys as the MachineID and value as the MachineName
//
//	for all machine's who where possibilly created by this ProviderSpec
func (p *Provider) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been recieved for %q", req.MachineClass.Name)
	defer klog.V(2).Infof("List machines request has been processed for %q", req.MachineClass.Name)

	// Extract projectId from Secret
	projectID := string(req.Secret.Data["projectId"])

	// Call STACKIT API to list all servers
	servers, err := p.client.ListServers(ctx, projectID)
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

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// REQUEST PARAMETERS (driver.GetVolumeIDsRequest)
// PVSpecList            []*corev1.PersistentVolumeSpec       PVSpecsList is a list PV specs for whom volume-IDs are required.
//
// RESPONSE PARAMETERS (driver.GetVolumeIDsResponse)
// VolumeIDs             []string                             VolumeIDs is a repeated list of VolumeIDs.
func (p *Provider) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been recieved for %q", req.PVSpecs)
	defer klog.V(2).Infof("GetVolumeIDs request has been processed successfully for %q", req.PVSpecs)

	return &driver.GetVolumeIDsResponse{}, status.Error(codes.Unimplemented, "")
}

// GenerateMachineClassForMigration helps in migration of one kind of machineClass CR to another kind.
// For instance a machineClass custom resource of `AWSMachineClass` to `MachineClass`.
// Implement this functionality only if something like this is desired in your setup.
// If you don't require this functionality leave is as is. (return Unimplemented)
//
// The following are the tasks typically expected out of this method
// 1. Validate if the incoming classSpec is valid one for migration (e.g. has the right kind).
// 2. Migrate/Copy over all the fields/spec from req.ProviderSpecificMachineClass to req.MachineClass
// For an example refer
//
//	https://github.com/prashanth26/machine-controller-manager-provider-gcp/blob/migration/pkg/gcp/machine_controller.go#L222-L233
//
// REQUEST PARAMETERS (driver.GenerateMachineClassForMigration)
// ProviderSpecificMachineClass    interface{}                             ProviderSpecificMachineClass is provider specfic machine class object (E.g. AWSMachineClass). Typecasting is required here.
// MachineClass 				   *v1alpha1.MachineClass                  MachineClass is the machine class object that is to be filled up by this method.
// ClassSpec                       *v1alpha1.ClassSpec                     Somemore classSpec details useful while migration.
//
// RESPONSE PARAMETERS (driver.GenerateMachineClassForMigration)
// NONE
func (p *Provider) GenerateMachineClassForMigration(ctx context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been recieved for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	return &driver.GenerateMachineClassForMigrationResponse{}, status.Error(codes.Unimplemented, "")
}
