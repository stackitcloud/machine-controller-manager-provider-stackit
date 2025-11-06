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
			MachineType: "c2i.2",
			ImageID:     "550e8400-e29b-41d4-a716-446655440000",
		}
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"projectId":    []byte("11111111-2222-3333-4444-555555555555"),
				"stackitToken": []byte("test-token"),
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
			Expect(errors[0].Error()).To(ContainSubstring("projectId"))
		})

		It("should fail when projectId is empty in secret", func() {
			secret.Data["projectId"] = []byte("")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("projectId"))
		})

		It("should fail when projectId is not a valid UUID", func() {
			secret.Data["projectId"] = []byte("invalid-uuid")
			errors := ValidateProviderSpecNSecret(providerSpec, secret)
			Expect(errors).NotTo(BeEmpty())
			Expect(errors[0].Error()).To(ContainSubstring("projectId' must be a valid UUID"))
		})
	})
})
