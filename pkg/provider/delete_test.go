// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/client"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("DeleteMachine", func() {
	var (
		ctx          context.Context
		provider     *Provider
		mockClient   *mockStackitClient
		req          *driver.DeleteMachineRequest
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
				"projectId":         []byte("11111111-2222-3333-4444-555555555555"),
				"serviceAccountKey": []byte(`{"credentials":{"iss":"test"}}`),
				"networkId":         []byte("770e8400-e29b-41d4-a716-446655440000"),
			},
		}

		// Create ProviderSpec
		providerSpec := &api.ProviderSpec{
			MachineType: "c2i.2",
			ImageID:     "image-uuid-123",
			Region:      "eu01",
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

		// Create Machine with ProviderID (set by CreateMachine)
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
		req = &driver.DeleteMachineRequest{
			Machine:      machine,
			MachineClass: machineClass,
			Secret:       secret,
		}
	})

	Context("with valid inputs", func() {
		It("should successfully delete a machine", func() {
			mockClient.deleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
		})

		It("should call STACKIT API with correct parameters", func() {
			var capturedProjectID string
			var capturedServerID string

			mockClient.deleteServerFunc = func(_ context.Context, projectID, _, serverID string) error {
				capturedProjectID = projectID
				capturedServerID = serverID
				return nil
			}

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedProjectID).To(Equal("11111111-2222-3333-4444-555555555555"))
			Expect(capturedServerID).To(Equal("550e8400-e29b-41d4-a716-446655440000"))
		})
	})

	Context("with missing or invalid ProviderID", func() {
		It("should still delete the machine when ProviderID is missing", func() {
			machine.Spec.ProviderID = ""

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return InvalidArgument when ProviderID has invalid format", func() {
			machine.Spec.ProviderID = "invalid-format"

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.InvalidArgument))
		})
	})

	Context("when machine not found", func() {
		It("should return success if machine does not exist (idempotent)", func() {
			mockClient.deleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return fmt.Errorf("%w: status 404", client.ErrServerNotFound)
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
		})
	})

	Context("when STACKIT API fails", func() {
		It("should return error when API call fails", func() {
			mockClient.deleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return fmt.Errorf("API connection failed")
			}

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.Internal))
		})
	})
})
