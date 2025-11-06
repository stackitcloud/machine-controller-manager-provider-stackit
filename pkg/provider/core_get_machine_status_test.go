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

var _ = Describe("GetMachineStatus", func() {
	var (
		ctx          context.Context
		provider     *Provider
		mockClient   *mockStackitClient
		req          *driver.GetMachineStatusRequest
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
			},
		}

		// Create ProviderSpec
		providerSpec := &api.ProviderSpec{
			MachineType: "c2i.2",
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

		// Create Machine with ProviderID
		machine = &v1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-machine",
				Namespace: "default",
			},
			Spec: v1alpha1.MachineSpec{
				ProviderID: "stackit://11111111-2222-3333-4444-555555555555/550e8400-e29b-41d4-a716-446655440000",
			},
		}

		// Create request
		req = &driver.GetMachineStatusRequest{
			Machine:      machine,
			MachineClass: machineClass,
			Secret:       secret,
		}
	})

	Context("with valid inputs", func() {
		It("should successfully get machine status when server exists", func() {
			mockClient.getServerFunc = func(ctx context.Context, token, projectID, serverID string) (*Server, error) {
				return &Server{
					ID:     serverID,
					Name:   "test-machine",
					Status: "RUNNING",
				}, nil
			}

			resp, err := provider.GetMachineStatus(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.ProviderID).To(Equal("stackit://11111111-2222-3333-4444-555555555555/550e8400-e29b-41d4-a716-446655440000"))
			Expect(resp.NodeName).To(Equal("test-machine"))
		})

		It("should call STACKIT API with correct parameters", func() {
			var capturedProjectID string
			var capturedServerID string

			mockClient.getServerFunc = func(ctx context.Context, token, projectID, serverID string) (*Server, error) {
				capturedProjectID = projectID
				capturedServerID = serverID
				return &Server{
					ID:     serverID,
					Name:   "test-machine",
					Status: "RUNNING",
				}, nil
			}

			_, err := provider.GetMachineStatus(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedProjectID).To(Equal("11111111-2222-3333-4444-555555555555"))
			Expect(capturedServerID).To(Equal("550e8400-e29b-41d4-a716-446655440000"))
		})
	})

	Context("with missing or invalid ProviderID", func() {
		It("should return NotFound when ProviderID is missing (machine not created yet)", func() {
			machine.Spec.ProviderID = ""

			_, err := provider.GetMachineStatus(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.NotFound))
		})

		It("should return InvalidArgument when ProviderID has invalid format", func() {
			machine.Spec.ProviderID = "invalid-format"

			_, err := provider.GetMachineStatus(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.InvalidArgument))
		})

		It("should return InvalidArgument when ProviderID is missing server ID", func() {
			machine.Spec.ProviderID = "stackit://11111111-2222-3333-4444-555555555555/"

			_, err := provider.GetMachineStatus(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Context("when server does not exist", func() {
		It("should return NotFound when server is not found", func() {
			mockClient.getServerFunc = func(ctx context.Context, token, projectID, serverID string) (*Server, error) {
				return nil, fmt.Errorf("%w: status 404", ErrServerNotFound)
			}

			_, err := provider.GetMachineStatus(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.NotFound))
		})
	})

	Context("when STACKIT API fails", func() {
		It("should return Internal error on API failure", func() {
			mockClient.getServerFunc = func(ctx context.Context, token, projectID, serverID string) (*Server, error) {
				return nil, fmt.Errorf("API connection failed")
			}

			_, err := provider.GetMachineStatus(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.Internal))
		})
	})
})
