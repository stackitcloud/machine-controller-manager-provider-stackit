// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package provider contains the cloud provider specific implementations to manage machines
package provider

import (
	"github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
)

// Provider is the struct that implements the driver interface
// It is used to implement the basic driver functionalities
type Provider struct {
	SPI    spi.SessionProviderInterface
	client StackitClient // STACKIT API client (can be mocked for testing)
}

// NewProvider returns an empty provider object
func NewProvider(spi spi.SessionProviderInterface) driver.Driver {
	return &Provider{
		SPI:    spi,
		client: newSDKStackitClient(), // Initialize SDK client for STACKIT API
	}
}
