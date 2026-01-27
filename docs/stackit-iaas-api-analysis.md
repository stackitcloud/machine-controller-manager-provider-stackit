# STACKIT IAAS API Analysis for MCM Provider

> **Generated:** 2025-10-28
> **Source:** STACKIT API mock server OpenAPI spec

## Overview

This document analyzes the STACKIT IAAS API to inform the design of our Machine Controller Manager provider implementation.

## API Endpoints for Server (VM) Management

### Core CRUD Operations

| Operation     | HTTP Method | Endpoint                                      | MCM Method         |
| ------------- | ----------- | --------------------------------------------- | ------------------ |
| List servers  | GET         | `/v1/projects/{projectId}/servers`            | ListMachines()     |
| Create server | POST        | `/v1/projects/{projectId}/servers`            | CreateMachine()    |
| Get server    | GET         | `/v1/projects/{projectId}/servers/{serverId}` | GetMachineStatus() |
| Update server | PATCH       | `/v1/projects/{projectId}/servers/{serverId}` | (optional)         |
| Delete server | DELETE      | `/v1/projects/{projectId}/servers/{serverId}` | DeleteMachine()    |

### Lifecycle Operations

Additional endpoints available (may be useful for future enhancements):

- `/v1/projects/{projectId}/servers/{serverId}/start` - Start stopped server
- `/v1/projects/{projectId}/servers/{serverId}/stop` - Stop running server
- `/v1/projects/{projectId}/servers/{serverId}/reboot` - Reboot server
- `/v1/projects/{projectId}/servers/{serverId}/resize` - Change machine type
- `/v1/projects/{projectId}/servers/{serverId}/deallocate` - Deallocate resources
- `/v1/projects/{projectId}/servers/{serverId}/rescue` - Enter rescue mode
- `/v1/projects/{projectId}/servers/{serverId}/console` - Access console
- `/v1/projects/{projectId}/servers/{serverId}/log` - Get server logs

### Networking Operations

- `/v1/projects/{projectId}/servers/{serverId}/nics` - Manage NICs
- `/v1/projects/{projectId}/servers/{serverId}/networks/{networkId}` - Attach/detach networks
- `/v1/projects/{projectId}/servers/{serverId}/public-ips/{publicIpId}` - Manage public IPs
- `/v1/projects/{projectId}/servers/{serverId}/security-groups/{securityGroupId}` - Manage security groups

### Storage Operations

- `/v1/projects/{projectId}/servers/{serverId}/volume-attachments` - Manage volume attachments

### Service Account Operations

- `/v1/projects/{projectId}/servers/{serverId}/service-accounts` - Manage service account access

## CreateServerPayload Schema

### Required Fields

- **`name`** (string) - Server name (MCM will use Machine CR name)
- **`machineType`** (string) - Machine/instance type (e.g., "c2i.2", "m2i.8")

### Optional Fields

**Compute Configuration:**

- `imageId` (UUID) - OS image to use for boot disk
- `availabilityZone` (string) - Availability zone for server placement
- `affinityGroup` (UUID) - Affinity/anti-affinity group for server placement

**Storage Configuration:**

- `bootVolume` (object) - Boot disk configuration
  - Likely includes size, type, etc. (TBD: check nested schema)
- `volumes` (UUID[]) - Additional volume IDs to attach at creation

**Networking Configuration:**

- `networking` (object) - Network configuration
  - Two variants: `CreateServerNetworking` or `CreateServerNetworkingWithNics`
  - TBD: Investigate exact structure
- `securityGroups` (string[]) - Security group names (writeOnly)

**Access Configuration:**

- `keypairName` (string) - SSH keypair name for access
- `serviceAccountMails` (string[]) - Service account emails for server identity

**Metadata & Customization:**

- `labels` (object/map) - Key-value labels for tagging and identification
- `metadata` (object/map) - Additional metadata
- `userData` (string) - Cloud-init/user data script (base64 encoded?)

**Agent Configuration:**

- `agent` (object) - STACKIT agent configuration
  - TBD: Investigate purpose and structure

### Read-Only Response Fields

These fields are returned when getting/listing servers but cannot be set on creation:

**Identifiers & Status:**

- `id` (UUID) - Server unique identifier
- `status` (string) - Server lifecycle status
- `powerStatus` (string) - Power state of the server

**Network Information:**

- `nics` (array) - Network interface card details

**Timestamps:**

- `createdAt` (ISO 8601) - Server creation timestamp
- `launchedAt` (ISO 8601) - Server launch timestamp
- `updatedAt` (ISO 8601) - Last update timestamp

**Maintenance & Errors:**

- `maintenanceWindow` (object) - Maintenance schedule
- `errorMessage` (string) - Error details if server is in error state

## Supporting Resources

### Machine Types

- Endpoint: `/v1/projects/{projectId}/machine-types`
- Get specific type: `/v1/projects/{projectId}/machine-types/{machineType}`
- Used to validate `machineType` field

### Images

- Endpoint: `/v1/projects/{projectId}/images`
- Get specific image: `/v1/projects/{projectId}/images/{imageId}`
- Used to validate `imageId` field

### Keypairs

- Endpoint: `/v1/keypairs`
- Get specific keypair: `/v1/keypairs/{keypairName}`
- Used to validate `keypairName` field

### Networks

- Endpoint: `/v1/projects/{projectId}/networks`
- Required for networking configuration

### Availability Zones

- Endpoint: `/v1/availability-zones`
- Used to validate `availabilityZone` field

## ProviderSpec Design Recommendations

Based on the API analysis, our ProviderSpec should include:

```go
type ProviderSpec struct {
    // Required fields
    MachineType string `json:"machineType"` // e.g., "c2i.2", "m2i.8"

    // Compute configuration
    ImageID          string `json:"imageId"`          // OS image UUID
    AvailabilityZone string `json:"availabilityZone,omitempty"` // AZ name
    AffinityGroup    string `json:"affinityGroup,omitempty"`    // Affinity group UUID

    // Storage configuration
    BootVolume *BootVolumeSpec `json:"bootVolume,omitempty"` // Boot disk config
    Volumes    []string        `json:"volumes,omitempty"`    // Additional volume UUIDs

    // Networking configuration
    Networking     *NetworkingSpec `json:"networking"`                // Network config
    SecurityGroups []string        `json:"securityGroups,omitempty"` // Security group names

    // Access configuration
    KeypairName         string   `json:"keypairName,omitempty"`         // SSH key
    ServiceAccountMails []string `json:"serviceAccountMails,omitempty"` // Service accounts

    // Metadata & customization
    Labels   map[string]string `json:"labels,omitempty"`   // For tagging/identification
    Metadata map[string]string `json:"metadata,omitempty"` // Additional metadata
    UserData string            `json:"userData,omitempty"` // Cloud-init script

    // Agent configuration
    Agent *AgentSpec `json:"agent,omitempty"` // STACKIT agent config
}

// Nested types (TBD: Define based on API spec)
type BootVolumeSpec struct {
    // TODO: Define fields from API schema
}

type NetworkingSpec struct {
    // TODO: Define fields from CreateServerNetworking schema
}

type AgentSpec struct {
    // TODO: Define fields from ServerAgent schema
}
```

## Server Identification Strategy

### ProviderID Format

Format: `stackit://<projectId>/<serverId>`

Example: `stackit://my-project-123/550e8400-e29b-41d4-a716-446655440000`

**Rationale:**

- Unique across STACKIT projects
- Contains both project and server ID for easy API calls
- Follows pattern used by other cloud providers (aws://, azure://)

### Server Tagging via Labels

Use the `labels` field for MCM identification and mapping:

| Label Key                    | Value             | Purpose                                         |
| ---------------------------- | ----------------- | ----------------------------------------------- |
| `kubernetes.io/machine`      | Machine CR name   | Map server to Kubernetes Machine                |
| `kubernetes.io/machineclass` | MachineClass name | Map server to MachineClass for orphan detection |

Example labels:

```json
{
  "kubernetes.io/machine": "worker-pool-a-12345",
  "kubernetes.io/machineclass": "worker-pool-a"
}
```

**Critical for:**

- ListMachines() - Filter servers by MachineClass
- Orphan VM detection - Identify servers without corresponding Machine CRs
- Debugging - Trace servers back to Kubernetes objects

## Server Status Mapping

Need to map STACKIT server status values to MCM/Kubernetes states.

**TODO:** Document exact status values from:

1. Real STACKIT API documentation
2. Testing with mock server
3. Observing real server lifecycle

Expected status values (to be confirmed):

- `CREATING` / `BUILDING` - Server is being created
- `ACTIVE` / `RUNNING` - Server is running
- `STOPPED` / `SHUTOFF` - Server is stopped
- `DELETING` - Server is being deleted
- `ERROR` - Server encountered an error
- `UNKNOWN` - Status cannot be determined

**MCM Status Codes:**

- Use `codes.OK` for running servers
- Use `codes.NotFound` for deleted/not-found servers
- Use `codes.Unknown` for error states
- Use `codes.Unavailable` for servers that are starting/stopping

## Authentication & Authorization

**Project ID:** Required in all API paths (`/v1/projects/{projectId}/...`)

**Authentication Methods (TBD):**

- API tokens
- Service account credentials
- OAuth 2.0

**Secret Structure (proposed):**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: stackit-credentials
type: Opaque
stringData:
  projectId: "my-project-123"
  # One of:
  apiToken: "..."
  # OR
  serviceAccountEmail: "..."
  serviceAccountKey: "..." # JSON key
```

**Environment Variables (for e2e tests):**

- `STACKIT_API_ENDPOINT` - API base URL
- `STACKIT_PROJECT_ID` - Project ID
- `STACKIT_NO_AUTH=true` - Bypass auth for mock server

## Next Steps

### Immediate (Phase 1.2 - API Research)

- [ ] Investigate nested schemas:
  - [ ] `BootVolume` structure
  - [ ] `CreateServerNetworking` vs `CreateServerNetworkingWithNics`
  - [ ] `ServerAgent` structure
- [ ] Document server status enum values
- [ ] Check real STACKIT API documentation for:
  - [ ] Authentication methods
  - [ ] Rate limiting
  - [ ] Error response formats
  - [ ] Pagination for list operations
- [ ] Test mock server endpoints:
  - [ ] Create server request/response
  - [ ] List servers with filtering
  - [ ] Get server by ID
  - [ ] Delete server

### Phase 1.3 - ProviderSpec Design

- [ ] Define complete ProviderSpec with all nested types
- [ ] Create example `samples/machine-class.yaml`
- [ ] Create example `samples/secret.yaml`
- [ ] Define validation rules for each field
- [ ] Write validation unit tests (TDD)

### Phase 1.4 - Technical Design

- [ ] Document error handling strategy
- [ ] Define retry/backoff policies
- [ ] Create sequence diagrams for:
  - [ ] CreateMachine flow
  - [ ] DeleteMachine flow
  - [ ] GetMachineStatus flow
  - [ ] ListMachines flow

## References

- **Mock Server Repository:** `github.com/stackit-controllers-k8s/stackit-api-mockservers`
- **OpenAPI Spec Location:** `config/apis/iaas/specs/openapi/iaas.json`
- **Real API Specs:** `github.com/stackitcloud/stackit-api-specifications/tree/main/services/iaas`
- **MCM Documentation:** https://gardener.cloud/docs/other-components/machine-controller-manager/
- **Provider Implementation Guide:** https://gardener.cloud/docs/other-components/machine-controller-manager/cp_support_new/
