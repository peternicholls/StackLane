package guidance

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/onboarding"
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
	if ctx.ProjectEnvPreview == nil {
		t.Fatal("Collect did not build project settings preview")
	}
	if !strings.Contains(ctx.ProjectEnvPreview.Body, "STAGESERVE_STACK=20i") {
		t.Fatalf("preview body did not use shared env renderer:\n%s", ctx.ProjectEnvPreview.Body)
	}
}

func TestShellViewRendersCoreSurfaces(t *testing.T) {
	plan := Plan(baseContext())
	view := renderShellView(plan, 80, 0, true, true)
	for _, want := range []string{"StageServe", "This project is ready to run.", "Key facts", "What you can do", "Details", "stage up", "q quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("shell view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "\x1b") {
		t.Fatalf("no-color shell view contains ANSI escape sequences:\n%s", view)
	}
}

func TestMachineReadinessFromResultUsesOnboardingSemantics(t *testing.T) {
	result := onboarding.BuildResult([]onboarding.StepResult{
		{ID: "docker.binary", Label: "Docker CLI", Status: onboarding.StatusReady, Message: "docker found"},
		{ID: "docker.daemon", Label: "Docker daemon", Status: onboarding.StatusNeedsAction, Message: "Docker daemon is not reachable"},
		{ID: "state.dir", Label: "State directory", Status: onboarding.StatusError, Message: "state dir cannot be read"},
	}, nil, nil)

	summary := MachineReadinessFromResult(result)
	if !summary.Checked || !summary.Blocked {
		t.Fatalf("summary checked=%v blocked=%v", summary.Checked, summary.Blocked)
	}
	if summary.Status != string(onboarding.OverallError) {
		t.Fatalf("status=%q want %q", summary.Status, onboarding.OverallError)
	}
	if summary.NextFixLabel != "Docker daemon" || summary.NextCommand != "stage setup" {
		t.Fatalf("next fix=%q command=%q", summary.NextFixLabel, summary.NextCommand)
	}
	if len(summary.WorkItems) != 3 {
		t.Fatalf("work items=%d want 3", len(summary.WorkItems))
	}
	if summary.WorkItems[0].Status != "ready" {
		t.Fatalf("first status=%q want ready", summary.WorkItems[0].Status)
	}
	if summary.WorkItems[1].Status != "needs attention" || summary.WorkItems[1].DirectCommand != "stage setup" {
		t.Fatalf("second item=%+v", summary.WorkItems[1])
	}
	if summary.WorkItems[2].Status != "error" {
		t.Fatalf("third status=%q want error", summary.WorkItems[2].Status)
	}
}
