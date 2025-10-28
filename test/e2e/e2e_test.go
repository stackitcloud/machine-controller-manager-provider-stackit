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
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
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

	Context("Machine lifecycle (placeholder)", func() {
		// TODO: Implement once provider CreateMachine/DeleteMachine are implemented
		PIt("should create a Machine and call IAAS API", func() {
			// This test will create a MachineClass and Machine CR
			// and verify that the provider attempts to create a VM via mock IAAS API
		})

		PIt("should delete a Machine and call IAAS API", func() {
			// This test will delete a Machine CR
			// and verify that the provider attempts to delete the VM via mock IAAS API
		})

		PIt("should report Machine status correctly", func() {
			// This test will verify GetMachineStatus returns correct VM state
		})

		PIt("should list Machines with proper filtering", func() {
			// This test will verify ListMachines returns VMs with correct labels/tags
		})
	})
})
