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
})
