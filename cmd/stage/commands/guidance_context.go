package commands

import (
	"context"
	"os"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/docker"
)

func collectGuidedContext(ctx context.Context, cfg config.ProjectConfig, capability guidance.TUICapability) guidance.GuidedContext {
	return guidance.Collect(ctx, cfg, guidance.CollectOptions{
		Capability:    capability,
		RuntimeStatus: guidedRuntimeStatus,
	})
}

func collectRootGuidedContext(ctx context.Context, cfg config.ProjectConfig, capability guidance.TUICapability) guidance.GuidedContext {
	return guidance.Collect(ctx, cfg, guidance.CollectOptions{
		Capability:       capability,
		MachineReadiness: cheapReadinessHeuristic,
		RuntimeStatus:    guidedRuntimeStatus,
	})
}

// cheapReadinessHeuristic avoids running slow Docker/DNS/port probes on every
// bare `stage` invocation. It only checks whether the StageServe state
// directory exists — a fast filesystem stat. If the directory is missing the
// machine is almost certainly not set up yet. If it exists, the machine is
// assumed healthy for the initial guided render; the user can press Enter or
// run `stage setup` / 'd' to trigger the full diagnostics.
func cheapReadinessHeuristic(_ context.Context, cfg config.ProjectConfig) guidance.MachineReadinessSummary {
	if cfg.StateDir == "" {
		return guidance.MachineReadinessSummary{}
	}
	if _, err := os.Stat(cfg.StateDir); os.IsNotExist(err) {
		return guidance.MachineReadinessSummary{
			Checked:      true,
			Blocked:      true,
			Status:       "needs_action",
			NextFixLabel: "Set up this computer",
			NextCommand:  "stage setup",
			WorkItems: []guidance.WorkItem{
				{
					Label:         "Set up this computer",
					Status:        "needs attention",
					Description:   "StageServe is not set up on this computer yet. Run stage setup to check and fix each requirement.",
					DirectCommand: "stage setup",
				},
			},
		}
	}
	return guidance.MachineReadinessSummary{}
}

func guidedRuntimeStatus(ctx context.Context, cfg config.ProjectConfig, record *state.Record) guidance.RuntimeSummary {
	if record == nil || record.AttachmentState != state.StateDown {
		return guidance.RuntimeSummary{}
	}

	containers, err := docker.NewSDKClient().ListContainersByLabel(ctx, map[string]string{
		"com.docker.compose.project": cfg.ComposeProjectName,
	})
	if err != nil {
		return guidance.RuntimeSummary{}
	}

	summary := guidance.RuntimeSummary{
		Checked: true,
		Running: len(containers) > 0,
		Status:  "stopped",
	}
	if summary.Running {
		summary.Status = "running"
	}
	return summary
}
