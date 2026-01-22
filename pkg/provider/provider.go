package provider

import (
	"fmt"
	"sync"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	client2 "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/client"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/spi"
	"k8s.io/klog/v2"
)

// Provider is the struct that implements the driver interface
// It is used to implement the basic driver functionalities
//
// Architecture: Single-tenant design
// - Each provider instance is deployed per Gardener shoot (cluster)
// - The STACKIT IaaS client is initialized lazily on first request using credentials from the Secret
// - All subsequent requests reuse the same client (SDK handles token refresh automatically)
// - Credential rotation requires pod restart (standard Kubernetes pattern)
type Provider struct {
	SPI                 spi.SessionProviderInterface
	client              client2.StackitClient // STACKIT API client (can be mocked for testing)
	clientOnce          sync.Once             // Ensures client is initialized exactly once
	clientErr           error                 // Stores initialization error if any
	capturedCredentials string                // Service account key used for initialization (for defensive checks)
}

// NewProvider returns an empty provider object
func NewProvider(i spi.SessionProviderInterface) driver.Driver {
	return &Provider{
		SPI: i,
	}
}

// ensureClient initializes the STACKIT client on first use (lazy initialization)
// This is called by all methods that need to interact with STACKIT API
// Thread-safe via sync.Once
//
// Design: Single-credential lifecycle
// - The serviceAccountKey parameter is captured and used ONLY on the first call
// - Subsequent calls reuse the same client regardless of the serviceAccountKey passed
// - Credential rotation requires pod restart (standard Kubernetes pattern)
// - If a client is already set (e.g., mock client in tests), initialization is skipped
func (p *Provider) ensureClient(serviceAccountKey string) error {
	// If client is already set (e.g., mock client in tests), skip initialization
	if p.client != nil {
		return nil
	}

	p.clientOnce.Do(func() {
		client, err := client2.NewStackitClient(serviceAccountKey)
		if err != nil {
			p.clientErr = fmt.Errorf("failed to initialize STACKIT client: %w", err)
			return
		}
		p.client = client
		p.capturedCredentials = serviceAccountKey
	})

	// Defensive check: warn if credentials changed after initialization
	// This indicates the Secret was updated, which requires pod restart
	if p.clientErr == nil && p.capturedCredentials != serviceAccountKey {
		klog.Warning("Service account credentials changed after client initialization. Credential rotation requires pod restart. Continuing with original credentials.")
	}

	return p.clientErr
}
