// Package onboarding defines the shared domain model for setup, doctor, and
// init command results. Every output mode (text, JSON, TUI) projects from
// these types.
package onboarding

import (
	"io"
)

// Status is the per-step result status.
type Status string

const (
	StatusReady       Status = "ready"
	StatusNeedsAction Status = "needs_action"
	StatusError       Status = "error"
)

// OverallStatus is derived from the full step set.
type OverallStatus string

const (
	OverallReady       OverallStatus = "ready"
	OverallNeedsAction OverallStatus = "needs_action"
	OverallError       OverallStatus = "error"
)

// ExitCode mirrors the contract exit-code semantics.
type ExitCode int

const (
	ExitReady         ExitCode = 0
	ExitNeedsAction   ExitCode = 1
	ExitError         ExitCode = 2
	ExitUnsupportedOS ExitCode = 3
)

// StepResult is one ordered command step.
type StepResult struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`
	Status      Status         `json:"status"`
	Message     string         `json:"message"`
	Remediation *string        `json:"remediation"`
	Code        string         `json:"code,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

// CommandResult is the top-level envelope returned by every onboarding command.
type CommandResult struct {
	OverallStatus OverallStatus  `json:"overall_status"`
	ExitCode      ExitCode       `json:"exit_code"`
	Steps         []StepResult   `json:"steps"`
	Result        map[string]any `json:"result,omitempty"`
	NextSteps     []string       `json:"next_steps,omitempty"`
}

// OutputMode selects the output projection adapter.
type OutputMode int

const (
	OutputModeAuto OutputMode = iota // detect TTY; fall back to text
	OutputModeTUI                    // force Bubble Tea TUI
	OutputModeText                   // force plain text
	OutputModeJSON                   // emit JSON only; implies non-interactive
)

// remediationPtr returns a *string pointer to s, or nil when s is empty.
// Useful when building StepResult values with optional remediation.
func remediationPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Projector is the common interface for rendering CommandResult in any output mode.
type Projector interface {
	Project(r CommandResult) error
}

// NewProjector creates and returns a Projector for the given output mode.
func NewProjector(mode OutputMode, w io.Writer) Projector {
	switch mode {
	case OutputModeJSON:
		return &JSONProjector{W: w}
	case OutputModeTUI:
		return &TUIProjector{W: w}
	default:
		return &TextProjector{W: w}
	}
}
