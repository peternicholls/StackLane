package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
	"github.com/peternicholls/stageserve/core/onboarding"
)

func runGuidedTUI(ctx context.Context, cfg config.ProjectConfig, plan guidance.NextActionPlan, capability guidance.TUICapability, output io.Writer) error {
	handler := func(action guidance.GuidedAction) (guidance.NextActionPlan, string, error) {
		return handleGuidedAction(ctx, cfg, capability, action)
	}
	return guidance.RunShell(plan, capability, os.Stdin, output, guidance.WithActionHandler(handler))
}

func handleGuidedAction(ctx context.Context, cfg config.ProjectConfig, capability guidance.TUICapability, action guidance.GuidedAction) (guidance.NextActionPlan, string, error) {
	switch action.ID {
	case "init", "init_here":
		result, err := onboarding.WriteProjectEnv(cfg.Dir, cfg.Name, cfg.DocRootRelative, false)
		if err != nil {
			return guidance.NextActionPlan{}, "", err
		}
		context := guidance.Collect(ctx, cfg, guidance.CollectOptions{Capability: capability})
		message := projectEnvActionMessage(result, filepath.Join(cfg.Dir, ".env.stageserve"))
		return guidance.Plan(context), message, nil
	default:
		return guidance.Plan(guidance.Collect(ctx, cfg, guidance.CollectOptions{Capability: capability})), fmt.Sprintf("%s is not available in guided mode yet.", action.Label), nil
	}
}

func projectEnvActionMessage(action onboarding.InitAction, path string) string {
	switch action {
	case onboarding.InitActionCreated:
		return "Created project settings at " + path + "."
	case onboarding.InitActionSkipped:
		return "Project settings already exist at " + path + "."
	case onboarding.InitActionOverwritten:
		return "Updated project settings at " + path + "."
	default:
		return "Project settings checked at " + path + "."
	}
}
