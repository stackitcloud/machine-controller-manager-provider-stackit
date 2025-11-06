module github.com/aoepeople/machine-controller-manager-provider-stackit

go 1.23.0

require (
	github.com/gardener/machine-controller-manager v0.36.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/ginkgo/v2 v2.27.2
	github.com/onsi/gomega v1.38.2
	github.com/spf13/pflag v1.0.5
	github.com/stackitcloud/stackit-sdk-go/core v0.18.0
	github.com/stackitcloud/stackit-sdk-go/services/iaas v1.0.0
	k8s.io/api v0.0.0-20190918155943-95b840bb6a1f
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090
	k8s.io/klog v0.4.0
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.2.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20180513044358-24b0969c4cb7 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/json-iterator/go v1.1.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910 // indirect
	github.com/prometheus/common v0.0.0-20181126121408-4724e9255275 // indirect
	github.com/prometheus/procfs v0.0.0-20181204211112-1dc9a6cbc91a // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	golang.org/x/tools v0.36.0 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	google.golang.org/protobuf v1.36.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad // indirect
	k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // indirect
	k8s.io/cluster-bootstrap v0.0.0-20190918163108-da9fdfce26bb // indirect
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf // indirect
	k8s.io/utils v0.0.0-20190801114015-581e00157fb1 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f // kubernetes-1.16.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655 // kubernetes-1.16.0
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad // kubernetes-1.16.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // kubernetes-1.16.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190918163108-da9fdfce26bb // kubernetes-1.16.0
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269 // kubernetes-1.16.0
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)
