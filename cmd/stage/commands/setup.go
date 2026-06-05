// stage setup: run ordered machine-readiness checks and one-time setup.
package commands

import (
	"fmt"

	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/spf13/cobra"
)

// validSuffixes lists the allowed values for the --suffix flag.
var validSuffixes = map[string]bool{
	"":        true, // empty = use stack default
	"develop": true,
	"dev":     true,
	"test":    true,
}

// setupFlags holds setup-specific CLI flags.
type setupFlags struct {
	Suffix         string
	NonInteractive bool
	NotUI          bool
	CLI            bool
	NoTUI          bool
	JSON           bool
}

func NewSetup(shared *SharedFlags) *cobra.Command {
	f := &setupFlags{}
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Check whether this computer is ready",
		Long:  "Checks whether this computer is ready to run StageServe. It verifies Docker, local DNS, ports, the StageServe state directory, and mkcert, then reports exact fixes without changing machine settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate --suffix value.
			if !validSuffixes[f.Suffix] {
				return fmt.Errorf("invalid --suffix value %q: must be one of 'develop', 'dev', 'test', or empty", f.Suffix)
			}

			mode := resolveOutputMode(f.JSON, plainTextOutputRequested(f.NotUI, f.CLI, f.NoTUI), f.NonInteractive)

			result, err := buildMachineReadinessResult(shared, f.Suffix)
			if err != nil {
				return err
			}

			projector := onboarding.NewProjector(mode, cmd.OutOrStdout(), onboarding.ProjectorOptions{
				Title:    "StageServe Setup",
				Detailed: true,
			})
			if err := projector.Project(result); err != nil {
				return err
			}

			// Return an exit-code-carrying error for non-zero exit codes.
			if result.ExitCode != 0 {
				return &setupExitError{code: int(result.ExitCode)}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Suffix, "suffix", "", "Site suffix (develop|dev|test)")
	cmd.Flags().BoolVar(&f.NonInteractive, "non-interactive", false, "Suppress prompts and use plain text output")
	addPlainTextOutputFlags(cmd, &f.NotUI, &f.CLI, &f.NoTUI)
	cmd.Flags().BoolVar(&f.JSON, "json", false, "Emit JSON envelope only")
	return cmd
}

// setupExitError wraps a non-zero exit code as an error so callers can
// distinguish "checks not ready" from "command failed".
type setupExitError struct{ code int }

func (e *setupExitError) Error() string {
	return fmt.Sprintf("setup finished with exit code %d", e.code)
}

func (e *setupExitError) ExitCode() int {
	return e.code
}

func (e *setupExitError) Silent() bool {
	return true
}
