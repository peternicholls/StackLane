// stage up: bring the project's stack online and attach it to the
// shared gateway.
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewUp(flags *SharedFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Bring a project's stack online",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig(flags)
			if err != nil {
				return err
			}
			if flags.DryRun {
				fmt.Fprintf(os.Stdout, "DRY RUN: would up project %s (%s) at %s\n", cfg.Slug, cfg.Hostname, cfg.Dir)
				return nil
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
			return orch.Up(ctx, cfg)
		},
	}
}

// contextWithSignal returns a cancellable context tied to the cobra context.
// (Real signal handling lives at the binary level; here we just propagate.)
func contextWithSignal(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithCancel(ctx)
}
