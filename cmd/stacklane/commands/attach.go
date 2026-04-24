// stacklane attach / detach: route or unroute a project at the shared gateway
// without changing whether its containers are running.
package commands

import (
	"github.com/spf13/cobra"
)

func NewAttach(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "attach",
		Short: "Route a running project through the shared gateway",
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
			return orch.Attach(ctx, cfg)
		},
	}
}

func NewDetach(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "detach",
		Short: "Stop routing a project at the shared gateway",
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
