package commands

import (
	"context"

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
