package guidance

import "github.com/peternicholls/stageserve/core/onboarding"

// MachineReadinessFromResult maps existing onboarding readiness semantics into
// the compact shape the guided planner consumes.
func MachineReadinessFromResult(result onboarding.CommandResult) MachineReadinessSummary {
	summary := MachineReadinessSummary{
		Checked: true,
		Status:  string(result.OverallStatus),
	}
	for _, step := range result.Steps {
		item := WorkItem{
			Label:       step.Label,
			Status:      readinessStatus(step.Status),
			Description: step.Message,
		}
		if step.Status != onboarding.StatusReady && !summary.Blocked {
			summary.Blocked = true
			summary.NextFixLabel = step.Label
			summary.NextCommand = "stage setup"
			item.DirectCommand = "stage setup"
		}
		summary.WorkItems = append(summary.WorkItems, item)
	}
	if summary.Blocked && summary.NextFixLabel == "" {
		summary.NextFixLabel = "Run setup checks"
		summary.NextCommand = "stage setup"
	}
	return summary
}

func readinessStatus(status onboarding.Status) string {
	switch status {
	case onboarding.StatusReady:
		return "ready"
	case onboarding.StatusError:
		return "error"
	default:
		return "needs attention"
	}
}
