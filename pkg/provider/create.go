package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/client"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis/validation"
	"k8s.io/klog/v2"
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

	// Check if incoming provider in the MachineClass is a provider we support
	if req.MachineClass.Provider != StackitProviderName {
		err := fmt.Errorf("requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, StackitProviderName)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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
	projectID, serviceAccountKey := extractSecretCredentials(req.Secret.Data)

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	// check if server already exists
	server, err := p.getServerByName(ctx, projectID, providerSpec.Region, req.Machine.Name)
	if err != nil {
		klog.Errorf("Failed to fetch server for machine %q: %v", req.Machine.Name, err)
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("failed to fetch server: %v", err))
	}

	if server == nil {
		// Call STACKIT API to create server
		server, err = p.client.CreateServer(ctx, projectID, providerSpec.Region, p.createServerRequest(req, providerSpec))
		if err != nil {
			klog.Errorf("Failed to create server for machine %q: %v", req.Machine.Name, err)

			// Check for resource exhaustion errors to avoid spamming the API
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no valid host") || strings.Contains(errMsg, "quota exceeded") {
				return nil, status.Error(codes.ResourceExhausted, fmt.Sprintf("failed to create server: %v", err))
			}

			return nil, status.Error(codes.Unavailable, fmt.Sprintf("failed to create server: %v", err))
		}
	}

	if err := p.patchNetworkInterface(ctx, projectID, server.ID, providerSpec); err != nil {
		klog.Errorf("Failed to patch network interface for server %q: %v", req.Machine.Name, err)
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("failed to patch network interface for server: %v", err))
	}

	// Generate ProviderID in format: stackit://<projectId>/<serverId>
	providerID := fmt.Sprintf("%s://%s/%s", StackitProviderName, projectID, server.ID)
	klog.V(2).Infof("Successfully created server %q with ID %q for machine %q", server.Name, server.ID, req.Machine.Name)

	return &driver.CreateMachineResponse{
		ProviderID: providerID,
		NodeName:   req.Machine.Name,
	}, nil
}

//nolint:gocyclo // TODO: will be fixed next PR
func (p *Provider) createServerRequest(req *driver.CreateMachineRequest, providerSpec *api.ProviderSpec) *client.CreateServerRequest {
	// Build labels: merge ProviderSpec labels with MCM-specific labels
	labels := make(map[string]string)
	// Start with user-provided labels from ProviderSpec
	if providerSpec.Labels != nil {
		for k, v := range providerSpec.Labels {
			labels[k] = v
		}
	}
	// Add MCM-specific labels for server identification and orphan VM detection
	labels[StackitMachineLabel] = req.Machine.Name
	labels[StackitMachineClassLabel] = req.MachineClass.Name
	labels[StackitRoleLabel] = "node"

	// Create server request
	createReq := &client.CreateServerRequest{
		Name:        req.Machine.Name,
		MachineType: providerSpec.MachineType,
		ImageID:     providerSpec.ImageID,
		Labels:      labels,
	}

	// Add networking configuration (required in v2 API)
	// If not specified in ProviderSpec, try to use networkId from Secret, or use empty
	if providerSpec.Networking != nil {
		createReq.Networking = &client.ServerNetworkingRequest{
			NetworkID: providerSpec.Networking.NetworkID,
			NICIDs:    providerSpec.Networking.NICIDs,
		}
	} else {
		// v2 API requires networking field - use networkId from Secret if available
		// This allows tests/deployments to specify a default network without modifying each MachineClass
		networkID := string(req.Secret.Data["networkId"])
		createReq.Networking = &client.ServerNetworkingRequest{
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
		createReq.BootVolume = &client.BootVolumeRequest{
			DeleteOnTermination: providerSpec.BootVolume.DeleteOnTermination,
			PerformanceClass:    providerSpec.BootVolume.PerformanceClass,
			Size:                providerSpec.BootVolume.Size,
		}

		// Add boot volume source if specified
		if providerSpec.BootVolume.Source != nil {
			createReq.BootVolume.Source = &client.BootVolumeSourceRequest{
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
		createReq.Agent = &client.AgentRequest{
			Provisioned: providerSpec.Agent.Provisioned,
		}
	}

	// Add metadata if specified
	if len(providerSpec.Metadata) > 0 {
		createReq.Metadata = providerSpec.Metadata
	}

	return createReq
}

func (p *Provider) getServerByName(ctx context.Context, projectID, region, serverName string) (*client.Server, error) {
	// Check if the server got already created
	labelSelector := map[string]string{
		StackitMachineLabel: serverName,
	}
	servers, err := p.client.ListServers(ctx, projectID, region, labelSelector)
	if err != nil {
		return nil, fmt.Errorf("SDK ListServers with labelSelector: %v failed: %w", labelSelector, err)
	}

	if len(servers) > 1 {
		return nil, fmt.Errorf("%v servers found for server name %v", len(servers), serverName)
	}

	if len(servers) == 1 {
		return servers[0], nil
	}

	// no servers found len == 0
	return nil, nil
}

func (p *Provider) patchNetworkInterface(ctx context.Context, projectID, serverID string, providerSpec *api.ProviderSpec) error {
	if len(providerSpec.AllowedAddresses) == 0 {
		return nil
	}

	nics, err := p.client.GetNICsForServer(ctx, projectID, providerSpec.Region, serverID)
	if err != nil {
		return fmt.Errorf("failed to get NICs for server %q: %w", serverID, err)
	}

	if len(nics) == 0 {
		return fmt.Errorf("failed to find NIC for server %q", serverID)
	}

	for _, nic := range nics {
		// if networking is not set, server is inside the default network
		// just patch the interface since the server should only have one
		if providerSpec.Networking != nil {
			// only process interfaces that are either in the configured network (NetworkID) or are defined in NICIDs
			if providerSpec.Networking.NetworkID != nic.NetworkID && !slices.Contains(providerSpec.Networking.NICIDs, nic.ID) {
				continue
			}
		}

		updateNic := false
		// check if every cidr in providerspec.allowedAddresses is inside the nic allowedAddresses
		for _, allowedAddress := range providerSpec.AllowedAddresses {
			if !slices.Contains(nic.AllowedAddresses, allowedAddress) {
				nic.AllowedAddresses = append(nic.AllowedAddresses, allowedAddress)
				updateNic = true
			}
		}

		if !updateNic {
			continue
		}

		if _, err := p.client.UpdateNIC(ctx, projectID, providerSpec.Region, nic.NetworkID, nic.ID, nic.AllowedAddresses); err != nil {
			return fmt.Errorf("failed to update allowed addresses for NIC %s: %w", nic.ID, err)
		}

		klog.V(2).Infof("Updated allowed addresses for NIC %s to %v", nic.ID, nic.AllowedAddresses)
	}

	return nil
}
