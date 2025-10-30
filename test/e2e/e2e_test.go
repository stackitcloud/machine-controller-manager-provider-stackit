/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MCM Provider STACKIT", func() {
	Context("Basic functionality", func() {
		It("should have MCM controller manager running", func() {
			cmd := exec.Command("kubectl", "get", "deployment",
				"machine-controller-manager", "-n", testNamespace, "-o", "jsonpath={.status.availableReplicas}")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.TrimSpace(string(output))).To(Equal("1"), "MCM deployment should have 1 available replica")
		})

		It("should have IAAS mock server running", func() {
			cmd := exec.Command("kubectl", "get", "deployment",
				"iaas", "-n", "stackitcloud", "-o", "jsonpath={.status.availableReplicas}")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(strings.TrimSpace(string(output))).To(Equal("1"), "IAAS mock should have 1 available replica")
		})

		It("should have MCM CRDs installed", func() {
			crds := []string{
				"machines.machine.sapcloud.io",
				"machineclasses.machine.sapcloud.io",
				"machinesets.machine.sapcloud.io",
				"machinedeployments.machine.sapcloud.io",
			}

			for _, crd := range crds {
				cmd := exec.Command("kubectl", "get", "crd", crd)
				output, err := cmd.CombinedOutput()
				Expect(err).NotTo(HaveOccurred(), "CRD %s should exist: %s", crd, string(output))
			}
		})

		It("should be able to access IAAS mock API", func() {
			curlPodName := generateResourceName("curl-test")
			curlPod := fmt.Sprintf(`
apiVersion: v1
kind: Pod
metadata:
  name: %s
  namespace: %s
spec:
  containers:
  - name: curl
    image: curlimages/curl:latest
    command: ["sleep", "60"]
  restartPolicy: Never
`, curlPodName, testNamespace)

			createAndTrackResource("pod", curlPodName, testNamespace, curlPod)

			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "pod", curlPodName, "-n", testNamespace, "-o", "jsonpath={.status.phase}")
				output, _ := cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, 60*time.Second, 2*time.Second).Should(Equal("Running"))

			cmd := exec.Command("kubectl", "exec", curlPodName, "-n", testNamespace, "--",
				"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
				"http://iaas.stackitcloud.svc.cluster.local")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			httpCode := strings.TrimSpace(string(output))
			Expect(httpCode).To(MatchRegexp("^[2-5]\\d{2}$"), "Mock API should respond with HTTP status code")
		})
	})

	Context("Machine lifecycle", func() {
		It("should create a Machine and call IAAS API", func() {
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
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set")

			By("verifying Machine exists and has correct ProviderID format")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "json")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("stackit://12345678-1234-1234-1234-123456789012/"))
		})

		It("should delete a Machine and call IAAS API", func() {
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

			By("waiting for Machine to have ProviderID set (via CreateMachine)")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set")

			By("deleting the Machine CR")
			cmd := exec.Command("kubectl", "delete", "machine", machineName, "-n", testNamespace, "--wait=false")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to initiate Machine deletion: %s", string(output))

			By("verifying Machine deletion was initiated (has deletionTimestamp)")
			Eventually(func() string {
				cmd = exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.metadata.deletionTimestamp}")
				output, _ = cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, "10s", "2s").ShouldNot(BeEmpty(), "Machine should have deletionTimestamp set")

			By("verifying DeleteMachine was attempted (check phase or operation)")
			Eventually(func() string {
				cmd = exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.status.lastOperation.type}")
				output, _ = cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, "15s", "2s").Should(Equal("Delete"), "Machine should have Delete operation attempted")

			// Note: We don't wait for complete deletion as the IAAS API may return errors (e.g., 400)
			// which prevent finalizer removal. We only verify the delete was attempted.
		})

		It("should report Machine status correctly", func() {
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

			By("waiting for Machine to have ProviderID set (via CreateMachine)")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.spec.providerID}")
				output, _ := cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set")

			By("verifying Machine status is reported (phase is set)")
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.status.currentStatus.phase}")
				output, _ := cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, "10s", "2s").ShouldNot(BeEmpty(), "Machine should have a status phase set")

			By("verifying Machine has a lastOperation type")
			cmd := exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.status.lastOperation.type}")
			output, _ := cmd.CombinedOutput()
			Expect(strings.TrimSpace(string(output))).To(Equal("Create"), "Machine should have Create operation type")

			// Note: With mock IAAS API, machines may stay in Pending/Creating state.
			// We verify status reporting works, not that machines reach Running state.
		})

		PIt("should list Machines with proper filtering", func() {
			// This test will verify ListMachines returns VMs with correct labels/tags
		})
	})
})
