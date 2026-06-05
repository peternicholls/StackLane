package onboarding

import "strings"

// splitSteps separates steps into those needing attention and those that passed.
func splitSteps(steps []StepResult) (attention []StepResult, ready []StepResult) {
	for _, step := range steps {
		if step.Status == StatusReady {
			ready = append(ready, step)
			continue
		}
		attention = append(attention, step)
	}
	return attention, ready
}

// statusLabel returns the human-readable label for a per-step status.
func statusLabel(status Status) string {
	switch status {
	case StatusReady:
		return "Ready"
	case StatusNeedsAction:
		return "Needs action"
	case StatusError:
		return "Error"
	default:
		return string(status)
	}
}

// followUpCommand returns the logical next command for the overall status.
func followUpCommand(status OverallStatus) string {
	if status == OverallReady {
		return "stage up"
	}
	return "stage doctor"
}

// plural returns sing when n == 1, plur otherwise.
func plural(n int, sing, plur string) string {
	if n == 1 {
		return sing
	}
	return plur
}

// sectionHeader returns a section title padded with ─ to fill ~40 columns.
// Example: "── Needs fixing ───────────────────────"
func sectionHeader(title string) string {
	const total = 40
	fill := total - 3 - len(title) - 1
	if fill < 2 {
		fill = 2
	}
	return "── " + title + " " + strings.Repeat("─", fill)
}

// cleanRemediation strips a leading "Run: " or "run: " prefix so the
// projector can display just the copy-pasteable command after "To fix:".
func cleanRemediation(s string) string {
	s = strings.TrimPrefix(s, "Run: ")
	s = strings.TrimPrefix(s, "run: ")
	return s
}

// checkDescription returns one sentence explaining why a check matters.
// Used in the detailed doctor/setup report to give context to each issue.
func checkDescription(id string) string {
	switch id {
	case "docker.binary":
		return "Docker CLI — the command-line tool used to manage containers."
	case "docker.daemon":
		return "The Docker daemon must be running before any container can start."
	case "state.dir":
		return "Stores StageServe runtime data: ports, certs, project registry."
	case "stack.shared_file":
		return "Shared runtime file required to start local URL routing."
	case "stack.project_file":
		return "Project runtime file required to run each project stack."
	case "port.80":
		return "Port 80 must be free for the local HTTP gateway to bind to it."
	case "port.443":
		return "Port 443 must be free for the local HTTPS gateway to bind to it."
	case "dns.resolver":
		return "Routes *.test domains to your stack — needs dnsmasq configured."
	case "mkcert.binary":
		return "Creates trusted local HTTPS certificates without browser warnings."
	default:
		return ""
	}
}

// compactMessage returns a short status string for use in compact summary tables.
func compactMessage(s StepResult) string {
	switch s.ID {
	case "state.dir":
		return "exists"
	case "stack.shared_file", "stack.project_file":
		return "available"
	case "port.80", "port.443":
		return "available"
	case "mkcert.binary":
		return "installed"
	case "dns.resolver":
		if s.Status == StatusReady {
			return "configured"
		}
	}
	return s.Message
}

// unused below — kept for compatibility with any external callers

type statusCounts struct {
	ready       int
	needsAction int
	errorCount  int
}

func countStatuses(steps []StepResult) statusCounts {
	counts := statusCounts{}
	for _, step := range steps {
		switch step.Status {
		case StatusReady:
			counts.ready++
		case StatusNeedsAction:
			counts.needsAction++
		case StatusError:
			counts.errorCount++
		}
	}
	return counts
}
