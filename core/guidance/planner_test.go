package guidance

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
		SiteSuffix:       "test",
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

func TestPlanProjectDownUsesAttachWhenRuntimeAlreadyRunning(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateDown}
	ctx.Runtime = RuntimeSummary{Checked: true, Running: true, Status: "running"}

	plan := Plan(ctx)

	if plan.Situation != SituationProjectDown {
		t.Fatalf("situation=%s want %s", plan.Situation, SituationProjectDown)
	}
	if plan.StatusHeader != "This project isn't added to StageServe right now." {
		t.Fatalf("status header=%q", plan.StatusHeader)
	}
	if len(plan.DecisionItems) < 2 {
		t.Fatalf("decision items=%+v", plan.DecisionItems)
	}
	if plan.DecisionItems[0].ID != "attach" || plan.DecisionItems[0].Label != "Add this project to StageServe" {
		t.Fatalf("primary action=%+v", plan.DecisionItems[0])
	}
	if plan.DecisionItems[0].DirectCommand != "stage attach" {
		t.Fatalf("attach direct command=%q", plan.DecisionItems[0].DirectCommand)
	}
	if plan.DecisionItems[1].ID != "status" || plan.DecisionItems[1].DirectCommand != "stage status" {
		t.Fatalf("secondary action=%+v", plan.DecisionItems[1])
	}
	if len(plan.DirectCommands) < 2 || plan.DirectCommands[0] != "stage attach" || plan.DirectCommands[2] != "stage down" {
		t.Fatalf("direct commands=%+v", plan.DirectCommands)
	}
}

func TestPlanProjectRunningKeepsDetachOutOfPrimaryChoices(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	ctx.Runtime = RuntimeSummary{Checked: true, Running: true, Status: "running", Services: []RuntimeServiceSummary{
		{ServiceName: "nginx", ContainerName: "stage-demo-nginx", Status: "running", EligibleForLogs: true, EligibleForRestart: true},
		{ServiceName: "apache", ContainerName: "stage-demo-apache", Status: "running", EligibleForLogs: true, EligibleForRestart: true},
	}}

	plan := Plan(ctx)

	if len(plan.DecisionItems) < 5 {
		t.Fatalf("decision items=%+v", plan.DecisionItems)
	}
	if plan.DecisionItems[0].ID != "logs_select" {
		t.Fatalf("running-project default should be logs selector, got %+v", plan.DecisionItems[0])
	}
	if plan.DecisionItems[1].ID != "open_browser" || plan.DecisionItems[2].ID != "status" {
		t.Fatalf("open/status should follow logs before mutations: %+v", plan.DecisionItems)
	}
	if !containsAction(plan.DecisionItems, "restart_service") || !containsAction(plan.DecisionItems, "more") {
		t.Fatalf("running-project actions missing restart or More: %+v", plan.DecisionItems)
	}
	for _, item := range plan.DecisionItems {
		if item.ID == "detach" {
			t.Fatalf("running-project choices should not expose detach directly: %+v", plan.DecisionItems)
		}
	}
	if len(plan.DirectCommands) < 4 || plan.DirectCommands[len(plan.DirectCommands)-1] != "stage down" {
		t.Fatalf("direct commands=%+v", plan.DirectCommands)
	}
	for _, command := range plan.DirectCommands {
		if strings.Contains(command, "<project-") {
			t.Fatalf("direct command is not copy-pasteable: %q", command)
		}
	}
}

func containsAction(actions []GuidedAction, id string) bool {
	for _, action := range actions {
		if action.ID == id {
			return true
		}
	}
	return false
}

func TestPlanProjectRunningAutoSelectsSingleService(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	ctx.Runtime = RuntimeSummary{Checked: true, Running: true, Status: "running", Services: []RuntimeServiceSummary{
		{ServiceName: "apache", ContainerName: "stage-demo-apache", Status: "running", EligibleForLogs: true, EligibleForRestart: true},
	}}

	plan := Plan(ctx)

	if plan.DecisionItems[0].ID != "logs" || plan.DecisionItems[0].Inputs["service"] != "apache" {
		t.Fatalf("logs action=%+v", plan.DecisionItems[0])
	}
	restartFound := false
	for _, action := range plan.DecisionItems {
		if action.ID == "restart_service" {
			restartFound = true
			if action.Inputs["service"] != "apache" || !action.RequiresConfirmation {
				t.Fatalf("restart action=%+v", action)
			}
		}
	}
	if !restartFound {
		t.Fatalf("restart action missing: %+v", plan.DecisionItems)
	}
}

func TestRenderTextUsesPlainLanguageWhileKeepingDirectCommands(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateDown}
	ctx.Runtime = RuntimeSummary{Checked: true, Running: true, Status: "running"}

	buf := &bytes.Buffer{}
	if err := RenderText(buf, Plan(ctx)); err != nil {
		t.Fatalf("render text: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Add this project to StageServe", "Check this project's status", "More:", "Direct commands:", "Plain text output: stage --cli", "Advanced and troubleshooting: stage doctor", "stage attach", "stage down"} {
		if !strings.Contains(out, want) {
			t.Fatalf("text output missing %q:\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"Attach", "Detach", "current state"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("text output leaked %q:\n%s", unwanted, out)
		}
	}
}

func TestRenderTextIncludesConfirmationImpactCopy(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	plan := Plan(ctx)

	buf := &bytes.Buffer{}
	if err := RenderText(buf, plan); err != nil {
		t.Fatalf("render text: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"Confirmation:", "Stops http://demo.test; project files and settings are not changed."} {
		if !strings.Contains(out, want) {
			t.Fatalf("text output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderTextUsesSeverityWithoutANSI(t *testing.T) {
	buf := &bytes.Buffer{}
	ctx := baseContext()
	ctx.MachineReadiness = MachineReadinessSummary{Checked: true, Blocked: true, NextFixLabel: "Restore runtime file", NextCommand: "stage setup"}
	if err := RenderText(buf, Plan(ctx)); err != nil {
		t.Fatalf("render text: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "StageServe\nNeeds attention\n") {
		t.Fatalf("severity header missing:\n%s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("text output contains ANSI escape sequences:\n%s", out)
	}
}

func TestPlanUnknownErrorExposesOrderedRecoverySteps(t *testing.T) {
	plan := Plan(ContextFromError("/tmp/demo", TUICapability{}, os.ErrInvalid))

	if plan.Situation != SituationUnknownError {
		t.Fatalf("situation=%s want %s", plan.Situation, SituationUnknownError)
	}
	if len(plan.WorkItems) != 5 {
		t.Fatalf("work items=%+v", plan.WorkItems)
	}
	if plan.WorkItems[0].DirectCommand != "stage doctor" || plan.WorkItems[1].DirectCommand != "stage status" || plan.WorkItems[2].DirectCommand != "stage logs" {
		t.Fatalf("recovery order=%+v", plan.WorkItems)
	}
	if len(plan.DecisionItems) != 5 || plan.DecisionItems[0].ID != "doctor" || plan.DecisionItems[1].ID != "status" || plan.DecisionItems[2].ID != "logs" {
		t.Fatalf("decision items=%+v", plan.DecisionItems)
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
	for _, want := range []string{"StageServe", "Project", "This project is ready to run.", "Key facts", "What you can do", "Details", "stage up", "q quit"} {
		if !strings.Contains(view, want) {
			t.Fatalf("shell view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "\x1b") {
		t.Fatalf("no-color shell view contains ANSI escape sequences:\n%s", view)
	}
}

func TestShellRecoveryViewPrioritizesRecoveryStepsBeforeChoices(t *testing.T) {
	ctx := baseContext()
	ctx.Warnings = []string{"Project settings do not match the recorded project path."}
	view := renderShellView(Plan(ctx), 80, 0, false, true)

	headerLine := ""
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "StageServe") {
			headerLine = line
			break
		}
	}
	if !strings.Contains(headerLine, "Recovery") {
		t.Fatalf("header missing recovery surface label:\n%s", view)
	}
	if recoveryIndex, choicesIndex := strings.Index(view, "Recovery steps"), strings.Index(view, "What you can do"); recoveryIndex == -1 || choicesIndex == -1 || recoveryIndex > choicesIndex {
		t.Fatalf("recovery checklist should appear before choices:\n%s", view)
	}
}

func TestShellNarrowViewStacksKeyFacts(t *testing.T) {
	view := renderShellView(Plan(baseContext()), 48, 0, false, true)
	for _, want := range []string{"Project", "Site name\n    demo", "Local URL\n    http://demo.test", "↑/↓ move • enter choose • c commands • d doctor • q"} {
		if !strings.Contains(view, want) {
			t.Fatalf("narrow shell view missing %q:\n%s", want, view)
		}
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

func driveShellCommand(t *testing.T, model shellModel, cmd tea.Cmd) shellModel {
	t.Helper()
	queue := []tea.Cmd{cmd}
	for step := 0; len(queue) > 0 && step < 16; step++ {
		current := queue[0]
		queue = queue[1:]
		if current == nil {
			continue
		}
		msg := current()
		if batch, ok := msg.(tea.BatchMsg); ok {
			queue = append([]tea.Cmd(batch), queue...)
			continue
		}
		updated, nextCmd := model.Update(msg)
		next, ok := updated.(shellModel)
		if !ok {
			t.Fatalf("updated model type %T", updated)
		}
		model = next
		if nextCmd != nil {
			queue = append(queue, nextCmd)
		}
		if !model.loading {
			return model
		}
	}
	return model
}

func TestShellMachineReadinessEnterUsesSetupAction(t *testing.T) {
	ctx := baseContext()
	ctx.MachineReadiness = MachineReadinessSummary{
		Checked:   true,
		Blocked:   true,
		WorkItems: []WorkItem{{Label: "Docker daemon", Status: "needs attention", DirectCommand: "stage setup"}},
	}
	model := newShellModel(Plan(ctx), true)
	called := false
	model.actionHandler = func(_ context.Context, action GuidedAction) (ActionResult, error) {
		called = true
		if action.ID != "setup" {
			t.Fatalf("action id=%q want setup", action.ID)
		}
		return ActionResult{Plan: Plan(ctx)}, nil
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := driveShellCommand(t, updated.(shellModel), cmd)
	if !called {
		t.Fatal("expected setup action to run")
	}
	if !strings.Contains(next.View(), "c commands") || !strings.Contains(next.View(), "d doctor") {
		t.Fatalf("machine readiness footer missing review hint:\n%s", next.View())
	}
}

func TestShellCommandShortcutShowsDirectCommands(t *testing.T) {
	model := newShellModel(Plan(baseContext()), true)

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cmd != nil {
		t.Fatal("command shortcut should not start async work")
	}
	next := updated.(shellModel)
	if next.utility == nil || next.utility.Title != "More..." {
		t.Fatalf("command utility=%+v", next.utility)
	}
	view := next.View()
	for _, want := range []string{"More...", "Direct commands", "Plain text output", "Advanced and troubleshooting", "stage up", "stage status"} {
		if !strings.Contains(view, want) {
			t.Fatalf("command utility missing %q:\n%s", want, view)
		}
	}
}

func TestShellMoreActionShowsAdvancedFallback(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	model := newShellModel(Plan(ctx), true)
	for index, action := range model.plan.DecisionItems {
		if action.ID == "more" {
			model.cursor = index
			break
		}
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("More action should not start async work")
	}
	next := updated.(shellModel)
	if next.utility == nil || next.utility.Title != "More..." {
		t.Fatalf("More utility=%+v", next.utility)
	}
	view := next.View()
	for _, want := range []string{"Direct commands", "stage logs", "Plain text output", "stage --cli", "Advanced and troubleshooting"} {
		if !strings.Contains(view, want) {
			t.Fatalf("More view missing %q:\n%s", want, view)
		}
	}
}

func TestShellRestartConfirmationNamesService(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	ctx.Runtime = RuntimeSummary{Checked: true, Running: true, Status: "running", Services: []RuntimeServiceSummary{
		{ServiceName: "apache", Status: "running", EligibleForLogs: true, EligibleForRestart: true},
	}}
	model := newShellModel(Plan(ctx), true)
	for index, action := range model.plan.DecisionItems {
		if action.ID == "restart_service" {
			model.cursor = index
			break
		}
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("opening restart confirmation should not run action")
	}
	next := updated.(shellModel)
	if !next.confirming {
		t.Fatal("expected restart confirmation")
	}
	view := next.View()
	for _, want := range []string{"Confirm restart", "Restart apache", "Restarts only apache; project files and settings are not changed."} {
		if !strings.Contains(view, want) {
			t.Fatalf("restart confirmation missing %q:\n%s", want, view)
		}
	}
}

func TestShellDoctorShortcutRunsFooterDiagnostics(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	model := newShellModel(Plan(ctx), true)
	called := false
	model.actionHandler = func(_ context.Context, action GuidedAction) (ActionResult, error) {
		called = true
		if action.ID != "doctor" {
			t.Fatalf("action id=%q want doctor", action.ID)
		}
		return ActionResult{
			Plan:    Plan(ctx),
			Utility: &UtilitySurface{Title: "Doctor report", Body: "StageServe Doctor", DismissOnAnyKey: true},
		}, nil
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	next := driveShellCommand(t, updated.(shellModel), cmd)
	if !called {
		t.Fatal("expected doctor shortcut to run")
	}
	if next.utility == nil || next.utility.Title != "Doctor report" {
		t.Fatalf("doctor utility=%+v", next.utility)
	}
}

func TestShellInitialEditorOptionStartsInEditMode(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectEnvExists = false
	model := newShellModelWithConfig(Plan(ctx), true, shellConfig{startEditing: true})

	if !model.editing {
		t.Fatal("expected initial editor mode")
	}
	view := model.View()
	for _, want := range []string{"Project settings", "Local URL", "Saving updates the preview only"} {
		if !strings.Contains(view, want) {
			t.Fatalf("initial editor view missing %q:\n%s", want, view)
		}
	}
}

func TestShellConfirmationRequiresExplicitConfirm(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectEnvExists = false
	plan := Plan(ctx)
	called := false

	model := newShellModel(plan, true)
	model.actionHandler = func(_ context.Context, action GuidedAction) (ActionResult, error) {
		called = true
		return ActionResult{Plan: Plan(baseContext()), Message: "Created project settings."}, nil
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("opening confirmation should not quit")
	}
	next, ok := updated.(shellModel)
	if !ok {
		t.Fatalf("updated model type %T", updated)
	}
	if called {
		t.Fatal("action ran before confirmation")
	}
	if !next.confirming {
		t.Fatal("expected confirmation state")
	}
	if view := next.View(); !strings.Contains(view, "Set up this directory as a project") || !strings.Contains(view, "enter confirm") {
		t.Fatalf("confirmation view missing expected copy:\n%s", view)
	}

	updated, cmd = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next = driveShellCommand(t, updated.(shellModel), cmd)
	if !called {
		t.Fatal("action did not run after confirmation")
	}
	if next.confirming {
		t.Fatal("confirmation state should close after action")
	}
	if next.plan.Situation != SituationProjectReadyToRun {
		t.Fatalf("situation=%s want %s", next.plan.Situation, SituationProjectReadyToRun)
	}
}

func TestShellConfirmationCanCancelWithoutAction(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectEnvExists = false
	model := newShellModel(Plan(ctx), true)
	model.actionHandler = func(context.Context, GuidedAction) (ActionResult, error) {
		t.Fatal("action should not run after cancel")
		return ActionResult{}, nil
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := updated.(shellModel)
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyEsc})
	next = updated.(shellModel)

	if next.confirming {
		t.Fatal("confirmation state should close after cancel")
	}
	if !strings.Contains(next.View(), "No changes made.") {
		t.Fatalf("cancel message missing:\n%s", next.View())
	}
}

func TestShellDriftRecoveryStopConfirmationUsesStopCopy(t *testing.T) {
	ctx := baseContext()
	ctx.Warnings = []string{"Project settings do not match the recorded project path."}
	model := newShellModel(Plan(ctx), true)
	for index, action := range model.plan.DecisionItems {
		if action.ID == "down" {
			model.cursor = index
			break
		}
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("opening confirmation should not quit")
	}
	next := updated.(shellModel)
	if !next.confirming {
		t.Fatal("expected confirmation state")
	}
	view := next.View()
	for _, want := range []string{"Confirm stop", "Step 5: stop this project first", "Stops http://demo.test; project files and settings are not changed.", "Default: cancel", "y confirm", "enter cancel"} {
		if !strings.Contains(view, want) {
			t.Fatalf("confirmation view missing %q:\n%s", want, view)
		}
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next = updated.(shellModel)
	if next.confirming {
		t.Fatal("confirmation state should close after cancel")
	}
	if !strings.Contains(next.View(), "No changes made. The project is still running.") {
		t.Fatalf("cancel message missing after recovery confirmation cancel:\n%s", next.View())
	}
}

func TestShellDestructiveConfirmationRequiresY(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	model := newShellModel(Plan(ctx), true)
	for index, action := range model.plan.DecisionItems {
		if action.ID == "down" {
			model.cursor = index
			break
		}
	}
	called := false
	model.actionHandler = func(_ context.Context, action GuidedAction) (ActionResult, error) {
		called = true
		if action.ID != "down" {
			t.Fatalf("action id=%q want down", action.ID)
		}
		return ActionResult{Plan: Plan(ctx), Message: "Project is stopped."}, nil
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := updated.(shellModel)
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next = updated.(shellModel)
	if called {
		t.Fatal("enter should cancel destructive confirmation")
	}
	if !strings.Contains(next.View(), "No changes made. The project is still running.") {
		t.Fatalf("destructive enter cancel missing copy:\n%s", next.View())
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next = updated.(shellModel)
	updated, cmd := next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	next = driveShellCommand(t, updated.(shellModel), cmd)
	if !called {
		t.Fatal("y should confirm destructive action")
	}
	if !strings.Contains(next.View(), "Project is stopped.") {
		t.Fatalf("confirmed action message missing:\n%s", next.View())
	}
}

func TestShellLifecycleActionShowsProgressWithinLatencyBudget(t *testing.T) {
	model := newShellModel(Plan(baseContext()), true)
	model.actionHandler = func(context.Context, GuidedAction) (ActionResult, error) {
		t.Fatal("handler should not run during initial update")
		return ActionResult{}, nil
	}

	started := time.Now()
	updated, cmd := model.runAction(GuidedAction{ID: "up", Label: "Run this project"})
	elapsed := time.Since(started)
	if cmd == nil {
		t.Fatal("expected async command")
	}
	if elapsed > 250*time.Millisecond {
		t.Fatalf("initial progress took %s, want <=250ms", elapsed)
	}
	next := updated.(shellModel)
	if !next.loading {
		t.Fatal("expected loading state")
	}
	view := next.View()
	for _, want := range []string{"Starting this project...", "Current step", "In progress", "esc cancel", "Press esc to request cancellation"} {
		if !strings.Contains(view, want) {
			t.Fatalf("loading view missing %q:\n%s", want, view)
		}
	}
}

func TestShellCancelRequestsContextCancellationAndIgnoresStaleResult(t *testing.T) {
	ctx := baseContext()
	model := newShellModel(Plan(ctx), true)
	canceled := false
	model.actionHandler = func(actionCtx context.Context, action GuidedAction) (ActionResult, error) {
		<-actionCtx.Done()
		canceled = true
		return ActionResult{Plan: Plan(ctx), Message: "stale success"}, nil
	}

	updated, cmd := model.runAction(GuidedAction{ID: "up", Label: "Run this project"})
	next := updated.(shellModel)
	batchMsg := cmd().(tea.BatchMsg)
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyEsc})
	canceledModel := updated.(shellModel)
	if canceledModel.loading {
		t.Fatal("loading should close after cancel")
	}
	if !strings.Contains(canceledModel.View(), "Cancellation requested") {
		t.Fatalf("cancel message missing:\n%s", canceledModel.View())
	}
	msg := batchMsg[1]().(actionCompleteMsg)
	if !canceled {
		t.Fatal("action context was not canceled")
	}
	updated, _ = canceledModel.Update(msg)
	finalModel := updated.(shellModel)
	if strings.Contains(finalModel.View(), "stale success") {
		t.Fatalf("stale action result should be ignored:\n%s", finalModel.View())
	}
}

func TestShellProjectSettingsEditorUpdatesPreviewAndActionInputs(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectEnvExists = false
	model := newShellModel(Plan(ctx), true)
	model.cursor = 1

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("opening edit form should not quit")
	}
	next, ok := updated.(shellModel)
	if !ok {
		t.Fatalf("updated model type %T", updated)
	}
	if !next.editing {
		t.Fatal("expected edit state")
	}
	if view := next.View(); !strings.Contains(view, "Project settings") || !strings.Contains(view, "Local URL") {
		t.Fatalf("edit view missing expected copy:\n%s", view)
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	next = updated.(shellModel)
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyTab})
	next = updated.(shellModel)
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyTab})
	next = updated.(shellModel)
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-local")})
	next = updated.(shellModel)
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next = updated.(shellModel)

	if next.editing {
		t.Fatal("edit state should close after save")
	}
	if view := next.View(); !strings.Contains(view, "http://demox.test-local") || !strings.Contains(view, "Nothing has been written yet") {
		t.Fatalf("saved preview missing expected values:\n%s", view)
	}

	var captured GuidedAction
	next.actionHandler = func(_ context.Context, action GuidedAction) (ActionResult, error) {
		captured = action
		return ActionResult{Plan: Plan(baseContext()), Message: "Created project settings."}, nil
	}
	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next = updated.(shellModel)
	updated, cmd = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next = driveShellCommand(t, updated.(shellModel), cmd)

	if captured.Inputs["site_name"] != "demox" {
		t.Fatalf("site_name input=%q", captured.Inputs["site_name"])
	}
	if captured.Inputs["docroot"] != "public_html" {
		t.Fatalf("docroot input=%q", captured.Inputs["docroot"])
	}
	if captured.Inputs["site_suffix"] != "test-local" {
		t.Fatalf("site_suffix input=%q", captured.Inputs["site_suffix"])
	}
	if next.plan.Situation != SituationProjectReadyToRun {
		t.Fatalf("situation=%s want %s", next.plan.Situation, SituationProjectReadyToRun)
	}
}

func TestShellUtilitySurfaceOpensAndCloses(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	plan := Plan(ctx)
	model := newShellModel(plan, true)
	model.actionHandler = func(context.Context, GuidedAction) (ActionResult, error) {
		return ActionResult{
			Plan: plan,
			Utility: &UtilitySurface{
				Title:  "demo logs",
				Body:   "10:42:13 GET / 200 12ms",
				Footer: "q exit logs • esc exit logs",
			},
		}, nil
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := driveShellCommand(t, updated.(shellModel), cmd)
	if next.utility == nil || !strings.Contains(next.View(), "demo logs") {
		t.Fatalf("utility view missing:\n%s", next.View())
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyEsc})
	next = updated.(shellModel)
	if next.utility != nil {
		t.Fatal("utility surface should close after esc")
	}
}
