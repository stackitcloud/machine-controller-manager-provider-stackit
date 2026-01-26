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
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/client/mock"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("ListMachines", func() {
	var (
		ctx          context.Context
		provider     *Provider
		mockClient   *mock.StackitClient
		req          *driver.ListMachinesRequest
		secret       *corev1.Secret
		machineClass *v1alpha1.MachineClass
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockClient = &mock.StackitClient{}
		provider = &Provider{
			client: mockClient,
		}

		// Create secret with projectId
		secret = &corev1.Secret{
			Data: map[string][]byte{
				"project-id":          []byte("11111111-2222-3333-4444-555555555555"),
				"serviceaccount.json": []byte(`{"credentials":{"iss":"test"}}`),
			},
		}

		// Create ProviderSpec
		providerSpec := &api.ProviderSpec{
			MachineType: "c2i.2",
			ImageID:     "image-uuid-123",
			Region:      "eu01",
		}
		providerSpecRaw, _ := mock.EncodeProviderSpec(providerSpec)

		// Create MachineClass
		machineClass = &v1alpha1.MachineClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-machine-class",
			},
			ProviderSpec: runtime.RawExtension{
				Raw: providerSpecRaw,
			},
		}

		// Create request
		req = &driver.ListMachinesRequest{
			MachineClass: machineClass,
			Secret:       secret,
		}
	})

	Context("with valid inputs", func() {
		It("should list machines filtered by MachineClass label", func() {
			mockClient.ListServersFunc = func(_ context.Context, _, _ string, selector map[string]string) ([]*client.Server, error) {
				Expect(selector["kubernetes.io/machineclass"]).To(Equal("test-machine-class"))

				return []*client.Server{
					{
						ID:   "server-1",
						Name: "machine-1",
						Labels: map[string]string{
							"kubernetes.io/machineclass": "test-machine-class",
							"kubernetes.io/machine":      "machine-1",
						},
					},
					{
						ID:   "server-2",
						Name: "machine-2",
						Labels: map[string]string{
							"kubernetes.io/machineclass": "test-machine-class",
							"kubernetes.io/machine":      "machine-2",
						},
					},
				}, nil
			}

			resp, err := provider.ListMachines(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.MachineList).To(HaveLen(2))
			Expect(resp.MachineList).To(HaveKeyWithValue("stackit://11111111-2222-3333-4444-555555555555/server-1", "machine-1"))
			Expect(resp.MachineList).To(HaveKeyWithValue("stackit://11111111-2222-3333-4444-555555555555/server-2", "machine-2"))
			Expect(resp.MachineList).NotTo(HaveKey("stackit://11111111-2222-3333-4444-555555555555/server-3"))
		})

		It("should return empty list when no servers match", func() {
			mockClient.ListServersFunc = func(_ context.Context, _, _ string, _ map[string]string) ([]*client.Server, error) {
				return []*client.Server{}, nil
			}

			resp, err := provider.ListMachines(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.MachineList).To(BeEmpty())
		})

		It("should return empty list when no servers exist", func() {
			mockClient.ListServersFunc = func(_ context.Context, _, _ string, _ map[string]string) ([]*client.Server, error) {
				return []*client.Server{}, nil
			}

			resp, err := provider.ListMachines(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.MachineList).To(BeEmpty())
		})
	})

	Context("when STACKIT API fails", func() {
		It("should return Internal error on API failure", func() {
			mockClient.ListServersFunc = func(_ context.Context, _, _ string, _ map[string]string) ([]*client.Server, error) {
				return nil, fmt.Errorf("API connection failed")
			}

			_, err := provider.ListMachines(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.Internal))
		})
	})
})
