// stage attach: route a project through the shared gateway.
package commands

import (
	"github.com/spf13/cobra"
)

func NewAttach(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "attach",
		Short: "Add this project to StageServe",
		Long:  "Adds the current project back to StageServe so its local project URL works again without changing your project files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			if err := ensureProjectEnvFile(cfg, flags); err != nil {
				return err
			}
			orch, err := buildOrchestrator(cfg)
			if err != nil {
				return err
			}
			ctx, cancel := contextWithSignal(cmd.Context())
			defer cancel()
			return orch.Attach(ctx, cfg)
		},
	}
}
