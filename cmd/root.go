package cmd

import (
	"github.com/spf13/cobra"
	"github.com/mumoshu/falco-operator/pkg/app"
)

var Root = &cobra.Command{
	Use:   "falco-operator",
	Short: "Manages Sysdig Falco behavioral monitoring rules with Kubernetes Custom Resources",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Operate()
	},
}
