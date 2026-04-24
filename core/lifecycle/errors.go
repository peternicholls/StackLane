// Typed step errors for the lifecycle. Every error surfaced to the operator
// MUST be a StepError so the cobra layer can render the failing step name,
// the affected project, and a stated next action (FR-013).
package lifecycle

import (
	"errors"
	"fmt"
)

// StepError carries the context the operator needs to know what failed and
// what to try next.
type StepError struct {
	Step    string // e.g. "compose-up", "wait-healthy", "gateway-reload"
	Project string // affected project slug or "" for shared infra
	Cause   error  // underlying error
	Remedy  string // operator-facing next action
}

func (e *StepError) Error() string {
	if e == nil {
		return "<nil>"
	}
	switch {
	case e.Project != "" && e.Remedy != "":
		return fmt.Sprintf("step %s failed for project %s: %v\n  next: %s", e.Step, e.Project, e.Cause, e.Remedy)
	case e.Project != "":
		return fmt.Sprintf("step %s failed for project %s: %v", e.Step, e.Project, e.Cause)
	case e.Remedy != "":
		return fmt.Sprintf("step %s failed: %v\n  next: %s", e.Step, e.Cause, e.Remedy)
	default:
		return fmt.Sprintf("step %s failed: %v", e.Step, e.Cause)
	}
}

func (e *StepError) Unwrap() error { return e.Cause }

// Wrap returns a StepError. cause may be nil for "I have nothing concrete to
// add" but that is discouraged.
func Wrap(step, project string, cause error, remedy string) *StepError {
	return &StepError{Step: step, Project: project, Cause: cause, Remedy: remedy}
}

// AsStepError reports whether err is or wraps a *StepError.
func AsStepError(err error) (*StepError, bool) {
	var se *StepError
	if errors.As(err, &se) {
		return se, true
	}
	return nil, false
}

// ErrPortConflict is returned when a port collision is detected before any
// docker action runs. It is wrapped by a StepError at the call site.
var ErrPortConflict = errors.New("port conflict")
