// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package provider contains the cloud provider specific implementations to manage machines
package provider

import (
	"fmt"
	"sync"

	"github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
)

// Provider is the struct that implements the driver interface
// It is used to implement the basic driver functionalities
//
// Architecture: Single-tenant design
// - Each provider instance is deployed per Gardener shoot (cluster)
// - The STACKIT IaaS client is initialized lazily on first request using credentials from the Secret
// - All subsequent requests reuse the same client (SDK handles token refresh automatically)
type Provider struct {
	SPI        spi.SessionProviderInterface
	client     StackitClient // STACKIT API client (can be mocked for testing)
	clientOnce sync.Once     // Ensures client is initialized exactly once
	clientErr  error         // Stores initialization error if any
}

// NewProvider returns an empty provider object
func NewProvider(spi spi.SessionProviderInterface) driver.Driver {
	return &Provider{
		SPI: spi,
	}
}

// ensureClient initializes the STACKIT client on first use (lazy initialization)
// This is called by all methods that need to interact with STACKIT API
// Thread-safe via sync.Once
// If a client is already set (e.g., mock client in tests), initialization is skipped
func (p *Provider) ensureClient(serviceAccountKey string) error {
	// If client is already set (e.g., mock client in tests), skip initialization
	if p.client != nil {
		return nil
	}

	p.clientOnce.Do(func() {
		client, err := NewStackitClient(serviceAccountKey)
		if err != nil {
			p.clientErr = fmt.Errorf("failed to initialize STACKIT client: %w", err)
			return
		}
		p.client = client
	})
	return p.clientErr
}
