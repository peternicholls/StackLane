package guidance

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/state"
)

func baseContext() GuidedContext {
	return GuidedContext{
		ProjectEnvExists: true,
		ProjectEnvValid:  true,
		StackID:          "20i",
		SiteName:         "demo",
		WebFolder:        "public_html",
		LocalURL:         "http://demo.test",
		ProjectEnvPath:   "/sites/demo/.env.stageserve",
	}
}

func TestPlanClassifiesCoreSituations(t *testing.T) {
	tests := []struct {
		name string
		ctx  GuidedContext
		want Situation
	}{
		{
			name: "machine not ready",
			ctx: func() GuidedContext {
				ctx := baseContext()
				ctx.MachineReadiness = MachineReadinessSummary{Checked: true, Blocked: true, NextFixLabel: "Set up this computer", NextCommand: "stage setup"}
				return ctx
			}(),
			want: SituationMachineNotReady,
		},
		{
			name: "not project",
			ctx: func() GuidedContext {
				ctx := baseContext()
				ctx.NotProject = true
				return ctx
			}(),
			want: SituationNotProject,
		},
		{
			name: "missing config",
			ctx: func() GuidedContext {
				ctx := baseContext()
				ctx.ProjectEnvExists = false
				return ctx
			}(),
			want: SituationProjectMissingConf,
		},
		{
			name: "ready to run",
			ctx:  baseContext(),
			want: SituationProjectReadyToRun,
		},
		{
			name: "running",
			ctx: func() GuidedContext {
				ctx := baseContext()
				ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
				return ctx
			}(),
			want: SituationProjectRunning,
		},
		{
			name: "down",
			ctx: func() GuidedContext {
				ctx := baseContext()
				ctx.ProjectState = &state.Record{AttachmentState: state.StateDown}
				return ctx
			}(),
			want: SituationProjectDown,
		},
		{
			name: "does not match expected",
			ctx: func() GuidedContext {
				ctx := baseContext()
				ctx.Warnings = []string{"Project settings do not match the recorded project path."}
				return ctx
			}(),
			want: SituationDriftDetected,
		},
		{
			name: "unknown error",
			ctx:  ContextFromError("/tmp/demo", TUICapability{}, os.ErrInvalid),
			want: SituationUnknownError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plan := Plan(test.ctx)
			if plan.Situation != test.want {
				t.Fatalf("situation=%s want %s", plan.Situation, test.want)
			}
			if plan.StatusHeader == "" {
				t.Fatal("status header must not be empty")
			}
			if len(plan.DecisionItems) == 0 && len(plan.WorkItems) == 0 {
				t.Fatal("plan should expose at least one safe next step")
			}
		})
	}
}

func TestCollectDoesNotCreateProjectOrStateFiles(t *testing.T) {
	projectDir := t.TempDir()
	stateDir := filepath.Join(t.TempDir(), "state")
	cfg := config.ProjectConfig{
		Name:          "demo",
		Slug:          "demo",
		Dir:           projectDir,
		StateDir:      stateDir,
		StackKind:     "20i",
		StackHome:     filepath.Dir(stateDir),
		Hostname:      "demo.test",
		SiteSuffix:    "test",
		DocRoot:       projectDir,
		SharedGateway: config.SharedGateway{HTTPSPort: 443},
	}

	ctx := Collect(context.Background(), cfg, CollectOptions{})
	if ctx.ProjectEnvExists {
		t.Fatal("project env should not exist")
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".env.stageserve")); !os.IsNotExist(err) {
		t.Fatalf("Collect should not create .env.stageserve, stat err=%v", err)
	}
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Fatalf("Collect should not create state dir, stat err=%v", err)
	}
}
