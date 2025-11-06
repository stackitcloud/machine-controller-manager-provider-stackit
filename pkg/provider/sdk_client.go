// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/stackitcloud/stackit-sdk-go/core/config"
	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
)

// sdkStackitClient is an SDK implementation of StackitClient
// It is stateless and creates SDK API clients per-request to support
// different credentials per MachineClass
type sdkStackitClient struct{}

// newSDKStackitClient creates a new stateless SDK STACKIT client wrapper
func newSDKStackitClient() *sdkStackitClient {
	return &sdkStackitClient{}
}

var (
	// ErrServerNotFound indicates the server was not found (404)
	ErrServerNotFound = errors.New("server not found")
)

// createIAASClient creates a new STACKIT SDK IAAS API client for a request
// This allows different MachineClasses to use different credentials
//
// Authentication: Uses ServiceAccount Key Flow (recommended by STACKIT)
// - Automatically generates JWT tokens from service account credentials
// - Handles token refresh before expiration
// - More secure than static tokens (short-lived, rotating)
func (c *sdkStackitClient) createIAASClient(serviceAccountKey string) (*iaas.APIClient, error) {
	// Configure SDK with custom base URL if provided (for testing with mock server)
	baseURL := os.Getenv("STACKIT_API_ENDPOINT")

	var opts []config.ConfigurationOption

	// Use ServiceAccount Key Flow (production-recommended authentication)
	// The SDK will:
	// 1. Parse the service account key JSON
	// 2. Use the private key to sign JWT tokens
	// 3. Automatically fetch access tokens from STACKIT token API
	// 4. Refresh tokens before expiration (with 5s leeway)
	opts = append(opts, config.WithServiceAccountKey(serviceAccountKey))

	if baseURL != "" {
		opts = append(opts, config.WithEndpoint(baseURL))
	}

	iaasClient, err := iaas.NewAPIClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create STACKIT SDK API client: %w", err)
	}

	return iaasClient, nil
}

// validRegionPattern matches STACKIT region formats like "eu01-1", "eu01-2"
// Pattern: <2 letters><2 digits>-<zone number>
var validRegionPattern = regexp.MustCompile(`^[a-z]{2}\d{2}-\d+$`)

// extractRegion extracts and validates the region from the secret data
// Region is required by STACKIT SDK v1.0.0+
func extractRegion(secretData map[string][]byte) (string, error) {
	region, ok := secretData["region"]
	if !ok || len(region) == 0 {
		return "", fmt.Errorf("'region' field is required in Secret (e.g., 'eu01-1')")
	}

	regionStr := string(region)

	// Validate region format
	if !validRegionPattern.MatchString(regionStr) {
		return "", fmt.Errorf("invalid region format '%s': must match pattern '<location><number>-<zone>' (e.g., 'eu01-1', 'eu01-2')", regionStr)
	}

	return regionStr, nil
}

// CreateServer creates a new server via STACKIT SDK
func (c *sdkStackitClient) CreateServer(ctx context.Context, serviceAccountKey, projectID, region string, req *CreateServerRequest) (*Server, error) {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(serviceAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDK client: %w", err)
	}

	// Convert our request to SDK payload
	payload := &iaas.CreateServerPayload{
		Name:        ptr(req.Name),
		MachineType: ptr(req.MachineType),
	}

	// ImageID (optional - can be nil if booting from snapshot/volume)
	if req.ImageID != "" {
		payload.ImageId = ptr(req.ImageID)
	}

	// Labels
	if req.Labels != nil {
		payload.Labels = convertLabelsToSDK(req.Labels)
	}

	// Networking - Required in v2 API, SDK uses union type: either NetworkId OR NicIds
	//
	// Precedence (mutually exclusive):
	//   1. If NetworkID is set (non-empty), use it (ignores NICIDs if both are set)
	//   2. If NICIDs is set (non-empty slice), use it
	//   3. Otherwise, create empty networking object (v2 API requirement)
	//
	// This matches the STACKIT API behavior where you can either:
	//   - Auto-create NICs in a network (NetworkID)
	//   - Use pre-created NICs (NICIDs)
	if req.Networking != nil {
		if req.Networking.NetworkID != "" {
			// Use CreateServerNetworking (with networkId)
			// This will auto-create a NIC in the specified network
			networking := iaas.NewCreateServerNetworking()
			networking.SetNetworkId(req.Networking.NetworkID)
			payload.SetNetworking(iaas.CreateServerNetworkingAsCreateServerPayloadAllOfNetworking(networking))
		} else if len(req.Networking.NICIDs) > 0 {
			// Use CreateServerNetworkingWithNics (with nicIds)
			// This attaches pre-existing NICs to the server
			networking := iaas.NewCreateServerNetworkingWithNics()
			networking.SetNicIds(req.Networking.NICIDs)
			payload.SetNetworking(iaas.CreateServerNetworkingWithNicsAsCreateServerPayloadAllOfNetworking(networking))
		} else {
			// Empty networking object (v2 API requires networking field even if empty)
			networking := iaas.NewCreateServerNetworking()
			payload.SetNetworking(iaas.CreateServerNetworkingAsCreateServerPayloadAllOfNetworking(networking))
		}
	}

	// Security Groups
	if len(req.SecurityGroups) > 0 {
		payload.SecurityGroups = convertStringSliceToSDK(req.SecurityGroups)
	}

	// UserData - SDK expects *[]byte (base64-encoded bytes)
	if req.UserData != "" {
		userDataBytes := []byte(req.UserData)
		payload.SetUserData(userDataBytes)
	}

	// Boot Volume
	if req.BootVolume != nil {
		bootVolume := iaas.NewServerBootVolume()
		if req.BootVolume.Size > 0 {
			bootVolume.SetSize(int64(req.BootVolume.Size))
		}
		if req.BootVolume.PerformanceClass != "" {
			bootVolume.SetPerformanceClass(req.BootVolume.PerformanceClass)
		}
		if req.BootVolume.DeleteOnTermination != nil {
			bootVolume.SetDeleteOnTermination(*req.BootVolume.DeleteOnTermination)
		}
		if req.BootVolume.Source != nil {
			source := iaas.NewBootVolumeSource(req.BootVolume.Source.ID, req.BootVolume.Source.Type)
			bootVolume.SetSource(*source)
		}
		payload.SetBootVolume(*bootVolume)
	}

	// Volumes
	if len(req.Volumes) > 0 {
		payload.Volumes = convertStringSliceToSDK(req.Volumes)
	}

	// KeypairName
	if req.KeypairName != "" {
		payload.KeypairName = ptr(req.KeypairName)
	}

	// AvailabilityZone
	if req.AvailabilityZone != "" {
		payload.AvailabilityZone = ptr(req.AvailabilityZone)
	}

	// AffinityGroup
	if req.AffinityGroup != "" {
		payload.AffinityGroup = ptr(req.AffinityGroup)
	}

	// ServiceAccountMails
	if len(req.ServiceAccountMails) > 0 {
		payload.ServiceAccountMails = convertStringSliceToSDK(req.ServiceAccountMails)
	}

	// Agent
	if req.Agent != nil {
		payload.Agent = &iaas.ServerAgent{
			Provisioned: req.Agent.Provisioned,
		}
	}

	// Metadata
	if req.Metadata != nil {
		payload.Metadata = convertMetadataToSDK(req.Metadata)
	}

	// Call SDK
	sdkServer, err := iaasClient.CreateServer(ctx, projectID, region).
		CreateServerPayload(*payload).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("SDK CreateServer failed: %w", err)
	}

	// Convert SDK server to our Server type
	server := &Server{
		ID:     getStringValue(sdkServer.Id),
		Name:   getStringValue(sdkServer.Name),
		Status: getStringValue(sdkServer.Status),
		Labels: convertLabelsFromSDK(sdkServer.Labels),
	}

	return server, nil
}

// GetServer retrieves a server by ID via STACKIT SDK
func (c *sdkStackitClient) GetServer(ctx context.Context, serviceAccountKey, projectID, region, serverID string) (*Server, error) {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(serviceAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDK client: %w", err)
	}

	sdkServer, err := iaasClient.GetServer(ctx, projectID, region, serverID).Execute()
	if err != nil {
		// Check if error is 404 Not Found
		if isNotFoundError(err) {
			return nil, fmt.Errorf("%w: %v", ErrServerNotFound, err)
		}
		return nil, fmt.Errorf("SDK GetServer failed: %w", err)
	}

	// Convert SDK server to our Server type
	server := &Server{
		ID:     getStringValue(sdkServer.Id),
		Name:   getStringValue(sdkServer.Name),
		Status: getStringValue(sdkServer.Status),
		Labels: convertLabelsFromSDK(sdkServer.Labels),
	}

	return server, nil
}

// DeleteServer deletes a server by ID via STACKIT SDK
func (c *sdkStackitClient) DeleteServer(ctx context.Context, serviceAccountKey, projectID, region, serverID string) error {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(serviceAccountKey)
	if err != nil {
		return fmt.Errorf("failed to create SDK client: %w", err)
	}

	err = iaasClient.DeleteServer(ctx, projectID, region, serverID).Execute()
	if err != nil {
		// Check if error is 404 Not Found - this is OK (idempotent)
		if isNotFoundError(err) {
			return fmt.Errorf("%w: %v", ErrServerNotFound, err)
		}
		return fmt.Errorf("SDK DeleteServer failed: %w", err)
	}

	return nil
}

// ListServers lists all servers in a project via STACKIT SDK
func (c *sdkStackitClient) ListServers(ctx context.Context, serviceAccountKey, projectID, region string) ([]*Server, error) {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(serviceAccountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SDK client: %w", err)
	}

	sdkResponse, err := iaasClient.ListServers(ctx, projectID, region).Execute()
	if err != nil {
		return nil, fmt.Errorf("SDK ListServers failed: %w", err)
	}

	// Convert SDK servers to our Server type
	servers := make([]*Server, 0)
	if sdkResponse.Items != nil {
		for _, sdkServer := range *sdkResponse.Items {
			server := &Server{
				ID:     getStringValue(sdkServer.Id),
				Name:   getStringValue(sdkServer.Name),
				Status: getStringValue(sdkServer.Status),
				Labels: convertLabelsFromSDK(sdkServer.Labels),
			}
			servers = append(servers, server)
		}
	}

	return servers, nil
}

// Helper functions

// getStringValue safely dereferences a string pointer, returning empty string if nil
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// isNotFoundError checks if an error is a 404 Not Found error from the SDK
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Use the SDK's structured error type to check the HTTP status code
	var oapiErr *oapierror.GenericOpenAPIError
	if errors.As(err, &oapiErr) {
		return oapiErr.StatusCode == 404
	}
	return false
}
