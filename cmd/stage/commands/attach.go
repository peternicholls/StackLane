// stage attach / detach: route a project through the shared gateway or
// tear it down and remove its record.
package commands

import (
	"github.com/spf13/cobra"
)

func NewAttach(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "attach",
		Short: "Add this project to StageServe",
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

func NewDetach(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "detach",
		Short: "Remove this project from StageServe",
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
}
