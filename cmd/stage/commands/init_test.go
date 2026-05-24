package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/onboarding"
)

// isInitExitError returns true for nil or a readiness-class exit error.
func isInitExitError(err error) bool {
	var e *initExitError
	return err == nil || errors.As(err, &e)
}

// TestInit_FlagsAccepted verifies that init flags are wired and accepted.
func TestInit_FlagsAccepted(t *testing.T) {
	dir := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--non-interactive", "--no-tui"})
	err := root.Execute()
	if !isInitExitError(err) {
		t.Fatalf("unexpected error (want nil or initExitError): %v", err)
	}
}

// TestInit_CreatesEnvFile verifies that running init in a fresh directory
// creates a .env.stageserve file.
func TestInit_CreatesEnvFile(t *testing.T) {
	dir := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--non-interactive", "--no-tui"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	envFile := filepath.Join(dir, ".env.stageserve")
	if _, err := os.Stat(envFile); err != nil {
		t.Errorf("expected .env.stageserve to be created, but stat failed: %v", err)
	}
}

func TestInit_WritesSiteSuffixInPlainMode(t *testing.T) {
	dir := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--site-suffix", "develop", "init", "--project-dir", dir, "--non-interactive", "--notui"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(dir, ".env.stageserve"))
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	if !strings.Contains(string(body), `SITE_SUFFIX="develop"`) {
		t.Fatalf("env file missing site suffix:\n%s", string(body))
	}
}

func TestInit_CLIPlainModeDoesNotRenderGuidedChoices(t *testing.T) {
	dir := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--cli"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, unwanted := range []string{"What you can do", "Project settings", "Edit before writing"} {
		if strings.Contains(out, unwanted) {
			t.Fatalf("plain mode rendered guided copy %q:\n%s", unwanted, out)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".env.stageserve")); err != nil {
		t.Fatalf("expected .env.stageserve to be created: %v", err)
	}
}

// TestInit_SkipsExistingWithoutForce verifies that init without --force does
// not overwrite an existing .env.stageserve.
func TestInit_SkipsExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env.stageserve")
	original := "# existing config\n"
	if err := os.WriteFile(envFile, []byte(original), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--non-interactive", "--no-tui"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(envFile)
	if string(got) != original {
		t.Error("expected existing .env.stageserve to be preserved, but content changed")
	}
}

// TestInit_OverwritesWithForce verifies that --force overwrites an existing file.
func TestInit_OverwritesWithForce(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env.stageserve")
	if err := os.WriteFile(envFile, []byte("# old\n"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--force", "--non-interactive", "--no-tui"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(envFile)
	if strings.Contains(string(got), "# old") {
		t.Error("expected --force to overwrite old .env.stageserve, but old content remains")
	}
}

// TestInit_JSONOutput verifies --json produces valid JSON.
func TestInit_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"overall_status"`) {
		t.Errorf("expected JSON output with overall_status, got: %s", out)
	}
}

func TestInit_NextStepsTextOutputDoesNotEmbedNewline(t *testing.T) {
	dir := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--non-interactive", "--no-tui"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Next: stage up\n") {
		t.Fatalf("expected clean next step, got: %q", out)
	}
	if strings.Contains(out, "stage up\n\n") {
		t.Fatalf("next step contains embedded newline: %q", out)
	}
}

func TestInit_NextStepsJSONOutputDoesNotEmbedNewline(t *testing.T) {
	dir := t.TempDir()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"init", "--project-dir", dir, "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result onboarding.CommandResult
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("json output did not unmarshal: %v\n%s", err, buf.String())
	}
	if got, want := fmt.Sprint(result.NextSteps), "[stage up]"; got != want {
		t.Fatalf("next_steps=%s, want %s", got, want)
	}
}
