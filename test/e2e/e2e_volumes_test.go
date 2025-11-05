// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/aoepeople/machine-controller-manager-provider-stackit/test/utils"
)

var _ = Describe("MCM Provider STACKIT", func() {
	Context("Machine volume configuration", func() {
		It("should create a Machine with BootVolume configuration", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Secret with minimal config
			secretYAML := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  projectId: "12345678-1234-1234-1234-123456789012"
  stackitToken: "mock-token-for-e2e-tests"
  region: "eu01-1"
  userData: |
    #cloud-config
    runcmd:
      - echo "Base bootstrap from Secret"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			// MachineClass with BootVolume configuration
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  bootVolume:
    size: 100
    performanceClass: "premium"
  labels:
    test: "bootvolume"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource("machineclass", machineClassName, testNamespace, machineClassYAML)

			machineYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: %s
  namespace: %s
spec:
  class:
    kind: MachineClass
    name: %s
`, machineName, testNamespace, machineClassName)
			createAndTrackResource("machine", machineName, testNamespace, machineYAML)

			By("waiting for Machine to get a ProviderID")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := utils.Run(cmd)
				return output
			}, MediumTimeout, StandardPoll).ShouldNot(BeEmpty(), "Machine should get a ProviderID")

			By("verifying Machine has ProviderID in correct format")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
			providerID, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(providerID).To(HavePrefix("stackit://"), "ProviderID should have stackit:// prefix")
		})

		It("should create a Machine with additional Volumes", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Secret with minimal config
			secretYAML := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  projectId: "12345678-1234-1234-1234-123456789012"
  stackitToken: "mock-token-for-e2e-tests"
  region: "eu01-1"
  userData: |
    #cloud-config
    runcmd:
      - echo "Base bootstrap from Secret"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			// MachineClass with additional volumes
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  volumes:
    - "550e8400-e29b-41d4-a716-446655440000"
    - "660e8400-e29b-41d4-a716-446655440001"
  labels:
    test: "volumes"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource("machineclass", machineClassName, testNamespace, machineClassYAML)

			machineYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: %s
  namespace: %s
spec:
  class:
    kind: MachineClass
    name: %s
`, machineName, testNamespace, machineClassName)
			createAndTrackResource("machine", machineName, testNamespace, machineYAML)

			By("waiting for Machine to get a ProviderID")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := utils.Run(cmd)
				return output
			}, MediumTimeout, StandardPoll).ShouldNot(BeEmpty(), "Machine should get a ProviderID")

			By("verifying Machine has ProviderID in correct format")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
			providerID, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(providerID).To(HavePrefix("stackit://"), "ProviderID should have stackit:// prefix")
		})

		It("should create a Machine with both BootVolume and Volumes", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Secret with minimal config
			secretYAML := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  projectId: "12345678-1234-1234-1234-123456789012"
  stackitToken: "mock-token-for-e2e-tests"
  region: "eu01-1"
  userData: |
    #cloud-config
    runcmd:
      - echo "Base bootstrap from Secret"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			// MachineClass with both BootVolume and Volumes
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  bootVolume:
    size: 50
    performanceClass: "standard"
  volumes:
    - "550e8400-e29b-41d4-a716-446655440000"
  labels:
    test: "bootvolume-and-volumes"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource("machineclass", machineClassName, testNamespace, machineClassYAML)

			machineYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: %s
  namespace: %s
spec:
  class:
    kind: MachineClass
    name: %s
`, machineName, testNamespace, machineClassName)
			createAndTrackResource("machine", machineName, testNamespace, machineYAML)

			By("waiting for Machine to get a ProviderID")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := utils.Run(cmd)
				return output
			}, MediumTimeout, StandardPoll).ShouldNot(BeEmpty(), "Machine should get a ProviderID")

			By("verifying Machine has ProviderID in correct format")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
			providerID, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(providerID).To(HavePrefix("stackit://"), "ProviderID should have stackit:// prefix")
		})

		It("should create a Machine with BootVolume from snapshot (no imageId)", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Secret with minimal config
			secretYAML := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: Opaque
stringData:
  projectId: "12345678-1234-1234-1234-123456789012"
  stackitToken: "mock-token-for-e2e-tests"
  region: "eu01-1"
  userData: |
    #cloud-config
    runcmd:
      - echo "Base bootstrap from Secret"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			// MachineClass WITHOUT imageId, using bootVolume.source instead
			// This tests the critical "imageId OR bootVolume.source" validation logic
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  bootVolume:
    size: 80
    performanceClass: "premium"
    source:
      type: "snapshot"
      id: "660e8400-e29b-41d4-a716-446655440000"
  labels:
    test: "bootvolume-from-snapshot"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource("machineclass", machineClassName, testNamespace, machineClassYAML)

			machineYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: %s
  namespace: %s
spec:
  class:
    kind: MachineClass
    name: %s
`, machineName, testNamespace, machineClassName)
			createAndTrackResource("machine", machineName, testNamespace, machineYAML)

			By("waiting for Machine to get a ProviderID")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := utils.Run(cmd)
				return output
			}, MediumTimeout, StandardPoll).ShouldNot(BeEmpty(), "Machine should get a ProviderID when booting from snapshot")

			By("verifying Machine has ProviderID in correct format")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
			providerID, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(providerID).To(HavePrefix("stackit://"), "ProviderID should have stackit:// prefix")

			By("verifying MachineClass has no imageId field")
			cmd = exec.Command("kubectl", "get", "machineclass", machineClassName, "-n", testNamespace, "-o", "jsonpath={.providerSpec.imageId}")
			imageID, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(imageID).To(BeEmpty(), "MachineClass should not have imageId when using bootVolume.source")

			By("verifying MachineClass has bootVolume.source configuration")
			cmd = exec.Command("kubectl", "get", "machineclass", machineClassName, "-n", testNamespace, "-o", "jsonpath={.providerSpec.bootVolume.source.type}")
			sourceType, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(sourceType).To(Equal("snapshot"), "BootVolume source type should be 'snapshot'")

			cmd = exec.Command("kubectl", "get", "machineclass", machineClassName, "-n", testNamespace, "-o", "jsonpath={.providerSpec.bootVolume.source.id}")
			sourceID, err := utils.Run(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(sourceID).To(Equal("660e8400-e29b-41d4-a716-446655440000"), "BootVolume source ID should match snapshot UUID")
		})
	})
})
