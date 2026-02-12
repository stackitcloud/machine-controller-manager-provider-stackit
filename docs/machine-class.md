# MachineClass ProviderSpec

This document describes the STACKIT MachineClass ProviderSpec schema and validation rules used by the machine-controller-manager-provider-stackit. It is generated based on the [providerSpec source code](../pkg/provider/apis/provider_spec.go).

## Overview

A MachineClass defines how STACKIT servers should be created. The ProviderSpec is the STACKIT-specific section of the MachineClass and contains all server configuration fields.

## Required Fields

- `region` (string): STACKIT region such as "eu01" or "eu02".
- `machineType` (string): STACKIT server type such as "c2i.2" or "m2i.8".
- `imageId` (string): UUID of the image to boot from, unless `bootVolume.source` is set.
- `networking` (object): Must be set and must specify either `networkId` or `nicIds`.

## ProviderSpec Fields

| Field                 | Type                   | Required | Description                                                   |
| --------------------- | ---------------------- | -------- | ------------------------------------------------------------- |
| `region`              | string                 | Yes      | STACKIT region (e.g., "eu01", "eu02").                        |
| `machineType`         | string                 | Yes      | STACKIT server type (e.g., "c2i.2", "m2i.8").                 |
| `imageId`             | string                 | Yes\*    | Image UUID. Required unless `bootVolume.source` is specified. |
| `labels`              | map[string]string      | No       | Labels for server identification.                             |
| `networking`          | NetworkingSpec         | Yes      | Network configuration (either `networkId` or `nicIds`).       |
| `allowedAddresses`    | []string               | No       | CIDR ranges allowed for anti-spoofing bypass.                 |
| `securityGroups`      | []string               | No       | Security group UUIDs.                                         |
| `userData`            | string                 | No       | Cloud-init user data (overrides Secret.userData).             |
| `bootVolume`          | BootVolumeSpec         | No       | Boot disk configuration.                                      |
| `volumes`             | []string               | No       | UUIDs of existing volumes to attach.                          |
| `keypairName`         | string                 | No       | SSH keypair name.                                             |
| `availabilityZone`    | string                 | No       | Availability zone (e.g., "eu01-1").                           |
| `affinityGroup`       | string                 | No       | UUID of affinity group.                                       |
| `serviceAccountMails` | []string               | No       | Service account emails (max 1).                               |
| `agent`               | AgentSpec              | No       | STACKIT agent configuration.                                  |
| `metadata`            | map[string]interface{} | No       | Freeform metadata.                                            |

## NetworkingSpec

Exactly one of the following must be set:

- `networkId` (string): UUID of the network to attach.
- `nicIds` ([]string): UUIDs of pre-created NICs.

## BootVolumeSpec

- `deleteOnTermination` (bool, optional): Delete boot volume with server. Default is true.
- `performanceClass` (string, optional): Storage performance tier (for example, "standard", "premium").
- `size` (int, optional): Size in GB. Must be at least the image size.
- `source` (BootVolumeSourceSpec, optional): Use this instead of `imageId`.

### BootVolumeSourceSpec

- `type` (string): One of "image", "snapshot", or "volume".
- `id` (string): UUID of the source object.

## AgentSpec

- `provisioned` (bool, optional): Whether the STACKIT agent is installed.

## Validation Rules

- `region` must match `^[a-z0-9]+$` (example: "eu01").
- `machineType` must match `^[a-z]+\d+[a-z]*\.\d+[a-z]*(\.[a-z]+\d+)*$` (examples: "c2i.2", "m2i.8").
- `imageId`, `volumes[]`, and `affinityGroup` must be valid UUIDs.
- `availabilityZone` must match `^[a-z0-9]+-\d+$` (example: "eu01-1").
- `keypairName` maximum length is 127 and may contain only `A-Z`, `a-z`, `0-9`, `@`, `.`, `_`, `-`.
- `labels` keys and values follow Kubernetes label rules and are limited to 63 characters.
- `allowedAddresses` entries must be valid CIDR blocks.
- `serviceAccountMails` allows a maximum of 1 entry, and each must be a valid email address.
- `networking` is required and must set exactly one of `networkId` or `nicIds`.

## Secret Requirements

MachineClass references a Secret via `secretRef`. The Secret must include:

- `project-id`: STACKIT project UUID.
- `serviceaccount.json`: Service account key JSON.
- `userData` (optional): Default cloud-init user data. Can be overridden by ProviderSpec `userData`.

## Examples

Minimal example:

```yaml
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: minimal-mc
  namespace: default
providerSpec:
  region: "eu01"
  machineType: "c2i.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    networkId: "770e8400-e29b-41d4-a716-446655440000"
secretRef:
  name: test-secret
  namespace: default
```

Extended example:

```yaml
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: full-example-mc
  namespace: default
providerSpec:
  region: "eu01"
  machineType: "c2i.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    networkId: "770e8400-e29b-41d4-a716-446655440000"
  securityGroups:
    - "660e8400-e29b-41d4-a716-446655440000"
  userData: |
    #cloud-config
    runcmd:
      - echo "Bootstrapped"
  bootVolume:
    size: 50
    performanceClass: "standard"
  volumes:
    - "880e8400-e29b-41d4-a716-446655440000"
  keypairName: "my-ssh-key"
  availabilityZone: "eu01-1"
  affinityGroup: "880e8400-e29b-41d4-a716-446655440000"
  serviceAccountMails:
    - "my-service@sa.stackit.cloud"
  agent:
    provisioned: true
  metadata:
    environment: "production"
    cost-center: "12345"
secretRef:
  name: test-secret
  namespace: default
```
