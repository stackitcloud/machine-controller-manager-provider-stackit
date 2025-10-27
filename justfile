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
IMAGE_REPOSITORY := env_var_or_default('IMAGE_REPOSITORY', '<link-to-image-repo>')
IMAGE_TAG := `cat VERSION`

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
