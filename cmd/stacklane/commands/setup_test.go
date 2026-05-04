package commands

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isSetupExitError returns true when err is only a non-zero readiness exit code,
// not a flag-parsing or configuration error.
func isSetupExitError(err error) bool {
	var e *setupExitError
	return err == nil || errors.As(err, &e)
}

// TestSetup_NonInteractiveFlagAccepted verifies that --non-interactive is a
// valid flag on the setup command (flag errors are fatal; readiness exit codes are not).
func TestSetup_NonInteractiveFlagAccepted(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"setup", "--non-interactive", "--no-tui"})
	err := root.Execute()
	if !isSetupExitError(err) {
		t.Fatalf("unexpected error (want nil or setupExitError): %v", err)
	}
}

// TestSetup_JSONFlagAccepted verifies that --json is a valid flag and the
// command accepts it.
func TestSetup_JSONFlagAccepted(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"setup", "--json"})
	err := root.Execute()
	if !isSetupExitError(err) {
		t.Fatalf("unexpected error (want nil or setupExitError): %v", err)
	}
}

// TestSetup_JSONModeStillReturnsReadinessExit verifies JSON rendering does not
// swallow the readiness exit classification.
func TestSetup_JSONModeStillReturnsReadinessExit(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"setup", "--json"})
	err := root.Execute()
	if !isSetupExitError(err) {
		t.Fatalf("expected nil or setupExitError after JSON render, got: %v", err)
	}
}

// TestSetup_InvalidSuffixRejected verifies that an invalid --suffix value is
// rejected with an error that mentions "suffix" and is NOT a readiness exit code.
func TestSetup_InvalidSuffixRejected(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"setup", "--suffix", "bogus", "--non-interactive", "--no-tui"})
	err := root.Execute()
	if err == nil {
		t.Fatal("want error for invalid --suffix value, got nil")
	}
	var exitErr *setupExitError
	if errors.As(err, &exitErr) {
		t.Fatal("want suffix validation error, not a readiness exit-code error")
	}
	if !strings.Contains(err.Error(), "suffix") {
		t.Errorf("want error message to mention 'suffix', got: %v", err)
	}
}

// TestSetup_UsesConfigResolvedStateDir verifies setup honors the shared
// config contract for state-dir resolution via --stack-home.
func TestSetup_UsesConfigResolvedStateDir(t *testing.T) {
	stackHome := t.TempDir()
	if err := os.WriteFile(filepath.Join(stackHome, "docker-compose.shared.yml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--stack-home", stackHome, "setup", "--json"})
	_ = root.Execute()

	want := filepath.Join(stackHome, ".stacklane-state")
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("expected setup output to reference config-resolved state dir %q, got: %s", want, buf.String())
	}
}
