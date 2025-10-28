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
# Development
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

# ==============================================================================
# Kind Cluster Helpers
# ==============================================================================

# Create kind cluster if it doesn't exist
[group('kind')]
kind-create-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters 2>/dev/null | grep -qx "{{ KIND_CLUSTER_NAME }}"; then
      echo "kind cluster '{{ KIND_CLUSTER_NAME }}' already exists; skipping create"
    else
      echo "Creating kind cluster '{{ KIND_CLUSTER_NAME }}'..."
      kind create cluster --name "{{ KIND_CLUSTER_NAME }}" --image "docker.io/kindest/node:{{ KIND_CLUSTER_VERSION }}"
    fi

# Delete kind cluster if it exists
[group('kind')]
kind-delete-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters 2>/dev/null | grep -qx "{{ KIND_CLUSTER_NAME }}"; then
      echo "Deleting kind cluster '{{ KIND_CLUSTER_NAME }}'..."
      kind delete cluster --name "{{ KIND_CLUSTER_NAME }}"
    else
      echo "kind cluster '{{ KIND_CLUSTER_NAME }}' does not exist; nothing to delete"
    fi

# Load Docker image into kind cluster
[group('kind')]
kind-load-image:
    @echo "Loading image {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }} into kind cluster..."
    kind load docker-image "{{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}" --name "{{ KIND_CLUSTER_NAME }}"

# Export kind cluster kubeconfig (sets kubectl context)
[group('kind')]
kind-kubeconfig:
    kind export kubeconfig --name {{ KIND_CLUSTER_NAME }}

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
# Local Development with Kind
# ==============================================================================

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
     kind-create-cluster \
     save-kubeconfig \
     install-crds \
     kind-load-image
    @echo ""
    @echo "✓ Development environment ready!"
    @echo ""
    @echo "  Kind cluster: {{ KIND_CLUSTER_NAME }}"
    @echo "  Kubeconfig: {{ TARGET_KUBECONFIG }}"
    @echo "  Image loaded: {{ IMAGE_REPOSITORY }}:{{ IMAGE_TAG }}"
    @echo ""
    @echo "Next steps:"
    @echo "  1. Deploy sample resources from kubernetes/ directory"
    @echo "  2. Run 'just start' to start the provider controller"
    @echo ""

# Clean up the local development environment
[group('dev')]
dev-clean: kind-delete-cluster
    @echo "Cleaning up dev directory..."
    rm -rf dev/
    @echo "✓ Development environment cleaned up"
