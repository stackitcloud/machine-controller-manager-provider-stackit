// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MCM Provider STACKIT", func() {
	Context("Machine networking configuration", func() {
		It("should create a Machine with networkId in networking spec", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

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
      - echo "Machine bootstrapped"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    networkId: "770e8400-e29b-41d4-a716-446655440000"
  labels:
    test: "networking-networkid"
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

			By("waiting for Machine to have ProviderID set")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set with networking configuration")

			By("verifying Machine was created successfully")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Should be able to get Machine: %s", string(output))
		})

		It("should create a Machine with nicIds in networking spec", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

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
      - echo "Machine bootstrapped"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    nicIds:
      - "880e8400-e29b-41d4-a716-446655440000"
      - "990e8400-e29b-41d4-a716-446655440000"
  labels:
    test: "networking-nicids"
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

			By("waiting for Machine to have ProviderID set")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set with NIC IDs configuration")

			By("verifying Machine was created successfully")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Should be able to get Machine: %s", string(output))
		})

		It("should create a Machine with securityGroups", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

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
      - echo "Machine bootstrapped"
`, secretName, testNamespace)
			createAndTrackResource("secret", secretName, testNamespace, secretYAML)

			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    networkId: "770e8400-e29b-41d4-a716-446655440000"
  securityGroups:
    - "default"
    - "web-servers"
  labels:
    test: "networking-securitygroups"
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

			By("waiting for Machine to have ProviderID set")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set with security groups configuration")

			By("verifying Machine was created successfully")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace)
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Should be able to get Machine: %s", string(output))
		})
	})
})
