import '../golang/vars.just'

export KUBERNETES_VERSION := env("KUBERNETES_VERSION", `go list -m -f "{{ .Version }}" k8s.io/api 2>/dev/null | awk -F'[v.]' '{ printf "1.%d", $3 }' || echo ""`)
export ENVTEST_VERSION := env("ENVTEST_VERSION", `go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime 2>/dev/null | awk -F'[v.]' '{ printf "release-%d.%d", $2, $3 }' || echo ""`)
export ENVTEST_K8S_VERSION := env("ENVTEST_K8S_VERSION", KUBERNETES_VERSION)
export KIND_CLUSTER_NAME := env("KIND_CLUSTER_NAME", go-mod-short)
export KIND_CLUSTER_VERSION := env("KIND_CLUSTER_VERSION", `go list -m -f "{{ .Version }}" k8s.io/api 2>/dev/null | sed -E 's/^v[0-9]+/v1/' || echo ""`)
export IMAGE_REGISTRY := env("IMAGE_REGISTRY", "ghcr.io")
export IMAGE_REPOSITORY := env("IMAGE_REPOSITORY", go-mod-path)
export IMAGE_NAME := env("IMAGE_NAME", "controller")
export IMAGE_TAG := env("IMAGE_TAG", "latest")
export IMAGE := env("IMAGE", IMAGE_REGISTRY + "/" + IMAGE_REPOSITORY + "/" + IMAGE_NAME + ":" + IMAGE_TAG)
export IMAGE_CACHE := env("IMAGE_CACHE", IMAGE_REGISTRY + "/" + IMAGE_REPOSITORY + "/" + IMAGE_NAME + ":cache")
