// Package onboarding provides the shared step-execution and result-derivation
// runtime for setup, doctor, and init commands.
package onboarding

// ReduceExitCode derives the canonical exit code from a set of step results.
// Precedence: ExitUnsupportedOS (3) > ExitError (2) > ExitNeedsAction (1) > ExitReady (0).
func ReduceExitCode(steps []StepResult) ExitCode {
	code := ExitReady
	for _, s := range steps {
		if s.Code == "unsupported-os" {
			return ExitUnsupportedOS
		}
		switch s.Status {
		case StatusError:
			if code < ExitError {
				code = ExitError
			}
		case StatusNeedsAction:
			if code < ExitNeedsAction {
				code = ExitNeedsAction
			}
		}
	}
	return code
}

// DeriveOverallStatus computes the overall status from a step slice.
// error dominates needs_action dominates ready.
func DeriveOverallStatus(steps []StepResult) OverallStatus {
	overall := OverallReady
	for _, s := range steps {
		switch s.Status {
		case StatusError:
			return OverallError
		case StatusNeedsAction:
			overall = OverallNeedsAction
		}
	}
	return overall
}

// BuildResult assembles a CommandResult from an ordered step slice, an
// optional result payload, and an optional next-steps slice.
func BuildResult(steps []StepResult, result map[string]any, nextSteps []string) CommandResult {
	if steps == nil {
		steps = []StepResult{}
	}
	return CommandResult{
		OverallStatus: DeriveOverallStatus(steps),
		ExitCode:      ReduceExitCode(steps),
		Steps:         steps,
		Result:        result,
		NextSteps:     nextSteps,
	}
}
