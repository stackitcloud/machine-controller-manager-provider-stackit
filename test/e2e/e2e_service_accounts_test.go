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
	Context("Machine service account configuration", func() {
		It("should create a Machine with serviceAccountMails", func() {
			secretName := generateResourceName("secret")
			machineClassName := generateResourceName("machineclass")
			machineName := generateResourceName("machine")

			// Create Secret
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

			// Create MachineClass with serviceAccountMails
			// Note: OpenAPI spec currently limits to maxItems: 1
			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  serviceAccountMails:
    - "my-service@sa.stackit.cloud"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource("machineclass", machineClassName, testNamespace, machineClassYAML)

			// Create Machine
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

			// Wait for Machine to get a ProviderID (indicates successful creation)
			By("waiting for Machine to have ProviderID set")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := cmd.CombinedOutput()
				return string(output)
			}, StandardTimeout, QuickPoll).Should(ContainSubstring("stackit://"), "Machine should have ProviderID set")

			// Verify Machine has ProviderID in correct format
			By("verifying Machine has ProviderID in correct format")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			providerID := string(output)
			Expect(providerID).To(MatchRegexp(`^stackit://[^/]+/[a-f0-9-]+$`), "ProviderID should match format stackit://<project>/<serverID>")

			// Note: We cannot easily verify the serviceAccountMails were actually passed to the API
			// without inspecting the mock server logs. The test verifies that:
			// 1. Machine creates successfully with serviceAccountMails specified
			// 2. No validation errors occur
			// Unit tests verify the API request includes the serviceAccountMails field
		})
	})
})
