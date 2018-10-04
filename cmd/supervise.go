package cmd

import (
	"github.com/spf13/cobra"
	"github.com/mumoshu/falco-operator/pkg/app"
	"time"
)

var supervise *cobra.Command

func init() {
	opts := app.SuperviseOpts{}

	supervise = &cobra.Command{
		Use:          "supervise",
		Short:        "Supervisor the targeted application and restart it on config changes",
		Long:         ``,
		RunE:         func(cmd *cobra.Command, args []string) error {
			return app.Supervise(args, opts)
		},
		SilenceUsage: true,
	}

	Root.AddCommand(supervise)
	supervise.Flags().DurationVar(&opts.RestarGracePeriod, "restart-grace-period", 15 * time.Second, "period of time until config changes finally trigger application restart. the timer is reset on another changes within the period")
	supervise.Flags().DurationVar(&opts.StopGracePeriod, "stop-grace-period", 5 * time.Second, "period of time until config changes trigger application restart")
	supervise.Flags().DurationVar(&opts.WatchInterval, "watch-interval", 5 * time.Second, "interval between the watch is triggered")
	supervise.Flags().StringSliceVar(&opts.Set, "set", []string{}, "set value in the file according to config changes")
	supervise.Flags().StringSliceVar(&opts.Watch, "watch", []string{}, "Watch a file for changes")
	supervise.Flags().BoolVar(&opts.Restart, "restart", true, "restart the application after config changes")
}
