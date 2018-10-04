package cmd

import (
	"github.com/spf13/cobra"
	"github.com/mumoshu/falco-operator/pkg/app"
)

var opts = app.OperateOpts{}

var Root = &cobra.Command{
	Use:   "falco-operator",
	Short: "Manages Sysdig Falco behavioral monitoring rules with Kubernetes Custom Resources",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Operate(opts)
	},
}

func init() {
	Root.Flags().StringVar(&opts.ConfigMapName, "configmap-name", "falco-operator", "the name of the configmap to which this operator writes the concatenated falco rules")
	Root.Flags().StringVarP(&opts.ConfigMapNamespace, "configmap-namespace", "n", "kube-system", "namespace in which falco and falco-operator are running")
	Root.Flags().StringVarP(&opts.WatchNamespace, "watch-namespace", "w", "", "namespaces on which the operator watches for changes")
}
