# Phase 0: STACKIT SDK Evaluation Report

**Date:** 2025-11-05
**Status:** ✅ APPROVED FOR MIGRATION
**Recommendation:** Proceed with SDK migration as planned

---

## Executive Summary

The STACKIT SDK Go is **mature, well-maintained, and fully supports** all operations required for this provider. The SDK covers all four critical operations (CreateServer, GetServer, DeleteServer, ListServers) and provides a clean, type-safe API with built-in authentication handling.

**Key Finding:** Migration is low-risk and will simplify our codebase significantly.

---

## 1. SDK Maturity Assessment ⭐⭐⭐⭐☆

### Indicators

| Metric | Value | Assessment |
|--------|-------|------------|
| **GitHub Stars** | 78 | Moderate community interest |
| **Contributors** | 26 | Active development team |
| **Release Frequency** | Bi-weekly (every 2-3 weeks) | Very active maintenance |
| **Last Release** | October 29, 2025 | Current and maintained |
| **License** | Apache 2.0 | Enterprise-friendly |
| **Go Version** | 1.21+ | Modern Go support |
| **Breaking Changes** | Well-documented with 6-12 month deprecation windows | Mature change management |

### Stability Notes

- **IAAS v1.0.0 released October 29, 2025** - Major milestone indicating API stability
- Breaking changes are infrequent and well-communicated
- Deprecation notices provide 6-12 months for migration
- Active bug fixes and feature additions

**Verdict:** SDK is production-ready and stable. ✅

---

## 2. API Coverage Verification ✅

### Required Operations

All operations needed by this provider are fully supported:

| Operation | HTTP Client (Current) | SDK Method | Status |
|-----------|----------------------|------------|--------|
| **CreateServer** | `CreateServer(token, projectID, req)` | `CreateServer(ctx, projectID, region).CreateServerPayload(payload).Execute()` | ✅ Supported |
| **GetServer** | `GetServer(token, projectID, serverID)` | `GetServer(ctx, projectID, region, serverID).Execute()` | ✅ Supported |
| **DeleteServer** | `DeleteServer(token, projectID, serverID)` | `DeleteServer(ctx, projectID, region, serverID).Execute()` | ✅ Supported |
| **ListServers** | `ListServers(token, projectID)` | `ListServers(ctx, projectID, region).Execute()` | ✅ Supported |

### API Method Signatures

#### CreateServer
```go
// SDK Pattern
iaasClient.CreateServer(ctx, projectId, region).
    CreateServerPayload(payload).
    Execute()

// Returns: (*iaas.Server, error)
```

#### GetServer
```go
iaasClient.GetServer(ctx, projectId, region, serverId).Execute()
// Returns: (*iaas.Server, error)
```

#### DeleteServer
```go
iaasClient.DeleteServer(ctx, projectId, region, serverId).Execute()
// Returns: error
```

#### ListServers
```go
iaasClient.ListServers(ctx, projectId, region).Execute()
// Returns: (*iaas.ServerListResponse, error)
```

### Additional Operations Available

Beyond our core needs, the SDK provides:

- **Server Management:** UpdateServer, ResizeServer, RescueServer
- **Server Monitoring:** GetServerLog, GetServerConsoleUrl
- **Volume Operations:** CreateVolume, AttachVolume, DetachVolume
- **Networking:** Network creation, NIC management, security groups
- **Public IPs:** Allocation, attachment
- **Keypairs:** SSH key management
- **Images:** Image listing and management
- **Affinity Groups:** Server placement control
- **Service Accounts:** Identity management

**Verdict:** SDK exceeds our requirements. ✅

---

## 3. Authentication Comparison

### Current Implementation (HTTP Client)

```go
// Manual bearer token handling
httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

// Token extracted from Kubernetes Secret
token := string(req.Secret.Data["stackitToken"])
```

**Issues:**
- No token refresh
- Manual header management
- No key-based authentication support

### SDK Implementation

The SDK supports **three authentication methods**:

#### Option 1: Token Flow (Simple Migration Path)
```go
token := string(req.Secret.Data["stackitToken"])
iaasClient, err := iaas.NewAPIClient(config.WithToken(token))
```

**Pros:**
- ✅ Direct drop-in replacement for current approach
- ✅ Minimal code changes
- ✅ Works with existing secrets

**Cons:**
- ⚠️ Long-lived tokens (less secure)
- ⚠️ Manual token rotation required

#### Option 2: Key Flow (Recommended for Production)
```go
saKeyPath := "/path/to/service_account_key.json"
iaasClient, err := iaas.NewAPIClient(
    config.WithServiceAccountKeyPath(saKeyPath),
)
```

**Pros:**
- ✅ Short-lived tokens (better security)
- ✅ Automatic token refresh
- ✅ Supports both STACKIT-generated and custom RSA keys

**Cons:**
- ⚠️ Requires updating deployment secrets (one-time)
- ⚠️ Requires secret volume mount for key file

#### Option 3: Environment Variables (Development)
```bash
export STACKIT_SERVICE_ACCOUNT_KEY_PATH=/path/to/key.json
# or
export STACKIT_SERVICE_ACCOUNT_TOKEN=your-token
```

```go
iaasClient, err := iaas.NewAPIClient()  // Auto-detects credentials
```

### Recommended Migration Strategy

**Phase 1 (This Migration):** Start with Token Flow
- Preserve existing `stackitToken` secret format
- Minimal changes to deployment
- Easy validation

**Phase 2 (Future Enhancement):** Document Key Flow
- Provide migration guide for users
- Keep token flow supported for backward compatibility
- Recommend key flow for new deployments

**Verdict:** Authentication migration is straightforward. ✅

---

## 4. Data Model Comparison

### Field Mapping Analysis

#### CreateServerRequest: HTTP Client → SDK

| Current Field | SDK Field | Type Change | Notes |
|---------------|-----------|-------------|-------|
| `Name` (string) | `Name` (*string) | ⚠️ Pointer | Required field |
| `MachineType` (string) | `MachineType` (*string) | ⚠️ Pointer | Required field |
| `ImageID` (string) | `ImageId` (*string) | ⚠️ Pointer, casing | Optional |
| `Labels` (map[string]string) | `Labels` (*map[string]interface{}) | ⚠️ Pointer, interface{} | Type conversion needed |
| `Networking` (*ServerNetworkingRequest) | `Networking` (*CreateServerPayloadAllOfNetworking) | ⚠️ Different type | Field remapping |
| `SecurityGroups` ([]string) | `SecurityGroups` (*[]string) | ⚠️ Pointer | Simple wrapper |
| `UserData` (string) | `UserData` (*[]byte) | ⚠️ Pointer, base64 bytes | Type conversion needed |
| `BootVolume` (*BootVolumeRequest) | `BootVolume` (*ServerBootVolume) | ⚠️ Different type | Field remapping |
| `Volumes` ([]string) | `Volumes` (*[]string) | ⚠️ Pointer | Simple wrapper |
| `KeypairName` (string) | `KeypairName` (*string) | ⚠️ Pointer | Simple wrapper |
| `AvailabilityZone` (string) | `AvailabilityZone` (*string) | ⚠️ Pointer | Simple wrapper |
| `AffinityGroup` (string) | `AffinityGroup` (*string) | ⚠️ Pointer | Simple wrapper |
| `ServiceAccountMails` ([]string) | `ServiceAccountMails` (*[]string) | ⚠️ Pointer | Simple wrapper |
| `Agent` (*AgentRequest) | `Agent` (*ServerAgent) | ⚠️ Different type | Field remapping |
| `Metadata` (map[string]interface{}) | `Metadata` (*map[string]interface{}) | ⚠️ Pointer | Simple wrapper |

**Key Observations:**

1. **All SDK fields are pointers** - Need helper function:
   ```go
   func ptr[T any](v T) *T {
       return &v
   }
   ```

2. **Type conversions required:**
   - `Labels`: `map[string]string` → `*map[string]interface{}`
   - `UserData`: Base64 string → `*[]byte`

3. **Field naming differences:**
   - `ImageID` → `ImageId` (casing)

4. **Nested structure changes:**
   - `Networking`: Field-level differences (NetworkID vs NetworkId)
   - `BootVolume`: Field-level remapping needed
   - `Agent`: Field-level remapping needed

#### Server Response: HTTP Client → SDK

| Current Field | SDK Field | Change |
|---------------|-----------|--------|
| `ID` (string) | `Id` (*string) | Pointer |
| `Name` (string) | `Name` (*string) | Pointer |
| `Status` (string) | `Status` (*string) | Pointer |
| `Labels` (map[string]string) | `Labels` (*map[string]interface{}) | Pointer + interface{} |

**Verdict:** Mapping is straightforward with helper functions. ✅

---

## 5. SDK Usage Patterns

### Client Initialization Pattern

```go
// Option 1: With explicit token
iaasClient, err := iaas.NewAPIClient(config.WithToken(token))

// Option 2: With service account key
iaasClient, err := iaas.NewAPIClient(
    config.WithServiceAccountKeyPath(keyPath),
)

// Option 3: Auto-discovery from environment
iaasClient, err := iaas.NewAPIClient()
```

### Request Builder Pattern

The SDK uses a **fluent builder pattern**:

```go
// 1. Call method with required parameters
request := iaasClient.CreateServer(ctx, projectId, region)

// 2. Chain payload/options
request = request.CreateServerPayload(payload)

// 3. Execute
result, err := request.Execute()
```

**Benefits:**
- Type-safe parameter validation
- Clear separation of required vs optional parameters
- Consistent API across all operations

### Error Handling

```go
server, err := iaasClient.GetServer(ctx, projectId, region, serverId).Execute()
if err != nil {
    // SDK returns structured errors with HTTP status codes
    return fmt.Errorf("failed to get server: %w", err)
}
```

**SDK provides:**
- Wrapped HTTP errors with status codes
- JSON error response parsing
- Context cancellation support

### Async Operations with Wait Handlers

```go
server, err := iaasClient.CreateServer(ctx, projectId, region).
    CreateServerPayload(payload).
    Execute()

// Wait for server to become ready
err = wait.CreateServerWaitHandler(ctx, iaasClient, projectId, region, *server.Id).
    WaitWithContext(context.Background())
```

**Note:** Wait handlers are optional and useful for E2E tests or scripts.

---

## 6. Current Implementation Analysis

### HTTP Client Features

**File:** `pkg/provider/http_client.go` (230 lines)

**Features we need to preserve:**

1. **Base URL configuration** via `STACKIT_API_ENDPOINT` environment variable
   - Default: `https://api.stackit.cloud`
   - SDK equivalent: Configuration option

2. **Timeout configuration** via `STACKIT_API_TIMEOUT` environment variable
   - Default: 30 seconds
   - SDK equivalent: HTTP client customization

3. **Error handling:**
   - 404 detection for `ErrServerNotFound`
   - HTTP status code parsing
   - SDK equivalent: Built-in error types

4. **Response parsing:**
   - JSON unmarshal of server responses
   - List response unwrapping (`{"items": []}`)
   - SDK equivalent: Automatic

### StackitClient Interface

**File:** `pkg/provider/stackit_client.go` (76 lines)

```go
type StackitClient interface {
    CreateServer(ctx context.Context, token, projectID string, req *CreateServerRequest) (*Server, error)
    GetServer(ctx context.Context, token, projectID, serverID string) (*Server, error)
    DeleteServer(ctx context.Context, token, projectID, serverID string) error
    ListServers(ctx context.Context, token, projectID string) ([]*Server, error)
}
```

**Migration changes:**

```go
// NEW interface (SDK-based)
type StackitClient interface {
    CreateServer(ctx context.Context, projectID, region string, req *iaas.CreateServerPayload) (*iaas.Server, error)
    GetServer(ctx context.Context, projectID, region, serverID string) (*iaas.Server, error)
    DeleteServer(ctx context.Context, projectID, region, serverID string) error
    ListServers(ctx context.Context, projectID, region string) (*iaas.ServerListResponse, error)
}
```

**Key differences:**
- ❌ Remove `token` parameter (SDK handles internally)
- ✅ Add `region` parameter (new in IAAS v1.0.0)
- ✅ Use SDK types for requests/responses

### Provider Core Logic

**File:** `pkg/provider/core.go`

**Current usage pattern:**
```go
// Extract credentials
projectID := string(req.Secret.Data["projectId"])
token := string(req.Secret.Data["stackitToken"])

// Call client
server, err := p.client.CreateServer(ctx, token, projectID, createReq)
```

**SDK migration pattern:**
```go
// Extract credentials
projectID := string(req.Secret.Data["projectId"])
region := string(req.Secret.Data["region"]) // NEW
token := string(req.Secret.Data["stackitToken"])

// Initialize SDK client (in provider initialization)
iaasClient, err := iaas.NewAPIClient(config.WithToken(token))
p.client = &sdkStackitClient{iaasClient: iaasClient, projectID: projectID, region: region}

// Call client
server, err := p.client.CreateServer(ctx, projectID, region, payload)
```

---

## 7. Migration Impact Assessment

### Code Changes Required

| Component | Files | LOC Current | LOC After | Change |
|-----------|-------|-------------|-----------|--------|
| Interface | `stackit_client.go` | 76 | 80 | +4 (region param) |
| HTTP Client | `http_client.go` | 230 | **0** | -230 (deleted) |
| SDK Client | `sdk_client.go` | 0 | 150 | +150 (new) |
| Core Logic | `core.go` | ~300 | ~320 | +20 (type conversions) |
| Unit Tests | `*_test.go` | ~800 | ~750 | -50 (simpler mocks) |
| **TOTAL** | | **1,406** | **1,300** | **-106 LOC** |

**Net result:** ~7% code reduction + improved type safety

### Risk Assessment

| Risk | Severity | Likelihood | Mitigation | Status |
|------|----------|------------|------------|--------|
| SDK API differs from HTTP | Medium | Low | Verified API coverage in Phase 0 | ✅ Mitigated |
| Breaking E2E tests | Medium | Low | Mock server uses same REST API | ✅ Low risk |
| Region parameter missing | High | Medium | Add to secret validation | ⚠️ Action required |
| Authentication breaks deploys | High | Low | Token flow maintains compatibility | ✅ Mitigated |
| Pointer conversion bugs | Low | Medium | Write tests for helper functions | ⚠️ Action required |
| Performance degradation | Low | Very Low | SDK is thin wrapper | ✅ Low risk |

### New Requirements Discovered

#### Region Parameter

**Issue:** IAAS SDK v1.0.0 requires `region` parameter for all operations.

**Current:** Not present in secrets or ProviderSpec
```yaml
apiVersion: v1
kind: Secret
data:
  projectId: ...
  stackitToken: ...
  # region: MISSING
```

**Required:**
```yaml
apiVersion: v1
kind: Secret
data:
  projectId: ...
  stackitToken: ...
  region: ZXUwMS0x  # base64: "eu01-1"
```

**Migration Actions:**
1. ✅ Add `region` to secret validation
2. ✅ Update sample secrets
3. ✅ Update documentation
4. ✅ Add validation error if missing
5. ⚠️ **BREAKING CHANGE** - Existing deployments must add region to secrets

**Recommendation:** Provide sensible default (e.g., `"eu01-1"`) if region is missing to avoid breaking existing deployments, but log a deprecation warning.

---

## 8. Testing Impact

### Unit Tests

**Current:** 800 LOC of unit tests with manual mocks

**Changes required:**
- Update mock interface signatures (add region, remove token)
- Update mock return types to SDK types
- Delete HTTP client tests (~200 LOC removed)
- Update field assertions for pointer types

**Expected effort:** 2-4 hours

### E2E Tests

**Current:** Mock STACKIT API server in Kubernetes

**Expected impact:** ⚠️ **Minimal to none**

**Reason:** The SDK makes identical HTTP requests to the mock server. The REST API contract is unchanged.

**Validation approach:**
1. Run E2E tests after migration
2. Compare HTTP request logs before/after
3. Verify mock server sees same requests

**Potential issues:**
- Request header differences (e.g., User-Agent)
- JSON field ordering (shouldn't matter but could break strict matchers)

**Recommendation:** Run E2E tests early in Phase 5 to validate assumptions.

---

## 9. Examples and Documentation

### SDK Examples Available

The SDK repository includes comprehensive examples:

**IaaS Examples:** (https://github.com/stackitcloud/stackit-sdk-go/tree/main/examples/iaas)
- `server/` - Server creation, update, deletion
- `attach_nic/` - Network interface management
- `attach_public_ip/` - Public IP association
- `attach_security_group/` - Security group management
- `attach_service_account/` - Service account linking
- `attach_volume/` - Volume attachment
- `network/` - Network creation
- `network_area/` - Network area configuration
- `routing_tables/` - Routing table management
- `volume/` - Volume operations
- `publicip/` - Public IP management

**Authentication Examples:**
- Token flow
- Key flow (STACKIT-generated)
- Key flow (custom RSA)
- Environment variable auto-detection

### Reference Implementation

**STACKIT Cloud Provider:** https://github.com/stackitcloud/cloud-provider-stackit

This is an official STACKIT project that uses the SDK for:
- Cloud Controller Manager (CCM)
- Container Storage Interface (CSI)

**Key learnings:**
- Uses `mockgen` for testing (we'll continue with manual mocks)
- Extensive SDK integration patterns
- Production-grade error handling

---

## 10. Recommendations

### ✅ Proceed with Migration

**Confidence Level:** High

The SDK is mature, well-documented, and fully covers our requirements. The migration will:
- ✅ Reduce code by ~100 LOC
- ✅ Improve type safety
- ✅ Add automatic token refresh capability (with key flow)
- ✅ Provide better error handling
- ✅ Future-proof the provider for new STACKIT features

### 📋 Action Items Before Phase 1

1. **Add Region Support**
   - [ ] Update secret validation to require `region`
   - [ ] Add default region fallback for backward compatibility
   - [ ] Update `samples/secret.yaml` with region field
   - [ ] Document region values (eu01-1, eu01-2, etc.)

2. **Pin SDK Version**
   - [ ] Use IAAS v1.0.0 (latest stable major version)
   - [ ] Pin in `go.mod`: `github.com/stackitcloud/stackit-sdk-go/services/iaas v1.0.0`
   - [ ] Document quarterly SDK update schedule

3. **Create Helper Functions**
   - [ ] Write `ptr[T any](v T) *T` helper
   - [ ] Write label conversion: `map[string]string` → `*map[string]interface{}`
   - [ ] Write userData conversion: `string` → `*[]byte` (base64)
   - [ ] Add unit tests for helpers

4. **Update Documentation**
   - [ ] Add SDK references to CLAUDE.md
   - [ ] Update ROADMAP_sdk_migration.md with Phase 0 findings
   - [ ] Document region parameter requirement

### 🎯 Updated Timeline Estimate

| Phase | Original Estimate | Updated Estimate | Change |
|-------|-------------------|------------------|--------|
| Phase 0 | Not estimated | ✅ Complete | n/a |
| Phase 1 | 1-2 hours | 2-3 hours | +1 hour (region support) |
| Phase 2 | 2-3 hours | 2-3 hours | Unchanged |
| Phase 3 | 3-4 hours | 4-5 hours | +1 hour (pointer conversions) |
| Phase 4 | 2-4 hours | 2-4 hours | Unchanged |
| Phase 5 | 1-2 hours | 1-2 hours | Unchanged |
| Phase 6 | 1-2 hours | 2-3 hours | +1 hour (region docs) |
| **Total** | **10-17 hours** | **14-20 hours** | **+3 hours** |

**New risks discovered:** Region parameter requirement (minor breaking change)
**New benefits discovered:** SDK v1.0.0 stability, comprehensive examples

---

## 11. Conclusion

**APPROVED FOR MIGRATION** ✅

The STACKIT SDK Go is production-ready and well-suited for this migration. The SDK provides all required operations, excellent documentation, and a clean API. The migration will simplify the codebase while improving maintainability and type safety.

**Next Steps:**
1. Complete action items listed in Section 10
2. Update ROADMAP_sdk_migration.md with Phase 0 findings
3. Begin Phase 1: Dependency & Authentication Setup

**Confidence:** High (95%)
**Risk Level:** Low
**Expected Effort:** 14-20 hours
**Expected Value:** High

---

**Evaluated by:** Claude Code
**Approval Status:** ✅ Recommended for implementation
**Date:** 2025-11-05
