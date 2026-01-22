package main

import (
	_ "github.com/gardener/machine-controller-manager/pkg/util/client/metrics/prometheus" // for client metric registration
	"github.com/gardener/machine-controller-manager/pkg/util/provider/app"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/app/options"
	_ "github.com/gardener/machine-controller-manager/pkg/util/reflector/prometheus" // for reflector metric registration
	_ "github.com/gardener/machine-controller-manager/pkg/util/workqueue/prometheus" // for workqueue metric registration
	"github.com/spf13/pflag"
	cp "github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/provider"
	"github.com/stackitcloud/machine-controller-manager-provider-stackit/pkg/spi"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog/v2"
)

func main() {
	s := options.NewMCServer()
	s.AddFlags(pflag.CommandLine)

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	provider := cp.NewProvider(&spi.PluginSPIImpl{})

	if err := app.Run(s, provider); err != nil {
		klog.Fatalf("failed to run application: %v", err)
	}
}
