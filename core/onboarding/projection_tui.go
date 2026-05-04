package onboarding

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TUIProjector renders a CommandResult using lipgloss styled output.
// It does not run a full Bubble Tea event loop; it produces styled text
// suitable for non-interactive TUI rendering in the setup/doctor/init flows.
type TUIProjector struct {
	W io.Writer
}

var (
	styleReady       = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // green
	styleNeedsAction = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
	styleDim         = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // grey
	styleBold        = lipgloss.NewStyle().Bold(true)
)

// Project renders the result with lipgloss styling.
func (p *TUIProjector) Project(r CommandResult) error {
	for _, s := range r.Steps {
		var styledStatus string
		switch s.Status {
		case StatusReady:
			styledStatus = styleReady.Render("✓ ready")
		case StatusNeedsAction:
			styledStatus = styleNeedsAction.Render("! needs_action")
		case StatusError:
			styledStatus = styleError.Render("✗ error")
		}
		_, err := fmt.Fprintf(p.W, "%s  %s\n", styledStatus, styleBold.Render(s.Label))
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(p.W, "   %s\n", s.Message)
		if err != nil {
			return err
		}
		if s.Remediation != nil && *s.Remediation != "" {
			_, err := fmt.Fprintf(p.W, "   %s\n", styleDim.Render("run: "+*s.Remediation))
			if err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(p.W, "\n%s %s (exit %d)\n",
		styleBold.Render("Status:"),
		strings.ToUpper(string(r.OverallStatus)),
		r.ExitCode,
	)
	if err != nil {
		return err
	}
	if len(r.NextSteps) > 0 {
		_, err := fmt.Fprintf(p.W, "%s %s\n",
			styleBold.Render("Next:"),
			strings.Join(r.NextSteps, "  |  "),
		)
		if err != nil {
			return err
		}
	}
	return nil
}
