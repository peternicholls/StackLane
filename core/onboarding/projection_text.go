package onboarding

import (
	"fmt"
	"io"
	"strings"
)

// TextProjector writes a CommandResult to w in human-readable plain text.
type TextProjector struct {
	W io.Writer
}

// Project renders the result as plain text.
func (p *TextProjector) Project(r CommandResult) error {
	for _, s := range r.Steps {
		icon := iconFor(s.Status)
		_, err := fmt.Fprintf(p.W, "%s [%s] %s — %s\n", icon, s.ID, s.Label, s.Message)
		if err != nil {
			return err
		}
		if s.Remediation != nil && *s.Remediation != "" {
			_, err := fmt.Fprintf(p.W, "   run: %s\n", *s.Remediation)
			if err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(p.W, "\nStatus: %s (exit %d)\n", r.OverallStatus, r.ExitCode)
	if err != nil {
		return err
	}
	if len(r.NextSteps) > 0 {
		_, err := fmt.Fprintf(p.W, "Next: %s\n", strings.Join(r.NextSteps, "  |  "))
		if err != nil {
			return err
		}
	}
	return nil
}

func iconFor(s Status) string {
	switch s {
	case StatusReady:
		return "✓"
	case StatusNeedsAction:
		return "!"
	case StatusError:
		return "✗"
	}
	return "?"
}
