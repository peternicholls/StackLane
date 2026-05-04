// stage init: generate a starter .env.stageserve in the project root.
package commands

import (
	"fmt"
	"os"

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
	NoTUI          bool
	JSON           bool
}

func NewInit(shared *SharedFlags) *cobra.Command {
	f := &initFlags{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a project for StageServe",
		Long:  "Creates a starter .env.stageserve with documented defaults. Validates docroot/site settings and protects existing config from accidental overwrite.",
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := resolveOutputMode(f.JSON, f.NoTUI, false, f.NonInteractive)

			// Determine project directory.
			projectDir := f.ProjectDir
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
			if f.DocRoot != "" {
				if err := onboarding.ValidateDocroot(root, f.DocRoot); err != nil {
					return err
				}
			}

			// Write the project env file.
			action, writeErr := onboarding.WriteProjectEnv(root, f.SiteName, f.DocRoot, f.Force)

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
			default:
				step = onboarding.StepResult{
					ID:      "init.env_file",
					Label:   ".env.stageserve",
					Status:  onboarding.StatusReady,
					Message: fmt.Sprintf(".env.stageserve %s in %s", action, root),
				}
			}

			result := onboarding.BuildResult([]onboarding.StepResult{step}, nil, []string{"stage up"})

			switch mode {
			case onboarding.OutputModeJSON:
				p := onboarding.JSONProjector{W: cmd.OutOrStdout()}
				if err := p.Project(result); err != nil {
					return err
				}
			case onboarding.OutputModeTUI:
				p := onboarding.TUIProjector{W: cmd.OutOrStdout()}
				p.Project(result)
			default:
				p := onboarding.TextProjector{W: cmd.OutOrStdout()}
				p.Project(result)
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
	cmd.Flags().BoolVar(&f.NoTUI, "no-tui", false, "Force plain-text output")
	cmd.Flags().BoolVar(&f.JSON, "json", false, "Emit JSON envelope only")
	return cmd
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
