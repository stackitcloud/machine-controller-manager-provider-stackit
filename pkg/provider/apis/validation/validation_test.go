// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	api "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis"
	. "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis/validation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
			MachineType: "c1.2",
			ImageID:     "550e8400-e29b-41d4-a716-446655440000",
		}
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"projectId": []byte("test-project"),
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

		It("should fail when ImageID is empty", func() {
			providerSpec.ImageID = ""
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("imageId"))
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
	})

	Context("Networking validation", func() {
		It("should succeed with valid NetworkID", func() {
			providerSpec.Networking = &api.NetworkingSpec{
				NetworkID: "550e8400-e29b-41d4-a716-446655440000",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with valid NICIDs", func() {
			providerSpec.Networking = &api.NetworkingSpec{
				NICIDs: []string{
					"550e8400-e29b-41d4-a716-446655440000",
					"660e8400-e29b-41d4-a716-446655440001",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when Networking is nil", func() {
			providerSpec.Networking = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when Networking has neither NetworkID nor NICIDs", func() {
			providerSpec.Networking = &api.NetworkingSpec{}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("must specify either networkId or nicIds"))
		})

		It("should fail when Networking has both NetworkID and NICIDs", func() {
			providerSpec.Networking = &api.NetworkingSpec{
				NetworkID: "550e8400-e29b-41d4-a716-446655440000",
				NICIDs: []string{
					"660e8400-e29b-41d4-a716-446655440001",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("mutually exclusive"))
		})

		It("should fail when NetworkID has invalid UUID format", func() {
			providerSpec.Networking = &api.NetworkingSpec{
				NetworkID: "invalid-uuid",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("valid UUID"))
		})

		It("should fail when NICIDs contains invalid UUID format", func() {
			providerSpec.Networking = &api.NetworkingSpec{
				NICIDs: []string{
					"550e8400-e29b-41d4-a716-446655440000",
					"invalid-uuid",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("valid UUID"))
		})

		It("should fail when NICIDs contains empty string", func() {
			providerSpec.Networking = &api.NetworkingSpec{
				NICIDs: []string{
					"550e8400-e29b-41d4-a716-446655440000",
					"",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("cannot be empty"))
		})
	})

	Context("SecurityGroups validation", func() {
		It("should succeed with valid SecurityGroups", func() {
			providerSpec.SecurityGroups = []string{"default", "web-servers"}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when SecurityGroups is nil", func() {
			providerSpec.SecurityGroups = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when SecurityGroups is empty array", func() {
			providerSpec.SecurityGroups = []string{}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when SecurityGroups contains empty string", func() {
			providerSpec.SecurityGroups = []string{"default", ""}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("cannot be empty"))
		})
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
			Expect(errors[0].Error()).To(ContainSubstring("projectId"))
		})

		It("should fail when projectId is empty in secret", func() {
			secret.Data["projectId"] = []byte("")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("projectId"))
		})
	})
})
