// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/stackitcloud/stackit-sdk-go/core/config"
	"github.com/stackitcloud/stackit-sdk-go/services/iaas"
)

// Note: ErrServerNotFound is defined in http_client.go and shared by both clients

// sdkStackitClient is an SDK implementation of StackitClient
// It is stateless and creates SDK API clients per-request to support
// different credentials per MachineClass
type sdkStackitClient struct{}

// newSDKStackitClient creates a new stateless SDK STACKIT client wrapper
func newSDKStackitClient() *sdkStackitClient {
	return &sdkStackitClient{}
}

// createIAASClient creates a new STACKIT SDK IAAS API client for a request
// This allows different MachineClasses to use different credentials
func (c *sdkStackitClient) createIAASClient(token string) (*iaas.APIClient, error) {
	// Configure SDK with custom base URL if provided (for testing with mock server)
	baseURL := os.Getenv("STACKIT_API_ENDPOINT")

	var opts []config.ConfigurationOption
	opts = append(opts, config.WithToken(token))

	if baseURL != "" {
		opts = append(opts, config.WithEndpoint(baseURL))
	}

	iaasClient, err := iaas.NewAPIClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create STACKIT SDK API client: %w", err)
	}

	return iaasClient, nil
}

// extractRegion extracts the region from the secret data
// Region is required by STACKIT SDK v1.0.0+
func extractRegion(secretData map[string][]byte) (string, error) {
	region, ok := secretData["region"]
	if !ok || len(region) == 0 {
		// Provide a helpful error message
		return "", fmt.Errorf("'region' field is required in Secret (e.g., 'eu01-1')")
	}
	return string(region), nil
}

// CreateServer creates a new server via STACKIT SDK
func (c *sdkStackitClient) CreateServer(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(token)
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

	// Networking
	if req.Networking != nil {
		payload.Networking = &iaas.CreateServerPayloadAllOfNetworking{}
		if req.Networking.NetworkID != "" {
			payload.Networking.NetworkId = ptr(req.Networking.NetworkID)
		}
		if len(req.Networking.NICIDs) > 0 {
			payload.Networking.NicIds = convertStringSliceToSDK(req.Networking.NICIDs)
		}
	}

	// Security Groups
	if len(req.SecurityGroups) > 0 {
		payload.SecurityGroups = convertStringSliceToSDK(req.SecurityGroups)
	}

	// UserData
	if req.UserData != "" {
		payload.UserData = convertUserDataToSDK(req.UserData)
	}

	// Boot Volume
	if req.BootVolume != nil {
		payload.BootVolume = &iaas.ServerBootVolume{
			Size:                ptr(int64(req.BootVolume.Size)),
			PerformanceClass:    ptr(req.BootVolume.PerformanceClass),
			DeleteOnTermination: req.BootVolume.DeleteOnTermination,
		}
		if req.BootVolume.Source != nil {
			payload.BootVolume.Source = &iaas.ServerBootVolumeSource{
				Type: ptr(req.BootVolume.Source.Type),
				Id:   ptr(req.BootVolume.Source.ID),
			}
		}
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
func (c *sdkStackitClient) GetServer(ctx context.Context, token, projectID, region, serverID string) (*Server, error) {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(token)
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
func (c *sdkStackitClient) DeleteServer(ctx context.Context, token, projectID, region, serverID string) error {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(token)
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
func (c *sdkStackitClient) ListServers(ctx context.Context, token, projectID, region string) ([]*Server, error) {
	// Create SDK client for this request
	iaasClient, err := c.createIAASClient(token)
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
				Status:getStringValue(sdkServer.Status),
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
	// The SDK returns structured errors that contain HTTP status codes
	// Check error message for common 404 indicators
	errStr := err.Error()
	return contains(errStr, "404") || 
		   contains(errStr, "not found") || 
		   contains(errStr, "NotFound")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || len(s) > len(substr) && 
		   (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		   findSubstring(s, substr)))
}

// findSubstring searches for a substring in a string
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
