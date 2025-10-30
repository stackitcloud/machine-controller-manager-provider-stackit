#!/usr/bin/env just --justfile

import 'vendir/just-pantry/lib/base.just'

# Import golang module for standard Go tasks
mod? golang 'vendir/just-pantry/modules/golang/module.just'

[private]
default: list

# Machine Controller Manager settings
export CONTROL_NAMESPACE := env_var_or_default('CONTROL_NAMESPACE', 'default')
export CONTROL_KUBECONFIG := env_var_or_default('CONTROL_KUBECONFIG', 'dev/target-kubeconfig.yaml')
export TARGET_KUBECONFIG := env_var_or_default('TARGET_KUBECONFIG', 'dev/target-kubeconfig.yaml')

# Build settings
IMAGE_REPOSITORY := env_var_or_default('IMAGE_REPOSITORY', 'localhost/machine-controller-manager-provider-stackit')
IMAGE_TAG := `cat VERSION`

# Kind cluster settings
KIND_CLUSTER_NAME := env_var_or_default('KIND_CLUSTER_NAME', 'mcm-provider-stackit')
KIND_CLUSTER_VERSION := env_var_or_default('KIND_CLUSTER_VERSION', 'v1.31.2')

# E2E test cluster settings (separate from dev cluster)
KIND_E2E_CLUSTER_NAME := env_var_or_default('KIND_E2E_CLUSTER_NAME', 'mcm-provider-stackit-e2e')
KIND_E2E_CLUSTER_VERSION := env_var_or_default('KIND_E2E_CLUSTER_VERSION', 'v1.31.2')
MCM_NAMESPACE := env_var_or_default('MCM_NAMESPACE', 'machine-controller-manager')

# ==============================================================================
# Build
# ==============================================================================

# Build the machine controller binary
[group('build')]
build:
    ./scripts/build.sh

# Build binary inside Docker (no local Go needed)
[group('build')]
build-docker:
    ./scripts/build-docker.sh

# Build Docker image
[group('build')]
docker-build:
    @echo "Building Docker image {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}..."
    docker build -t {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }} .

# Push Docker image to registry
[group('build')]
docker-push:
    @echo "Pushing Docker image {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}..."
    docker push {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}

# Clean all build artifacts
[group('build')]
clean:
    @echo "Cleaning build artifacts..."
    rm -rf build/
    rm -f *linux-amd64 *darwin-amd64
    rm -f *.out
    @echo "✓ Clean complete"

# ==============================================================================
# Development + Dev Cluster Management
# ==============================================================================

# Run the machine controller locally
[group('dev')]
start:
    @echo "Starting machine controller..."
    @echo "  CONTROL_KUBECONFIG: {{ CONTROL_KUBECONFIG }}"
    @echo "  TARGET_KUBECONFIG: {{ TARGET_KUBECONFIG }}"
    @echo "  CONTROL_NAMESPACE: {{ CONTROL_NAMESPACE }}"
    @echo ""
    GO111MODULE=on go run -mod=vendor cmd/machine-controller/main.go \
        --control-kubeconfig={{ CONTROL_KUBECONFIG }} \
        --target-kubeconfig={{ TARGET_KUBECONFIG }} \
        --namespace={{ CONTROL_NAMESPACE }} \
        --machine-creation-timeout=20m \
        --machine-drain-timeout=5m \
        --machine-health-timeout=10m \
        --machine-pv-detach-timeout=2m \
        --machine-safety-apiserver-statuscheck-timeout=30s \
        --machine-safety-apiserver-statuscheck-period=1m \
        --machine-safety-orphan-vms-period=30m \
        --v=3

# Create dev cluster if it doesn't exist
[group('dev')]
dev-create-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters 2>/dev/null | grep -qx "{{ KIND_CLUSTER_NAME }}"; then
      echo "Dev cluster '{{ KIND_CLUSTER_NAME }}' already exists; skipping create"
    else
      echo "Creating dev cluster '{{ KIND_CLUSTER_NAME }}'..."
      kind create cluster --name "{{ KIND_CLUSTER_NAME }}" --image "docker.io/kindest/node:{{ KIND_CLUSTER_VERSION }}"
    fi

# Delete dev cluster if it exists
[group('dev')]
dev-delete-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters 2>/dev/null | grep -qx "{{ KIND_CLUSTER_NAME }}"; then
      echo "Deleting dev cluster '{{ KIND_CLUSTER_NAME }}'..."
      kind delete cluster --name "{{ KIND_CLUSTER_NAME }}"
    else
      echo "Dev cluster '{{ KIND_CLUSTER_NAME }}' does not exist; nothing to delete"
    fi
    rm -rf dev/

# Load Docker image into dev cluster
[group('dev')]
dev-load-image:
    @echo "Loading image {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }} into dev cluster..."
    kind load docker-image "{{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}" --name "{{ KIND_CLUSTER_NAME }}"

# Export dev cluster kubeconfig (sets kubectl context)
[group('dev')]
dev-kubeconfig:
    kind export kubeconfig --name {{ KIND_CLUSTER_NAME }}

# Install MCM CRDs into the kind cluster
[group('dev')]
install-crds:
    @echo "Installing MCM CRDs..."
    kubectl apply -k config/crd/

# Save kind cluster kubeconfig to dev directory
[group('dev')]
save-kubeconfig:
    @echo "Saving kubeconfig to {{ TARGET_KUBECONFIG }}..."
    @mkdir -p dev
    kind get kubeconfig --name {{ KIND_CLUSTER_NAME }} > {{ TARGET_KUBECONFIG }}

# Set up local development environment with kind cluster
[group('dev')]
dev: docker-build \
     dev-create-cluster \
     save-kubeconfig \
     dev-load-image \
     dev-deploy
    @echo ""
    @echo "✓ Development environment ready!"
    @echo ""
    @echo "  Dev cluster: {{ KIND_CLUSTER_NAME }}"
    @echo "  Kubeconfig: {{ TARGET_KUBECONFIG }}"
    @echo "  Image loaded: {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}"
    @echo ""
    @echo "Check deployment status:"
    @echo "  kubectl get pods -n default"
    @echo ""
    @echo "For local debugging, run 'just start' instead of deploying to cluster"
    @echo ""

# Deploy MCM provider to development cluster
[group('dev')]
dev-deploy:
    @echo "Deploying MCM provider (development overlay)..."
    kubectl apply -k config/overlays/development

# ==============================================================================
# Testing
# ==============================================================================

# Run unit tests (excludes e2e tests which require a cluster)
# E2E tests should be run explicitly with test-e2e
[group('test')]
test:
    go test $(go list ./... | grep -v /e2e) -coverprofile cover.out

# ==============================================================================
# Dependencies
# ==============================================================================

# Update vendored dependencies
[group('deps')]
revendor: (recipe 'golang::mod-vendor') (recipe 'golang::mod-tidy')

# Update all dependencies to latest versions
[group('deps')]
update-deps:
    @echo "Updating dependencies to latest versions..."
    GO111MODULE=on go get -u
    @echo "Running go mod tidy..."
    GO111MODULE=on go mod tidy

# ==============================================================================
# E2E Test Cluster Management
# ==============================================================================

# Create e2e test cluster if it doesn't exist
[group('test')]
e2e-create-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters 2>/dev/null | grep -qx "{{ KIND_E2E_CLUSTER_NAME }}"; then
      echo "E2E cluster '{{ KIND_E2E_CLUSTER_NAME }}' already exists; skipping create"
    else
      echo "Creating E2E cluster '{{ KIND_E2E_CLUSTER_NAME }}'..."
      kind create cluster --name "{{ KIND_E2E_CLUSTER_NAME }}" --image "docker.io/kindest/node:{{ KIND_E2E_CLUSTER_VERSION }}"
    fi

# Delete e2e test cluster if it exists
[group('test')]
e2e-delete-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters 2>/dev/null | grep -qx "{{ KIND_E2E_CLUSTER_NAME }}"; then
      echo "Deleting E2E cluster '{{ KIND_E2E_CLUSTER_NAME }}'..."
      kind delete cluster --name "{{ KIND_E2E_CLUSTER_NAME }}"
    else
      echo "E2E cluster '{{ KIND_E2E_CLUSTER_NAME }}' does not exist; nothing to delete"
    fi

# Load Docker image into e2e test cluster
[group('test')]
e2e-load-image:
    @echo "Loading image {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }} into E2E cluster..."
    kind load docker-image "{{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}" --name "{{ KIND_E2E_CLUSTER_NAME }}"

# Run e2e tests in isolated kind cluster
# Usage: just test-e2e [focus]
# Example: just test-e2e "should create Machine"
[group('test')]
test-e2e focus="": docker-build e2e-create-cluster e2e-load-image && e2e-delete-cluster
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running e2e tests in isolated cluster {{ KIND_E2E_CLUSTER_NAME }}..."
    if [ -n "{{ focus }}" ]; then
        echo "Focus: {{ focus }}"
        KIND_CLUSTER_NAME={{ KIND_E2E_CLUSTER_NAME }} MCM_NAMESPACE={{ MCM_NAMESPACE }} go test ./test/e2e/... -v -ginkgo.v -ginkgo.focus="{{ focus }}" -timeout=15m
    else
        KIND_CLUSTER_NAME={{ KIND_E2E_CLUSTER_NAME }} MCM_NAMESPACE={{ MCM_NAMESPACE }} go test ./test/e2e/... -v -ginkgo.v -timeout=15m
    fi

# Run e2e tests and preserve cluster and resources for debugging
# Usage: just test-e2e-preserve [focus]
# Example: just test-e2e-preserve "negative test"
[group('test')]
test-e2e-preserve focus="": docker-build e2e-create-cluster e2e-load-image
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Running e2e tests (cluster and resources will be preserved)..."
    if [ -n "{{ focus }}" ]; then
        echo "Focus: {{ focus }}"
        KIND_CLUSTER_NAME={{ KIND_E2E_CLUSTER_NAME }} MCM_NAMESPACE={{ MCM_NAMESPACE }} SKIP_CLUSTER_CLEANUP=true SKIP_RESOURCE_CLEANUP=true go test ./test/e2e/... -v -ginkgo.v -ginkgo.focus="{{ focus }}" -timeout=15m
    else
        KIND_CLUSTER_NAME={{ KIND_E2E_CLUSTER_NAME }} MCM_NAMESPACE={{ MCM_NAMESPACE }} SKIP_CLUSTER_CLEANUP=true SKIP_RESOURCE_CLEANUP=true go test ./test/e2e/... -v -ginkgo.v -timeout=15m
    fi
    echo ""
    echo "E2E cluster '{{ KIND_E2E_CLUSTER_NAME }}' and test resources preserved for debugging."
    echo "MCM namespace: {{ MCM_NAMESPACE }}"
    echo "To inspect resources: kubectl get machines,secrets,machineclasses -n {{ MCM_NAMESPACE }}"
    echo "To clean up cluster: just e2e-delete-cluster"

# Export dev cluster kubeconfig (sets kubectl context)
[group('test')]
e2e-kubeconfig:
    kind export kubeconfig --name {{ KIND_E2E_CLUSTER_NAME }}
