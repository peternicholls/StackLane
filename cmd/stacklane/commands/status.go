// stacklane status: print runtime state for the current project, or every
// recorded project when --all is passed. Drift is reported per FR-010.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/peternicholls/stacklane/core/state"
	"github.com/peternicholls/stacklane/infra/docker"
	"github.com/peternicholls/stacklane/observability/status"
)

func NewStatus(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show project status",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			one, err := r.One(ctx, cfg.Slug)
			if err != nil {
				return err
			}
			fmt.Fprint(os.Stdout, status.Render(one))
			return nil
		},
	}
}
