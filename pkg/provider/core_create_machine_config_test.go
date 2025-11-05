// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"

	api "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

		// Create secret with projectId
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"projectId":    []byte("11111111-2222-3333-4444-555555555555"),
				"stackitToken": []byte("test-token-123"),
				"region":       []byte("eu01-1"),
			},
		}

		// Create ProviderSpec
		providerSpec := &api.ProviderSpec{
			MachineType: "c1.2",
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

	Context("with keypairName", func() {
		It("should pass KeypairName to API when specified", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				KeypairName: "my-ssh-key",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.KeypairName).To(Equal("my-ssh-key"))
		})

		It("should not send KeypairName when empty", func() {
			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.KeypairName).To(BeEmpty())
		})
	})

	Context("with availabilityZone", func() {
		It("should pass AvailabilityZone to API when specified", func() {
			providerSpec := &api.ProviderSpec{
				MachineType:      "c1.2",
				ImageID:          "12345678-1234-1234-1234-123456789abc",
				AvailabilityZone: "eu01-1",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.AvailabilityZone).To(Equal("eu01-1"))
		})

		It("should not send AvailabilityZone when empty", func() {
			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.AvailabilityZone).To(BeEmpty())
		})
	})

	Context("with affinityGroup", func() {
		It("should pass AffinityGroup to API when specified", func() {
			providerSpec := &api.ProviderSpec{
				MachineType:   "c1.2",
				ImageID:       "12345678-1234-1234-1234-123456789abc",
				AffinityGroup: "880e8400-e29b-41d4-a716-446655440000",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.AffinityGroup).To(Equal("880e8400-e29b-41d4-a716-446655440000"))
		})

		It("should not send AffinityGroup when empty", func() {
			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.AffinityGroup).To(BeEmpty())
		})

		It("should pass ServiceAccountMails to API when specified", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				ServiceAccountMails: []string{
					"my-service@sa.stackit.cloud",
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.ServiceAccountMails).To(Equal([]string{
				"my-service@sa.stackit.cloud",
			}))
		})

		It("should not send ServiceAccountMails when empty", func() {
			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.ServiceAccountMails).To(BeNil())
		})

		It("should pass Agent to API when specified", func() {
			provisioned := true
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Agent: &api.AgentSpec{
					Provisioned: &provisioned,
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.Agent).NotTo(BeNil())
			Expect(*capturedReq.Agent.Provisioned).To(BeTrue())
		})

		It("should not send Agent when nil", func() {
			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.Agent).To(BeNil())
		})

		It("should pass Metadata to API when specified", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				Metadata: map[string]interface{}{
					"environment": "production",
					"cost-center": "12345",
					"count":       42,
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.Metadata).NotTo(BeNil())
			Expect(capturedReq.Metadata).To(HaveLen(3))
			Expect(capturedReq.Metadata["environment"]).To(Equal("production"))
			Expect(capturedReq.Metadata["cost-center"]).To(Equal("12345"))
			// JSON marshaling converts int to float64
			Expect(capturedReq.Metadata["count"]).To(BeNumerically("==", 42))
		})

		It("should not send Metadata when nil", func() {
			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, token, projectID, region string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.Metadata).To(BeNil())
		})
	})
})
