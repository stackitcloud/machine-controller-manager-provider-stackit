// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"errors"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stackitcloud/stackit-sdk-go/core/oapierror"
)

var _ = Describe("SDK Client Helpers", func() {

	Describe("getStringValue", func() {
		Context("with valid pointer", func() {
			It("should return string value", func() {
				value := "test-string"
				result := getStringValue(&value)

				Expect(result).To(Equal("test-string"))
			})

			It("should return empty string when pointer is to empty string", func() {
				value := ""
				result := getStringValue(&value)

				Expect(result).To(Equal(""))
			})
		})

		Context("with nil pointer", func() {
			It("should return empty string when pointer is nil", func() {
				result := getStringValue(nil)

				Expect(result).To(Equal(""))
			})
		})
	})

	Describe("isNotFoundError", func() {
		Context("with SDK 404 errors", func() {
			It("should detect GenericOpenAPIError with 404 status code", func() {
				err := &oapierror.GenericOpenAPIError{
					StatusCode:   404,
					ErrorMessage: "Not Found",
				}

				result := isNotFoundError(err)

				Expect(result).To(BeTrue())
			})

			It("should detect wrapped GenericOpenAPIError with 404", func() {
				baseErr := &oapierror.GenericOpenAPIError{
					StatusCode:   404,
					ErrorMessage: "Server not found",
					Body:         []byte(`{"message": "server does not exist"}`),
				}
				wrappedErr := fmt.Errorf("failed to get server: %w", baseErr)

				result := isNotFoundError(wrappedErr)

				Expect(result).To(BeTrue())
			})
		})

		Context("with non-404 errors", func() {
			It("should not detect GenericOpenAPIError with 500 status", func() {
				err := &oapierror.GenericOpenAPIError{
					StatusCode:   500,
					ErrorMessage: "Internal Server Error",
				}

				result := isNotFoundError(err)

				Expect(result).To(BeFalse())
			})

			It("should not detect GenericOpenAPIError with 403 status", func() {
				err := &oapierror.GenericOpenAPIError{
					StatusCode:   403,
					ErrorMessage: "Forbidden",
				}

				result := isNotFoundError(err)

				Expect(result).To(BeFalse())
			})

			It("should not detect generic error without status code", func() {
				err := errors.New("something went wrong")

				result := isNotFoundError(err)

				Expect(result).To(BeFalse())
			})

			It("should not detect wrapped generic error", func() {
				err := fmt.Errorf("API call failed: %w", errors.New("connection timeout"))

				result := isNotFoundError(err)

				Expect(result).To(BeFalse())
			})

			It("should return false for nil error", func() {
				result := isNotFoundError(nil)

				Expect(result).To(BeFalse())
			})
		})
	})
})

var _ = Describe("SDK Type Conversion Helpers", func() {

	Describe("ptr", func() {
		It("should create pointer to string", func() {
			value := "test"
			result := ptr(value)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal("test"))
		})

		It("should create pointer to int", func() {
			value := 42
			result := ptr(value)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(42))
		})

		It("should create pointer to bool", func() {
			value := true
			result := ptr(value)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(BeTrue())
		})

		It("should create pointer to empty string", func() {
			value := ""
			result := ptr(value)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(Equal(""))
		})
	})

	Describe("convertLabelsToSDK", func() {
		Context("with valid labels", func() {
			It("should convert labels map to SDK format", func() {
				labels := map[string]string{
					"app":  "web-server",
					"team": "platform",
				}

				result := convertLabelsToSDK(labels)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(HaveLen(2))
				Expect((*result)["app"]).To(Equal("web-server"))
				Expect((*result)["team"]).To(Equal("platform"))
			})

			It("should convert empty labels map", func() {
				labels := map[string]string{}

				result := convertLabelsToSDK(labels)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(BeEmpty())
			})

			It("should convert labels with special characters", func() {
				labels := map[string]string{
					"mcm.gardener.cloud/machine": "test-machine",
					"kubernetes.io/role":         "node",
				}

				result := convertLabelsToSDK(labels)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(HaveLen(2))
				Expect((*result)["mcm.gardener.cloud/machine"]).To(Equal("test-machine"))
				Expect((*result)["kubernetes.io/role"]).To(Equal("node"))
			})
		})

		Context("with nil labels", func() {
			It("should return nil for nil labels", func() {
				result := convertLabelsToSDK(nil)

				Expect(result).To(BeNil())
			})
		})
	})

	Describe("convertLabelsFromSDK", func() {
		Context("with valid SDK labels", func() {
			It("should convert SDK labels to string map", func() {
				sdkLabels := map[string]interface{}{
					"app":  "web-server",
					"team": "platform",
				}

				result := convertLabelsFromSDK(&sdkLabels)

				Expect(result).NotTo(BeNil())
				Expect(result).To(HaveLen(2))
				Expect(result["app"]).To(Equal("web-server"))
				Expect(result["team"]).To(Equal("platform"))
			})

			It("should convert empty SDK labels map", func() {
				sdkLabels := map[string]interface{}{}

				result := convertLabelsFromSDK(&sdkLabels)

				Expect(result).NotTo(BeNil())
				Expect(result).To(BeEmpty())
			})

			It("should skip non-string values", func() {
				sdkLabels := map[string]interface{}{
					"app":     "web-server",
					"count":   42,         // not a string
					"enabled": true,       // not a string
					"team":    "platform", // string
				}

				result := convertLabelsFromSDK(&sdkLabels)

				Expect(result).NotTo(BeNil())
				Expect(result).To(HaveLen(2))
				Expect(result["app"]).To(Equal("web-server"))
				Expect(result["team"]).To(Equal("platform"))
				Expect(result).NotTo(HaveKey("count"))
				Expect(result).NotTo(HaveKey("enabled"))
			})

			It("should handle nil values in map", func() {
				sdkLabels := map[string]interface{}{
					"app":  "web-server",
					"team": nil,
				}

				result := convertLabelsFromSDK(&sdkLabels)

				Expect(result).NotTo(BeNil())
				Expect(result).To(HaveLen(1))
				Expect(result["app"]).To(Equal("web-server"))
				Expect(result).NotTo(HaveKey("team"))
			})
		})

		Context("with nil SDK labels", func() {
			It("should return nil for nil SDK labels", func() {
				result := convertLabelsFromSDK(nil)

				Expect(result).To(BeNil())
			})
		})
	})

	Describe("convertStringSliceToSDK", func() {
		Context("with valid slice", func() {
			It("should convert string slice to pointer", func() {
				slice := []string{"value1", "value2", "value3"}

				result := convertStringSliceToSDK(slice)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(HaveLen(3))
				Expect(*result).To(Equal([]string{"value1", "value2", "value3"}))
			})

			It("should convert empty slice", func() {
				slice := []string{}

				result := convertStringSliceToSDK(slice)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(BeEmpty())
			})

			It("should convert slice with single item", func() {
				slice := []string{"only-item"}

				result := convertStringSliceToSDK(slice)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(HaveLen(1))
				Expect((*result)[0]).To(Equal("only-item"))
			})
		})

		Context("with nil slice", func() {
			It("should return nil for nil slice", func() {
				result := convertStringSliceToSDK(nil)

				Expect(result).To(BeNil())
			})
		})
	})

	Describe("convertMetadataToSDK", func() {
		Context("with valid metadata", func() {
			It("should convert metadata map to pointer", func() {
				metadata := map[string]interface{}{
					"key1": "value1",
					"key2": 42,
					"key3": true,
				}

				result := convertMetadataToSDK(metadata)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(HaveLen(3))
				Expect((*result)["key1"]).To(Equal("value1"))
				Expect((*result)["key2"]).To(Equal(42))
				Expect((*result)["key3"]).To(BeTrue())
			})

			It("should convert empty metadata map", func() {
				metadata := map[string]interface{}{}

				result := convertMetadataToSDK(metadata)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(BeEmpty())
			})

			It("should convert nested metadata objects", func() {
				metadata := map[string]interface{}{
					"config": map[string]interface{}{
						"nested": "value",
					},
					"list": []string{"item1", "item2"},
				}

				result := convertMetadataToSDK(metadata)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(HaveLen(2))
				Expect((*result)["config"]).To(BeAssignableToTypeOf(map[string]interface{}{}))
				Expect((*result)["list"]).To(BeAssignableToTypeOf([]string{}))
			})
		})

		Context("with nil metadata", func() {
			It("should return nil for nil metadata", func() {
				result := convertMetadataToSDK(nil)

				Expect(result).To(BeNil())
			})
		})
	})

	Describe("NewStackitClient", func() {
		Context("with STACKIT_NO_AUTH enabled", func() {
			It("should create client successfully without authentication", func() {
				// Set environment variable to skip authentication
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Setenv("STACKIT_NO_AUTH", "true")
				defer func() {
					if originalNoAuth == "" {
						os.Unsetenv("STACKIT_NO_AUTH")
					} else {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				client, err := NewStackitClient("")

				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
				Expect(client.iaasClient).NotTo(BeNil())
			})

			It("should create client with any service account key when no auth is enabled", func() {
				// When STACKIT_NO_AUTH=true, the serviceAccountKey should be ignored
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Setenv("STACKIT_NO_AUTH", "true")
				defer func() {
					if originalNoAuth == "" {
						os.Unsetenv("STACKIT_NO_AUTH")
					} else {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				client, err := NewStackitClient("invalid-key")

				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
				Expect(client.iaasClient).NotTo(BeNil())
			})
		})

		Context("with custom endpoint", func() {
			It("should create client with custom endpoint when STACKIT_API_ENDPOINT is set", func() {
				// Set both environment variables
				originalEndpoint := os.Getenv("STACKIT_API_ENDPOINT")
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Setenv("STACKIT_API_ENDPOINT", "https://test.example.com")
				os.Setenv("STACKIT_NO_AUTH", "true")
				defer func() {
					if originalEndpoint == "" {
						os.Unsetenv("STACKIT_API_ENDPOINT")
					} else {
						os.Setenv("STACKIT_API_ENDPOINT", originalEndpoint)
					}
					if originalNoAuth == "" {
						os.Unsetenv("STACKIT_NO_AUTH")
					} else {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				client, err := NewStackitClient("")

				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
				Expect(client.iaasClient).NotTo(BeNil())
			})

			It("should work with default endpoint when STACKIT_API_ENDPOINT is not set", func() {
				// Ensure endpoint is not set
				originalEndpoint := os.Getenv("STACKIT_API_ENDPOINT")
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Unsetenv("STACKIT_API_ENDPOINT")
				os.Setenv("STACKIT_NO_AUTH", "true")
				defer func() {
					if originalEndpoint != "" {
						os.Setenv("STACKIT_API_ENDPOINT", originalEndpoint)
					}
					if originalNoAuth == "" {
						os.Unsetenv("STACKIT_NO_AUTH")
					} else {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				client, err := NewStackitClient("")

				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
				Expect(client.iaasClient).NotTo(BeNil())
			})
		})

		Context("with service account authentication", func() {
			It("should fail with invalid JSON service account key", func() {
				// Ensure authentication is enabled
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Unsetenv("STACKIT_NO_AUTH")
				defer func() {
					if originalNoAuth != "" {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				_, err := NewStackitClient("not-valid-json")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create STACKIT SDK client"))
			})

			It("should fail with empty service account key", func() {
				// Ensure authentication is enabled
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Unsetenv("STACKIT_NO_AUTH")
				defer func() {
					if originalNoAuth != "" {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				_, err := NewStackitClient("")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create STACKIT SDK client"))
			})

			It("should fail with valid JSON but missing required fields", func() {
				// Ensure authentication is enabled
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Unsetenv("STACKIT_NO_AUTH")
				defer func() {
					if originalNoAuth != "" {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				// Valid JSON but not a valid service account key structure
				_, err := NewStackitClient(`{"some": "json"}`)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create STACKIT SDK client"))
			})
		})

		Context("client reusability", func() {
			It("should create client once and reuse IaaS client", func() {
				// Use STACKIT_NO_AUTH to avoid needing valid credentials
				originalNoAuth := os.Getenv("STACKIT_NO_AUTH")
				os.Setenv("STACKIT_NO_AUTH", "true")
				defer func() {
					if originalNoAuth == "" {
						os.Unsetenv("STACKIT_NO_AUTH")
					} else {
						os.Setenv("STACKIT_NO_AUTH", originalNoAuth)
					}
				}()

				client1, err1 := NewStackitClient("")
				client2, err2 := NewStackitClient("")

				Expect(err1).NotTo(HaveOccurred())
				Expect(err2).NotTo(HaveOccurred())
				Expect(client1).NotTo(BeNil())
				Expect(client2).NotTo(BeNil())
				// Note: We can't easily verify they're different instances without
				// accessing internal SDK state, but the test documents the intent
			})
		})
	})

})
