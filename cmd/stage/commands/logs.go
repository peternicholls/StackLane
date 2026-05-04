// stage logs: stream logs for one service.
package commands

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/peternicholls/stageserve/infra/docker"
	"github.com/peternicholls/stageserve/observability/logs"
)

func NewLogs(flags *SharedFlags) *cobra.Command {
	var (
		service string
		follow  bool
	)
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Stream container logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			s := &logs.Streamer{Docker: docker.NewSDKClient()}
			if service == "" {
				service = "nginx"
			}
			return s.Stream(ctx, cfg.ComposeProjectName, service, follow, os.Stdout)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "Compose service name (default: nginx)")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	return cmd
}
