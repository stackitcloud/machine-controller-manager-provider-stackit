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
			Networking: &api.NetworkingSpec{
				NetworkID: "770e8400-e29b-41d4-a716-446655440000",
			},
		}
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"project-id":          []byte("11111111-2222-3333-4444-555555555555"),
				"serviceaccount.json": []byte(`{"credentials":{"iss":"test"}}`),
			},
		}
	})

	Context("KeypairName validation", func() {
		It("should succeed with valid keypairName", func() {
			providerSpec.KeypairName = "my-ssh-key"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when keypairName is empty", func() {
			providerSpec.KeypairName = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with keypairName containing allowed characters", func() {
			providerSpec.KeypairName = "my-key_2024@test.com"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when keypairName contains invalid characters", func() {
			providerSpec.KeypairName = "my key!"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("invalid characters"))
		})

		It("should fail when keypairName exceeds max length", func() {
			providerSpec.KeypairName = string(make([]byte, 128)) // 128 > 127 max
			for i := range providerSpec.KeypairName {
				providerSpec.KeypairName = providerSpec.KeypairName[:i] + "a" + providerSpec.KeypairName[i+1:]
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("maximum length"))
		})
	})

	Context("Region validation", func() {
		It("should succeed with valid region", func() {
			providerSpec.Region = "eu01"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when region is empty", func() {
			providerSpec.Region = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors[0].Error()).To(ContainSubstring("Region cannot be empty"))
		})

		It("should succeed with various region formats", func() {
			testCases := []string{
				"eu01",
				"eu02",
				"eu10",
				"us02",
			}
			for _, r := range testCases {
				providerSpec.Region = r
				errors := ValidateProviderSpecNSecret(providerSpec, secret)
				Expect(errors).To(BeEmpty(), "Region %q should be valid", r)
			}
		})

		It("should fail with various region formats", func() {
			testCases := []string{
				"eu01-1",
				"eu02-b",
				"us-east-01",
			}
			for _, r := range testCases {
				providerSpec.Region = r
				errors := ValidateProviderSpecNSecret(providerSpec, secret)
				Expect(errors[0].Error()).To(ContainSubstring("Region has invalid format"))
			}
		})
	})

	Context("AvailabilityZone validation", func() {
		It("should succeed with valid availabilityZone", func() {
			providerSpec.AvailabilityZone = "eu01-1"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when availabilityZone is empty", func() {
			providerSpec.AvailabilityZone = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with various AZ formats", func() {
			testCases := []string{
				"eu01-1",
				"eu01-2",
				"eu02-1",
				"eu02-4",
			}
			for _, az := range testCases {
				providerSpec.AvailabilityZone = az
				errors := ValidateProviderSpecNSecret(providerSpec, secret)
				Expect(errors).To(BeEmpty(), "AvailabilityZone %q should be valid", az)
			}
		})
	})

	Context("AffinityGroup validation", func() {
		It("should succeed with valid affinityGroup UUID", func() {
			providerSpec.AffinityGroup = "880e8400-e29b-41d4-a716-446655440000"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when affinityGroup is empty", func() {
			providerSpec.AffinityGroup = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when affinityGroup has invalid UUID format", func() {
			providerSpec.AffinityGroup = "invalid-uuid-format"
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("valid UUID"))
		})
	})

	Context("ServiceAccountMails validation", func() {
		It("should succeed with valid service account email", func() {
			providerSpec.ServiceAccountMails = []string{
				"my-service@sa.stackit.cloud",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when serviceAccountMails is empty", func() {
			providerSpec.ServiceAccountMails = []string{}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when serviceAccountMails is nil", func() {
			providerSpec.ServiceAccountMails = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when serviceAccountMails contains invalid email format", func() {
			providerSpec.ServiceAccountMails = []string{
				"invalid-email-format",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("valid email"))
		})

		It("should fail when serviceAccountMails contains empty string", func() {
			providerSpec.ServiceAccountMails = []string{
				"",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("empty"))
		})

		It("should fail when serviceAccountMails has more than 1 item", func() {
			providerSpec.ServiceAccountMails = []string{
				"first@sa.stackit.cloud",
				"second@sa.stackit.cloud",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("maximum of 1"))
		})
	})

	Context("Agent validation", func() {
		It("should succeed when agent is nil", func() {
			providerSpec.Agent = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with agent provisioned true", func() {
			provisioned := true
			providerSpec.Agent = &api.AgentSpec{
				Provisioned: &provisioned,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with agent provisioned false", func() {
			provisioned := false
			providerSpec.Agent = &api.AgentSpec{
				Provisioned: &provisioned,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with agent provisioned nil", func() {
			providerSpec.Agent = &api.AgentSpec{
				Provisioned: nil,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})
	})

	Context("Metadata validation", func() {
		It("should succeed when metadata is nil", func() {
			providerSpec.Metadata = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with empty metadata", func() {
			providerSpec.Metadata = map[string]interface{}{}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with valid metadata", func() {
			providerSpec.Metadata = map[string]interface{}{
				"environment": "production",
				"cost-center": "12345",
				"owner":       "team-a",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with nested metadata objects", func() {
			providerSpec.Metadata = map[string]interface{}{
				"tags": map[string]interface{}{
					"env":  "prod",
					"tier": "backend",
				},
				"count": 42,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})
	})
})
