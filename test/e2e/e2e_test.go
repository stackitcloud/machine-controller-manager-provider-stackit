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
	"bytes"
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
				"machine-controller-manager", "-n", "default", "-o", "jsonpath={.status.availableReplicas}")
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
			// Create a test pod to curl the mock API
			curlPod := `
apiVersion: v1
kind: Pod
metadata:
  name: test-curl
  namespace: default
spec:
  containers:
  - name: curl
    image: curlimages/curl:latest
    command: ["sleep", "60"]
  restartPolicy: Never
`
			// Apply the pod
			cmd := exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = strings.NewReader(curlPod)
			_, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			// Wait for pod to be ready
			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "pod", "test-curl", "-n", "default", "-o", "jsonpath={.status.phase}")
				output, _ := cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, 60*time.Second, 2*time.Second).Should(Equal("Running"))

			// Test connectivity to mock API
			cmd = exec.Command("kubectl", "exec", "test-curl", "-n", "default", "--",
				"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
				"http://iaas.stackitcloud.svc.cluster.local")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			httpCode := strings.TrimSpace(string(output))
			// Prism mock server returns various codes, we just want to verify it's responding
			Expect(httpCode).To(MatchRegexp("^[2-5]\\d{2}$"), "Mock API should respond with HTTP status code")

			// Cleanup
			cmd = exec.Command("kubectl", "delete", "pod", "test-curl", "-n", "default")
			cmd.Run()
		})
	})

	Context("Machine lifecycle", func() {
		It("should create a Machine and call IAAS API", func() {
			var (
				cmd    *exec.Cmd
				err    error
				output []byte
			)

			By("creating a Secret with projectId and userData")
			secretYAML := `
apiVersion: v1
kind: Secret
metadata:
  name: test-secret
  namespace: default
type: Opaque
stringData:
  projectId: "12345678-1234-1234-1234-123456789012"
  userData: |
    #cloud-config
    runcmd:
      - echo "Machine bootstrapped"
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(secretYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create secret: %s", string(output))

			By("creating a MachineClass with minimal ProviderSpec")
			machineClassYAML := `
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: test-machineclass
  namespace: default
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
secretRef:
  name: test-secret
  namespace: default
provider: STACKIT
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(machineClassYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create MachineClass: %s", string(output))

			By("creating a Machine CR")
			machineYAML := `
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: test-machine
  namespace: default
spec:
  class:
    kind: MachineClass
    name: test-machineclass
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(machineYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create Machine: %s", string(output))

			By("waiting for Machine to have ProviderID set")
			Eventually(func() string {
				cmd = exec.Command("kubectl", "get", "machine", "test-machine", "-n", "default", "-o", "jsonpath={.spec.providerID}")
				output, _ = cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set")

			By("verifying Machine exists and has correct ProviderID format")
			cmd = exec.Command("kubectl", "get", "machine", "test-machine", "-n", "default", "-o", "json")
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("stackit://12345678-1234-1234-1234-123456789012/"))

			By("cleaning up test resources")
			deleteK8sResource("machine", "test-machine", "default")
			verifyK8sResourceDeleted("machine", "test-machine", "default")
			deleteK8sResource("machineclass", "test-machineclass", "default")
			verifyK8sResourceDeleted("machineclass", "test-machineclass", "default")
			deleteK8sResource("secret", "test-secret", "default")
			verifyK8sResourceDeleted("secret", "test-secret", "default")
		})

		It("should delete a Machine and call IAAS API", func() {
			var (
				cmd    *exec.Cmd
				err    error
				output []byte
			)

			By("creating a Secret with projectId")
			secretYAML := `
apiVersion: v1
kind: Secret
metadata:
  name: test-delete-secret
  namespace: default
type: Opaque
stringData:
  projectId: "12345678-1234-1234-1234-123456789012"
  userData: |
    #cloud-config
    runcmd:
      - echo "Machine bootstrapped"
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(secretYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create secret: %s", string(output))

			By("creating a MachineClass")
			machineClassYAML := `
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: test-delete-machineclass
  namespace: default
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
secretRef:
  name: test-delete-secret
  namespace: default
provider: STACKIT
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(machineClassYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create MachineClass: %s", string(output))

			By("creating a Machine CR")
			machineYAML := `
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: test-delete-machine
  namespace: default
spec:
  class:
    kind: MachineClass
    name: test-delete-machineclass
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(machineYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create Machine: %s", string(output))

			By("waiting for Machine to have ProviderID set (via CreateMachine)")
			Eventually(func() string {
				cmd = exec.Command("kubectl", "get", "machine", "test-delete-machine", "-n", "default", "-o", "jsonpath={.spec.providerID}")
				output, _ = cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set")

			By("deleting the Machine CR")
			cmd = exec.Command("kubectl", "delete", "machine", "test-delete-machine", "-n", "default")
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to delete Machine: %s", string(output))

			By("verifying Machine is deleted from Kubernetes")
			Eventually(func() string {
				cmd = exec.Command("kubectl", "get", "machine", "test-delete-machine", "-n", "default", "--ignore-not-found=true")
				output, _ = cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, "60s", "2s").Should(BeEmpty(), "Machine should be deleted from Kubernetes")

			By("cleaning up test resources")
			deleteK8sResource("machineclass", "test-delete-machineclass", "default")
			verifyK8sResourceDeleted("machineclass", "test-delete-machineclass", "default")
			deleteK8sResource("secret", "test-delete-secret", "default")
			verifyK8sResourceDeleted("secret", "test-delete-secret", "default")
		})

		It("should report Machine status correctly", func() {
			var (
				cmd    *exec.Cmd
				err    error
				output []byte
			)

			By("creating a Secret with projectId")
			secretYAML := `
apiVersion: v1
kind: Secret
metadata:
  name: test-status-secret
  namespace: default
type: Opaque
stringData:
  projectId: "12345678-1234-1234-1234-123456789012"
  userData: |
    #cloud-config
    runcmd:
      - echo "Machine bootstrapped"
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(secretYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create secret: %s", string(output))

			By("creating a MachineClass")
			machineClassYAML := `
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: test-status-machineclass
  namespace: default
providerSpec:
  machineType: "c1.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
secretRef:
  name: test-status-secret
  namespace: default
provider: STACKIT
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(machineClassYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create MachineClass: %s", string(output))

			By("creating a Machine CR")
			machineYAML := `
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: test-status-machine
  namespace: default
spec:
  class:
    kind: MachineClass
    name: test-status-machineclass
`
			cmd = exec.Command("kubectl", "apply", "-f", "-")
			cmd.Stdin = bytes.NewBufferString(machineYAML)
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), "Failed to create Machine: %s", string(output))

			By("waiting for Machine to have ProviderID set (via CreateMachine)")
			Eventually(func() string {
				cmd = exec.Command("kubectl", "get", "machine", "test-status-machine", "-n", "default", "-o", "jsonpath={.spec.providerID}")
				output, _ = cmd.CombinedOutput()
				return string(output)
			}, "60s", "2s").Should(ContainSubstring("stackit://"), "Machine should have ProviderID set")

			By("verifying Machine status is reported correctly")
			// The machine controller should be calling GetMachineStatus periodically
			// We verify that the Machine CR is in Running phase
			Eventually(func() string {
				cmd = exec.Command("kubectl", "get", "machine", "test-status-machine", "-n", "default", "-o", "jsonpath={.status.currentStatus.phase}")
				output, _ = cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, "60s", "2s").Should(Equal("Running"), "Machine should be in Running phase")

			By("cleaning up test resources")
			deleteK8sResource("machine", "test-status-machine", "default")
			verifyK8sResourceDeleted("machine", "test-status-machine", "default")
			deleteK8sResource("machineclass", "test-status-machineclass", "default")
			verifyK8sResourceDeleted("machineclass", "test-status-machineclass", "default")
			deleteK8sResource("secret", "test-status-secret", "default")
			verifyK8sResourceDeleted("secret", "test-status-secret", "default")
		})

		PIt("should list Machines with proper filtering", func() {
			// This test will verify ListMachines returns VMs with correct labels/tags
		})
	})
})
