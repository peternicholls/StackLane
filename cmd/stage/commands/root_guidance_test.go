package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/guidance"
)

func TestRootNoArgsPrintsGuidanceWithoutMutatingProject(t *testing.T) {
	projectDir := t.TempDir()
	stackHome := t.TempDir()
	stateDir := filepath.Join(stackHome, ".stageserve-state")

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
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Fatalf("bare stage should not create state dir, stat err=%v", err)
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

	plan, message, err := handleGuidedAction(context.Background(), cfg, guidance.TUICapability{}, guidance.GuidedAction{ID: "init"})
	if err != nil {
		t.Fatalf("handleGuidedAction: %v", err)
	}
	if !strings.Contains(message, "Created project settings") {
		t.Fatalf("message=%q", message)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".env.stageserve")); err != nil {
		t.Fatalf("expected .env.stageserve to be written: %v", err)
	}
	if plan.Situation != guidance.SituationProjectReadyToRun {
		t.Fatalf("situation=%s want %s", plan.Situation, guidance.SituationProjectReadyToRun)
	}
	if len(plan.DecisionItems) == 0 || plan.DecisionItems[0].Label != "Run this project" {
		t.Fatalf("expected run action after init, got %+v", plan.DecisionItems)
	}
}
