package e2e

import (
	"context"
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

		It("should be able to access IAAS mock API", func(ctx context.Context) {
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

			createAndTrackResource(ctx, "pod", curlPodName, testNamespace, curlPod)

			Eventually(func() string {
				cmd := exec.Command("kubectl", "get", "pod", curlPodName, "-n", testNamespace, "-o", "jsonpath={.status.phase}")
				output, _ := cmd.CombinedOutput()
				return strings.TrimSpace(string(output))
			}, 60*time.Second, 2*time.Second).Should(Equal("Running"))

			cmd := exec.CommandContext(ctx, "kubectl", "exec", curlPodName, "-n", testNamespace, "--",
				"curl", "-s", "-o", "/dev/null", "-w", "%{http_code}",
				"http://iaas.stackitcloud.svc.cluster.local")
			output, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			httpCode := strings.TrimSpace(string(output))
			Expect(httpCode).To(MatchRegexp("^[2-5]\\d{2}$"), "Mock API should respond with HTTP status code")
		})
	})
})
