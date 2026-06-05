// stage init: generate a starter .env.stageserve in the project root.
package commands

import (
	"fmt"
	"os"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/spf13/cobra"
)

// initFlags holds init-specific CLI flags.
type initFlags struct {
	DocRoot        string
	SiteName       string
	ProjectDir     string
	Force          bool
	NonInteractive bool
	NotUI          bool
	CLI            bool
	NoTUI          bool
	JSON           bool
}

func NewInit(shared *SharedFlags) *cobra.Command {
	f := &initFlags{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create project settings for this folder",
		Long:  "Creates a starter .env.stageserve with documented defaults. In an interactive terminal, stage init opens the guided settings form before it writes anything. It also validates the web folder and protects existing project settings from accidental overwrite.",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := resolveOutputMode(f.JSON, plainTextOutputRequested(f.NotUI, f.CLI, f.NoTUI), f.NonInteractive)

			// Determine project directory.
			projectDir := initProjectDir(shared, f)
			if projectDir == "" {
				var err error
				projectDir, err = os.Getwd()
				if err != nil {
					return fmt.Errorf("cannot determine current directory: %w", err)
				}
			}

			// Validate project root.
			root, err := onboarding.ValidateProjectRoot(projectDir)
			if err != nil {
				return err
			}

			// Validate docroot (only if supplied).
			docRoot := initDocRoot(shared, f)
			if docRoot != "" {
				if err := onboarding.ValidateDocroot(root, docRoot); err != nil {
					return err
				}
			}

			capability := guidance.DetectCapability(os.Stdin, os.Stdout, os.Stderr, f.NotUI || f.NoTUI, f.CLI)
			if mode == onboarding.OutputModeTUI && capability.AllowsTUI() && !f.Force {
				cfg, err := loadInitGuidedConfig(shared, f)
				if err != nil {
					return err
				}
				context := collectGuidedContext(cmd.Context(), cfg, capability)
				plan := buildInitGuidedPlan(context)
				return runGuidedTUI(cmd.Context(), cfg, plan, capability, cmd.OutOrStdout(), guidance.WithInitialEditor())
			}

			settings := projectEnvSettingsFromInitFlags(shared, f)

			// Write the project env file unless dry-run is requested.
			action := onboarding.InitActionSkipped
			var writeErr error
			if shared.DryRun {
				action, writeErr = plannedProjectEnvAction(root, settings, f.Force)
			} else {
				action, writeErr = onboarding.WriteProjectEnvWithSettings(root, settings, f.Force)
			}

			// Build a result step from the write outcome.
			var step onboarding.StepResult
			switch {
			case writeErr != nil:
				step = onboarding.StepResult{
					ID:      "init.env_file",
					Label:   ".env.stageserve",
					Status:  onboarding.StatusError,
					Message: writeErr.Error(),
				}
			case action == onboarding.InitActionSkipped:
				step = onboarding.StepResult{
					ID:      "init.env_file",
					Label:   ".env.stageserve",
					Status:  onboarding.StatusReady,
					Message: ".env.stageserve already exists (use --force to overwrite)",
				}
			case shared.DryRun && action == onboarding.InitActionOverwritten:
				step = onboarding.StepResult{
					ID:      "init.env_file",
					Label:   ".env.stageserve",
					Status:  onboarding.StatusReady,
					Message: fmt.Sprintf(".env.stageserve would be overwritten in %s", root),
				}
			case shared.DryRun && action == onboarding.InitActionCreated:
				step = onboarding.StepResult{
					ID:      "init.env_file",
					Label:   ".env.stageserve",
					Status:  onboarding.StatusReady,
					Message: fmt.Sprintf(".env.stageserve would be created in %s", root),
				}
			default:
				step = onboarding.StepResult{
					ID:      "init.env_file",
					Label:   ".env.stageserve",
					Status:  onboarding.StatusReady,
					Message: fmt.Sprintf(".env.stageserve %s in %s", action, root),
				}
			}

			result := onboarding.BuildResult([]onboarding.StepResult{step}, nil, []string{"stage up"})

			projector := onboarding.NewProjector(mode, cmd.OutOrStdout())
			if err := projector.Project(result); err != nil {
				return err
			}

			if writeErr != nil {
				return &initExitError{code: int(onboarding.ExitError)}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&f.DocRoot, "docroot", "", "Document root inside the project")
	cmd.Flags().StringVar(&f.SiteName, "site-name", "", "Site name override")
	cmd.Flags().StringVar(&f.ProjectDir, "project-dir", "", "Project directory (defaults to cwd)")
	cmd.Flags().BoolVar(&f.Force, "force", false, "Overwrite existing .env.stageserve")
	cmd.Flags().BoolVar(&f.NonInteractive, "non-interactive", false, "Suppress interactive prompts")
	addPlainTextOutputFlags(cmd, &f.NotUI, &f.CLI, &f.NoTUI)
	cmd.Flags().BoolVar(&f.JSON, "json", false, "Emit JSON envelope only")
	return cmd
}

func buildInitGuidedPlan(ctx guidance.GuidedContext) guidance.NextActionPlan {
	plan := guidance.Plan(ctx)
	if !ctx.ProjectEnvExists {
		return plan
	}

	plan.StatusHeader = "This folder already has StageServe settings."
	plan.Summary = "StageServe can update .env.stageserve with the values shown here."
	plan.DecisionItems = []guidance.GuidedAction{
		{
			ID:                   "overwrite_init",
			Kind:                 "choose",
			Label:                "Update project settings",
			Description:          "Overwrite .env.stageserve with the values shown here.",
			MutatesState:         true,
			RequiresConfirmation: true,
			DirectCommand:        "stage init --force",
		},
		{
			ID:            "edit_config",
			Kind:          "choose",
			Label:         "Edit before updating",
			Description:   "Change the project name, web folder, or local address first.",
			DirectCommand: "stage init --force",
		},
	}
	plan.DirectCommands = []string{"stage init --force"}
	for itemIndex := range plan.VisibleDefaults {
		if plan.VisibleDefaults[itemIndex].Label == "Settings file" {
			plan.VisibleDefaults[itemIndex].Note = "will be updated after confirmation"
		}
	}
	return plan
}

func initProjectDir(shared *SharedFlags, flags *initFlags) string {
	if flags.ProjectDir != "" {
		return flags.ProjectDir
	}
	return shared.ProjectDir
}

func initSiteName(shared *SharedFlags, flags *initFlags) string {
	if flags.SiteName != "" {
		return flags.SiteName
	}
	return shared.SiteName
}

func initDocRoot(shared *SharedFlags, flags *initFlags) string {
	if flags.DocRoot != "" {
		return flags.DocRoot
	}
	return shared.DocRoot
}

func loadInitGuidedConfig(shared *SharedFlags, flags *initFlags) (config.ProjectConfig, error) {
	merged := *shared
	merged.ProjectDir = initProjectDir(shared, flags)
	merged.SiteName = initSiteName(shared, flags)
	merged.DocRoot = initDocRoot(shared, flags)
	cfg, err := loadConfig(&merged)
	if err != nil {
		return config.ProjectConfig{}, err
	}
	cfg.DryRun = shared.DryRun
	return cfg, nil
}

func projectEnvSettingsFromInitFlags(shared *SharedFlags, flags *initFlags) onboarding.ProjectEnvSettings {
	return onboarding.ProjectEnvSettings{
		SiteName:   initSiteName(shared, flags),
		DocRoot:    normalizeProjectEnvDocRoot(initDocRoot(shared, flags)),
		SiteSuffix: normalizeProjectEnvSuffix(shared.SiteSuffix),
	}
}

func plannedProjectEnvAction(projectRoot string, settings onboarding.ProjectEnvSettings, force bool) (onboarding.InitAction, error) {
	preview, err := onboarding.PreviewProjectEnvWithSettings(projectRoot, settings)
	if err != nil {
		return "", err
	}
	if preview.Exists {
		if force {
			return onboarding.InitActionOverwritten, nil
		}
		return onboarding.InitActionSkipped, nil
	}
	return onboarding.InitActionCreated, nil
}

// initExitError wraps a non-zero exit code for the init command.
type initExitError struct{ code int }

func (e *initExitError) Error() string {
	return fmt.Sprintf("init finished with exit code %d", e.code)
}

func (e *initExitError) ExitCode() int {
	return e.code
}

func (e *initExitError) Silent() bool {
	return true
}
