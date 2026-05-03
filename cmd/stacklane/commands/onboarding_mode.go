package commands

import (
	"os"

	"github.com/mattn/go-isatty"

	"github.com/peternicholls/stacklane/core/onboarding"
)

// resolveOutputMode maps the shared onboarding CLI flags to an OutputMode.
// Priority (highest wins):
//  1. --json     → OutputModeJSON
//  2. --no-tui   → OutputModeText
//  3. --tui      → OutputModeTUI
//  4. --non-interactive → OutputModeText
//  5. TTY detected → OutputModeTUI (auto)
//  6. no TTY → OutputModeText
func resolveOutputMode(json, noTUI, forceTUI, nonInteractive bool) onboarding.OutputMode {
	switch {
	case json:
		return onboarding.OutputModeJSON
	case noTUI:
		return onboarding.OutputModeText
	case forceTUI:
		return onboarding.OutputModeTUI
	case nonInteractive:
		return onboarding.OutputModeText
	case isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()):
		return onboarding.OutputModeTUI
	default:
		return onboarding.OutputModeText
	}
}
