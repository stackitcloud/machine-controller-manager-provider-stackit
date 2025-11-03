// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// httpStackitClient is an HTTP implementation of StackitClient
type httpStackitClient struct {
	baseURL    string
	httpClient *http.Client
}

// newHTTPStackitClient creates a new HTTP STACKIT client
func newHTTPStackitClient() *httpStackitClient {
	baseURL := os.Getenv("STACKIT_API_ENDPOINT")
	if baseURL == "" {
		baseURL = "https://api.stackit.cloud" // Default to production
	}

	return &httpStackitClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// CreateServer creates a new server via HTTP API
func (c *httpStackitClient) CreateServer(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
	// Build API path
	url := fmt.Sprintf("%s/v1/projects/%s/servers", c.baseURL, projectID)

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var server Server
	if err := json.Unmarshal(respBody, &server); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &server, nil
}

// GetServer retrieves a server by ID via HTTP API
func (c *httpStackitClient) GetServer(ctx context.Context, projectID, serverID string) (*Server, error) {
	// Build API path
	url := fmt.Sprintf("%s/v1/projects/%s/servers/%s", c.baseURL, projectID, serverID)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("server not found: 404")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var server Server
	if err := json.Unmarshal(respBody, &server); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &server, nil
}

// DeleteServer deletes a server by ID via HTTP API
func (c *httpStackitClient) DeleteServer(ctx context.Context, projectID, serverID string) error {
	// Build API path
	url := fmt.Sprintf("%s/v1/projects/%s/servers/%s", c.baseURL, projectID, serverID)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body (for error messages)
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	// Success: 204 No Content (according to STACKIT API spec)
	if resp.StatusCode == 204 {
		return nil
	}

	// Not Found: Server already deleted - this is OK (idempotent)
	if resp.StatusCode == 404 {
		return fmt.Errorf("server not found: 404")
	}

	// All other status codes are errors
	return fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(respBody))
}

// ListServers lists all servers in a project via HTTP API
func (c *httpStackitClient) ListServers(ctx context.Context, projectID string) ([]*Server, error) {
	// Build API path
	url := fmt.Sprintf("%s/v1/projects/%s/servers", c.baseURL, projectID)

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response - the API returns an object with "items" array
	var listResponse struct {
		Items []*Server `json:"items"`
	}
	if err := json.Unmarshal(respBody, &listResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return listResponse.Items, nil
}
