// stage logs: stream logs for one service.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/docker"
	"github.com/peternicholls/stageserve/observability/logs"
)

func NewLogs(flags *SharedFlags) *cobra.Command {
	var (
		service         string
		projectSelector string
		follow          bool
	)
	cmd := &cobra.Command{
		Use:   "logs [service]",
		Short: "View project logs",
		Long:  "Streams logs for the current project by default. Use --project to select a recorded project and either [service] or --service to choose a service.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceName, err := resolveLogService(service, args)
			if err != nil {
				return err
			}
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			if projectSelector != "" {
				store, err := state.NewStore(cfg.StateDir)
				if err != nil {
					return err
				}
				rec, _, err := store.StateFileForSelector(projectSelector)
				if err != nil {
					return fmt.Errorf("logs: project %q not found: %w", projectSelector, err)
				}
				cfg = rec.Project
			}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			s := &logs.Streamer{Docker: docker.NewSDKClient()}
			return s.Stream(ctx, cfg.ComposeProjectName, serviceName, follow, os.Stdout)
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "Compose service name (default: nginx)")
	cmd.Flags().StringVar(&projectSelector, "project", "", "Recorded project selector (slug, name, hostname, or path)")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	return cmd
}

func resolveLogService(flagValue string, args []string) (string, error) {
	service := flagValue
	if len(args) > 0 {
		if service != "" && service != args[0] {
			return "", fmt.Errorf("logs: use either [service] or --service, not both")
		}
		service = args[0]
	}
	if service == "" {
		return "nginx", nil
	}
	return service, nil
}
