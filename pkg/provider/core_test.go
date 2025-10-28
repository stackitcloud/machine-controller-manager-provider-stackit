// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
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

// mockStackitClient is a mock implementation of StackitClient for testing
type mockStackitClient struct {
	createServerFunc func(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error)
}

func (m *mockStackitClient) CreateServer(ctx context.Context, projectID string, req *CreateServerRequest) (*Server, error) {
	if m.createServerFunc != nil {
		return m.createServerFunc(ctx, projectID, req)
	}
	return &Server{
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   req.Name,
		Status: "CREATING",
	}, nil
}

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
})

// Helper function to encode ProviderSpec
func encodeProviderSpec(spec *api.ProviderSpec) ([]byte, error) {
	// For now, use a simple JSON encoding
	// In real implementation, this would use proper encoding
	return []byte(fmt.Sprintf(`{"machineType":"%s","imageId":"%s"}`, spec.MachineType, spec.ImageID)), nil
}
