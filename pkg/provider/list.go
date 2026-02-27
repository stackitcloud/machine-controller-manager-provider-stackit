package provider

import (
	"context"
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/klog/v2"
)

// ListMachines lists all STACKIT servers that belong to the specified MachineClass
//
// This method retrieves all servers in the STACKIT project and filters them based on
// the "kubernetes.io/machineclass" label. This enables the MCM safety controller
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
	projectID, serviceAccountKey := extractSecretCredentials(req.Secret.Data)

	// Initialize client on first use (lazy initialization)
	if err := p.ensureClient(serviceAccountKey); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to initialize STACKIT client: %v", err))
	}

	// Decode ProviderSpec from MachineClass
	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Call STACKIT API to list all servers
	labelSelector := map[string]string{
		p.GetMachineClassLabelKey(): req.MachineClass.Name,
	}
	servers, err := p.client.ListServers(ctx, projectID, providerSpec.Region, labelSelector)
	if err != nil {
		klog.Errorf("Failed to list servers for MachineClass %q: %v", req.MachineClass.Name, err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list servers: %v", err))
	}

	// Filter servers by MachineClass label
	// We use the label to identify which servers belong to this MachineClass
	machineList := make(map[string]string)
	for _, server := range servers {
		// Generate ProviderID in format: stackit://<projectId>/<serverId>
		providerID := fmt.Sprintf("stackit://%s/%s", projectID, server.ID)

		// Get machine name from labels (fallback to server name if not found)
		machineName := server.Name
		if machineLabel, ok := server.Labels[p.GetMachineLabelKey()]; ok {
			machineName = machineLabel
		}

		machineList[providerID] = machineName
	}

	klog.V(2).Infof("Found %d machines for MachineClass %q", len(machineList), req.MachineClass.Name)

	return &driver.ListMachinesResponse{
		MachineList: machineList,
	}, nil
}
