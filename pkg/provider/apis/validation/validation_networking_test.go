// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

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
})
