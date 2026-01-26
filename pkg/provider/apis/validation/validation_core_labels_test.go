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

	Context("ProviderSpec validation", func() {
		It("should succeed with valid ProviderSpec", func() {
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when MachineType is empty", func() {
			providerSpec.MachineType = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("machineType"))
		})

		It("should fail when MachineType has invalid format", func() {
			providerSpec.MachineType = "InvalidFormat"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("machineType has invalid format"))
		})

		It("should succeed when MachineType has valid format", func() {
			providerSpec.MachineType = "c2i.2"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when ImageID is empty", func() {
			providerSpec.ImageID = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("imageId"))
		})

		It("should fail when ImageID is not a valid UUID", func() {
			providerSpec.ImageID = "invalid-uuid"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("imageId must be a valid UUID"))
		})

		It("should fail when both required fields are empty", func() {
			providerSpec.MachineType = ""
			providerSpec.ImageID = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(HaveLen(2))
		})

		It("should succeed when Labels is nil", func() {
			providerSpec.Labels = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when Labels is empty map", func() {
			providerSpec.Labels = map[string]string{}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when Labels has valid key-value pairs", func() {
			providerSpec.Labels = map[string]string{
				"environment": "production",
				"team":        "platform",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with label keys containing allowed characters", func() {
			providerSpec.Labels = map[string]string{
				"app.kubernetes.io_component": "worker",
				"environment-type":            "prod",
				"version":                     "v1.2.3",
				"app/component":               "core",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with label values containing allowed characters", func() {
			providerSpec.Labels = map[string]string{
				"env":     "production-env_01.test",
				"version": "v1.2.3-alpha",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with empty label value", func() {
			providerSpec.Labels = map[string]string{
				"optional-label": "",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when label key exceeds 63 characters", func() {
			longKey := string(make([]byte, 64))
			for i := range longKey {
				longKey = longKey[:i] + "a" + longKey[i+1:]
			}
			providerSpec.Labels = map[string]string{
				longKey: "value",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("exceeds maximum length of 63 characters"))
		})

		It("should fail when label value exceeds 63 characters", func() {
			longValue := string(make([]byte, 64))
			for i := range longValue {
				longValue = longValue[:i] + "a" + longValue[i+1:]
			}
			providerSpec.Labels = map[string]string{
				"key": longValue,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("exceeds maximum length of 63 characters"))
		})

		It("should fail when label key starts with non-alphanumeric", func() {
			providerSpec.Labels = map[string]string{
				"-invalid-key": "value",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("invalid format"))
		})

		It("should fail when label key ends with non-alphanumeric", func() {
			providerSpec.Labels = map[string]string{
				"invalid-key-": "value",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("invalid format"))
		})

		It("should succeed with label keys containing slashes", func() {
			providerSpec.Labels = map[string]string{
				"mycompany.com/environment": "prod",
				"app.io/version":            "v2",
				"team/project/name":         "web",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when label key contains invalid characters", func() {
			providerSpec.Labels = map[string]string{
				"invalid@key": "value",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("invalid format"))
		})

		It("should fail when label value starts with non-alphanumeric", func() {
			providerSpec.Labels = map[string]string{
				"key": "-invalid-value",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("invalid format"))
		})

		It("should fail when label value ends with non-alphanumeric", func() {
			providerSpec.Labels = map[string]string{
				"key": "invalid-value-",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("invalid format"))
		})

		It("should fail when label value contains invalid characters", func() {
			providerSpec.Labels = map[string]string{
				"key": "invalid@value",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("invalid format"))
		})

		It("should fail with multiple label validation errors", func() {
			longKey := string(make([]byte, 64))
			for i := range longKey {
				longKey = longKey[:i] + "a" + longKey[i+1:]
			}
			providerSpec.Labels = map[string]string{
				longKey:        "value1",
				"-invalid-key": "value2",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(HaveLen(2))
		})
	})
})
