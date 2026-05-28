package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
	"github.com/peternicholls/stageserve/core/lifecycle"
	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/docker"
	obslogs "github.com/peternicholls/stageserve/observability/logs"
	obsstatus "github.com/peternicholls/stageserve/observability/status"
)

type guidedLifecycleRunner interface {
	Up(context.Context, config.ProjectConfig) error
	Attach(context.Context, config.ProjectConfig) error
	Down(context.Context, config.ProjectConfig, bool) error
	Detach(context.Context, config.ProjectConfig) error
	Status(context.Context, config.ProjectConfig) (string, error)
	Logs(context.Context, config.ProjectConfig, string) (string, error)
}

type guidedRuntimeRunner struct {
	orch *lifecycle.Orchestrator
}

func (r guidedRuntimeRunner) Up(ctx context.Context, cfg config.ProjectConfig) error {
	return r.orch.Up(ctx, cfg)
}

func (r guidedRuntimeRunner) Attach(ctx context.Context, cfg config.ProjectConfig) error {
	return r.orch.Attach(ctx, cfg)
}

func (r guidedRuntimeRunner) Down(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	return r.orch.Down(ctx, cfg, removeVolumes)
}

func (r guidedRuntimeRunner) Detach(ctx context.Context, cfg config.ProjectConfig) error {
	return r.orch.Detach(ctx, cfg)
}

func (r guidedRuntimeRunner) Status(ctx context.Context, cfg config.ProjectConfig) (string, error) {
	store, err := state.NewStore(cfg.StateDir)
	if err != nil {
		return "", err
	}
	reporter := &obsstatus.Reporter{State: store, Docker: docker.NewSDKClient()}
	projectStatus, err := reporter.One(ctx, cfg.Slug)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(obsstatus.Render(projectStatus)), nil
}

func (r guidedRuntimeRunner) Logs(ctx context.Context, cfg config.ProjectConfig, service string) (string, error) {
	var output bytes.Buffer
	streamer := &obslogs.Streamer{Docker: docker.NewSDKClient()}
	if err := streamer.Stream(ctx, cfg.ComposeProjectName, service, false, &output); err != nil {
		return "", err
	}
	body := strings.TrimSpace(output.String())
	if body == "" {
		return "No log output yet.", nil
	}
	return body, nil
}

func runGuidedTUI(ctx context.Context, cfg config.ProjectConfig, plan guidance.NextActionPlan, capability guidance.TUICapability, output io.Writer) error {
	orch, err := buildOrchestrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}
	runner := guidedRuntimeRunner{orch: orch}
	handler := func(action guidance.GuidedAction) (guidance.ActionResult, error) {
		return handleGuidedAction(ctx, cfg, runner, capability, action)
	}
	return guidance.RunShell(plan, capability, os.Stdin, output, guidance.WithActionHandler(handler))
}

func handleGuidedAction(ctx context.Context, cfg config.ProjectConfig, runner guidedLifecycleRunner, capability guidance.TUICapability, action guidance.GuidedAction) (guidance.ActionResult, error) {
	switch action.ID {
	case "init", "init_here":
		settings := projectEnvSettingsFromGuidedAction(cfg, action)
		result, err := onboarding.WriteProjectEnvWithSettings(cfg.Dir, settings, false)
		if err != nil {
			return guidance.ActionResult{}, err
		}
		nextConfig := reloadGuidedConfig(cfg)
		context := collectGuidedContext(ctx, nextConfig, capability)
		message := projectEnvActionMessage(result, filepath.Join(cfg.Dir, ".env.stageserve"))
		return guidance.ActionResult{Plan: guidance.Plan(context), Message: message}, nil
	case "up", "attach", "down", "detach":
		if err := executeGuidedAction(action.ID, ctx, cfg, runner); err != nil {
			return guidance.ActionResult{}, fmt.Errorf("failed to %s project: %w", guidedActionVerb(action.ID), err)
		}
		nextConfig := reloadGuidedConfig(cfg)
		context := collectGuidedContext(ctx, nextConfig, capability)
		return guidance.ActionResult{Plan: guidance.Plan(context), Message: guidedActionMessage(action.ID)}, nil
	case "status":
		body, err := runner.Status(ctx, cfg)
		if err != nil {
			return guidance.ActionResult{}, err
		}
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{
			Plan: nextPlan,
			Utility: &guidance.UtilitySurface{
				Title:           cfg.Name + " status",
				Body:            body,
				DismissOnAnyKey: true,
			},
		}, nil
	case "logs":
		body, err := runner.Logs(ctx, cfg, "nginx")
		if err != nil {
			return guidance.ActionResult{}, err
		}
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{
			Plan: nextPlan,
			Utility: &guidance.UtilitySurface{
				Title:  cfg.Name + " logs",
				Body:   body,
				Footer: "q exit logs • esc exit logs",
			},
		}, nil
	default:
		return guidance.ActionResult{Plan: guidance.Plan(collectGuidedContext(ctx, cfg, capability)), Message: fmt.Sprintf("%s is not available in guided mode yet.", action.Label)}, nil
	}
}

func executeGuidedAction(actionID string, ctx context.Context, cfg config.ProjectConfig, runner guidedLifecycleRunner) error {
	switch actionID {
	case "up":
		if runner == nil {
			return fmt.Errorf("guided lifecycle runner is not available")
		}
		return runner.Up(ctx, cfg)
	case "attach":
		if runner == nil {
			return fmt.Errorf("guided lifecycle runner is not available")
		}
		return runner.Attach(ctx, cfg)
	case "down":
		if runner == nil {
			return fmt.Errorf("guided lifecycle runner is not available")
		}
		return runner.Down(ctx, cfg, false)
	case "detach":
		if runner == nil {
			return fmt.Errorf("guided lifecycle runner is not available")
		}
		return runner.Detach(ctx, cfg)
	default:
		return fmt.Errorf("action %q not implemented", actionID)
	}
}

func guidedActionVerb(actionID string) string {
	switch actionID {
	case "attach":
		return "add"
	case "down":
		return "stop"
	case "detach":
		return "remove"
	default:
		return "run"
	}
}

func guidedActionMessage(actionID string) string {
	switch actionID {
	case "attach":
		return "Project is now added to StageServe."
	case "down":
		return "Project is stopped."
	case "detach":
		return "Project was removed from StageServe."
	default:
		return "Project is now running."
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
