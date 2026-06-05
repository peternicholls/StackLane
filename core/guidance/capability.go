package guidance

import (
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

// DetectCapability inspects only the real terminal and shell environment. It
// intentionally does not read .env.stageserve, because STAGESERVE_NO_TUI is a
// shell-level control.
func DetectCapability(stdin, stdout, stderr *os.File, notUIFlag, cliFlag bool) TUICapability {
	capability := TUICapability{
		NotUIFlag:     notUIFlag,
		CLIFlag:       cliFlag,
		NoTUIShellEnv: ShellEnvTruthy("STAGESERVE_NO_TUI"),
		NoColor:       os.Getenv("NO_COLOR") != "",
		Term:          os.Getenv("TERM"),
	}
	if stdin != nil {
		capability.StdinTTY = isTerminal(stdin)
	}
	if stdout != nil {
		capability.StdoutTTY = isTerminal(stdout)
	}
	if stderr != nil {
		capability.StderrTTY = isTerminal(stderr)
	}
	capability.Reason = capabilityReason(capability)
	return capability
}

func isTerminal(file *os.File) bool {
	return isatty.IsTerminal(file.Fd()) || isatty.IsCygwinTerminal(file.Fd())
}

func ShellEnvTruthy(key string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return false
	}
	switch strings.ToLower(value) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func capabilityReason(capability TUICapability) string {
	switch {
	case capability.NotUIFlag:
		return "plain text requested with --notui"
	case capability.CLIFlag:
		return "plain text requested with --cli"
	case capability.NoTUIShellEnv:
		return "plain text requested with STAGESERVE_NO_TUI"
	case !capability.StdinTTY || !capability.StdoutTTY:
		return "plain text output because this terminal is not interactive"
	default:
		return "interactive terminal detected"
	}
}
