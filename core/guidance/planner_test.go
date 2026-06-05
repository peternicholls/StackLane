package guidance

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

	plan := Plan(ctx)

	for _, item := range plan.DecisionItems {
		if item.ID == "detach" {
			t.Fatalf("running-project choices should not expose detach directly: %+v", plan.DecisionItems)
		}
	}
	if len(plan.DirectCommands) != 3 || plan.DirectCommands[2] != "stage down" {
		t.Fatalf("direct commands=%+v", plan.DirectCommands)
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
	for _, want := range []string{"Add this project to StageServe", "Check this project's status", "Direct commands:", "stage attach", "stage down"} {
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

func TestPlanUnknownErrorExposesOrderedRecoverySteps(t *testing.T) {
	plan := Plan(ContextFromError("/tmp/demo", TUICapability{}, os.ErrInvalid))

	if plan.Situation != SituationUnknownError {
		t.Fatalf("situation=%s want %s", plan.Situation, SituationUnknownError)
	}
	if len(plan.WorkItems) != 4 {
		t.Fatalf("work items=%+v", plan.WorkItems)
	}
	if plan.WorkItems[0].DirectCommand != "stage status" || plan.WorkItems[1].DirectCommand != "stage logs" {
		t.Fatalf("recovery order=%+v", plan.WorkItems)
	}
	if len(plan.DecisionItems) != 4 || plan.DecisionItems[0].ID != "status" || plan.DecisionItems[1].ID != "logs" {
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
	model.actionHandler = func(action GuidedAction) (ActionResult, error) {
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
	if next.utility == nil || next.utility.Title != "Direct commands" {
		t.Fatalf("command utility=%+v", next.utility)
	}
	view := next.View()
	for _, want := range []string{"Direct commands", "stage up", "stage status"} {
		if !strings.Contains(view, want) {
			t.Fatalf("command utility missing %q:\n%s", want, view)
		}
	}
}

func TestShellDoctorShortcutRunsFooterDiagnostics(t *testing.T) {
	ctx := baseContext()
	ctx.ProjectState = &state.Record{AttachmentState: state.StateAttached}
	model := newShellModel(Plan(ctx), true)
	called := false
	model.actionHandler = func(action GuidedAction) (ActionResult, error) {
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
	model.actionHandler = func(action GuidedAction) (ActionResult, error) {
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
	if view := next.View(); !strings.Contains(view, "Create project settings") || !strings.Contains(view, "enter confirm") {
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
	model.actionHandler = func(action GuidedAction) (ActionResult, error) {
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
	model.cursor = 3

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("opening confirmation should not quit")
	}
	next := updated.(shellModel)
	if !next.confirming {
		t.Fatal("expected confirmation state")
	}
	view := next.View()
	for _, want := range []string{"Step 4: stop this project first", "StageServe will stop this project.", "Your files will not be touched.", "http://demo.test will no longer respond until you run it again."} {
		if !strings.Contains(view, want) {
			t.Fatalf("confirmation view missing %q:\n%s", want, view)
		}
	}

	updated, _ = next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")})
	next = updated.(shellModel)
	if next.confirming {
		t.Fatal("confirmation state should close after cancel")
	}
	if !strings.Contains(next.View(), "No changes made.") {
		t.Fatalf("cancel message missing after recovery confirmation cancel:\n%s", next.View())
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
	next.actionHandler = func(action GuidedAction) (ActionResult, error) {
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
	model.actionHandler = func(action GuidedAction) (ActionResult, error) {
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
