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

var _ = Describe("CreateMachine", func() {
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

		// Create secret with projectId and networkId (required for v2 API)
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"projectId":         []byte("11111111-2222-3333-4444-555555555555"),
				"serviceAccountKey": []byte(`{"credentials":{"iss":"test"}}`),
				"region":            []byte("eu01-1"),
				"networkId":         []byte("770e8400-e29b-41d4-a716-446655440000"),
			},
		}

		// Create ProviderSpec
		providerSpec := &api.ProviderSpec{
			MachineType: "c2i.2",
			ImageID:     "12345678-1234-1234-1234-123456789abc",
		}
		providerSpecRaw, _ := encodeProviderSpec(providerSpec)

		// Create MachineClass
		machineClass = &v1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine-class",
			},
			ProviderSpec: runtime.RawExtension{
				Raw: providerSpecRaw,
			},
		}

		// Create Machine
		machine = &v1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-machine",
				Namespace: "default",
			},
		}

		// Create request
		req = &driver.CreateMachineRequest{
			Machine:      machine,
			MachineClass: machineClass,
			Secret:       secret,
		}
	})

	Context("with volumes", func() {
		It("should pass BootVolume with all fields to API", func() {
			deleteOnTermination := true
			providerSpec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				BootVolume: &api.BootVolumeSpec{
					DeleteOnTermination: &deleteOnTermination,
					PerformanceClass:    "premium",
					Size:                100,
					Source: &api.BootVolumeSourceSpec{
						Type: "image",
						ID:   "550e8400-e29b-41d4-a716-446655440000",
					},
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

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
			Expect(capturedReq).NotTo(BeNil())
			Expect(capturedReq.BootVolume).NotTo(BeNil())
			Expect(capturedReq.BootVolume.DeleteOnTermination).To(Equal(&deleteOnTermination))
			Expect(capturedReq.BootVolume.PerformanceClass).To(Equal("premium"))
			Expect(capturedReq.BootVolume.Size).To(Equal(100))
			Expect(capturedReq.BootVolume.Source).NotTo(BeNil())
			Expect(capturedReq.BootVolume.Source.Type).To(Equal("image"))
			Expect(capturedReq.BootVolume.Source.ID).To(Equal("550e8400-e29b-41d4-a716-446655440000"))
		})

		It("should pass BootVolume with minimal config to API", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				BootVolume: &api.BootVolumeSpec{
					Size: 50,
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

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
			Expect(capturedReq).NotTo(BeNil())
			Expect(capturedReq.BootVolume).NotTo(BeNil())
			Expect(capturedReq.BootVolume.Size).To(Equal(50))
		})

		It("should pass Volumes array to API", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Volumes: []string{
					"550e8400-e29b-41d4-a716-446655440000",
					"660e8400-e29b-41d4-a716-446655440001",
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

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
			Expect(capturedReq).NotTo(BeNil())
			Expect(capturedReq.Volumes).To(Equal([]string{
				"550e8400-e29b-41d4-a716-446655440000",
				"660e8400-e29b-41d4-a716-446655440001",
			}))
		})

		It("should pass both BootVolume and Volumes to API", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				BootVolume: &api.BootVolumeSpec{
					Size: 50,
				},
				Volumes: []string{
					"550e8400-e29b-41d4-a716-446655440000",
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

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
			Expect(capturedReq).NotTo(BeNil())
			Expect(capturedReq.BootVolume).NotTo(BeNil())
			Expect(capturedReq.BootVolume.Size).To(Equal(50))
			Expect(capturedReq.Volumes).To(HaveLen(1))
		})

		It("should not send volumes when not specified", func() {
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
			Expect(capturedReq).NotTo(BeNil())
			Expect(capturedReq.BootVolume).To(BeNil())
			Expect(capturedReq.Volumes).To(BeNil())
		})
	})
})
