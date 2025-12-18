// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("CreateMachine - Networking", func() {
	var (
		ctx          context.Context
		provider     *Provider
		mockClient   *mockStackitClient
		req          *driver.CreateMachineRequest
		secret       *corev1.Secret
		machineClass *v1alpha1.MachineClass
		machine      *v1alpha1.Machine
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockClient = &mockStackitClient{}
		provider = &Provider{
			client: mockClient,
		}

		// Create secret with basic required fields
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"project-id":          []byte("11111111-2222-3333-4444-555555555555"),
				"serviceaccount.json": []byte(`{"credentials":{"iss":"test"}}`),
			},
		}

		// Create Machine
		machine = &v1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-machine",
				Namespace: "default",
			},
		}
	})

	Context("with networking configuration in ProviderSpec", func() {
		It("should use networkId from ProviderSpec", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				Region:      "eu01",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Networking: &api.NetworkingSpec{
					NetworkID: "770e8400-e29b-41d4-a716-446655440000",
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)

			machineClass = &v1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-class",
				},
				ProviderSpec: runtime.RawExtension{
					Raw: providerSpecRaw,
				},
			}

			req = &driver.CreateMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       secret,
			}

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(_ context.Context, _, _ string, req *CreateServerRequest) (*Server, error) {
				capturedReq = req
				return &Server{
					ID:     "test-server-id",
					Name:   req.Name,
					Status: "CREATING",
				}, nil
			}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedReq.Networking).NotTo(BeNil())
			Expect(capturedReq.Networking.NetworkID).To(Equal("770e8400-e29b-41d4-a716-446655440000"))
			Expect(capturedReq.Networking.NICIDs).To(BeEmpty())
		})

		It("should use nicIds from ProviderSpec", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				Region:      "eu01",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Networking: &api.NetworkingSpec{
					NICIDs: []string{
						"880e8400-e29b-41d4-a716-446655440001",
						"990e8400-e29b-41d4-a716-446655440002",
					},
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)

			machineClass = &v1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-class",
				},
				ProviderSpec: runtime.RawExtension{
					Raw: providerSpecRaw,
				},
			}

			req = &driver.CreateMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       secret,
			}

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(_ context.Context, _, _ string, req *CreateServerRequest) (*Server, error) {
				capturedReq = req
				return &Server{
					ID:     "test-server-id",
					Name:   req.Name,
					Status: "CREATING",
				}, nil
			}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedReq.Networking).NotTo(BeNil())
			Expect(capturedReq.Networking.NetworkID).To(BeEmpty())
			Expect(capturedReq.Networking.NICIDs).To(HaveLen(2))
			Expect(capturedReq.Networking.NICIDs[0]).To(Equal("880e8400-e29b-41d4-a716-446655440001"))
			Expect(capturedReq.Networking.NICIDs[1]).To(Equal("990e8400-e29b-41d4-a716-446655440002"))
		})
	})

	Context("with networking fallback to Secret", func() {
		It("should use networkId from Secret when ProviderSpec.Networking is nil", func() {
			// Add networkId to Secret
			secret.Data["networkId"] = []byte("660e8400-e29b-41d4-a716-446655440000")

			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Region:      "eu01",
				// Networking is nil - should fall back to Secret
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)

			machineClass = &v1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-class",
				},
				ProviderSpec: runtime.RawExtension{
					Raw: providerSpecRaw,
				},
			}

			req = &driver.CreateMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       secret,
			}

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(_ context.Context, _, _ string, req *CreateServerRequest) (*Server, error) {
				capturedReq = req
				return &Server{
					ID:     "test-server-id",
					Name:   req.Name,
					Status: "CREATING",
				}, nil
			}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedReq.Networking).NotTo(BeNil())
			Expect(capturedReq.Networking.NetworkID).To(Equal("660e8400-e29b-41d4-a716-446655440000"))
			Expect(capturedReq.Networking.NICIDs).To(BeEmpty())
		})

		It("should create empty networking when neither ProviderSpec nor Secret has networkId", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Region:      "eu01",
				// Networking is nil
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)

			machineClass = &v1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-class",
				},
				ProviderSpec: runtime.RawExtension{
					Raw: providerSpecRaw,
				},
			}

			req = &driver.CreateMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       secret, // No networkId in secret
			}

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(_ context.Context, _, _ string, req *CreateServerRequest) (*Server, error) {
				capturedReq = req
				return &Server{
					ID:     "test-server-id",
					Name:   req.Name,
					Status: "CREATING",
				}, nil
			}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedReq.Networking).NotTo(BeNil())
			Expect(capturedReq.Networking.NetworkID).To(BeEmpty())
			Expect(capturedReq.Networking.NICIDs).To(BeEmpty())
		})
	})

	Context("with priority of networking configuration", func() {
		It("should prioritize ProviderSpec.Networking over Secret.networkId", func() {
			// Add networkId to Secret (should be ignored)
			secret.Data["networkId"] = []byte("880e8400-e29b-41d4-a716-446655440001")

			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Region:      "eu01",
				Networking: &api.NetworkingSpec{
					NetworkID: "990e8400-e29b-41d4-a716-446655440002",
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)

			machineClass = &v1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-class",
				},
				ProviderSpec: runtime.RawExtension{
					Raw: providerSpecRaw,
				},
			}

			req = &driver.CreateMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       secret,
			}

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(_ context.Context, _, _ string, req *CreateServerRequest) (*Server, error) {
				capturedReq = req
				return &Server{
					ID:     "test-server-id",
					Name:   req.Name,
					Status: "CREATING",
				}, nil
			}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedReq.Networking).NotTo(BeNil())
			Expect(capturedReq.Networking.NetworkID).To(Equal("990e8400-e29b-41d4-a716-446655440002"))
			Expect(capturedReq.Networking.NetworkID).NotTo(Equal("880e8400-e29b-41d4-a716-446655440001"))
		})

		It("should reject empty networking in ProviderSpec (validation)", func() {
			// Add networkId to Secret
			secret.Data["networkId"] = []byte("880e8400-e29b-41d4-a716-446655440001")

			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Region:      "eu01",
				Networking:  &api.NetworkingSpec{
					// Both NetworkID and NICIDs are empty - should fail validation
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)

			machineClass = &v1alpha1.MachineClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-machine-class",
				},
				ProviderSpec: runtime.RawExtension{
					Raw: providerSpecRaw,
				},
			}

			req = &driver.CreateMachineRequest{
				Machine:      machine,
				MachineClass: machineClass,
				Secret:       secret,
			}

			_, err := provider.CreateMachine(ctx, req)

			// Should fail validation because networking is specified but empty
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("networking must specify either networkId or nicIds"))
		})
	})
})
