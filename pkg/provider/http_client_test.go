// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTP Client", func() {
	var (
		server     *httptest.Server
		client     *httpStackitClient
		ctx        context.Context
		projectID  string
		serverID   string
	)

	BeforeEach(func() {
		ctx = context.Background()
		projectID = "test-project-123"
		serverID = "test-server-456"
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	Describe("CreateServer", func() {
		Context("with successful API response", func() {
			It("should create a server successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))
					Expect(r.URL.Path).To(Equal("/v1/projects/test-project-123/servers"))
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"id": "550e8400-e29b-41d4-a716-446655440000",
						"name": "test-machine",
						"status": "CREATING"
					}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				req := &CreateServerRequest{
					Name:        "test-machine",
					MachineType: "c1.2",
					ImageID:     "image-123",
				}

				result, err := client.CreateServer(ctx, projectID, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.ID).To(Equal("550e8400-e29b-41d4-a716-446655440000"))
				Expect(result.Name).To(Equal("test-machine"))
				Expect(result.Status).To(Equal("CREATING"))
			})

			It("should send correct request body", func() {
				var receivedBody []byte
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					receivedBody, _ = io.ReadAll(r.Body)
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"id":"123","name":"test","status":"OK"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				req := &CreateServerRequest{
					Name:        "my-server",
					MachineType: "c1.4",
					ImageID:     "img-789",
				}

				_, err := client.CreateServer(ctx, projectID, req)

				Expect(err).NotTo(HaveOccurred())
				Expect(string(receivedBody)).To(ContainSubstring(`"name":"my-server"`))
				Expect(string(receivedBody)).To(ContainSubstring(`"machineType":"c1.4"`))
				Expect(string(receivedBody)).To(ContainSubstring(`"imageId":"img-789"`))
			})
		})

		Context("with API errors", func() {
			It("should return error on 400 Bad Request", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error": "Invalid machine type"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				req := &CreateServerRequest{
					Name:        "test-machine",
					MachineType: "invalid",
					ImageID:     "image-123",
				}

				result, err := client.CreateServer(ctx, projectID, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 400"))
				Expect(result).To(BeNil())
			})

			It("should return error on 500 Internal Server Error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Internal server error"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				req := &CreateServerRequest{
					Name:        "test-machine",
					MachineType: "c1.2",
					ImageID:     "image-123",
				}

				result, err := client.CreateServer(ctx, projectID, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 500"))
				Expect(result).To(BeNil())
			})

			It("should return error on invalid JSON response", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{invalid json}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				req := &CreateServerRequest{
					Name:        "test-machine",
					MachineType: "c1.2",
					ImageID:     "image-123",
				}

				result, err := client.CreateServer(ctx, projectID, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse response"))
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("GetServer", func() {
		Context("with successful API response", func() {
			It("should get server status successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("GET"))
					Expect(r.URL.Path).To(Equal("/v1/projects/test-project-123/servers/test-server-456"))
					Expect(r.Header.Get("Accept")).To(Equal("application/json"))

					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{
						"id": "test-server-456",
						"name": "test-machine",
						"status": "RUNNING"
					}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, projectID, serverID)

				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.ID).To(Equal("test-server-456"))
				Expect(result.Name).To(Equal("test-machine"))
				Expect(result.Status).To(Equal("RUNNING"))
			})
		})

		Context("with API errors", func() {
			It("should return error on 404 Not Found", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"error": "Server not found"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("server not found: 404"))
				Expect(result).To(BeNil())
			})

			It("should return error on 403 Forbidden", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"error": "Access denied"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 403"))
				Expect(result).To(BeNil())
			})

			It("should return error on invalid JSON response", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`not valid json`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse response"))
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("DeleteServer", func() {
		Context("with successful API response", func() {
			It("should delete server successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("DELETE"))
					Expect(r.URL.Path).To(Equal("/v1/projects/test-project-123/servers/test-server-456"))

					w.WriteHeader(http.StatusOK)
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, projectID, serverID)

				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept 204 No Content response", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, projectID, serverID)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with API errors", func() {
			It("should return error on 404 Not Found (server already deleted)", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"error": "Server not found"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("server not found: 404"))
			})

			It("should return error on 403 Forbidden", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte(`{"error": "Access denied"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 403"))
			})

			It("should return error on 500 Internal Server Error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error": "Internal error"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 500"))
			})
		})
	})

	Describe("newHTTPStackitClient", func() {
		It("should use environment variable for base URL", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_ENDPOINT")
			defer func() {
				if originalEnv != "" {
					os.Setenv("STACKIT_API_ENDPOINT", originalEnv)
				} else {
					os.Unsetenv("STACKIT_API_ENDPOINT")
				}
			}()

			os.Setenv("STACKIT_API_ENDPOINT", "https://custom.api.example.com")
			client := newHTTPStackitClient()

			Expect(client.baseURL).To(Equal("https://custom.api.example.com"))
		})

		It("should use default URL when environment variable is not set", func() {
			// Save original env
			originalEnv := os.Getenv("STACKIT_API_ENDPOINT")
			defer func() {
				if originalEnv != "" {
					os.Setenv("STACKIT_API_ENDPOINT", originalEnv)
				} else {
					os.Unsetenv("STACKIT_API_ENDPOINT")
				}
			}()

			os.Unsetenv("STACKIT_API_ENDPOINT")
			client := newHTTPStackitClient()

			Expect(client.baseURL).To(Equal("https://api.stackit.cloud"))
		})
	})
})
