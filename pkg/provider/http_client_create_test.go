// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"

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

	Describe("CreateServer", func() {
		Context("with successful API response", func() {
			It("should create a server successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("POST"))
					Expect(r.URL.Path).To(Equal("/v1/projects/11111111-2222-3333-4444-555555555555/servers"))
					Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{
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

				result, err := client.CreateServer(ctx, token, projectID, req)

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
					_, _ = w.Write([]byte(`{"id":"123","name":"test","status":"OK"}`))
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

				_, err := client.CreateServer(ctx, token, projectID, req)

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
					_, _ = w.Write([]byte(`{"error": "Invalid machine type"}`))
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

				result, err := client.CreateServer(ctx, token, projectID, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 400"))
				Expect(result).To(BeNil())
			})

			It("should return error on 500 Internal Server Error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
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

				result, err := client.CreateServer(ctx, token, projectID, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 500"))
				Expect(result).To(BeNil())
			})

			It("should return error on invalid JSON response", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{invalid json}`))
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

				result, err := client.CreateServer(ctx, token, projectID, req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse response"))
				Expect(result).To(BeNil())
			})
		})
	})
})
