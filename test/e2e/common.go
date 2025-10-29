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
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/aoepeople/machine-controller-manager-provider-stackit/test/utils"
)

// Timeout constants used across E2E tests
const (
	// Short timeouts for quick operations
	QuickTimeout     = 30 * time.Second
	StandardTimeout  = 1 * time.Minute
	MediumTimeout    = 2 * time.Minute
	ContainerTimeout = 3 * time.Minute

	// Polling intervals
	QuickPoll    = 2 * time.Second
	StandardPoll = 5 * time.Second
	MediumPoll   = 10 * time.Second
	LongPoll     = 15 * time.Second
	HelmPoll     = 20 * time.Second
)

// verifyK8sResourceExists verifies that a Kubernetes resource exists
// This is a common pattern used across multiple E2E tests
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
	cmd := exec.Command("kubectl", "delete", resourceType, resourceName, "-n", namespace)
	_, err := utils.Run(cmd)
	gomega.Expect(err).NotTo(gomega.HaveOccurred(), fmt.Sprintf("%s resource should be deleted successfully",
		resourceType))
}
