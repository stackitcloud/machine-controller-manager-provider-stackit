package e2e

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MCM Provider STACKIT", func(ctx context.Context) {
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
  serviceAccountKey: "{}"
  region: "eu01-1"
  userData: |
    #cloud-config
    runcmd:
      - echo "Machine bootstrapped"
`, secretName, testNamespace)
			createAndTrackResource(ctx, "secret", secretName, testNamespace, secretYAML)

			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c2i.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    networkId: "770e8400-e29b-41d4-a716-446655440000"
  labels:
    application: "web-server"
    team: "platform"
    cost-center: "engineering"
    environment: "e2e-test"
    component: "worker-node"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource(ctx, "machineclass", machineClassName, testNamespace, machineClassYAML)

			machineYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: %s
  namespace: %s
  labels:
    test-type: "e2e"
    test-phase: "create"
    environment: "test"
    machine-role: "worker"
spec:
  class:
    kind: MachineClass
    name: %s
`, machineName, testNamespace, machineClassName)
			createAndTrackResource(ctx, "machine", machineName, testNamespace, machineYAML)

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

			By("verifying Machine has correct labels applied")
			cmd = exec.Command("kubectl", "get", "machine", machineName, "-n", testNamespace, "-o", "jsonpath={.metadata.labels}")
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			// Verify our test labels are present
			Expect(string(output)).To(ContainSubstring("test-type"))
			Expect(string(output)).To(ContainSubstring("e2e"))
			Expect(string(output)).To(ContainSubstring("test-phase"))
			Expect(string(output)).To(ContainSubstring("create"))
			Expect(string(output)).To(ContainSubstring("environment"))
			Expect(string(output)).To(ContainSubstring("test"))
			Expect(string(output)).To(ContainSubstring("machine-role"))
			Expect(string(output)).To(ContainSubstring("worker"))

			By("verifying MachineClass has correct ProviderSpec labels")
			cmd = exec.Command("kubectl", "get", "machineclass", machineClassName, "-n", testNamespace, "-o", "jsonpath={.providerSpec.labels}")
			output, err = cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			// Verify ProviderSpec labels are present
			Expect(string(output)).To(ContainSubstring("application"))
			Expect(string(output)).To(ContainSubstring("web-server"))
			Expect(string(output)).To(ContainSubstring("team"))
			Expect(string(output)).To(ContainSubstring("platform"))
			Expect(string(output)).To(ContainSubstring("cost-center"))
			Expect(string(output)).To(ContainSubstring("engineering"))
			Expect(string(output)).To(ContainSubstring("component"))
			Expect(string(output)).To(ContainSubstring("worker-node"))
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
  serviceAccountKey: "{}"
  region: "eu01-1"
  userData: |
    #cloud-config
    runcmd:
      - echo "Machine bootstrapped"
`, secretName, testNamespace)
			createAndTrackResource(ctx, "secret", secretName, testNamespace, secretYAML)

			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c2i.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    networkId: "770e8400-e29b-41d4-a716-446655440000"
  labels:
    application: "database"
    team: "backend"
    cost-center: "ops"
    environment: "e2e-delete-test"
    component: "storage-node"
    lifecycle: "temporary"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource(ctx, "machineclass", machineClassName, testNamespace, machineClassYAML)

			machineYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: %s
  namespace: %s
  labels:
    test-type: "e2e"
    test-phase: "delete"
    environment: "test"
    machine-role: "worker"
    lifecycle: "deletion-test"
spec:
  class:
    kind: MachineClass
    name: %s
`, machineName, testNamespace, machineClassName)
			createAndTrackResource(ctx, "machine", machineName, testNamespace, machineYAML)

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
  serviceAccountKey: "{}"
  region: "eu01-1"
  userData: |
    #cloud-config
    runcmd:
      - echo "Machine bootstrapped"
`, secretName, testNamespace)
			createAndTrackResource(ctx, "secret", secretName, testNamespace, secretYAML)

			machineClassYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: MachineClass
metadata:
  name: %s
  namespace: %s
providerSpec:
  machineType: "c2i.2"
  imageId: "550e8400-e29b-41d4-a716-446655440000"
  networking:
    networkId: "770e8400-e29b-41d4-a716-446655440000"
  labels:
    application: "monitoring"
    team: "sre"
    cost-center: "operations"
    environment: "e2e-status-test"
    component: "metrics-collector"
    monitoring: "enabled"
    alerting: "critical"
secretRef:
  name: %s
  namespace: %s
provider: STACKIT
`, machineClassName, testNamespace, secretName, testNamespace)
			createAndTrackResource(ctx, "machineclass", machineClassName, testNamespace, machineClassYAML)

			machineYAML := fmt.Sprintf(`
apiVersion: machine.sapcloud.io/v1alpha1
kind: Machine
metadata:
  name: %s
  namespace: %s
  labels:
    test-type: "e2e"
    test-phase: "status"
    environment: "test"
    machine-role: "worker"
    monitoring: "enabled"
spec:
  class:
    kind: MachineClass
    name: %s
`, machineName, testNamespace, machineClassName)
			createAndTrackResource(ctx, "machine", machineName, testNamespace, machineYAML)

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
	})
})
