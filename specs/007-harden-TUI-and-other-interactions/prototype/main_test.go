package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFixturesContainCanonicalPlannerSituations(t *testing.T) {
	plans := planFixtures()
	required := []situation{
		machineNotReady,
		projectMissingConfig,
		projectReadyToRun,
		projectRunning,
		projectDown,
		driftDetected,
		notProject,
		unknownError,
		doctorReportNeedsHelp,
	}
	for _, id := range required {
		if _, ok := plans[id]; !ok {
			t.Fatalf("missing prototype plan %q", id)
		}
	}
}

func TestEasyModeDoesNotExposeDiagnosticsAsDecisionItems(t *testing.T) {
	for id, plan := range planFixtures() {
		for _, item := range plan.Decisions {
			if strings.EqualFold(item.Label, "Find issues") || strings.EqualFold(item.Label, "Show commands") {
				t.Fatalf("%s exposes %q as a decision item", id, item.Label)
			}
		}
	}
}

func TestMachineSetupIsToolOwnedChecklist(t *testing.T) {
	plan := planFixtures()[machineNotReady]
	if len(plan.Decisions) != 0 {
		t.Fatalf("machine_not_ready decisions=%d want 0", len(plan.Decisions))
	}
	if len(plan.WorkItems) == 0 {
		t.Fatal("machine_not_ready has no work items")
	}
	if got := plan.WorkItems[plan.ActiveWorkIndex].Label; got != "Local DNS for .develop" {
		t.Fatalf("active work item=%q want Local DNS for .develop", got)
	}
}

func TestProjectSetupShowsVisibleDefaultsBeforeWrite(t *testing.T) {
	plan := planFixtures()[projectMissingConfig]
	assertContainsDefault(t, plan, "Site name", "pete-site")
	assertContainsDefault(t, plan, "Web folder", "./public_html")
	assertContainsDefault(t, plan, "Domain suffix", ".develop")
	assertContainsDefault(t, plan, "Local URL", "http://pete-site.develop")
	assertContainsDefault(t, plan, "Target file", prototypeProjectDir+"/.env.stageserve")
}

func TestEditedValuesCarryThroughProjectWorkflowCopy(t *testing.T) {
	values := editValues{SiteName: "client-demo", WebFolder: "./web", Suffix: ".test"}

	setup := projectSetupPlan(prototypeFooter(), values)
	assertContainsDefault(t, setup, "Site name", "client-demo")
	assertContainsDefault(t, setup, "Web folder", "./web")
	assertContainsDefault(t, setup, "Domain suffix", ".test")
	assertContainsDefault(t, setup, "Local URL", "http://client-demo.test")
	setupConfirm := strings.Join(setup.Decisions[0].ConfirmBody, "\n")
	for _, want := range []string{
		"Site name: client-demo",
		"Web folder: ./web",
		"Domain suffix: .test",
		"Local URL: http://client-demo.test",
		prototypeProjectDir + "/.env.stageserve",
	} {
		if !strings.Contains(setupConfirm, want) {
			t.Fatalf("setup confirmation missing %q:\n%s", want, setupConfirm)
		}
	}

	running := projectRunningPlan(prototypeFooter(), values)
	if running.StatusHeader != "client-demo is running" {
		t.Fatalf("running status header=%q", running.StatusHeader)
	}
	stopConfirm := strings.Join(running.Decisions[1].ConfirmBody, "\n")
	for _, want := range []string{"Stop client-demo?", "http://client-demo.test"} {
		if strings.Contains(stopConfirm, want) {
			continue
		}
		if want == "Stop client-demo?" {
			if running.Decisions[1].ConfirmTitle != want {
				t.Fatalf("running confirm title=%q want %q", running.Decisions[1].ConfirmTitle, want)
			}
			continue
		}
		t.Fatalf("running stop confirmation missing %q:\n%s", want, stopConfirm)
	}
	if !strings.Contains(strings.Join(running.Details, "\n"), "./web") {
		t.Fatalf("running details missing edited web folder: %v", running.Details)
	}

	drift := driftDetectedPlan(prototypeFooter(), values)
	if !strings.Contains(drift.Summary, "http://client-demo.test") {
		t.Fatalf("drift summary missing edited URL: %s", drift.Summary)
	}
	if !strings.Contains(strings.Join(drift.Details, "\n"), "client-demo.test") {
		t.Fatalf("drift details missing edited host: %v", drift.Details)
	}
}

func TestRunningProjectDefaultIsNonDestructive(t *testing.T) {
	plan := planFixtures()[projectRunning]
	if len(plan.Decisions) == 0 {
		t.Fatal("project_running has no decisions")
	}
	first := plan.Decisions[0]
	if first.ID != "logs" {
		t.Fatalf("running default id=%q want logs", first.ID)
	}
	if first.Mutates {
		t.Fatal("running default mutates state")
	}
	if strings.Contains(strings.ToLower(first.Label), "stop") {
		t.Fatalf("running default label=%q must not stop", first.Label)
	}
}

func TestTextFallbackUsesSurfaceLanguage(t *testing.T) {
	var b strings.Builder
	renderText(&b, planFixtures()[projectMissingConfig])
	text := b.String()
	for _, want := range []string{
		"StageServe Project setup",
		"This folder doesn't have StageServe settings yet.",
		"Key facts",
		"http://pete-site.develop",
		"What you can do",
		"Use these settings",
		"More…",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("text fallback missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "Visible defaults") || strings.Contains(text, "Decision bar") || strings.Contains(text, "Highlighted default") {
		t.Fatalf("text fallback uses old action model:\n%s", text)
	}
}

func TestGuidedPlansExposeMoreAsSecondaryAction(t *testing.T) {
	for _, id := range []situation{projectMissingConfig, projectReadyToRun, projectRunning, projectDown, driftDetected, unknownError} {
		plan := planFixtures()[id]
		found := false
		for _, item := range plan.Decisions {
			if item.Label == "More…" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s is missing More… as a secondary action", id)
		}
	}
}

func TestDoctorReportOffersAssistanceWithoutHidingCommands(t *testing.T) {
	plans := planFixtures()
	plan, ok := plans[doctorReportNeedsHelp]
	if !ok {
		t.Fatal("missing doctor report assistance scenario")
	}

	var b strings.Builder
	renderText(&b, plan)
	text := b.String()

	for _, want := range []string{
		"StageServe Doctor",
		"Not ready - 2 of 7 checks need attention.",
		"Needs fixing",
		"Port 443",
		"Something else on your computer is using port 443.",
		"StageServe needs elevated permission to identify the process.",
		"To fix: sudo lsof -nP -iTCP:443 -sTCP:LISTEN",
		"Local DNS resolver",
		"Your computer cannot yet open local project URLs.",
		"Local DNS is not set up yet.",
		"To fix: stage setup",
		"Assistance",
		"Help me fix these",
		"Walk through each issue one at a time.",
		"Leave it here",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("doctor assistance fallback missing %q:\n%s", want, text)
		}
	}
	for _, unwanted := range []string{"Highlighted default", "Decision bar"} {
		if strings.Contains(text, unwanted) {
			t.Fatalf("doctor assistance fallback should omit %q:\n%s", unwanted, text)
		}
	}
}

func TestAssistedDoctorFlowStartsWithOneFocusedBlocker(t *testing.T) {
	m := newModel(planFixtures(), doctorReportNeedsHelp)
	next := m.handleDecision(m.currentPlan().Decisions[0])

	if next.mode != modeAssist {
		t.Fatalf("mode=%v want modeAssist", next.mode)
	}
	view := next.View()
	for _, want := range []string{
		"Port 443",
		"Something else on your computer is using port 443.",
		"Check with sudo",
		"Run a read-only command to identify the process.",
		"Skip this issue",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("assist view missing %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Local DNS resolver") {
		t.Fatalf("assist view should focus on one blocker at a time:\n%s", view)
	}
}

func TestDoctorLeaveItHereReturnsQuitCommand(t *testing.T) {
	m := newModel(planFixtures(), doctorReportNeedsHelp)
	m.cursor = 1

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("leave decision should return a quit command")
	}
	next, ok := updated.(model)
	if !ok {
		t.Fatalf("updated model has unexpected type %T", updated)
	}
	if !next.currentPlan().Decisions[1].Quits {
		t.Fatal("leave decision should be marked as quitting")
	}
	if msg := cmd(); msg == nil {
		t.Fatal("quit command should return a quit message")
	} else if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("quit command returned %T want tea.QuitMsg", msg)
	}
}

func TestAssistEnterOpensReadOnlySudoConfirmation(t *testing.T) {
	m := newModel(planFixtures(), doctorReportNeedsHelp)
	m.mode = modeAssist

	updated, cmd := m.updateAssist("enter")
	if cmd != nil {
		t.Fatalf("assist enter returned unexpected command %v", cmd)
	}
	next, ok := updated.(model)
	if !ok {
		t.Fatalf("updated model has unexpected type %T", updated)
	}
	if next.mode != modeConfirm {
		t.Fatalf("mode=%v want modeConfirm", next.mode)
	}
	if !next.confirmYes {
		t.Fatal("sudo confirmation should default to yes")
	}
	if next.pending.Title != "Check port 443 with sudo?" {
		t.Fatalf("title=%q", next.pending.Title)
	}
	if next.pending.YesLabel != "Yes, check with sudo" {
		t.Fatalf("yes label=%q", next.pending.YesLabel)
	}
	if next.pending.NoLabel != "No, go back" {
		t.Fatalf("no label=%q", next.pending.NoLabel)
	}
	if !next.pending.YesDefault {
		t.Fatal("yes default should be true")
	}
	if next.pending.ResultTitle != "Read-only check approved" {
		t.Fatalf("result title=%q", next.pending.ResultTitle)
	}
	if next.pending.ResultBody != "Prototype only: StageServe would run sudo lsof to identify the process using port 443." {
		t.Fatalf("result body=%q", next.pending.ResultBody)
	}
	body := strings.Join(next.pending.Body, "\n")
	for _, want := range []string{
		"StageServe will run a read-only command to identify what is using port 443.",
		"Your computer will ask for your password because macOS hides this detail by default.",
		"Command: sudo lsof -nP -iTCP:443 -sTCP:LISTEN",
		"Prototype only: no command will be run.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("confirmation body missing %q:\n%s", want, body)
		}
	}
}

func TestDoctorReportReadySummariesMatchInTUIAndText(t *testing.T) {
	plan := planFixtures()[doctorReportNeedsHelp]
	tui := newModel(planFixtures(), doctorReportNeedsHelp).renderMain()

	var b strings.Builder
	renderText(&b, plan)
	text := b.String()

	for _, want := range []string{"running", "exists", "available", "installed"} {
		if !strings.Contains(tui, want) {
			t.Fatalf("tui output missing ready summary %q:\n%s", want, tui)
		}
		if !strings.Contains(text, want) {
			t.Fatalf("text output missing ready summary %q:\n%s", want, text)
		}
	}
	if strings.Contains(tui, "Docker Desktop       ready") {
		t.Fatalf("tui output should prefer fixture messages over generic status:\n%s", tui)
	}
}

func TestRenderTextOmitsNeedsFixingWhenOnlyReadyChecksExist(t *testing.T) {
	var b strings.Builder
	renderText(&b, plan{
		StatusHeader: "Doctor",
		ReportReady: []reportItem{
			{Label: "Docker Desktop", Status: statusReady, Message: "running"},
		},
	})
	text := b.String()

	if strings.Contains(text, "Needs fixing") {
		t.Fatalf("text output should not show Needs fixing without attention items:\n%s", text)
	}
	if !strings.Contains(text, "All clear") {
		t.Fatalf("text output missing All clear section:\n%s", text)
	}
}

func TestRenderMainKeepsProjectVerdictOutOfHeaderChrome(t *testing.T) {
	plan := planFixtures()[projectReadyToRun]
	view := newModel(planFixtures(), projectReadyToRun).renderMain()
	lines := strings.Split(view, "\n")
	if len(lines) == 0 {
		t.Fatal("renderMain returned no lines")
	}
	if strings.Contains(lines[0], plan.StatusHeader) {
		t.Fatalf("header line should not contain verdict text:\n%s", lines[0])
	}
	if !strings.Contains(lines[0], "Project") {
		t.Fatalf("header line should contain the surface label:\n%s", lines[0])
	}
	for _, want := range []string{
		plan.StatusHeader,
		"Key facts",
		"What you can do",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("project render missing %q:\n%s", want, view)
		}
	}
}

func TestProjectRunningTUIFooterIncludesBrowserHint(t *testing.T) {
	view := newModel(planFixtures(), projectRunning).renderMain()
	if !strings.Contains(view, "right open in browser") {
		t.Fatalf("running project footer missing browser hint:\n%s", view)
	}
}

func assertContainsDefault(t *testing.T, plan plan, label, value string) {
	t.Helper()
	for _, item := range plan.Defaults {
		if item.Label == label && item.Value == value {
			return
		}
	}
	t.Fatalf("%s default %q=%q not found in %#v", plan.Situation, label, value, plan.Defaults)
}
