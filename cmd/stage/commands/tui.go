package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
		settings := projectEnvSettingsFromGuidedAction(cfg, action)
		result, err := onboarding.WriteProjectEnvWithSettings(cfg.Dir, settings, false)
		if err != nil {
			return guidance.NextActionPlan{}, "", err
		}
		nextConfig := reloadGuidedConfig(cfg)
		context := guidance.Collect(ctx, nextConfig, guidance.CollectOptions{Capability: capability})
		message := projectEnvActionMessage(result, filepath.Join(cfg.Dir, ".env.stageserve"))
		return guidance.Plan(context), message, nil
	default:
		return guidance.Plan(guidance.Collect(ctx, cfg, guidance.CollectOptions{Capability: capability})), fmt.Sprintf("%s is not available in guided mode yet.", action.Label), nil
	}
}

func projectEnvSettingsFromGuidedAction(cfg config.ProjectConfig, action guidance.GuidedAction) onboarding.ProjectEnvSettings {
	return onboarding.ProjectEnvSettings{
		SiteName:   guidedActionInput(action, "site_name", cfg.Name),
		DocRoot:    normalizeProjectEnvDocRoot(guidedActionInput(action, "docroot", cfg.DocRootRelative)),
		SiteSuffix: normalizeProjectEnvSuffix(guidedActionInput(action, "site_suffix", cfg.SiteSuffix)),
	}
}

func guidedActionInput(action guidance.GuidedAction, key, fallback string) string {
	if action.Inputs == nil {
		return fallback
	}
	value := strings.TrimSpace(action.Inputs[key])
	if value == "" {
		return fallback
	}
	return value
}

func normalizeProjectEnvDocRoot(value string) string {
	value = strings.TrimSpace(value)
	if value == "." {
		return ""
	}
	return value
}

func normalizeProjectEnvSuffix(value string) string {
	return strings.TrimPrefix(strings.TrimSpace(value), ".")
}

func reloadGuidedConfig(cfg config.ProjectConfig) config.ProjectConfig {
	loader := config.NewLoader()
	loader.StackHomeOverride = cfg.StackHome
	nextConfig, err := loader.Load(cfg.Dir, config.CLIFlags{ProjectDir: cfg.Dir})
	if err != nil {
		return cfg
	}
	return nextConfig
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
