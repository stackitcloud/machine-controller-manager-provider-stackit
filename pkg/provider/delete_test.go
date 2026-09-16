package provider

import (
	"context"
	"fmt"
	"time"

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

var _ = Describe("DeleteMachine", func() {
	var (
		ctx          context.Context
		provider     *Provider
		mockClient   *mock.StackitClient
		req          *driver.DeleteMachineRequest
		secret       *corev1.Secret
		machineClass *v1alpha1.MachineClass
		machine      *v1alpha1.Machine
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockClient = &mock.StackitClient{
			GetServerFunc: func(_ context.Context, _, _, _ string) (*client.Server, error) {
				return nil, fmt.Errorf("%w: status 404", client.ErrServerNotFound)
			},
		}
		provider = &Provider{
			client:          mockClient,
			pollingInterval: 10 * time.Millisecond,
			pollingTimeout:  5 * time.Second,
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
			Networking: &api.NetworkingSpec{
				NetworkID: "770e8400-e29b-41d4-a716-446655440000",
			},
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
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}
			mockClient.GetServerFunc = func(_ context.Context, _, _, _ string) (*client.Server, error) {
				return nil, fmt.Errorf("%w: status 404", client.ErrServerNotFound)
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
		})

		It("should call STACKIT API with correct parameters", func() {
			var capturedProjectID string
			var capturedServerID string

			mockClient.DeleteServerFunc = func(_ context.Context, projectID, _, serverID string) error {
				capturedProjectID = projectID
				capturedServerID = serverID
				return nil
			}
			mockClient.GetServerFunc = func(_ context.Context, _, _, _ string) (*client.Server, error) {
				return nil, fmt.Errorf("%w: status 404", client.ErrServerNotFound)
			}

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(capturedProjectID).To(Equal("11111111-2222-3333-4444-555555555555"))
			Expect(capturedServerID).To(Equal("550e8400-e29b-41d4-a716-446655440000"))
		})

		It("should poll GetServer until server is deleted", func() {
			getServerCallCount := 0

			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}
			mockClient.GetServerFunc = func(_ context.Context, _, _, _ string) (*client.Server, error) {
				getServerCallCount++
				// First call returns server still exists, second call returns not found
				if getServerCallCount == 1 {
					return &client.Server{
						ID:     "550e8400-e29b-41d4-a716-446655440000",
						Name:   "test-machine",
						Status: "SHUTTING_DOWN",
					}, nil
				}
				return nil, fmt.Errorf("%w: status 404", client.ErrServerNotFound)
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(getServerCallCount).To(BeNumerically(">=", 2))
		})
	})

	Context("with missing or invalid ProviderID", func() {
		It("should still delete the machine when ProviderID is missing", func() {
			machine.Spec.ProviderID = ""

			mockClient.ListServersFunc = func(_ context.Context, _, _ string, selector map[string]string) ([]*client.Server, error) {
				Expect(selector).To(Equal(map[string]string{StackitMachineLabel: "test-machine"}))
				return []*client.Server{{
					ID:   "550e8400-e29b-41d4-a716-446655440000",
					Name: "test-machine",
				}}, nil
			}
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}

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

	Context("when deleting a migrated machine", func() {
		It("deletes all matching servers and NICs after an unfiltered lookup", func() {
			machine.Spec.ProviderID = ""
			machine.Annotations = map[string]string{migratedMachineAnnotation: "true"}
			var deletedServerIDs, deletedNICIDs []string
			mockClient.ListServersFunc = func(_ context.Context, _, _ string, selector map[string]string) ([]*client.Server, error) {
				Expect(selector).To(BeNil())
				return []*client.Server{
					{ID: "server-1", Name: "test-machine"},
					{ID: "other-server", Name: "another-machine"},
					{ID: "server-2", Name: "test-machine"},
				}, nil
			}
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, serverID string) error {
				deletedServerIDs = append(deletedServerIDs, serverID)
				return nil
			}
			mockClient.ListNICsFunc = func(_ context.Context, _, _, networkID string) ([]*client.NIC, error) {
				Expect(networkID).To(Equal("770e8400-e29b-41d4-a716-446655440000"))
				return []*client.NIC{
					{ID: "nic-1", NetworkID: networkID, Name: "test-machine"},
					{ID: "other-nic", NetworkID: networkID, Name: "another-machine"},
					{ID: "nic-2", NetworkID: networkID, Name: "test-machine"},
				}, nil
			}
			mockClient.DeleteNICFunc = func(_ context.Context, _, _, networkID, nicID string) error {
				Expect(networkID).To(Equal("770e8400-e29b-41d4-a716-446655440000"))
				deletedNICIDs = append(deletedNICIDs, nicID)
				return nil
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(deletedServerIDs).To(ConsistOf("server-1", "server-2"))
			Expect(deletedNICIDs).To(ConsistOf("nic-1", "nic-2"))
		})

		It("waits for server deletion before deleting NICs", func() {
			machine.Spec.ProviderID = ""
			machine.Annotations = map[string]string{migratedMachineAnnotation: "true"}

			var executionOrder []string
			mockClient.ListServersFunc = func(_ context.Context, _, _ string, _ map[string]string) ([]*client.Server, error) {
				return []*client.Server{
					{ID: "server-1", Name: "test-machine"},
				}, nil
			}
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, serverID string) error {
				executionOrder = append(executionOrder, "delete-server:"+serverID)
				return nil
			}
			mockClient.GetServerFunc = func(_ context.Context, _, _, serverID string) (*client.Server, error) {
				executionOrder = append(executionOrder, "wait-server:"+serverID)
				return nil, fmt.Errorf("%w: status 404", client.ErrServerNotFound)
			}
			mockClient.ListNICsFunc = func(_ context.Context, _, _, networkID string) ([]*client.NIC, error) {
				executionOrder = append(executionOrder, "list-nics")
				return []*client.NIC{
					{ID: "nic-1", NetworkID: networkID, Name: "test-machine"},
				}, nil
			}
			mockClient.DeleteNICFunc = func(_ context.Context, _, _, _, nicID string) error {
				executionOrder = append(executionOrder, "delete-nic:"+nicID)
				return nil
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(executionOrder).To(Equal([]string{
				"delete-server:server-1",
				"wait-server:server-1",
				"list-nics",
				"delete-nic:nic-1",
			}))
		})

		It("does not delete NICs if server deletion wait times out", func() {
			machine.Spec.ProviderID = ""
			machine.Annotations = map[string]string{migratedMachineAnnotation: "true"}
			provider.pollingTimeout = 20 * time.Millisecond

			mockClient.ListServersFunc = func(_ context.Context, _, _ string, _ map[string]string) ([]*client.Server, error) {
				return []*client.Server{
					{ID: "server-1", Name: "test-machine"},
				}, nil
			}
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}
			mockClient.GetServerFunc = func(_ context.Context, _, _, serverID string) (*client.Server, error) {
				return &client.Server{ID: serverID, Status: "SHUTTING_DOWN"}, nil
			}
			deleteNICCalled := false
			mockClient.DeleteNICFunc = func(_ context.Context, _, _, _, _ string) error {
				deleteNICCalled = true
				return nil
			}

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.DeadlineExceeded))
			Expect(deleteNICCalled).To(BeFalse())
		})

		It("safely skips NIC cleanup if Networking is nil", func() {
			machine.Spec.ProviderID = ""
			machine.Annotations = map[string]string{migratedMachineAnnotation: "true"}

			spec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "image-uuid-123",
				Region:      "eu01",
			}
			specRaw, _ := mock.EncodeProviderSpec(spec)
			machineClass.ProviderSpec.Raw = specRaw

			mockClient.ListServersFunc = func(_ context.Context, _, _ string, _ map[string]string) ([]*client.Server, error) {
				return []*client.Server{
					{ID: "server-1", Name: "test-machine"},
				}, nil
			}
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}
			listNICsCalled := false
			mockClient.ListNICsFunc = func(_ context.Context, _, _, _ string) ([]*client.NIC, error) {
				listNICsCalled = true
				return nil, nil
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(listNICsCalled).To(BeFalse())
		})

		It("safely skips NIC cleanup if NetworkID is empty", func() {
			machine.Spec.ProviderID = ""
			machine.Annotations = map[string]string{migratedMachineAnnotation: "true"}

			spec := &api.ProviderSpec{
				MachineType: "c2i.2",
				ImageID:     "image-uuid-123",
				Region:      "eu01",
				Networking: &api.NetworkingSpec{
					NICIDs: []string{"nic-123"},
				},
			}
			specRaw, _ := mock.EncodeProviderSpec(spec)
			machineClass.ProviderSpec.Raw = specRaw

			mockClient.ListServersFunc = func(_ context.Context, _, _ string, _ map[string]string) ([]*client.Server, error) {
				return []*client.Server{
					{ID: "server-1", Name: "test-machine"},
				}, nil
			}
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}
			listNICsCalled := false
			mockClient.ListNICsFunc = func(_ context.Context, _, _, _ string) ([]*client.NIC, error) {
				listNICsCalled = true
				return nil, nil
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(listNICsCalled).To(BeFalse())
		})
	})

	Context("when machine not found", func() {
		It("should return success if machine does not exist (idempotent)", func() {
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return fmt.Errorf("%w: status 404", client.ErrServerNotFound)
			}

			resp, err := provider.DeleteMachine(ctx, req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
		})
	})

	Context("when STACKIT API fails", func() {
		It("should return error when API call fails", func() {
			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return fmt.Errorf("API connection failed")
			}

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.Internal))
		})

		It("should return DeadlineExceeded when waiting for server deletion times out", func() {
			provider.pollingTimeout = 20 * time.Millisecond

			mockClient.DeleteServerFunc = func(_ context.Context, _, _, _ string) error {
				return nil
			}
			mockClient.GetServerFunc = func(_ context.Context, _, _, serverID string) (*client.Server, error) {
				return &client.Server{ID: serverID, Status: "SHUTTING_DOWN"}, nil
			}

			_, err := provider.DeleteMachine(ctx, req)

			Expect(err).To(HaveOccurred())
			statusErr, ok := status.FromError(err)
			Expect(ok).To(BeTrue())
			Expect(statusErr.Code()).To(Equal(codes.DeadlineExceeded))
		})
	})
})
