# Kubebuilder module just recipes
#
# This file provides common Go (golang) development tasks as just recipes.
#
# Usage examples:
#   mod? kb 'vendir/just-pantry/modules/kubebuilder/module.just'
#   just kb::generate
#
# For more information, see the project documentation, the just manual (https://just.systems/man/en/modules1190.html), or run `just --help`.

import '../../lib/module.just'
import '../golang/vars.just'
import 'vars.just'

golang-justfile := join(source_directory(), "../golang/module.just")

# Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
[group('kubebuilder')]
[no-cd]
manifests: (requires 'controller-gen')
    controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
[group('kubebuilder')]
[no-cd]
generate: (requires 'controller-gen')
    #!/usr/bin/env bash
    if [ -f hack/boilerplate.go.txt ]; then
        controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./...";
    else
        controller-gen object paths="./...";
    fi

# Run go fmt against code.
[group('kubebuilder')]
[no-cd]
fmt: (recipe-file golang-justfile 'format')

# Run go vet against code.
[group('kubebuilder')]
[no-cd]
vet: (recipe-file golang-justfile 'vet')

# Runs golangci-lint with default settings
[group('kubebuilder')]
[no-cd]
lint: (recipe-file golang-justfile 'lint')

# Runs golangci-lint with default settings
[group('kubebuilder')]
[no-cd]
lint-fix: (recipe-file golang-justfile 'lint-fix')

# Verify golangci-lint linter configuration
[group('kubebuilder')]
[no-cd]
lint-config: (recipe-file golang-justfile 'lint-config-verify')

# Run tests.
[group('kubebuilder')]
[no-cd]
test: (requires-vars 'ENVTEST_K8S_VERSION') (requires 'setup-envtest') manifests generate fmt vet setup-envtest
    KUBEBUILDER_ASSETS="$(setup-envtest use "{{ ENVTEST_K8S_VERSION }}" --bin-dir "{{ go-bin }}" -p path)" \
    go test $(go list ./... | grep -v /e2e) -coverprofile cover.out

# Download the binaries required for ENVTEST in the local bin directory.
[group('kubebuilder')]
[no-cd]
setup-envtest: (requires-vars 'ENVTEST_K8S_VERSION')
    @setup-envtest use "{{ ENVTEST_K8S_VERSION }}" --bin-dir "{{ go-bin }}" -p path || { \
      echo "Error: Failed to set up envtest binaries for version {{ ENVTEST_K8S_VERSION }}."; \
      exit 1; \
    }

# Build manager binary.
[group('kubebuilder')]
[no-cd]
build: manifests generate fmt vet
    go build -o bin/manager cmd/main.go

# Run a controller from your host.
[group('kubebuilder')]
[no-cd]
run: manifests generate fmt vet
    go run cmd/main.go

# Build docker image with the manager.
[group('kubebuilder')]
[no-cd]
docker-build file='./Dockerfile' *args='--load': (requires-vars 'IMAGE_NAME' 'IMAGE_TAG' 'IMAGE_REGISTRY' 'IMAGE_REPOSITORY') (requires 'docker' 'docker-buildx')
    #!/usr/bin/env bash
    docker buildx build . \
      --progress=plain \
      --cache-from=type=registry,ref={{ IMAGE_CACHE }} \
      --cache-to=type=inline \
      --tag={{ IMAGE }} \
      --tag={{ IMAGE_CACHE }} \
      --file="{{ file }}" \
      {{ args }}

# Push docker image with the manager.
[group('kubebuilder')]
[no-cd]
docker-push:
    docker push "{{ IMAGE }}"

# Install CRDs into the K8s cluster specified in ~/.kube/config.
[group('kubebuilder')]
[no-cd]
install: (requires 'kustomize' 'kubectl') manifests
    kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
[group('kubebuilder')]
[no-cd]
uninstall ingore-not-found="false": (requires 'kustomize' 'kubectl') manifests
    kustomize build config/crd | kubectl delete --ignore-not-found={{ ingore-not-found }} -f -

# Deploy controller in the K8s cluster specified in ~/.kube/config.
[group('kubebuilder')]
[no-cd]
deploy dir="config/default" *args='--server-side --force-conflicts': (requires 'kustomize' 'kubectl') manifests
    kustomize build {{ dir }} | kubectl apply {{ args }} -f -

# Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
[group('kubebuilder')]
[no-cd]
undeploy dir="config/default" ingore-not-found="false" *args='': (requires 'kustomize' 'kubectl')
    kustomize build {{ dir }} | kubectl delete {{ args }} --ignore-not-found={{ ingore-not-found }} -f -

# Set up a Kind cluster for e2e tests if it does not exist
[group('kubebuilder')]
setup-test-e2e: kind-create-cluster

# Tear down the Kind cluster used for e2e tests
[group('kubebuilder')]
cleanup-test-e2e: kind-delete-cluster

# Run the e2e tests. Expected an isolated environment using Kind.
[group('kubebuilder')]
[no-cd]
test-e2e: setup-test-e2e && cleanup-test-e2e
    KIND_CLUSTER="{{ KIND_CLUSTER_NAME }}" go test ./test/e2e/ -v -ginkgo.v

# Prints the image name that is used for the controller manager.
[group('kubebuilder')]
[no-cd]
image-name:
    @echo "{{ IMAGE }}"

# Load the docker image into the Kind cluster.
[group('kubebuilder')]
[no-cd]
kind-load-image: (requires-vars 'KIND_CLUSTER_NAME' 'IMAGE') (requires 'kind')
    kind load docker-image "{{ IMAGE }}" --name "{{ KIND_CLUSTER_NAME }}"

# Set up a Kind cluster for e2e tests if it does not exist
[group('kubebuilder')]
kind-create-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters | grep -qx "{{ KIND_CLUSTER_NAME }}"; then
      echo "kind cluster '{{ KIND_CLUSTER_NAME }}' already exists; skipping create"
    else
      kind create cluster --name "{{ KIND_CLUSTER_NAME }}" --image "docker.io/kindest/node:{{ KIND_CLUSTER_VERSION }}"
    fi

# Tear down the Kind cluster used for e2e tests (only if it exists)
[group('kubebuilder')]
kind-delete-cluster:
    #!/usr/bin/env bash
    set -euo pipefail
    if kind get clusters | grep -qx "{{ KIND_CLUSTER_NAME }}"; then
      kind delete cluster --name "{{ KIND_CLUSTER_NAME }}"
    else
      echo "kind cluster '{{ KIND_CLUSTER_NAME }}' does not exist; nothing to delete"
    fi
