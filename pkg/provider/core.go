// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package provider contains the cloud provider specific implementations to manage machines
package provider

import (
	"context"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/klog/v2"
)

const (
	StackitProviderName      = "stackit"
	StackitMachineLabel      = "mcm.gardener.cloud/machine"
	StackitMachineClassLabel = "mcm.gardener.cloud/machineclass"
	StackitRoleLabel         = "mcm.gardener.cloud/role"
)

// GetVolumeIDs extracts volume IDs from PersistentVolume specs
//
// This method is used by MCM to get volume IDs for persistent volumes.
// Currently unimplemented for STACKIT provider - volumes are managed directly
// through the ProviderSpec (bootVolume and volumes fields).
//
// Returns:
//   - Unimplemented: This functionality is not required for STACKIT provider
func (p *Provider) GetVolumeIDs(_ context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
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
func (p *Provider) GenerateMachineClassForMigration(_ context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been received for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	return &driver.GenerateMachineClassForMigrationResponse{}, status.Error(codes.Unimplemented, "")
}

// InitializeMachine handles VM initialization for STACKIT VM's. Currently, un-implemented.
func (p *Provider) InitializeMachine(context.Context, *driver.InitializeMachineRequest) (*driver.InitializeMachineResponse, error) {
	return nil, status.Error(codes.Unimplemented, "STACKIT Provider does not yet implement InitializeMachine")
}
