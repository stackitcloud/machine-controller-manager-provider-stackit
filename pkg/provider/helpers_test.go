package provider

import (
	"encoding/json"

	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Helpers", func() {
	Describe("parseProviderID", func() {
		Context("with valid ProviderIDs", func() {
			It("should parse a valid ProviderID", func() {
				projectID, serverID, err := parseProviderID("stackit://11111111-2222-3333-4444-555555555555/server-456")

				Expect(err).NotTo(HaveOccurred())
				Expect(projectID).To(Equal("11111111-2222-3333-4444-555555555555"))
				Expect(serverID).To(Equal("server-456"))
			})

			It("should parse ProviderID with UUID format", func() {
				projectID, serverID, err := parseProviderID("stackit://12345678-1234-1234-1234-123456789012/550e8400-e29b-41d4-a716-446655440000")

				Expect(err).NotTo(HaveOccurred())
				Expect(projectID).To(Equal("12345678-1234-1234-1234-123456789012"))
				Expect(serverID).To(Equal("550e8400-e29b-41d4-a716-446655440000"))
			})

			It("should parse ProviderID with alphanumeric IDs", func() {
				projectID, serverID, err := parseProviderID("stackit://proj-abc123/srv-xyz789")

				Expect(err).NotTo(HaveOccurred())
				Expect(projectID).To(Equal("proj-abc123"))
				Expect(serverID).To(Equal("srv-xyz789"))
			})
		})

		Context("with invalid ProviderIDs", func() {
			It("should fail when ProviderID is too short", func() {
				_, _, err := parseProviderID("stackit")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must start with 'stackit://'"))
			})

			It("should fail when ProviderID doesn't start with stackit://", func() {
				_, _, err := parseProviderID("aws://project/server")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must start with 'stackit://'"))
			})

			It("should fail when ProviderID is just the prefix", func() {
				_, _, err := parseProviderID("stackit://")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must have format"))
			})

			It("should fail when ProviderID has only projectId", func() {
				_, _, err := parseProviderID("stackit://project-123")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must have format"))
			})

			It("should fail when ProviderID has empty projectId", func() {
				_, _, err := parseProviderID("stackit:///server-456")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot be empty"))
			})

			It("should fail when ProviderID has empty serverId", func() {
				_, _, err := parseProviderID("stackit://project-123/")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot be empty"))
			})

			It("should fail when ProviderID has too many parts", func() {
				_, _, err := parseProviderID("stackit://project/server/extra")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("must have format"))
			})
		})
	})

	Describe("decodeProviderSpec", func() {
		Context("with valid MachineClass", func() {
			It("should decode a valid ProviderSpec", func() {
				providerSpecJSON := `{"machineType":"c2i.2","imageId":"image-123"}`
				machineClass := &v1alpha1.MachineClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-class",
					},
					ProviderSpec: runtime.RawExtension{
						Raw: []byte(providerSpecJSON),
					},
				}

				spec, err := decodeProviderSpec(machineClass)

				Expect(err).NotTo(HaveOccurred())
				Expect(spec).NotTo(BeNil())
				Expect(spec.MachineType).To(Equal("c2i.2"))
				Expect(spec.ImageID).To(Equal("image-123"))
			})
		})

		Context("with invalid MachineClass", func() {
			It("should fail when MachineClass is nil", func() {
				spec, err := decodeProviderSpec(nil)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("machineClass is nil"))
				Expect(spec).To(BeNil())
			})

			It("should fail when ProviderSpec is invalid JSON", func() {
				machineClass := &v1alpha1.MachineClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-class",
					},
					ProviderSpec: runtime.RawExtension{
						Raw: []byte(`{invalid json}`),
					},
				}

				spec, err := decodeProviderSpec(machineClass)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to decode"))
				Expect(spec).To(BeNil())
			})
		})
	})

	Describe("encodeProviderSpecForResponse", func() {
		Context("with valid ProviderSpec", func() {
			It("should encode a ProviderSpec to JSON", func() {
				spec := &api.ProviderSpec{
					MachineType: "c2i.2",
					ImageID:     "image-123",
				}

				data, err := encodeProviderSpecForResponse(spec)

				Expect(err).NotTo(HaveOccurred())
				Expect(data).NotTo(BeNil())

				// Verify it's valid JSON
				var decoded api.ProviderSpec
				err = json.Unmarshal(data, &decoded)
				Expect(err).NotTo(HaveOccurred())
				Expect(decoded.MachineType).To(Equal("c2i.2"))
				Expect(decoded.ImageID).To(Equal("image-123"))
			})

			It("should encode an empty ProviderSpec", func() {
				spec := &api.ProviderSpec{}

				data, err := encodeProviderSpecForResponse(spec)

				Expect(err).NotTo(HaveOccurred())
				Expect(data).NotTo(BeNil())
			})
		})
	})
})
