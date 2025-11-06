# STACKIT SDK Migration Roadmap

---

## Executive Summary

Migrate from custom HTTP client implementation to the official [STACKIT Go SDK](https://github.com/stackitcloud/stackit-sdk-go) for improved maintainability, type safety, and official support from STACKIT.

### Benefits

- ✅ **Official Support:** Use STACKIT's maintained SDK with ongoing updates
- ✅ **Better Authentication:** Built-in token refresh and key-based auth flow
- ✅ **Type Safety:** Compile-time validation of API structures
- ✅ **Error Handling:** Typed errors with better context and debugging
- ✅ **Future-Proof:** Automatic access to new STACKIT features
- ✅ **Reduced Maintenance:** ~200 lines of HTTP client code eliminated
- ✅ **Testing Utilities:** SDK includes testing support and utilities

---

## Objectives

### Primary Goals

1. Replace custom `httpStackitClient` with official STACKIT SDK
2. Maintain backward compatibility with existing deployments
3. Preserve all existing functionality and test coverage
4. Improve error handling and debugging capabilities

### Non-Goals

- ❌ Changing E2E test infrastructure (mock server remains)
- ❌ Switching to mockgen for unit tests (keeping current approach)
- ❌ Breaking changes to API or configuration
- ❌ Adding new features during migration

---

## Migration Phases

### Phase 1: Dependency & Authentication Setup ✅ **COMPLETE**

#### Tasks

- [x] Add STACKIT SDK dependencies to `go.mod`
  - ✅ Added `github.com/stackitcloud/stackit-sdk-go/core@v0.18.0`
  - ✅ Added `github.com/stackitcloud/stackit-sdk-go/services/iaas@v1.0.0`

- [x] Update vendor directory
  - ✅ Ran `just revendor` successfully
  - ✅ Verified SDK packages in `vendor/github.com/stackitcloud/stackit-sdk-go/`

- [x] Research SDK authentication methods
  - ✅ Token flow selected (simple migration, maintains compatibility)
  - ✅ Key flow documented for future enhancement
  - ✅ Reviewed cloud-provider-stackit for SDK usage patterns

- [x] Document authentication migration path for users
  - ✅ Using token flow via `stackitToken` in Secret
  - ✅ Region now required in Secret (breaking change documented)

#### Deliverables

- ✅ Updated `go.mod` and `go.sum`
- ✅ Vendored SDK packages (core + iaas)
- ✅ Authentication strategy: Token flow (Phase 0 evaluation completed)

---

### Phase 2: Client Interface Refactoring ✅ **COMPLETE**

#### Tasks

- [x] **Update `pkg/provider/stackit_client.go`** interface
  - ✅ **Key Decision**: Keep token/projectID as parameters (different MachineClasses use different credentials)
  - ✅ **Added `region` parameter** to all methods (required by SDK v1.0.0+)
  - ✅ Updated interface:
    ```go
    // After migration
    CreateServer(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error)
    GetServer(ctx context.Context, token, projectID, region, serverID string) (*Server, error)
    DeleteServer(ctx context.Context, token, projectID, region, serverID string) error
    ListServers(ctx context.Context, token, projectID, region string) ([]*Server, error)
    ```

- [x] **Create `pkg/provider/sdk_client.go`** - new SDK wrapper
  - ✅ Implemented stateless SDK client (creates API client per-request)
  - ✅ Supports different credentials per MachineClass
  - ✅ Handles authentication via token flow
  - ✅ Proper error wrapping with shared `ErrServerNotFound`

- [x] **Create `pkg/provider/helpers.go`** - SDK conversion helpers
  - ✅ `ptr[T any](v T) *T` - pointer helper for optional SDK fields
  - ✅ `convertLabelsToSDK/FromSDK` - label type conversions
  - ✅ `convertUserDataToSDK` - base64 encoding helper
  - ✅ `convertStringSliceToSDK` - slice pointer wrapper
  - ✅ `convertMetadataToSDK` - metadata pointer wrapper

- [x] **Update `pkg/provider/provider.go`** initialization
  - ✅ Changed from `newHTTPStackitClient()` to `newSDKStackitClient()`
  - ✅ SDK client is stateless (no stored credentials)

- [x] **Update `pkg/provider/http_client.go`** for compatibility
  - ✅ Added `region` parameter to all methods (ignored for HTTP API)
  - ✅ Kept HTTP client for backward compatibility during testing

#### Deliverables

- ✅ New SDK client wrapper implementation (`sdk_client.go`)
- ✅ Helper functions for SDK type conversions (`helpers.go`)
- ✅ Updated provider initialization
- ⏸️ HTTP client kept for now (will remove after E2E testing)

---

### Phase 3: Core Logic Migration ✅ **COMPLETE**

#### Tasks

- [x] **Update `pkg/provider/core.go` - Extract region from Secret**
  - ✅ Added `region := string(req.Secret.Data["region"])` in all driver methods
  - ✅ Region now required in Secret (validated by validation package)

- [x] **Update `pkg/provider/core.go` - CreateMachine**
  - ✅ Pass `region` to `p.client.CreateServer(ctx, token, projectID, region, createReq)`
  - ✅ SDK client handles conversion to `iaas.CreateServerPayload` internally
  - ✅ All field mappings implemented with helper functions

- [x] **Update `pkg/provider/core.go` - GetMachineStatus**
  - ✅ Pass `region` to `p.client.GetServer(ctx, token, projectID, region, serverID)`
  - ✅ SDK client handles response parsing

- [x] **Update `pkg/provider/core.go` - DeleteMachine**
  - ✅ Pass `region` to `p.client.DeleteServer(ctx, token, projectID, region, serverID)`
  - ✅ Error handling uses shared `ErrServerNotFound`

- [x] **Update `pkg/provider/core.go` - ListMachines**
  - ✅ Pass `region` to `p.client.ListServers(ctx, token, projectID, region)`
  - ✅ SDK client handles `ListServersResponse` parsing

#### Field Mapping Examples

```go
// Networking
createReq.Networking = &iaas.ServerNetworking{
    NetworkId: ptr(providerSpec.Networking.NetworkID),
    NicIds:    providerSpec.Networking.NICIDs,
}

// Boot Volume
createReq.BootVolume = &iaas.ServerBootVolume{
    Size:                ptr(int64(providerSpec.BootVolume.Size)),
    PerformanceClass:    ptr(providerSpec.BootVolume.PerformanceClass),
    DeleteOnTermination: providerSpec.BootVolume.DeleteOnTermination,
    Source: &iaas.ServerBootVolumeSource{
        Type: ptr(providerSpec.BootVolume.Source.Type),
        Id:   ptr(providerSpec.BootVolume.Source.ID),
    },
}
```

#### Deliverables

- All core provider methods using SDK
- Field mapping documentation
- Error handling improvements

---

### Phase 4: Unit Test Updates ⏳ **IN PROGRESS**

#### Tasks

- [x] **Update `pkg/provider/core_mocks_test.go`** mock implementation
  - ✅ Updated mock to include `region` parameter in all methods
  - ✅ Mock signatures:
    ```go
    type mockStackitClient struct {
        createServerFunc func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error)
        getServerFunc    func(ctx context.Context, token, projectID, region, serverID string) (*Server, error)
        deleteServerFunc func(ctx context.Context, token, projectID, region, serverID string) error
        listServersFunc  func(ctx context.Context, token, projectID, region string) ([]*Server, error)
    }
    ```

- [ ] **Update test secrets to include `region` field**
  - Files requiring updates:
    - [ ] `pkg/provider/core_create_machine_basic_test.go`
    - [ ] `pkg/provider/core_create_machine_config_test.go`
    - [ ] `pkg/provider/core_create_machine_storage_test.go`
    - [ ] `pkg/provider/core_create_machine_userdata_test.go`
    - [ ] `pkg/provider/core_delete_machine_test.go`
    - [ ] `pkg/provider/core_get_machine_status_test.go`
    - [ ] `pkg/provider/core_list_machines_test.go`
  - Required change in each file:
    ```go
    secret = &corev1.Secret{
        Data: map[string][]byte{
            "projectId":    []byte("11111111-2222-3333-4444-555555555555"),
            "stackitToken": []byte("test-token-123"),
            "region":       []byte("eu01-1"), // ADD THIS
        },
    }
    ```

- [ ] **Keep HTTP client tests for now**
  - ⏸️ Will remove after E2E validation with SDK
  - Files: `http_client_*_test.go` (4 files)

- [ ] Run unit tests and fix failures
  ```sh
  just golang::test
  ```

#### Note on Mocking Strategy

We will **continue using our current manual mocking approach** for simplicity and maintainability. The functional mock style fits our testing needs well with only one interface to mock.

**Future consideration:** If the project grows to require mocking many interfaces, we could adopt `mockgen` (as used by [cloud-provider-stackit](https://github.com/stackitcloud/cloud-provider-stackit/blob/main/Makefile#L130)) for automatic mock generation.

#### Deliverables

- All unit tests passing with SDK mocks
- Cleaned up obsolete test files
- Updated test documentation

---

### Phase 5: E2E Test Validation

#### Tasks

- [ ] **Verify E2E tests work without changes**
  ```sh
  just test-e2e
  ```

- [ ] **Check mock server compatibility**
  - SDK should make identical HTTP calls
  - Verify request/response formats match
  - Check authentication headers

- [ ] **Test all E2E scenarios**
  - Basic machine creation/deletion
  - User data handling
  - Volume attachments
  - SSH keypair configuration
  - Availability zones
  - Affinity groups
  - Service accounts
  - Agent configuration
  - Metadata handling
  - Networking (single and multi-NIC)

- [ ] **Debug any failures**
  - Check mock server logs: `kubectl logs -n stackitcloud -l app=iaas-mock`
  - Compare HTTP requests before/after migration
  - Update mock server if payload formats differ

#### Expected Outcome

E2E tests should **pass without modifications** since the SDK makes the same REST API calls. The mock server sees identical HTTP requests.

#### Deliverables

- All E2E tests passing
- Documentation of any mock server adjustments
- E2E test validation report

---

### Phase 6: Documentation & Release 📝 **PENDING**

#### Tasks

- [ ] **Update samples/secret.yaml**
  - [ ] Add `region` field to example
  - [ ] Add comments explaining region requirement
  - [ ] Document both token and key flow options

- [ ] **Update CLAUDE.md**
  - [ ] Document SDK migration completion
  - [ ] Update architecture section
  - [ ] Add SDK references

- [ ] **Update README.md** (if exists)
  - [ ] Add SDK authentication details
  - [ ] Document region requirement

- [ ] **Create MIGRATION.md** (for existing deployments)
  - [ ] Document breaking change: `region` required in Secret
  - [ ] Provide upgrade checklist
  - [ ] List valid region values (eu01-1, eu01-2, etc.)

- [ ] **Update CHANGELOG.md**
  - [ ] Document migration to SDK
  - [ ] List breaking changes: `region` field required
  - [ ] Note benefits: official SDK support, better maintainability

- [ ] **Add SDK references to docs**
  - [ ] Link to STACKIT SDK documentation
  - [ ] Add troubleshooting section for SDK-specific issues

#### Deliverables

- Updated documentation
- Migration guide for users
- CHANGELOG entry

---

## Testing Strategy

### Unit Tests

- **Approach:** Manual mocks (current approach)
- **Coverage Target:** Maintain 80%+ coverage
- **Validation:** `just test`

### E2E Tests

- **Approach:** Mock STACKIT API server (unchanged)
- **Validation:** `just test-e2e`
- **Focus:** End-to-end workflow validation

### Manual Testing

- **Test with real STACKIT project**
- **Validate all provider operations:**
  - Machine creation with various configurations
  - Machine deletion and cleanup
  - Error handling scenarios
  - Concurrent operations

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| SDK API differs from HTTP implementation | High | Low | Review SDK examples and documentation first |
| Breaking changes in unit tests | Medium | Medium | Update mocks incrementally, one file at a time |
| E2E tests fail with SDK | Medium | Low | Mock server unchanged, SDK uses same REST API |
| Authentication changes break deployments | High | Low | Support both token and key flow, document migration |
| SDK has bugs or limitations | Medium | Low | Have rollback plan, report issues to STACKIT |
| Performance degradation | Low | Very Low | SDK is lightweight wrapper, same HTTP calls |

---

## Success Criteria

- ✅ All unit tests passing with SDK client
- ✅ All E2E tests passing without changes
- ✅ No breaking changes for existing users
- ✅ Documentation updated and complete
- ✅ Manual testing validated with real STACKIT project
- ✅ Code review approved
- ✅ Performance metrics unchanged or improved

---

## References

### STACKIT SDK

- **GitHub Repository:** https://github.com/stackitcloud/stackit-sdk-go
- **IAAS Service Package:** https://github.com/stackitcloud/stackit-sdk-go/tree/main/services/iaas
- **Core Package:** https://github.com/stackitcloud/stackit-sdk-go/tree/main/core
- **Examples:** https://github.com/stackitcloud/stackit-sdk-go/tree/main/examples
- **Authentication Example:** https://github.com/stackitcloud/stackit-sdk-go/blob/main/examples/authentication/authentication.go
- **Releases:** https://github.com/stackitcloud/stackit-sdk-go/releases

### STACKIT Documentation

- **Service Accounts:** https://docs.stackit.cloud/stackit/en/service-accounts-134415819.html
- **Service Account Keys:** https://docs.stackit.cloud/stackit/en/usage-of-the-service-account-keys-in-stackit-175112464.html
- **Creating RSA Key-Pair:** https://docs.stackit.cloud/stackit/en/usage-of-the-service-account-keys-in-stackit-175112464.html#CreatinganRSAkey-pair
- **Assigning Permissions:** https://docs.stackit.cloud/stackit/en/assign-permissions-to-a-service-account-134415855.html
- **Tutorials for Service Accounts:** https://docs.stackit.cloud/stackit/en/tutorials-for-service-accounts-134415861.html
- **STACKIT API Documentation:** https://docs.stackit.cloud/

### Reference Implementations

- **STACKIT Cloud Provider (CSI & CCM):** https://github.com/stackitcloud/cloud-provider-stackit
  - Uses STACKIT SDK with mockgen for testing
  - Good reference for SDK patterns and testing strategies
  - Makefile with mock generation: https://github.com/stackitcloud/cloud-provider-stackit/blob/main/Makefile#L130

### Machine Controller Manager

- **MCM Repository:** https://github.com/gardener/machine-controller-manager
- **Driver Interface:** https://github.com/gardener/machine-controller-manager/blob/master/pkg/util/provider/driver/driver.go
- **Provider Development Guide:** https://github.com/gardener/machine-controller-manager/blob/master/docs/development/cp_support_new.md
- **Sample Provider:** https://github.com/gardener/machine-controller-manager-provider-sampleprovider

### Testing Tools

- **Ginkgo (BDD Framework):** https://onsi.github.io/ginkgo/
- **Gomega (Matcher Library):** https://onsi.github.io/gomega/
- **mockgen (Future Consideration):** https://github.com/uber-go/mock
  - Official Go mocking framework
  - Used by cloud-provider-stackit
  - Could be adopted if project scales

### Go Best Practices

- **Effective Go:** https://go.dev/doc/effective_go
- **Go Modules:** https://go.dev/ref/mod
- **Go Testing:** https://go.dev/doc/tutorial/add-a-test

---

## Key Decisions Made

1. **Interface Design:**
   - ✅ Keep `token`, `projectID`, `region` as explicit parameters (not stored in client)
   - ✅ Stateless SDK client wrapper (creates API client per-request)
   - ✅ Rationale: Different MachineClasses can use different credentials/projects

2. **Region Handling:**
   - ✅ Add `region` as explicit parameter to interface
   - ✅ Extract from Secret in core.go
   - ✅ **Breaking Change**: Region required in Secret (STACKIT SDK v1.0.0+ requirement)
   - ✅ Rationale: Clean API, no encoding/parsing hacks

3. **Authentication Strategy:**
   - ✅ Token flow (maintains compatibility)
   - ⏸️ Key flow documented for future enhancement
   - ✅ Decision: Start with token flow, users can migrate to key flow later

4. **Backward Compatibility:**
   - ✅ Keep HTTP client temporarily (testing safety net)
   - ✅ Updated HTTP client to accept `region` parameter (ignored)
   - ⏸️ Will remove after E2E validation

5. **SDK Version Pinning:**
   - ✅ Core: v0.18.0 (stable)
   - ✅ IAAS: v1.0.0 (first stable major version)
   - ✅ Decision: Pin to stable versions, update quarterly

## Breaking Changes

### For Users (Production Deployments)

**⚠️ BREAKING CHANGE: `region` field required in Secret**

Before SDK migration:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: stackit-credentials
data:
  projectId: <base64-encoded-project-id>
  stackitToken: <base64-encoded-token>
```

After SDK migration:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: stackit-credentials
data:
  projectId: <base64-encoded-project-id>
  stackitToken: <base64-encoded-token>
  region: ZXUwMS0x  # base64("eu01-1") - REQUIRED
```

**Valid region values**: `eu01-1`, `eu01-2`, etc.

### For Developers (Testing)

- All test secrets must include `region` field
- Mock client interface updated with `region` parameter
- 7 test files require secret updates
