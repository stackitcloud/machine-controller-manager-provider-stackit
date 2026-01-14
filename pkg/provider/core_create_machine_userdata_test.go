// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"encoding/base64"

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
				"project-id":          []byte("11111111-2222-3333-4444-555555555555"),
				"serviceaccount.json": []byte(`{"credentials":{"iss":"test"}}`),
				"networkId":           []byte("770e8400-e29b-41d4-a716-446655440000"),
			},
		}

		// Create ProviderSpec
		providerSpec := &api.ProviderSpec{
			MachineType: "c2i.2",
			ImageID:     "12345678-1234-1234-1234-123456789abc",
			Region:      "eu01",
		}
		providerSpecRaw, _ := encodeProviderSpec(providerSpec)

		// Create MachineClass
		machineClass = &v1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine-class",
			},
			Provider: "stackit",
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

	Context("with userData", func() {
		It("should pass userData from ProviderSpec to API", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				UserData:    "#cloud-config\nruncmd:\n  - echo 'Hello from ProviderSpec'",
				Region:      "eu01",
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
			expectedUserData := base64.StdEncoding.EncodeToString([]byte("#cloud-config\nruncmd:\n  - echo 'Hello from ProviderSpec'"))
			Expect(capturedReq.UserData).To(Equal(expectedUserData))
		})

		It("should pass userData from Secret to API when ProviderSpec.UserData is empty", func() {
			secret.Data["userData"] = []byte("#cloud-config\nruncmd:\n  - echo 'Hello from Secret'")

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
			expectedUserData := base64.StdEncoding.EncodeToString([]byte("#cloud-config\nruncmd:\n  - echo 'Hello from Secret'"))
			Expect(capturedReq.UserData).To(Equal(expectedUserData))
		})

		It("should prefer ProviderSpec.UserData over Secret.userData", func() {
			providerSpec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "12345678-1234-1234-1234-123456789abc",
				UserData:    "#cloud-config from ProviderSpec",
				Region:      "eu01",
			}
			providerSpecRaw, _ := encodeProviderSpec(providerSpec)
			req.MachineClass.ProviderSpec.Raw = providerSpecRaw
			secret.Data["userData"] = []byte("#cloud-config from Secret")

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
			expectedUserData := base64.StdEncoding.EncodeToString([]byte("#cloud-config from ProviderSpec"))
			Expect(capturedReq.UserData).To(Equal(expectedUserData))
		})

		It("should not send userData when neither ProviderSpec nor Secret have it", func() {
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
			Expect(capturedReq.UserData).To(BeEmpty())
		})

		It("should handle empty userData in Secret gracefully", func() {
			secret.Data["userData"] = []byte("")

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
			Expect(capturedReq.UserData).To(BeEmpty())
		})
	})
})
