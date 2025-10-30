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
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestE2E runs the end-to-end (e2e) test suite for the MCM provider.
// These tests execute in an isolated kind cluster to validate the provider
// with mock STACKIT IAAS API endpoints.
func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting MCM provider STACKIT integration test suite\n")
	RunSpecs(t, "e2e suite")
}

var _ = BeforeSuite(func() {
	var (
		cmd    *exec.Cmd
		err    error
		output []byte
	)

	// Get the cluster name from environment (set by just recipe)
	clusterName := os.Getenv("KIND_CLUSTER_NAME")
	if clusterName == "" {
		clusterName = "mcm-provider-stackit-e2e" // fallback default
	}
	_, _ = fmt.Fprintf(GinkgoWriter, "Using kind cluster: %s\n", clusterName)

	By("determining test namespace")
	// Use MCM_NAMESPACE env var or default to machine-controller-manager
	testNamespace = os.Getenv("MCM_NAMESPACE")
	if testNamespace == "" {
		testNamespace = "machine-controller-manager"
	}
	_, _ = fmt.Fprintf(GinkgoWriter, "Using test namespace: %s (set MCM_NAMESPACE to override)\n", testNamespace)

	By("deploying MCM provider with mock IAAS API")
	cmd = exec.Command("kubectl", "apply", "-k", "../../config/overlays/e2e")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "Failed to deploy: %s\n", string(output))
	}
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "Failed to deploy MCM provider with mock API")

	By("waiting for deployments to be ready")
	// Wait for MCM deployment
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=300s",
		"deployment/machine-controller-manager", "-n", testNamespace)
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "MCM deployment not ready: %s\n", string(output))
	}
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "MCM deployment did not become ready")

	// Wait for IAAS mock server
	cmd = exec.Command("kubectl", "wait", "--for=condition=available", "--timeout=120s",
		"deployment/iaas", "-n", "stackitcloud")
	output, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "IAAS mock not ready: %s\n", string(output))
	}
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "IAAS mock server did not become ready")

	_, _ = fmt.Fprintf(GinkgoWriter, "E2E environment setup complete\n")
})

var _ = AfterSuite(func() {
	skipResourceCleanup := os.Getenv("SKIP_RESOURCE_CLEANUP") == "true"
	skipClusterCleanup := os.Getenv("SKIP_CLUSTER_CLEANUP") == "true"

	if skipResourceCleanup {
		_, _ = fmt.Fprintf(GinkgoWriter, "Skipping resource cleanup (SKIP_RESOURCE_CLEANUP=true)\n")
		_, _ = fmt.Fprintf(GinkgoWriter, "Test resources preserved in namespace: %s\n", testNamespace)
		_, _ = fmt.Fprintf(GinkgoWriter, "To list test resources: kubectl get machines,secrets,machineclasses -n %s\n", testNamespace)
		return
	}

	By("uninstalling MCM provider")
	cmd := exec.Command("kubectl", "delete", "-k", "../../config/overlays/e2e", "--ignore-not-found=true")
	output, _ := cmd.CombinedOutput()
	_, _ = fmt.Fprintf(GinkgoWriter, "MCM uninstall output: %s\n", string(output))

	By("cleaning up tracked test resources after MCM removal")
	for i := len(testResources) - 1; i >= 0; i-- {
		resource := testResources[i]
		// Wrap each cleanup in a function with GinkgoRecover to prevent one failure
		// from stopping cleanup of remaining resources
		func(res TestResource) {
			defer GinkgoRecover()
			deleteK8sResource(res.Type, res.Name, res.Namespace)
			verifyK8sResourceDeleted(res.Type, res.Name, res.Namespace)
		}(resource)
	}

	// Note: We don't delete the namespace as it's managed by MCM deployment, not the tests

	if !skipClusterCleanup {
		_, _ = fmt.Fprintf(GinkgoWriter, "E2E cleanup complete. Set SKIP_CLUSTER_CLEANUP=true to preserve cluster.\n")
	}
})
