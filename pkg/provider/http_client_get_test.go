// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
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
		serverID  string
		token     string
	)

	BeforeEach(func() {
		ctx = context.Background()
		projectID = "11111111-2222-3333-4444-555555555555"
		serverID = "test-server-456"
		token = "test-token-12345"
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})


	Describe("GetServer", func() {
		Context("with successful API response", func() {
			It("should get server status successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("GET"))
					Expect(r.URL.Path).To(Equal("/v1/projects/11111111-2222-3333-4444-555555555555/servers/test-server-456"))
					Expect(r.Header.Get("Accept")).To(Equal("application/json"))

					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`{
						"id": "test-server-456",
						"name": "test-machine",
						"status": "RUNNING"
					}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, token, projectID, serverID)

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
					_, _ = w.Write([]byte(`{"error": "Server not found"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, token, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("server not found")))
				Expect(result).To(BeNil())
			})

			It("should return error on 403 Forbidden", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"error": "Access denied"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, token, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 403"))
				Expect(result).To(BeNil())
			})

			It("should return error on invalid JSON response", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(`not valid json`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				result, err := client.GetServer(ctx, token, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to parse response"))
				Expect(result).To(BeNil())
			})
		})
	})
})
