// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP Client", func() {
	var (
		server    *httptest.Server
		client    *httpStackitClient
		ctx       context.Context
		projectID string
		token     string
	)

	BeforeEach(func() {
		ctx = context.Background()
		projectID = "11111111-2222-3333-4444-555555555555"
		token = "test-token-12345"
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("newHTTPStackitClient", func() {
		It("should use environment variable for base URL", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_ENDPOINT")
			defer func() {
				if originalEnv != "" {
					_ = os.Setenv("STACKIT_API_ENDPOINT", originalEnv)
				} else {
					_ = os.Unsetenv("STACKIT_API_ENDPOINT")
				}
			}()

			_ = os.Setenv("STACKIT_API_ENDPOINT", "https://custom.api.example.com")
			client := newHTTPStackitClient()

			Expect(client.baseURL).To(Equal("https://custom.api.example.com"))
		})

		It("should use default URL when environment variable is not set", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_ENDPOINT")
			defer func() {
				if originalEnv != "" {
					_ = os.Setenv("STACKIT_API_ENDPOINT", originalEnv)
				} else {
					_ = os.Unsetenv("STACKIT_API_ENDPOINT")
				}
			}()

			_ = os.Unsetenv("STACKIT_API_ENDPOINT")
			client := newHTTPStackitClient()

			Expect(client.baseURL).To(Equal("https://api.stackit.cloud"))
		})
	})

	Describe("Timeout Configuration", func() {
		It("should have 30-second default timeout", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_TIMEOUT")
			defer func() {
				if originalEnv != "" {
					_ = os.Setenv("STACKIT_API_TIMEOUT", originalEnv)
				} else {
					_ = os.Unsetenv("STACKIT_API_TIMEOUT")
				}
			}()

			_ = os.Unsetenv("STACKIT_API_TIMEOUT")
			client := newHTTPStackitClient()

			Expect(client.httpClient.Timeout).To(Equal(30 * time.Second))
		})

		It("should use custom timeout from environment variable", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_TIMEOUT")
			defer func() {
				if originalEnv != "" {
					_ = os.Setenv("STACKIT_API_TIMEOUT", originalEnv)
				} else {
					_ = os.Unsetenv("STACKIT_API_TIMEOUT")
				}
			}()

			_ = os.Setenv("STACKIT_API_TIMEOUT", "60")
			client := newHTTPStackitClient()

			Expect(client.httpClient.Timeout).To(Equal(60 * time.Second))
		})

		It("should use default timeout when environment variable is invalid", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_TIMEOUT")
			defer func() {
				if originalEnv != "" {
					_ = os.Setenv("STACKIT_API_TIMEOUT", originalEnv)
				} else {
					_ = os.Unsetenv("STACKIT_API_TIMEOUT")
				}
			}()

			_ = os.Setenv("STACKIT_API_TIMEOUT", "invalid")
			client := newHTTPStackitClient()

			Expect(client.httpClient.Timeout).To(Equal(30 * time.Second))
		})

		It("should use default timeout when environment variable is negative", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_TIMEOUT")
			defer func() {
				if originalEnv != "" {
					_ = os.Setenv("STACKIT_API_TIMEOUT", originalEnv)
				} else {
					_ = os.Unsetenv("STACKIT_API_TIMEOUT")
				}
			}()

			_ = os.Setenv("STACKIT_API_TIMEOUT", "-10")
			client := newHTTPStackitClient()

			Expect(client.httpClient.Timeout).To(Equal(30 * time.Second))
		})

		It("should timeout on slow CreateServer request", func() {
			// Create a server that delays response longer than timeout
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond) // Delay longer than client timeout
				w.WriteHeader(http.StatusOK)
			}))

			client = &httpStackitClient{
				baseURL: server.URL,
				httpClient: &http.Client{
					Timeout: 10 * time.Millisecond, // Very short timeout for testing
				},
			}

			req := &CreateServerRequest{
				Name:        "test-machine",
				MachineType: "c1.2",
			}

			_, err := client.CreateServer(ctx, token, projectID, req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("HTTP request failed"))
		})
	})
})
