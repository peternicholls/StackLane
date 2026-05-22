package commands

import (
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/peternicholls/stageserve/core/onboarding"
)

// resolveOutputMode maps the shared onboarding CLI flags to an OutputMode.
// Priority (highest wins):
//  1. --json     -> OutputModeJSON
//  2. --notui / --cli / legacy --no-tui -> OutputModeText
//  3. --non-interactive or STAGESERVE_NO_TUI -> OutputModeText
//  4. TTY detected -> OutputModeTUI (auto)
//  5. no TTY -> OutputModeText
func resolveOutputMode(json, plainText, nonInteractive bool) onboarding.OutputMode {
	switch {
	case json:
		return onboarding.OutputModeJSON
	case plainText:
		return onboarding.OutputModeText
	case nonInteractive || tuiDisabledByEnv():
		return onboarding.OutputModeText
	case isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()):
		return onboarding.OutputModeTUI
	default:
		return onboarding.OutputModeText
	}
}

func addPlainTextOutputFlags(cmd *cobra.Command, notUI, cli, legacyNoTUI *bool) {
	cmd.Flags().BoolVar(notUI, "notui", false, "Use plain text output")
	cmd.Flags().BoolVar(cli, "cli", false, "Use plain text output")
	cmd.Flags().BoolVar(legacyNoTUI, "no-tui", false, "Use plain text output")
	_ = cmd.Flags().MarkHidden("no-tui")
}

func plainTextOutputRequested(notUI, cli, legacyNoTUI bool) bool {
	return notUI || cli || legacyNoTUI
}

func tuiDisabledByEnv() bool {
	v := strings.TrimSpace(os.Getenv("STAGESERVE_NO_TUI"))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}
