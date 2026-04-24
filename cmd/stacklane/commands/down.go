// stacklane down: stop a project's stack.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewDown(flags *SharedFlags) *cobra.Command {
	var removeVolumes bool
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Stop a project's stack",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			if flags.DryRun {
				fmt.Fprintf(os.Stdout, "DRY RUN: would down project %s\n", cfg.Slug)
				return nil
			}
			orch, err := buildOrchestrator(cfg)
			if err != nil {
				return err
			}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			return orch.Down(ctx, cfg, removeVolumes)
		},
	}
	cmd.Flags().BoolVarP(&removeVolumes, "volumes", "v", false, "Remove named volumes")
	return cmd
}
