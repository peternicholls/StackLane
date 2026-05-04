package commands

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
// creates a .env.stacklane file.
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
	envFile := filepath.Join(dir, ".env.stacklane")
	if _, err := os.Stat(envFile); err != nil {
		t.Errorf("expected .env.stacklane to be created, but stat failed: %v", err)
	}
}

// TestInit_SkipsExistingWithoutForce verifies that init without --force does
// not overwrite an existing .env.stacklane.
func TestInit_SkipsExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env.stacklane")
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
		t.Error("expected existing .env.stacklane to be preserved, but content changed")
	}
}

// TestInit_OverwritesWithForce verifies that --force overwrites an existing file.
func TestInit_OverwritesWithForce(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env.stacklane")
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
		t.Error("expected --force to overwrite old .env.stacklane, but old content remains")
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
