// stage status: print runtime state for the current project, or every
// recorded project when --all is passed. Drift is reported per FR-010.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/docker"
	"github.com/peternicholls/stageserve/observability/status"
)

func NewStatus(flags *SharedFlags) *cobra.Command {
	var projectSelector string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show project status",
		Long:  "Shows the current project by default, every recorded project with --all, or one recorded project selected by slug, name, hostname, or path with --project.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.All && projectSelector != "" {
				return fmt.Errorf("status: use either --all or --project, not both")
			}
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			store, err := state.NewStore(cfg.StateDir)
			if err != nil {
				return err
			}
			r := &status.Reporter{State: store, Docker: docker.NewSDKClient()}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			if flags.All {
				results, err := r.All(ctx)
				if err != nil {
					return err
				}
				if len(results) == 0 {
					fmt.Println("no projects recorded")
					return nil
				}
				for _, p := range results {
					fmt.Fprint(os.Stdout, status.Render(p))
				}
				return nil
			}
			var one status.ProjectStatus
			if projectSelector != "" {
				one, err = r.OneBySelector(ctx, projectSelector)
			} else {
				one, err = r.One(ctx, cfg.Slug)
			}
			if err != nil {
				return err
			}
			fmt.Fprint(os.Stdout, status.Render(one))
			return nil
		},
	}
	cmd.Flags().StringVar(&projectSelector, "project", "", "Recorded project selector (slug, name, hostname, or path)")
	return cmd
}
