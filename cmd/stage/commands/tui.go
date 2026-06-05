package commands

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

var buildGuidedSetupResult = buildMachineReadinessResultForConfig
var buildGuidedDoctorResult = buildMachineReadinessResultForConfig

type guidedLifecycleRunner interface {
	Up(context.Context, config.ProjectConfig) error
	Attach(context.Context, config.ProjectConfig) error
	Down(context.Context, config.ProjectConfig, bool) error
	Detach(context.Context, config.ProjectConfig) error
	RestartService(context.Context, config.ProjectConfig, string) error
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

func (r guidedRuntimeRunner) RestartService(ctx context.Context, cfg config.ProjectConfig, service string) error {
	return r.orch.RestartService(ctx, cfg, service)
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

func runGuidedTUI(ctx context.Context, cfg config.ProjectConfig, plan guidance.NextActionPlan, capability guidance.TUICapability, output io.Writer, opts ...guidance.ShellOption) error {
	orch, err := buildOrchestrator(cfg)
	if err != nil {
		return fmt.Errorf("failed to build orchestrator: %w", err)
	}
	runner := guidedRuntimeRunner{orch: orch}
	handler := func(actionCtx context.Context, action guidance.GuidedAction) (guidance.ActionResult, error) {
		return handleGuidedAction(actionCtx, cfg, runner, capability, action)
	}
	shellOpts := append([]guidance.ShellOption{guidance.WithActionHandler(handler)}, opts...)
	return guidance.RunShell(plan, capability, os.Stdin, output, shellOpts...)
}

func handleGuidedAction(ctx context.Context, cfg config.ProjectConfig, runner guidedLifecycleRunner, capability guidance.TUICapability, action guidance.GuidedAction) (guidance.ActionResult, error) {
	switch action.ID {
	case "setup":
		result := buildGuidedSetupResult(cfg)
		body, err := renderMachineReadinessReport("StageServe Setup", result)
		if err != nil {
			return guidance.ActionResult{}, err
		}
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{
			Plan: nextPlan,
			Utility: &guidance.UtilitySurface{
				Title:           "Setup report",
				Body:            strings.TrimSpace(body),
				Footer:          "press any key to return • q quit",
				DismissOnAnyKey: true,
			},
		}, nil
	case "doctor":
		result := buildGuidedDoctorResult(cfg)
		body, err := renderMachineReadinessReport("StageServe Doctor", result)
		if err != nil {
			return guidance.ActionResult{}, err
		}
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{
			Plan: nextPlan,
			Utility: &guidance.UtilitySurface{
				Title:           "Doctor report",
				Body:            strings.TrimSpace(body),
				Footer:          "press any key to return • q quit",
				DismissOnAnyKey: true,
			},
		}, nil
	case "init", "init_here", "overwrite_init":
		settings := projectEnvSettingsFromGuidedAction(cfg, action)
		force := action.ID == "overwrite_init"
		var message string
		if cfg.DryRun {
			plannedAction, err := plannedProjectEnvAction(cfg.Dir, settings, force)
			if err != nil {
				return guidance.ActionResult{}, err
			}
			message = projectEnvDryRunMessage(plannedAction, filepath.Join(cfg.Dir, ".env.stageserve"))
		} else {
			result, err := onboarding.WriteProjectEnvWithSettings(cfg.Dir, settings, force)
			if err != nil {
				return guidance.ActionResult{}, err
			}
			message = projectEnvActionMessage(result, filepath.Join(cfg.Dir, ".env.stageserve"))
		}
		nextConfig := reloadGuidedConfig(cfg)
		context := collectGuidedContext(ctx, nextConfig, capability)
		return guidance.ActionResult{Plan: guidance.Plan(context), Message: message}, nil
	case "up", "attach", "down", "detach":
		if err := executeGuidedActionWithInputs(action.ID, action.Inputs, ctx, cfg, runner); err != nil {
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
		service := guidedActionInput(action, "service", "nginx")
		body, err := runner.Logs(ctx, cfg, service)
		if err != nil {
			return guidance.ActionResult{}, err
		}
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{
			Plan: nextPlan,
			Utility: &guidance.UtilitySurface{
				Title:  cfg.Name + " logs",
				Body:   "Project: " + cfg.Name + "\nService: " + service + "\n\n" + body,
				Footer: "q exit logs • esc exit logs",
			},
		}, nil
	case "logs_select":
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{Plan: nextPlan, Utility: serviceChoiceSurface("Project log choices", "stage logs", nextPlan, true)}, nil
	case "restart_select":
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{Plan: nextPlan, Utility: serviceChoiceSurface("Restart choices", "restart", nextPlan, false)}, nil
	case "restart_service":
		service := guidedActionInput(action, "service", "")
		if runner == nil {
			return guidance.ActionResult{}, fmt.Errorf("guided lifecycle runner is not available")
		}
		if err := runner.RestartService(ctx, cfg, service); err != nil {
			return guidance.ActionResult{}, fmt.Errorf("failed to restart %s: %w", service, err)
		}
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		return guidance.ActionResult{Plan: nextPlan, Message: service + " was restarted."}, nil
	case "open_browser":
		url := action.DirectCommand
		if url == "" {
			url = "https://" + cfg.Hostname
		}
		nextPlan := guidance.Plan(collectGuidedContext(ctx, reloadGuidedConfig(cfg), capability))
		if err := openBrowser(url); err != nil {
			return guidance.ActionResult{
				Plan:    nextPlan,
				Message: "StageServe couldn't open the browser. Next: open " + url,
			}, nil
		}
		return guidance.ActionResult{
			Plan:    nextPlan,
			Message: "Opening " + url + " in your browser.",
		}, nil
	default:
		return guidance.ActionResult{Plan: guidance.Plan(collectGuidedContext(ctx, cfg, capability)), Message: fmt.Sprintf("%s is not available in guided mode yet.", action.Label)}, nil
	}
}

func serviceChoiceSurface(title, commandPrefix string, plan guidance.NextActionPlan, logs bool) *guidance.UtilitySurface {
	commands := []string{}
	for _, command := range plan.DirectCommands {
		if logs && strings.HasPrefix(command, commandPrefix+" ") {
			commands = append(commands, command)
		}
		if !logs && strings.Contains(command, " restart ") {
			commands = append(commands, command)
		}
	}
	if len(commands) == 0 {
		commands = []string{"No service-specific command is available yet."}
	}
	return &guidance.UtilitySurface{
		Title:           title,
		Body:            strings.Join(commands, "\n"),
		Footer:          "q return • esc return",
		DismissOnAnyKey: false,
	}
}

func renderMachineReadinessReport(title string, result onboarding.CommandResult) (string, error) {
	var output bytes.Buffer
	projector := onboarding.NewProjector(onboarding.OutputModeText, &output, onboarding.ProjectorOptions{
		Title:    title,
		Detailed: true,
	})
	if err := projector.Project(result); err != nil {
		return "", err
	}
	return output.String(), nil
}

func executeGuidedAction(actionID string, ctx context.Context, cfg config.ProjectConfig, runner guidedLifecycleRunner) error {
	return executeGuidedActionWithInputs(actionID, nil, ctx, cfg, runner)
}

func executeGuidedActionWithInputs(actionID string, inputs map[string]string, ctx context.Context, cfg config.ProjectConfig, runner guidedLifecycleRunner) error {
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
	case "restart_service":
		if runner == nil {
			return fmt.Errorf("guided lifecycle runner is not available")
		}
		service := strings.TrimSpace(inputs["service"])
		if service == "" {
			return fmt.Errorf("service is required")
		}
		return runner.RestartService(ctx, cfg, service)
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
	case "restart_service":
		return "restart"
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
	case "restart_service":
		return "Project service restarted."
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

func projectEnvDryRunMessage(action onboarding.InitAction, path string) string {
	switch action {
	case onboarding.InitActionCreated:
		return "Dry run: would create project settings at " + path + "."
	case onboarding.InitActionSkipped:
		return "Dry run: project settings already exist at " + path + "."
	case onboarding.InitActionOverwritten:
		return "Dry run: would update project settings at " + path + "."
	default:
		return "Dry run: checked project settings at " + path + "."
	}
}

// openBrowser opens url in the user's default browser.
var browserCommand = exec.Command

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = browserCommand("open", url)
	case "linux":
		cmd = browserCommand("xdg-open", url)
	default:
		return fmt.Errorf("opening a browser is not supported on %s", runtime.GOOS)
	}
	return cmd.Start()
}
