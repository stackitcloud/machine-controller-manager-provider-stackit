// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/base64"
	"fmt"

	api "github.com/aoepeople/machine-controller-manager-provider-stackit/pkg/provider/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
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
				"projectId": []byte("test-project-123"),
			},
		}

		// Create ProviderSpec
		providerSpec := &api.ProviderSpec{
			MachineType: "c1.2",
			ImageID:     "image-uuid-123",
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

	Context("with valid inputs", func() {
		It("should successfully create a machine", func() {
			resp, err := provider.CreateMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.ProviderID).To(Equal("stackit://test-project-123/550e8400-e29b-41d4-a716-446655440000"))
			Expect(resp.NodeName).To(Equal("test-machine"))
		})

		It("should call STACKIT API with correct parameters", func() {
			var capturedReq *CreateServerRequest
			var capturedProjectID string

			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
				capturedProjectID = projectID
				capturedReq = req
				return &Server{
					ID:     "test-server-id",
					Name:   req.Name,
					Status: "CREATING",
				}, nil
			}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedProjectID).To(Equal("test-project-123"))
			Expect(capturedReq).NotTo(BeNil())
			Expect(capturedReq.Name).To(Equal("test-machine"))
			Expect(capturedReq.MachineType).To(Equal("c1.2"))
			Expect(capturedReq.ImageID).To(Equal("image-uuid-123"))
		})
	})

	Context("with invalid ProviderSpec", func() {
		It("should fail when MachineType is missing", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "",
				ImageID:     "image-uuid-123",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.InvalidArgument))
		})

		It("should fail when ImageID is missing", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Context("with invalid Secret", func() {
		It("should fail when projectId is missing", func() {
			req.Secret.Data = map[string][]byte{}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Context("when STACKIT API fails", func() {
		It("should return Internal error on API failure", func() {
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
				return nil, fmt.Errorf("API connection failed")
			}

			_, err := provider.CreateMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.Internal))
		})
	})

	Context("with userData", func() {
		It("should pass userData from ProviderSpec to API", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "image-uuid-123",
				UserData:    "#cloud-config\nruncmd:\n  - echo 'Hello from ProviderSpec'",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
			expectedUserData := base64.StdEncoding.EncodeToString([]byte("#cloud-config\nruncmd:\n  - echo 'Hello from ProviderSpec'"))
			Expect(capturedReq.UserData).To(Equal(expectedUserData))
		})

		It("should pass userData from Secret to API when ProviderSpec.UserData is empty", func() {
			secret.Data["userData"] = []byte("#cloud-config\nruncmd:\n  - echo 'Hello from Secret'")

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
			expectedUserData := base64.StdEncoding.EncodeToString([]byte("#cloud-config\nruncmd:\n  - echo 'Hello from Secret'"))
			Expect(capturedReq.UserData).To(Equal(expectedUserData))
		})

		It("should prefer ProviderSpec.UserData over Secret.userData", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "image-uuid-123",
				UserData:    "#cloud-config from ProviderSpec",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw
			secret.Data["userData"] = []byte("#cloud-config from Secret")

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
			expectedUserData := base64.StdEncoding.EncodeToString([]byte("#cloud-config from ProviderSpec"))
			Expect(capturedReq.UserData).To(Equal(expectedUserData))
		})

		It("should not send userData when neither ProviderSpec nor Secret have it", func() {
			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.UserData).To(BeEmpty())
		})

		It("should handle empty userData in Secret gracefully", func() {
			secret.Data["userData"] = []byte("")

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
			Expect(capturedReq.UserData).To(BeEmpty())
		})
	})

	Context("with volumes", func() {
		It("should pass BootVolume with all fields to API", func() {
			deleteOnTermination := true
			providerSpec := &api.ProviderSpec{
				MachineType: "c1.2",
				ImageID:     "image-uuid-123",
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
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
				MachineType: "c1.2",
				ImageID:     "image-uuid-123",
				BootVolume: &api.BootVolumeSpec{
					Size: 50,
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
				MachineType: "c1.2",
				ImageID:     "image-uuid-123",
				Volumes: []string{
					"550e8400-e29b-41d4-a716-446655440000",
					"660e8400-e29b-41d4-a716-446655440001",
				},
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw

			var capturedReq *CreateServerRequest
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
				MachineType: "c1.2",
				ImageID:     "image-uuid-123",
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
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
			mockClient.createServerFunc = func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
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
