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
				"projectId":    []byte("11111111-2222-3333-4444-555555555555"),
				"stackitToken": []byte("test-token"),
			},
		}
	})

	Context("Volume validation", func() {
		It("should succeed with valid BootVolume configuration", func() {
			deleteOnTermination := true
			providerSpec.BootVolume = &api.BootVolumeSpec{
				DeleteOnTermination: &deleteOnTermination,
				PerformanceClass:    "premium",
				Size:                100,
				Source: &api.BootVolumeSourceSpec{
					Type: "image",
					ID:   "550e8400-e29b-41d4-a716-446655440000",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed with minimal BootVolume configuration", func() {
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Size: 50,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when BootVolume is nil", func() {
			providerSpec.BootVolume = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when BootVolume source type is invalid", func() {
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Source: &api.BootVolumeSourceSpec{
					Type: "invalid-type",
					ID:   "550e8400-e29b-41d4-a716-446655440000",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("source type must be"))
		})

		It("should fail when BootVolume source is missing type", func() {
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Source: &api.BootVolumeSourceSpec{
					ID: "550e8400-e29b-41d4-a716-446655440000",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("type"))
		})

		It("should fail when BootVolume source is missing ID", func() {
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Source: &api.BootVolumeSourceSpec{
					Type: "image",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("id"))
		})

		It("should fail when BootVolume source ID has invalid UUID format", func() {
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Source: &api.BootVolumeSourceSpec{
					Type: "image",
					ID:   "invalid-uuid",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("valid UUID"))
		})

		It("should fail when BootVolume size is negative", func() {
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Size: -10,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("size must be positive"))
		})

		It("should succeed with valid Volumes array", func() {
			providerSpec.Volumes = []string{
				"550e8400-e29b-41d4-a716-446655440000",
				"660e8400-e29b-41d4-a716-446655440001",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when Volumes is nil", func() {
			providerSpec.Volumes = nil
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should succeed when Volumes is empty array", func() {
			providerSpec.Volumes = []string{}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when Volumes contains invalid UUID", func() {
			providerSpec.Volumes = []string{
				"550e8400-e29b-41d4-a716-446655440000",
				"invalid-uuid",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("valid UUID"))
		})

		It("should fail when Volumes contains empty string", func() {
			providerSpec.Volumes = []string{
				"550e8400-e29b-41d4-a716-446655440000",
				"",
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("cannot be empty"))
		})

		It("should allow empty ImageID when BootVolume.Source is specified", func() {
			providerSpec.ImageID = ""
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Source: &api.BootVolumeSourceSpec{
					Type: "snapshot",
					ID:   "550e8400-e29b-41d4-a716-446655440000",
				},
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).To(BeEmpty())
		})

		It("should fail when both ImageID and BootVolume.Source are empty", func() {
			providerSpec.ImageID = ""
			providerSpec.BootVolume = &api.BootVolumeSpec{
				Size: 50,
			}
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("imageId or bootVolume.source"))
		})
	})
})
