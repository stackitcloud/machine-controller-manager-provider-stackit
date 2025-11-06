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
	Context("Machine userData configuration", func() {
		It("should create a Machine with userData in ProviderSpec", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Secret with base userData (required by MCM for node bootstrapping)
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
  networkId: "770e8400-e29b-41d4-a716-446655440000"
  userData: |
    #cloud-config
    runcmd:
      - echo "Base bootstrap from Secret"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			// MachineClass with userData in providerSpec (overrides Secret.userData)
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  userData: |
    #cloud-config
    runcmd:
      - echo "UserData from ProviderSpec (overrides Secret)"
  labels:
    test: "userdata-providerspec"
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

		It("should create a Machine with userData in Secret", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Secret with userData
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
  networkId: "770e8400-e29b-41d4-a716-446655440000"
  userData: |
    #cloud-config
    runcmd:
      - echo "UserData from Secret"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			// MachineClass without userData in providerSpec
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  labels:
    test: "userdata-secret"
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

		It("should prefer ProviderSpec userData over Secret userData", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Secret with userData (should be ignored)
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
  networkId: "770e8400-e29b-41d4-a716-446655440000"
  userData: |
    #cloud-config
    runcmd:
      - echo "UserData from Secret (should be ignored)"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			// MachineClass with userData in providerSpec (should take precedence)
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  userData: |
    #cloud-config
    runcmd:
      - echo "UserData from ProviderSpec (should take precedence)"
  labels:
    test: "userdata-precedence"
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
	})
})
