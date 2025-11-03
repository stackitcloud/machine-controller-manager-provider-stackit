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
	"math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/aoepeople/machine-controller-manager-provider-stackit/test/utils"
)

// Timeout constants used across E2E tests
const (
	// QuickTimeout - for operations that should complete in seconds (resource existence checks)
	QuickTimeout = 30 * time.Second
	// StandardTimeout - for basic CRUD operations on Kubernetes resources
	StandardTimeout = 1 * time.Minute
	// MediumTimeout - for operations involving API calls with retries or resource deletion
	MediumTimeout = 2 * time.Minute
	// ContainerTimeout - for container startup and initialization
	ContainerTimeout = 3 * time.Minute

	// Polling intervals for Eventually/Consistently checks
	QuickPoll    = 2 * time.Second  // For fast-changing resources
	StandardPoll = 5 * time.Second  // For typical resource state changes
	MediumPoll   = 10 * time.Second // For slower cloud operations
	LongPoll     = 15 * time.Second // For very slow operations
	HelmPoll     = 20 * time.Second // For Helm chart operations
)

// TestResource represents a Kubernetes resource created during tests
type TestResource struct {
	Type      string
	Name      string
	Namespace string
}

var (
	// testResources tracks all resources created during tests for cleanup
	testResources []TestResource
	// testNamespace is the dedicated namespace for e2e tests
	testNamespace string
)

// validResourceTypes contains the resource types we expect to manage in E2E tests
// This helps catch typos and unexpected resource types during test development
var validResourceTypes = map[string]bool{
	"machine":           true,
	"machineclass":      true,
	"machineset":        true,
	"machinedeployment": true,
	"secret":            true,
	"pod":               true,
	"configmap":         true,
	"service":           true,
	"deployment":        true,
}

// generateRandomString generates a random alphanumeric string of the specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// generateResourceName creates a unique resource name with a random suffix
func generateResourceName(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, generateRandomString(8))
}

// trackResource adds a resource to the cleanup list
func trackResource(resourceType, resourceName, namespace string) {
	if !validResourceTypes[resourceType] {
		ginkgo.GinkgoWriter.Printf("Warning: tracking unexpected resource type '%s' (resource: %s/%s in namespace %s)\n",
			resourceType, resourceType, resourceName, namespace)
	}
	testResources = append(testResources, TestResource{
		Type:      resourceType,
		Name:      resourceName,
		Namespace: namespace,
	})
	ginkgo.GinkgoWriter.Printf("Tracked resource for cleanup: %s/%s in namespace %s\n", resourceType, resourceName, namespace)
}

// createAndTrackResource creates a Kubernetes resource and tracks it for cleanup
func createAndTrackResource(resourceType, resourceName, namespace, yamlContent string) {
	ginkgo.By(fmt.Sprintf("creating %s: %s", resourceType, resourceName))
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(yamlContent)
	_, err := utils.Run(cmd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf("Failed to create %s: %s", resourceType, resourceName))
	
	trackResource(resourceType, resourceName, namespace)
}

// verifyK8sResourceExists verifies that a Kubernetes resource exists
// This is a common pattern used across multiple E2E tests
// nolint:unused // Reserved for future E2E tests
func verifyK8sResourceExists(resourceType, resourceName, namespace string) {
	ginkgo.By(fmt.Sprintf("ensuring %s resource exists: %s", resourceType, resourceName))
	cmd := exec.Command("kubectl", "get", resourceType, resourceName, "-n", namespace)
	_, err := utils.Run(cmd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf("%s resource should exist before deletion test",
		resourceType))
}

// verifyK8sResourceDeleted verifies that a Kubernetes resource is deleted within the timeout
// This is a common pattern used across multiple E2E tests
func verifyK8sResourceDeleted(resourceType, resourceName, namespace string, timeout ...time.Duration) {
	deleteTimeout := MediumTimeout // Default timeout
	if len(timeout) > 0 {
		deleteTimeout = timeout[0]
	}

	ginkgo.By(fmt.Sprintf("verifying %s resource is removed: %s", resourceType, resourceName))
	gomega.Eventually(func() bool {
		cmd := exec.Command("kubectl", "get", resourceType, resourceName, "-n", namespace)
		_, err := utils.Run(cmd)
		return err != nil // Error means resource not found
	}, deleteTimeout, StandardPoll).Should(gomega.BeTrue(), fmt.Sprintf("%s resource should be deleted",
		resourceType))
}

// deleteK8sResource deletes a Kubernetes resource
// This is a common pattern used across multiple E2E tests
func deleteK8sResource(resourceType, resourceName, namespace string) {
	ginkgo.By(fmt.Sprintf("deleting the %s custom resource: %s", resourceType, resourceName))
	cmd := exec.Command("kubectl", "delete", resourceType, resourceName, "-n", namespace, "--wait=false")
	_, err := utils.Run(cmd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf("%s resource should be deleted successfully",
		resourceType))

	// For machines, remove finalizers if stuck (IAAS API may return errors preventing cleanup)
	if resourceType == "machine" {
		removeMachineFinalizers(resourceName, namespace)
	}
}

// removeMachineFinalizers removes finalizers from a machine to force cleanup
// This is needed because IAAS API may return errors (e.g., 400) preventing normal deletion
func removeMachineFinalizers(machineName, namespace string) {
	ginkgo.By(fmt.Sprintf("removing finalizers from machine: %s", machineName))
	cmd := exec.Command("kubectl", "patch", "machine", machineName, "-n", namespace,
		"--type=json", "-p=[{\"op\":\"remove\",\"path\":\"/metadata/finalizers\"}]")
	_, _ = utils.Run(cmd) // Ignore errors - machine may already be gone
}

// extractServerIDFromProviderID extracts the server ID from a STACKIT ProviderID
// Expected format: stackit://<projectId>/<serverId>
// Returns the serverId portion, or empty string if format is invalid
// nolint:unused // Reserved for future E2E tests
func extractServerIDFromProviderID(providerID string) string {
	// Expected format: stackit://12345678-1234-1234-1234-123456789012/497f6eca-6276-4993-bfeb-53cbbbba6f08
	const prefix = "stackit://"

	if !strings.HasPrefix(providerID, prefix) {
		return ""
	}

	// Remove prefix: "12345678-1234-1234-1234-123456789012/497f6eca-6276-4993-bfeb-53cbbbba6f08"
	remainder := strings.TrimPrefix(providerID, prefix)

	// Split by '/': ["12345678-1234-1234-1234-123456789012", "497f6eca-6276-4993-bfeb-53cbbbba6f08"]
	parts := strings.Split(remainder, "/")

	if len(parts) != 2 {
		return ""
	}

	// Return serverID (second part)
	return parts[1]
}
