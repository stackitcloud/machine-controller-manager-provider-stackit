// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"encoding/base64"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SDK Client Helpers", func() {

	Describe("extractRegion", func() {
		Context("with valid region", func() {
			It("should extract region from secret data", func() {
				secretData := map[string][]byte{
					"region": []byte("eu01-1"),
				}

				region, err := extractRegion(secretData)

				Expect(err).NotTo(HaveOccurred())
				Expect(region).To(Equal("eu01-1"))
			})

			It("should extract region with different value", func() {
				secretData := map[string][]byte{
					"region": []byte("eu01-2"),
				}

				region, err := extractRegion(secretData)

				Expect(err).NotTo(HaveOccurred())
				Expect(region).To(Equal("eu01-2"))
			})

			It("should extract region when other fields are present", func() {
				secretData := map[string][]byte{
					"projectId":    []byte("11111111-2222-3333-4444-555555555555"),
					"stackitToken": []byte("test-token-123"),
					"region":       []byte("eu01-1"),
					"userData":     []byte("some-user-data"),
				}

				region, err := extractRegion(secretData)

				Expect(err).NotTo(HaveOccurred())
				Expect(region).To(Equal("eu01-1"))
			})
		})

		Context("with missing or invalid region", func() {
			It("should fail when region field is missing", func() {
				secretData := map[string][]byte{
					"projectId":    []byte("11111111-2222-3333-4444-555555555555"),
					"stackitToken": []byte("test-token-123"),
				}

				_, err := extractRegion(secretData)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'region' field is required"))
				Expect(err.Error()).To(ContainSubstring("eu01-1"))
			})

			It("should fail when region field is empty", func() {
				secretData := map[string][]byte{
					"region": []byte(""),
				}

				_, err := extractRegion(secretData)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'region' field is required"))
			})

			It("should fail when region field is nil bytes", func() {
				secretData := map[string][]byte{
					"region": nil,
				}

				_, err := extractRegion(secretData)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'region' field is required"))
			})

			It("should fail when secret data is empty", func() {
				secretData := map[string][]byte{}

				_, err := extractRegion(secretData)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'region' field is required"))
			})

			It("should fail when secret data is nil", func() {
				var secretData map[string][]byte

				_, err := extractRegion(secretData)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("'region' field is required"))
			})
		})
	})

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
		Context("with 404 errors", func() {
			It("should detect error with '404' in message", func() {
				err := errors.New("API returned 404 status code")

				result := isNotFoundError(err)

				Expect(result).To(BeTrue())
			})

			It("should detect error with 'not found' in message", func() {
				err := errors.New("server not found")

				result := isNotFoundError(err)

				Expect(result).To(BeTrue())
			})

			It("should detect error with 'NotFound' in message", func() {
				err := errors.New("ResourceNotFound: server does not exist")

				result := isNotFoundError(err)

				Expect(result).To(BeTrue())
			})

			It("should detect error with '404' at start", func() {
				err := errors.New("404: server missing")

				result := isNotFoundError(err)

				Expect(result).To(BeTrue())
			})

			It("should detect error with '404' at end", func() {
				err := errors.New("HTTP error 404")

				result := isNotFoundError(err)

				Expect(result).To(BeTrue())
			})

			It("should detect error with '404' in middle", func() {
				err := errors.New("API call failed with 404 status")

				result := isNotFoundError(err)

				Expect(result).To(BeTrue())
			})
		})

		Context("with non-404 errors", func() {
			It("should not detect 500 error", func() {
				err := errors.New("API returned 500 status code")

				result := isNotFoundError(err)

				Expect(result).To(BeFalse())
			})

			It("should not detect 403 error", func() {
				err := errors.New("403 Forbidden")

				result := isNotFoundError(err)

				Expect(result).To(BeFalse())
			})

			It("should not detect generic error", func() {
				err := errors.New("something went wrong")

				result := isNotFoundError(err)

				Expect(result).To(BeFalse())
			})

			It("should return false for nil error", func() {
				result := isNotFoundError(nil)

				Expect(result).To(BeFalse())
			})
		})
	})

	Describe("contains", func() {
		Context("with substring present", func() {
			It("should find substring at start", func() {
				result := contains("hello world", "hello")

				Expect(result).To(BeTrue())
			})

			It("should find substring at end", func() {
				result := contains("hello world", "world")

				Expect(result).To(BeTrue())
			})

			It("should find substring in middle", func() {
				result := contains("hello world", "lo wo")

				Expect(result).To(BeTrue())
			})

			It("should find exact match", func() {
				result := contains("404", "404")

				Expect(result).To(BeTrue())
			})

			It("should find single character", func() {
				result := contains("test", "e")

				Expect(result).To(BeTrue())
			})
		})

		Context("with substring not present", func() {
			It("should not find missing substring", func() {
				result := contains("hello world", "foo")

				Expect(result).To(BeFalse())
			})

			It("should not find when substring is longer", func() {
				result := contains("hi", "hello")

				Expect(result).To(BeFalse())
			})

			It("should handle empty string", func() {
				result := contains("", "test")

				Expect(result).To(BeFalse())
			})
		})
	})

	Describe("findSubstring", func() {
		Context("with substring present", func() {
			It("should find substring in string", func() {
				result := findSubstring("hello world", "world")

				Expect(result).To(BeTrue())
			})

			It("should find substring at start", func() {
				result := findSubstring("testing", "test")

				Expect(result).To(BeTrue())
			})

			It("should find substring in middle", func() {
				result := findSubstring("the quick brown fox", "quick")

				Expect(result).To(BeTrue())
			})
		})

		Context("with substring not present", func() {
			It("should not find missing substring", func() {
				result := findSubstring("hello", "world")

				Expect(result).To(BeFalse())
			})

			It("should handle empty string", func() {
				result := findSubstring("", "test")

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

	Describe("convertUserDataToSDK", func() {
		Context("with plain text userData", func() {
			It("should base64-encode plain text userData", func() {
				userData := "#cloud-config\nruncmd:\n  - echo 'test'"

				result := convertUserDataToSDK(userData)

				Expect(result).NotTo(BeNil())
				// Should be base64-encoded
				Expect(*result).NotTo(Equal(userData))
				Expect(*result).To(MatchRegexp(`^[A-Za-z0-9+/]+=*$`)) // base64 pattern
			})
		})

		Context("with base64-encoded userData", func() {
			It("should keep already base64-encoded userData", func() {
				// Generate base64-encoded userData
				originalData := "#cloud-config\nruncmd:\n  - echo 'test'"
				base64Data := base64.StdEncoding.EncodeToString([]byte(originalData))

				result := convertUserDataToSDK(base64Data)

				Expect(result).NotTo(BeNil())
				Expect(*result).To(Equal(base64Data))

				// Verify it decodes back to original
				decoded, err := base64.StdEncoding.DecodeString(*result)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(decoded)).To(Equal(originalData))
			})
		})

		Context("with empty userData", func() {
			It("should return nil for empty userData", func() {
				result := convertUserDataToSDK("")

				Expect(result).To(BeNil())
			})
		})
	})
})
