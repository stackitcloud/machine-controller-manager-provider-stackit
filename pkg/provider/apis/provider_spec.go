// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package api

// ProviderSpec is the spec to be used while parsing the calls.
// This is a minimal implementation for Slice #1.
type ProviderSpec struct {
	// MachineType is the STACKIT server type (e.g., "c1.2", "m1.4")
	// Required field for creating a server.
	MachineType string `json:"machineType"`

	// ImageID is the UUID of the OS image to use for the server
	// Required field for creating a server.
	ImageID string `json:"imageId"`

	// Labels are key-value pairs used to tag and identify servers
	// Used by MCM for mapping servers to MachineClasses and orphan VM detection
	// Optional field. MCM will automatically add standard labels.
	Labels map[string]string `json:"labels,omitempty"`
}
