// stacklane doctor: read-only machine diagnostics and drift detection.
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/peternicholls/stacklane/core/onboarding"
	"github.com/spf13/cobra"
)

// doctorFlags holds doctor-specific CLI flags.
type doctorFlags struct {
	JSON           bool
	NonInteractive bool
	NoTUI          bool
}

func NewDoctor(shared *SharedFlags) *cobra.Command {
	f := &doctorFlags{}
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose machine-readiness drift and suggest targeted fixes",
		Long:  "Read-only diagnostics: checks Docker, DNS, ports, state dir, and shared gateway. Reports ready/needs_action/error with exact remediation. Does not mutate machine state.",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := resolveOutputMode(f.JSON, f.NoTUI, false, f.NonInteractive)

			stateDir := os.Getenv("STACKLANE_STATE_DIR")
			if stateDir == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf("cannot determine home directory: %w", err)
				}
				stateDir = filepath.Join(home, ".stacklane-state")
			}

			steps := []onboarding.StepResult{
				onboarding.CheckDockerBinary(""),
				onboarding.CheckDockerDaemon(),
				onboarding.CheckStateDir(stateDir),
				onboarding.CheckPort("port.80", 80),
				onboarding.CheckPort("port.443", 443),
				onboarding.CheckDNS(""),
				onboarding.CheckMkcert(),
			}

			result := onboarding.BuildResult(steps, nil, nil)

			switch mode {
			case onboarding.OutputModeJSON:
				p := onboarding.JSONProjector{W: cmd.OutOrStdout()}
				return p.Project(result)
			case onboarding.OutputModeTUI:
				p := onboarding.TUIProjector{W: cmd.OutOrStdout()}
				p.Project(result)
			default:
				p := onboarding.TextProjector{W: cmd.OutOrStdout()}
				p.Project(result)
			}

			if result.ExitCode != 0 {
				return &doctorExitError{code: int(result.ExitCode)}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&f.JSON, "json", false, "Emit JSON envelope only")
	cmd.Flags().BoolVar(&f.NonInteractive, "non-interactive", false, "Suppress interactive prompts")
	cmd.Flags().BoolVar(&f.NoTUI, "no-tui", false, "Force plain-text output")
	return cmd
}

// doctorExitError wraps a non-zero readiness exit code.
type doctorExitError struct{ code int }

func (e *doctorExitError) Error() string {
	return fmt.Sprintf("doctor finished with exit code %d", e.code)
}
