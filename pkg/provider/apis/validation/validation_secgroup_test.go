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
		}
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"project-id":          []byte("11111111-2222-3333-4444-555555555555"),
				"serviceaccount.json": []byte(`{"credentials":{"iss":"test"}}`),
				"region":              []byte("eu01-1"),
			},
		}
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
})
