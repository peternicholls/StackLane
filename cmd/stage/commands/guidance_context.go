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

	items := []guidance.WorkItem{}
	addItem := func(label, description string) {
		items = append(items, guidance.WorkItem{
			Label:         label,
			Status:        "needs attention",
			Description:   description,
			DirectCommand: "stage setup",
		})
	}

	if _, err := os.Stat(cfg.StateDir); os.IsNotExist(err) {
		addItem("Set up this computer", "StageServe is not set up on this computer yet. Run stage setup to check and fix each requirement.")
	}
	if _, err := os.Stat(cfg.SharedFile); os.IsNotExist(err) {
		addItem("Restore shared runtime file", "StageServe could not find the shared runtime file for this stack home.")
	}
	if _, err := os.Stat(cfg.StackFile); os.IsNotExist(err) {
		addItem("Restore project runtime file", "StageServe could not find the project runtime file for this stack home.")
	}

	if len(items) == 0 {
		return guidance.MachineReadinessSummary{}
	}

	return guidance.MachineReadinessSummary{
		Checked:      true,
		Blocked:      true,
		Status:       "needs_action",
		NextFixLabel: items[0].Label,
		NextCommand:  "stage setup",
		WorkItems:    items,
	}
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
