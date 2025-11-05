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


	Describe("DeleteServer", func() {
		Context("with successful API response", func() {
			It("should delete server successfully", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					Expect(r.Method).To(Equal("DELETE"))
					Expect(r.URL.Path).To(Equal("/v1/projects/11111111-2222-3333-4444-555555555555/servers/test-server-456"))

					// Real STACKIT API returns 204 No Content on successful DELETE
					w.WriteHeader(http.StatusNoContent)
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, token, projectID, serverID)

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

				err := client.DeleteServer(ctx, token, projectID, serverID)

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with API errors", func() {
			It("should return error on 404 Not Found (server already deleted)", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(`{"error": "Server not found"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, token, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("server not found")))
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

				err := client.DeleteServer(ctx, token, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 403"))
			})

			It("should return error on 500 Internal Server Error", func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error": "Internal error"}`))
				}))

				client = &httpStackitClient{
					baseURL:    server.URL,
					httpClient: &http.Client{},
				}

				err := client.DeleteServer(ctx, token, projectID, serverID)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("API returned error status 500"))
			})
		})
	})
})
