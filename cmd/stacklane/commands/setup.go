// stacklane setup: run ordered machine-readiness checks and one-time setup.
package commands

import (
	"fmt"

	"github.com/peternicholls/stacklane/core/onboarding"
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
	TUI            bool
	NoTUI          bool
	JSON           bool
}

func NewSetup(shared *SharedFlags) *cobra.Command {
	f := &setupFlags{}
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Run machine-readiness checks and first-run setup",
		Long:  "Validates prerequisites (Docker, DNS, ports, state dir) and runs one-time setup steps. Reports each step as ready, needs_action, or error with exact remediation.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate --suffix value.
			if !validSuffixes[f.Suffix] {
				return fmt.Errorf("invalid --suffix value %q: must be one of 'develop', 'dev', 'test', or empty", f.Suffix)
			}

			mode := resolveOutputMode(f.JSON, f.NoTUI, f.TUI, f.NonInteractive)

			stateDir, err := resolveOnboardingStateDir(shared)
			if err != nil {
				return err
			}

			// Run ordered readiness checks.
			steps := []onboarding.StepResult{
				onboarding.CheckDockerBinary(""),
				onboarding.CheckDockerDaemon(),
				onboarding.CheckStateDir(stateDir),
				onboarding.CheckPort("port.80", 80),
				onboarding.CheckPort("port.443", 443),
				onboarding.CheckDNS(f.Suffix),
				onboarding.CheckMkcert(),
			}

			result := onboarding.BuildResult(steps, nil, nil)

			switch mode {
			case onboarding.OutputModeJSON:
				p := onboarding.JSONProjector{W: cmd.OutOrStdout()}
				if err := p.Project(result); err != nil {
					return err
				}
			case onboarding.OutputModeTUI:
				p := onboarding.TUIProjector{W: cmd.OutOrStdout()}
				p.Project(result)
			default:
				p := onboarding.TextProjector{W: cmd.OutOrStdout()}
				p.Project(result)
			}

			// Return an exit-code-carrying error for non-zero exit codes.
			if result.ExitCode != 0 {
				return &setupExitError{code: int(result.ExitCode)}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.Suffix, "suffix", "", "Site suffix (develop|dev|test)")
	cmd.Flags().BoolVar(&f.NonInteractive, "non-interactive", false, "Suppress prompts; implies --no-tui")
	cmd.Flags().BoolVar(&f.TUI, "tui", false, "Force TUI mode")
	cmd.Flags().BoolVar(&f.NoTUI, "no-tui", false, "Force plain-text output")
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
