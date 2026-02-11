# machine-controller-manager-provider-stackit

[![GitHub License](https://img.shields.io/github/license/stackitcloud/machine-controller-manager-provider-stackit)](https://www.apache.org/licenses/LICENSE-2.0)

Out of tree (controller based) implementation for `STACKIT` as a provider for Gardener.

A Machine Controller Manager (MCM) provider implementation for STACKIT cloud infrastructure. This provider enables Gardener to manage virtual machines on STACKIT using the declarative Kubernetes API.

The provider was built following the [MCM provider development guidelines](https://github.com/gardener/machine-controller-manager/blob/master/docs/development/cp_support_new.md) and bootstrapped from the [sample provider template](https://github.com/gardener/machine-controller-manager-provider-sampleprovider).

## Getting Started

### Deployment

See the [samples/](./samples/) directory for example manifests including:

- [`secret.yaml`](./samples/secret.yaml) - STACKIT credentials configuration
- [`machine-class.yaml`](./samples/machine-class.yaml) - MachineClass definition
- [`machine.yaml`](./samples/machine.yaml) - Individual Machine example
- [`machine-deployment.yaml`](./samples/machine-deployment.yaml) - MachineDeployment for scaled workloads
- [`deployment.yaml`](./samples/deployment.yaml) - Provider controller deployment

### Minimal MachineClass Example

Here's a bare minimum MachineClass configuration:

```yaml
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: stackit-machine
  namespace: default
providerSpec:
  region: eu01
  machineType: c2i.2
  imageId: "12345678-1234-1234-1234-123456789012"
secretRef:
  name: stackit-credentials
  namespace: default
```

For detailed information on all available configuration fields, see the [MachineClass documentation](./docs/machine-class.md).

## Local Testing & Development

Use the Makefile targets for development and testing:

```sh
# Run tests
make test

# Verify code formatting and run all checks
make verify

# Format code
make fmt

# Build container image
make image
```

## STACKIT SDK Integration

This provider uses the official [STACKIT Go SDK](https://github.com/stackitcloud/stackit-sdk-go) for all interactions with the STACKIT IaaS API. The SDK provides type-safe API access, built-in authentication handling, and is officially maintained by STACKIT.

Each provider instance is bound to a single STACKIT project via the service account credentials provided in the Secret. The SDK client is initialized once on first use and automatically handles token refresh. In Gardener deployments, each shoot cluster gets its own control plane with a dedicated MCM and provider instance.

### Authentication & Credentials

The provider requires STACKIT credentials to be provided via a Kubernetes Secret. The Secret must contain the following fields:

| Field                 | Required | Description                                                      |
| --------------------- | -------- | ---------------------------------------------------------------- |
| `project-id`          | Yes      | STACKIT project UUID                                             |
| `serviceaccount.json` | Yes      | STACKIT service account credentials (JSON format)                |
| `userData`            | No       | Default cloud-init user data (can be overridden in ProviderSpec) |

The service account key should be obtained from the STACKIT Portal (Project Settings → Service Accounts → Create Key) and contains JWT credentials and a private key for secure authentication.

**Credential Rotation:** The provider captures credentials on first use and reuses the same STACKIT SDK client for all subsequent requests (the SDK automatically handles token refresh). If the Secret is updated with new credentials, the provider pod must be restarted to pick up the changes. This follows the standard Kubernetes pattern for credential rotation.

### Environment Variables

The provider supports the following environment variables for configuration:

| Variable                | Default       | Description                                                        |
| ----------------------- | ------------- | ------------------------------------------------------------------ |
| `STACKIT_IAAS_ENDPOINT` | (SDK default) | Override STACKIT API endpoint URL (useful for testing)             |
| `STACKIT_TOKEN_BASEURL` | (SDK default) | Override STACKIT Token endpoint URL (useful for testing)           |
| `STACKIT_NO_AUTH`       | `false`       | Skip authentication (for testing with mock servers, set to `true`) |

**Note:** `STACKIT_NO_AUTH=true` is only intended for testing environments with mock servers. It skips the authenticaiton step and communicates with the STACKIT API without authenticating itself. Do not use in production.

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
- [IaaS API v2 Documentation](https://docs.api.stackit.cloud/documentation/iaas/version/v2) - STACKIT IaaS REST API reference
