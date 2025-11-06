# machine-controller-manager-provider-stackit

[![REUSE status](https://api.reuse.software/badge/github.com/aoepeople/machine-controller-manager-provider-stackit)](https://api.reuse.software/info/github.com/aoepeople/machine-controller-manager-provider-stackit)

[![REUSE status](https://api.reuse.software/badge/github.com/aoepeople/machine-controller-manager-provider-stackit)](https://api.reuse.software/info/github.com/aoepeople/machine-controller-manager-provider-stackit)

Out of tree (controller based) implementation for `STACKIT` as a provider for Gardener.

A Machine Controller Manager (MCM) external provider implementation for STACKIT cloud infrastructure. This provider enables Gardener to manage virtual machines on STACKIT using the declarative Kubernetes API.

The provider was built following the [MCM provider development guidelines](https://github.com/gardener/machine-controller-manager/blob/master/docs/development/cp_support_new.md) and bootstrapped from the [sample provider template](https://github.com/gardener/machine-controller-manager-provider-sampleprovider).Following are the basic principles kept in mind while developing the external plugin.

## Project Structure

```
machine-controller-manager-provider-stackit/
├── cmd/
│   └── machine-controller/
│       └── main.go                    # Provider entrypoint
├── pkg/
│   ├── provider/
│   │   ├── core.go                    # Core provider implementation
│   │   ├── provider.go                # Driver interface implementation
│   │   ├── stackit_client.go          # STACKIT client interface
│   │   ├── sdk_client.go              # STACKIT SDK wrapper implementation
│   │   ├── helpers.go                 # SDK type conversion utilities
│   │   ├── apis/
│   │   │   ├── provider_spec.go       # ProviderSpec CRD definitions
│   │   │   └── validation/            # Field validation logic
│   │   └── *_test.go                  # Unit tests
│   └── spi/
│       └── spi.go                     # Service provider interface
├── test/
│   └── e2e/                           # End-to-end integration tests
├── samples/                           # Example manifests
├── kubernetes/                        # Deployment manifests
└── vendor/                            # Go module dependencies
```

## Getting Started

### Prerequisites

- **[Hermit](https://cashapp.github.io/hermit/)** - Environment manager that provides isolated, reproducible tooling
- Access to a STACKIT project with API credentials
- Docker (for building container images)

### Tool Management

This project uses **Hermit** for reproducible development environments and **just** as the command runner.

**Hermit** automatically manages tool versions (Go, kubectl, kind, ginkgo, etc.) defined in `bin/hermit.hcl`. When you activate the Hermit environment, all required tools are available without manual installation:

```sh
# Activate hermit environment (tools auto-install on first use)
. ./bin/activate-hermit

# Or install hermit shell hooks for automatic activation
hermit shell-hooks
```

**just** is the task runner (defined in `justfile`). It provides a cleaner syntax than Make and better task organization:

```sh
# List all available commands
just --list

# Or just run 'just' with no arguments
just
```

### Quick Start

```sh
just build              # Build the provider binary
just test               # Run unit tests
just test-e2e           # Run end-to-end tests
just dev                # Complete local dev setup (cluster + deployment)
just start              # Run provider locally for debugging
just docker-build       # Build container image
```

**NOTE:** Run `just --list` for more information on all available commands.

### Deployment

See the [samples/](./samples/) directory for example manifests including:
- `secret.yaml` - STACKIT credentials configuration
- `machine-class.yaml` - MachineClass definition
- `machine.yaml` - Individual Machine example
- `machine-deployment.yaml` - MachineDeployment for scaled workloads
- `deployment.yaml` - Provider controller deployment

Deploy using standard kubectl commands:

```sh
kubectl apply -f samples/secret.yaml
kubectl apply -f samples/machine-class.yaml
kubectl apply -f samples/machine.yaml
```

## STACKIT SDK Integration

This provider uses the official [STACKIT Go SDK](https://github.com/stackitcloud/stackit-sdk-go) for all interactions with the STACKIT IaaS API. The SDK provides type-safe API access, built-in authentication handling, and is officially maintained by STACKIT.

The SDK client is stateless and supports different credentials per MachineClass, allowing multi-tenancy scenarios where different machine pools use different STACKIT projects.

### Authentication & Credentials

The provider requires STACKIT credentials to be provided via a Kubernetes Secret. The Secret must contain the following fields:

| Field | Required | Description |
|-------|----------|-------------|
| `projectId` | Yes | STACKIT project UUID |
| `stackitToken` | Yes | STACKIT API authentication token |
| `region` | Yes | STACKIT region (e.g., `eu01-1`, `eu01-2`) |
| `userData` | No | Default cloud-init user data (can be overridden in ProviderSpec) |
| `networkId` | No | Default network UUID (can be overridden in ProviderSpec) |

## Configuration Reference

### ProviderSpec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `machineType` | string | Yes | STACKIT server type (e.g., "c1.2", "m1.4") |
| `imageId` | string | Yes | UUID of the OS image |
| `labels` | map[string]string | No | Labels for server identification |
| `networking` | NetworkingSpec | No | Network configuration (NetworkID or NICIDs) |
| `securityGroups` | []string | No | Security group names |
| `userData` | string | No | Cloud-init user data (overrides Secret.userData) |
| `bootVolume` | BootVolumeSpec | No | Boot disk configuration |
| `volumes` | []string | No | UUIDs of additional volumes to attach |
| `keypairName` | string | No | SSH keypair name |
| `availabilityZone` | string | No | Availability zone (e.g., "eu01-1") |
| `affinityGroup` | string | No | UUID of affinity group |
| `serviceAccountMails` | []string | No | Service account email addresses (max 1) |
| `agent` | AgentSpec | No | STACKIT agent configuration |
| `metadata` | map[string]interface{} | No | Custom metadata key-value pairs |

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make changes and add tests
4. Run verification: `just test && just golang::lint`
5. Commit with meaningful messages
6. Push and create a Pull Request

### Local Testing

Use the local development environment for rapid iteration:

```sh
# Set up dev environment
just dev

# Or run provider locally for debugging
just start
```

## References

### Machine Controller Manager
- [Machine Controller Manager](https://github.com/gardener/machine-controller-manager) - Core MCM project
- [MCM Provider Development Guide](https://github.com/gardener/machine-controller-manager/blob/master/docs/development/cp_support_new.md) - Guidelines followed to build this provider
- [MCM Sample Provider](https://github.com/gardener/machine-controller-manager-provider-sampleprovider) - Original template used as starting point
- [MCM Driver Interface](https://github.com/gardener/machine-controller-manager/blob/master/pkg/util/provider/driver/driver.go) - Provider contract interface

### STACKIT SDK
- [STACKIT SDK Go](https://github.com/stackitcloud/stackit-sdk-go) - Official STACKIT Go SDK
- [IaaS Service Package](https://github.com/stackitcloud/stackit-sdk-go/tree/main/services/iaas) - IaaS service API documentation
- [SDK Core Package](https://github.com/stackitcloud/stackit-sdk-go/tree/main/core) - Core SDK configuration and authentication
- [SDK Examples](https://github.com/stackitcloud/stackit-sdk-go/tree/main/examples) - Code examples and usage patterns
- [SDK Releases](https://github.com/stackitcloud/stackit-sdk-go/releases) - Release notes and changelog

### STACKIT Platform
- [STACKIT Documentation](https://docs.stackit.cloud/) - STACKIT cloud platform documentation
- [STACKIT Portal](https://portal.stackit.cloud/) - STACKIT management console
- [Service Accounts](https://docs.stackit.cloud/stackit/en/service-accounts-134415819.html) - Creating and managing service accounts
- [Service Account Keys](https://docs.stackit.cloud/stackit/en/usage-of-the-service-account-keys-in-stackit-175112464.html) - API authentication setup
- [IaaS API Documentation](https://docs.stackit.cloud/) - STACKIT IaaS REST API reference

## License

Copyright 2024 SAP SE or an SAP affiliate company and Gardener contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
