package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
