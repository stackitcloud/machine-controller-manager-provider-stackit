package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	. "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis/validation"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("ValidateProviderSpecNSecret", func() {
	var (
		providerSpec *api.ProviderSpec
		secret       *corev1.Secret
	)

	BeforeEach(func() {
		// Set up valid defaults
		providerSpec = &api.ProviderSpec{
			MachineType: "c2i.2",
			ImageID:     "550e8400-e29b-41d4-a716-446655440000",
			Region:      "eu01",
		}
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"project-id":          []byte("11111111-2222-3333-4444-555555555555"),
				"serviceaccount.json": []byte(`{"credentials":{"iss":"test"}}`),
			},
		}
	})

	Context("Secret validation", func() {
		It("should fail when secret is nil", func() {
			errors := ValidateProviderSpecNSecret(providerSpec, nil)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("secret"))
		})

		It("should fail when projectId is missing from secret", func() {
			secret.Data = map[string][]byte{}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("project-id"))
		})

		It("should fail when projectId is empty in secret", func() {
			secret.Data["project-id"] = []byte("")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("project-id"))
		})

		It("should fail when projectId is not a valid UUID", func() {
			secret.Data["project-id"] = []byte("invalid-uuid")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("project-id' must be a valid UUID"))
		})

		It("should fail when serviceaccount.json is missing from secret", func() {
			delete(secret.Data, "serviceaccount.json")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("serviceaccount.json"))
		})

		It("should fail when serviceaccount.json is empty in secret", func() {
			secret.Data["serviceaccount.json"] = []byte("")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("serviceaccount.json"))
		})

		It("should fail when serviceaccount.json is not valid JSON", func() {
			secret.Data["serviceaccount.json"] = []byte("not-valid-json")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("must be valid JSON"))
		})

		It("should fail when serviceaccount.json is malformed JSON (missing closing brace)", func() {
			secret.Data["serviceaccount.json"] = []byte(`{"credentials":{"iss":"test"`)
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("must be valid JSON"))
		})

		It("should pass when serviceAccountKey is valid JSON with minimal structure", func() {
			secret.Data["serviceAccountKey"] = []byte(`{"credentials":{"iss":"test@sa.stackit.cloud"}}`)
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should pass when serviceAccountKey is valid JSON with full structure", func() {
			secret.Data["serviceAccountKey"] = []byte(`{
				"credentials": {
					"iss": "test@sa.stackit.cloud",
					"sub": "12345678-1234-1234-1234-123456789012",
					"aud": "stackit"
				},
				"privateKey": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQE\n-----END PRIVATE KEY-----"
			}`)
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should pass when serviceAccountKey is valid JSON array (edge case)", func() {
			// JSON validation should accept any valid JSON, even if not the expected structure
			// The SDK will validate the actual content
			secret.Data["serviceAccountKey"] = []byte(`[]`)
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			// Should not have JSON validation error (though SDK would fail with this content)
			for _, err := range errors {
				Expect(err.Error()).NotTo(ContainSubstring("must be valid JSON"))
			}
		})

		It("should pass when serviceAccountKey is valid JSON string (edge case)", func() {
			// JSON validation should accept any valid JSON, even if not the expected structure
			secret.Data["serviceAccountKey"] = []byte(`"some-string"`)
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			// Should not have JSON validation error (though SDK would fail with this content)
			for _, err := range errors {
				Expect(err.Error()).NotTo(ContainSubstring("must be valid JSON"))
			}
		})
	})
})
