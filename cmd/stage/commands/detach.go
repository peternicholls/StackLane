// stage detach: stop tracking a project in StageServe.
package commands

import "github.com/spf13/cobra"

func NewDetach(flags *SharedFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Remove this project from StageServe",
		Long:  "Stops the current project if needed, removes its StageServe record, and clears its local URL route without touching your project files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			orch, err := buildOrchestrator(cfg)
			if err != nil {
				return err
			}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			return orch.Detach(ctx, cfg)
		},
	}
	return cmd
}
