// stage doctor: read-only machine diagnostics and drift detection.
package commands

import (
	"fmt"

	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/spf13/cobra"
)

// doctorFlags holds doctor-specific CLI flags.
type doctorFlags struct {
	JSON           bool
	NonInteractive bool
	NotUI          bool
	CLI            bool
	NoTUI          bool
}

func NewDoctor(shared *SharedFlags) *cobra.Command {
	f := &doctorFlags{}
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose local readiness issues",
		Long:  "Checks the same local requirements as stage setup and reports exact fixes without changing machine settings.",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := resolveOutputMode(f.JSON, plainTextOutputRequested(f.NotUI, f.CLI, f.NoTUI), f.NonInteractive)

			result, err := buildMachineReadinessResult(shared, "")
			if err != nil {
				return err
			}

			switch mode {
			case onboarding.OutputModeJSON:
				p := onboarding.JSONProjector{W: cmd.OutOrStdout()}
				if err := p.Project(result); err != nil {
					return err
				}
			case onboarding.OutputModeTUI:
				p := onboarding.TUIProjector{W: cmd.OutOrStdout(), Title: "StageServe Doctor", Detailed: true}
				p.Project(result)
			default:
				p := onboarding.TextProjector{W: cmd.OutOrStdout(), Title: "StageServe Doctor", Detailed: true}
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
	addPlainTextOutputFlags(cmd, &f.NotUI, &f.CLI, &f.NoTUI)
	return cmd
}

// doctorExitError wraps a non-zero readiness exit code.
type doctorExitError struct{ code int }

func (e *doctorExitError) Error() string {
	return fmt.Sprintf("doctor finished with exit code %d", e.code)
}

func (e *doctorExitError) ExitCode() int {
	return e.code
}

func (e *doctorExitError) Silent() bool {
	return true
}
