package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
	"github.com/peternicholls/stageserve/core/onboarding"
)

type fakeGuidedLifecycleRunner struct {
	upCalls     int
	attachCalls int
	downCalls   int
	detachCalls int
	statusBody  string
	logsBody    string
}

func (f *fakeGuidedLifecycleRunner) Up(context.Context, config.ProjectConfig) error {
	f.upCalls++
	return nil
}

func (f *fakeGuidedLifecycleRunner) Attach(context.Context, config.ProjectConfig) error {
	f.attachCalls++
	return nil
}

func (f *fakeGuidedLifecycleRunner) Down(context.Context, config.ProjectConfig, bool) error {
	f.downCalls++
	return nil
}

func (f *fakeGuidedLifecycleRunner) Detach(context.Context, config.ProjectConfig) error {
	f.detachCalls++
	return nil
}

func (f *fakeGuidedLifecycleRunner) Status(context.Context, config.ProjectConfig) (string, error) {
	if f.statusBody == "" {
		return "demo (attached) — demo.test", nil
	}
	return f.statusBody, nil
}

func (f *fakeGuidedLifecycleRunner) Logs(context.Context, config.ProjectConfig, string) (string, error) {
	if f.logsBody == "" {
		return "10:42:13 GET / 200 12ms", nil
	}
	return f.logsBody, nil
}

func TestRootNoArgsPrintsGuidanceWithoutMutatingProject(t *testing.T) {
	projectDir := t.TempDir()
	stackHome := t.TempDir()
	stateDir := filepath.Join(stackHome, ".stageserve-state")

	// Pre-create the state directory to simulate a set-up machine so the cheap
	// readiness heuristic does not fire and the test exercises the
	// project_missing_config path.
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("create state dir: %v", err)
	}

	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--stack-home", stackHome, "--project-dir", projectDir, "--notui"})

	if err := root.Execute(); err != nil {
		t.Fatalf("root guidance: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"StageServe", "This folder doesn't have StageServe settings yet.", "stage init"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".env.stageserve")); !os.IsNotExist(err) {
		t.Fatalf("bare stage should not create .env.stageserve, stat err=%v", err)
	}
	// State dir was pre-created but bare stage should not write project files inside it.
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("read state dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("bare stage should not write project files into state dir, found: %v", entries)
	}
}

func TestRootCLIFlagUsesGuidanceFallback(t *testing.T) {
	projectDir := t.TempDir()
	stackHome := t.TempDir()

	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--stack-home", stackHome, "--project-dir", projectDir, "--cli"})

	if err := root.Execute(); err != nil {
		t.Fatalf("root guidance: %v", err)
	}
	if !strings.Contains(buf.String(), "Direct commands:") {
		t.Fatalf("expected plain guidance output, got:\n%s", buf.String())
	}
}

func TestGuidancePlanCommandEmitsInspectablePlanJSON(t *testing.T) {
	projectDir := t.TempDir()
	stackHome := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"guidance-plan", "--stack-home", stackHome, "--project-dir", projectDir, "--skip-readiness"})

	if err := root.Execute(); err != nil {
		t.Fatalf("guidance-plan: %v", err)
	}

	var view struct {
		Situation     string `json:"situation"`
		StatusHeader  string `json:"status_header"`
		DecisionItems []struct {
			ID    string `json:"id"`
			Label string `json:"label"`
		} `json:"decision_items"`
	}
	if err := json.Unmarshal(buf.Bytes(), &view); err != nil {
		t.Fatalf("guidance-plan output is not valid json: %v\n%s", err, buf.String())
	}
	if view.Situation != string(guidance.SituationProjectMissingConf) {
		t.Fatalf("situation=%q want %q", view.Situation, guidance.SituationProjectMissingConf)
	}
	if view.StatusHeader != "This folder doesn't have StageServe settings yet." {
		t.Fatalf("status header=%q", view.StatusHeader)
	}
	if len(view.DecisionItems) == 0 || view.DecisionItems[0].ID != "init" {
		t.Fatalf("decision items=%+v", view.DecisionItems)
	}
}

func TestGuidedInitActionWritesEnvAndReplans(t *testing.T) {
	projectDir := t.TempDir()
	cfg := config.ProjectConfig{
		Name:            "demo",
		Slug:            "demo",
		Dir:             projectDir,
		StateDir:        filepath.Join(t.TempDir(), "state"),
		StackKind:       "20i",
		StackHome:       t.TempDir(),
		Hostname:        "demo.test",
		SiteSuffix:      "test",
		DocRoot:         filepath.Join(projectDir, "public_html"),
		DocRootRelative: "public_html",
		SharedGateway:   config.SharedGateway{HTTPSPort: 443},
	}

	result, err := handleGuidedAction(context.Background(), cfg, nil, guidance.TUICapability{}, guidance.GuidedAction{ID: "init"})
	if err != nil {
		t.Fatalf("handleGuidedAction: %v", err)
	}
	if !strings.Contains(result.Message, "Created project settings") {
		t.Fatalf("message=%q", result.Message)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".env.stageserve")); err != nil {
		t.Fatalf("expected .env.stageserve to be written: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(projectDir, ".env.stageserve"))
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	if !strings.Contains(string(body), `SITE_SUFFIX="test"`) {
		t.Fatalf("expected site suffix in env file:\n%s", string(body))
	}
	if result.Plan.Situation != guidance.SituationProjectReadyToRun {
		t.Fatalf("situation=%s want %s", result.Plan.Situation, guidance.SituationProjectReadyToRun)
	}
	if len(result.Plan.DecisionItems) == 0 || result.Plan.DecisionItems[0].Label != "Run this project" {
		t.Fatalf("expected run action after init, got %+v", result.Plan.DecisionItems)
	}
}

func TestGuidedOverwriteInitActionOverwritesEnvAndReplans(t *testing.T) {
	projectDir := t.TempDir()
	envFile := filepath.Join(projectDir, ".env.stageserve")
	if err := os.WriteFile(envFile, []byte("# old settings\n"), 0o600); err != nil {
		t.Fatalf("seed env file: %v", err)
	}
	cfg := config.ProjectConfig{
		Name:            "demo",
		Slug:            "demo",
		Dir:             projectDir,
		StateDir:        filepath.Join(t.TempDir(), "state"),
		StackKind:       "20i",
		StackHome:       t.TempDir(),
		Hostname:        "demo.test",
		SiteSuffix:      "test",
		DocRoot:         filepath.Join(projectDir, "public_html"),
		DocRootRelative: "public_html",
		SharedGateway:   config.SharedGateway{HTTPSPort: 443},
	}

	result, err := handleGuidedAction(context.Background(), cfg, nil, guidance.TUICapability{}, guidance.GuidedAction{ID: "overwrite_init"})
	if err != nil {
		t.Fatalf("handleGuidedAction: %v", err)
	}
	if !strings.Contains(result.Message, "Updated project settings") {
		t.Fatalf("message=%q", result.Message)
	}
	body, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	if strings.Contains(string(body), "# old settings") {
		t.Fatalf("expected .env.stageserve to be overwritten, got:\n%s", string(body))
	}
	if result.Plan.Situation != guidance.SituationProjectReadyToRun {
		t.Fatalf("situation=%s want %s", result.Plan.Situation, guidance.SituationProjectReadyToRun)
	}
}

func TestExecuteGuidedActionRouting(t *testing.T) {
	projectDir := t.TempDir()
	cfg := config.ProjectConfig{
		Name:            "demo",
		Slug:            "demo",
		Dir:             projectDir,
		StateDir:        filepath.Join(t.TempDir(), "state"),
		StackKind:       "20i",
		StackHome:       t.TempDir(),
		Hostname:        "demo.test",
		SiteSuffix:      "test",
		DocRoot:         filepath.Join(projectDir, "public_html"),
		DocRootRelative: "public_html",
		SharedGateway:   config.SharedGateway{HTTPSPort: 443},
	}

	runner := &fakeGuidedLifecycleRunner{}

	if err := executeGuidedAction("up", context.Background(), cfg, runner); err != nil {
		t.Fatalf("up action: %v", err)
	}
	if runner.upCalls != 1 || runner.attachCalls != 0 {
		t.Fatalf("unexpected up routing counts: %+v", runner)
	}

	if err := executeGuidedAction("attach", context.Background(), cfg, runner); err != nil {
		t.Fatalf("attach action: %v", err)
	}
	if runner.upCalls != 1 || runner.attachCalls != 1 {
		t.Fatalf("unexpected attach routing counts: %+v", runner)
	}

	if err := executeGuidedAction("down", context.Background(), cfg, runner); err != nil {
		t.Fatalf("down action: %v", err)
	}
	if runner.downCalls != 1 {
		t.Fatalf("unexpected down routing counts: %+v", runner)
	}

	if err := executeGuidedAction("detach", context.Background(), cfg, runner); err != nil {
		t.Fatalf("detach action: %v", err)
	}
	if runner.detachCalls != 1 {
		t.Fatalf("unexpected detach routing counts: %+v", runner)
	}

	err := executeGuidedAction("unknown", context.Background(), cfg, nil)
	if err == nil {
		t.Fatalf("expected error for unknown action")
	}
	if err.Error() != "action \"unknown\" not implemented" {
		t.Fatalf("wrong error message for unknown action: %v", err)
	}
	if err := executeGuidedAction("up", context.Background(), cfg, nil); err == nil || err.Error() != "guided lifecycle runner is not available" {
		t.Fatalf("nil runner error = %v", err)
	}
}

func TestHandleGuidedActionReturnsUtilityForStatusAndLogs(t *testing.T) {
	projectDir := t.TempDir()
	cfg := config.ProjectConfig{
		Name:            "demo",
		Slug:            "demo",
		Dir:             projectDir,
		StateDir:        filepath.Join(t.TempDir(), "state"),
		StackKind:       "20i",
		StackHome:       t.TempDir(),
		Hostname:        "demo.test",
		SiteSuffix:      "test",
		DocRoot:         filepath.Join(projectDir, "public_html"),
		DocRootRelative: "public_html",
		SharedGateway:   config.SharedGateway{HTTPSPort: 443},
	}
	runner := &fakeGuidedLifecycleRunner{}

	statusResult, err := handleGuidedAction(context.Background(), cfg, runner, guidance.TUICapability{}, guidance.GuidedAction{ID: "status"})
	if err != nil {
		t.Fatalf("status action: %v", err)
	}
	if statusResult.Utility == nil || statusResult.Utility.Title != "demo status" {
		t.Fatalf("status utility=%+v", statusResult.Utility)
	}

	logsResult, err := handleGuidedAction(context.Background(), cfg, runner, guidance.TUICapability{}, guidance.GuidedAction{ID: "logs"})
	if err != nil {
		t.Fatalf("logs action: %v", err)
	}
	if logsResult.Utility == nil || logsResult.Utility.Title != "demo logs" {
		t.Fatalf("logs utility=%+v", logsResult.Utility)
	}
}

func TestHandleGuidedActionReturnsUtilityForSetup(t *testing.T) {
	projectDir := t.TempDir()
	cfg := config.ProjectConfig{
		Name:            "demo",
		Slug:            "demo",
		Dir:             projectDir,
		StateDir:        filepath.Join(t.TempDir(), "state"),
		StackKind:       "20i",
		StackHome:       t.TempDir(),
		Hostname:        "demo.test",
		SiteSuffix:      "test",
		DocRoot:         filepath.Join(projectDir, "public_html"),
		DocRootRelative: "public_html",
		SharedGateway:   config.SharedGateway{HTTPSPort: 443},
	}
	oldBuilder := buildGuidedSetupResult
	buildGuidedSetupResult = func(config.ProjectConfig) onboarding.CommandResult {
		return onboarding.BuildResult([]onboarding.StepResult{{
			ID:      "docker.daemon",
			Label:   "Docker daemon",
			Status:  onboarding.StatusNeedsAction,
			Message: "Docker daemon is not reachable",
		}}, nil, nil)
	}
	defer func() {
		buildGuidedSetupResult = oldBuilder
	}()

	result, err := handleGuidedAction(context.Background(), cfg, nil, guidance.TUICapability{}, guidance.GuidedAction{ID: "setup"})
	if err != nil {
		t.Fatalf("setup action: %v", err)
	}
	if result.Utility == nil || result.Utility.Title != "Setup report" {
		t.Fatalf("setup utility=%+v", result.Utility)
	}
	if !strings.Contains(result.Utility.Body, "Docker daemon") {
		t.Fatalf("setup report missing expected check:\n%s", result.Utility.Body)
	}
}

func TestHandleGuidedActionReturnsUtilityForDoctor(t *testing.T) {
	projectDir := t.TempDir()
	cfg := config.ProjectConfig{
		Name:            "demo",
		Slug:            "demo",
		Dir:             projectDir,
		StateDir:        filepath.Join(t.TempDir(), "state"),
		StackKind:       "20i",
		StackHome:       t.TempDir(),
		Hostname:        "demo.test",
		SiteSuffix:      "test",
		DocRoot:         filepath.Join(projectDir, "public_html"),
		DocRootRelative: "public_html",
		SharedGateway:   config.SharedGateway{HTTPSPort: 443},
	}
	oldBuilder := buildGuidedDoctorResult
	buildGuidedDoctorResult = func(config.ProjectConfig) onboarding.CommandResult {
		return onboarding.BuildResult([]onboarding.StepResult{{
			ID:      "port.443",
			Label:   "Port 443",
			Status:  onboarding.StatusNeedsAction,
			Message: "port 443 is already in use",
		}}, nil, nil)
	}
	defer func() {
		buildGuidedDoctorResult = oldBuilder
	}()

	result, err := handleGuidedAction(context.Background(), cfg, nil, guidance.TUICapability{}, guidance.GuidedAction{ID: "doctor"})
	if err != nil {
		t.Fatalf("doctor action: %v", err)
	}
	if result.Utility == nil || result.Utility.Title != "Doctor report" {
		t.Fatalf("doctor utility=%+v", result.Utility)
	}
	if !strings.Contains(result.Utility.Body, "StageServe Doctor") || !strings.Contains(result.Utility.Body, "Port 443") {
		t.Fatalf("doctor report missing expected content:\n%s", result.Utility.Body)
	}
}
