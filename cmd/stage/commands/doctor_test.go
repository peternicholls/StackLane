package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// isDoctorExitError returns true for nil or a readiness exit error from doctor.
func isDoctorExitError(err error) bool {
	var e *doctorExitError
	return err == nil || errors.As(err, &e)
}

// TestDoctor_FlagsAccepted verifies that the doctor command accepts its flags
// without flag-parse errors.
func TestDoctor_FlagsAccepted(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--non-interactive", "--no-tui"})
	err := root.Execute()
	if !isDoctorExitError(err) {
		t.Fatalf("unexpected error (want nil or doctorExitError): %v", err)
	}
}

// TestDoctor_JSONFlagAccepted verifies --json is a valid flag.
func TestDoctor_JSONFlagAccepted(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--json"})
	err := root.Execute()
	if !isDoctorExitError(err) {
		t.Fatalf("unexpected error (want nil or doctorExitError): %v", err)
	}
}

// TestDoctor_JSONModeStillReturnsReadinessExit verifies JSON rendering does not
// swallow the readiness exit classification.
func TestDoctor_JSONModeStillReturnsReadinessExit(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--json"})
	err := root.Execute()
	if !isDoctorExitError(err) {
		t.Fatalf("expected nil or doctorExitError after JSON render, got: %v", err)
	}
}

// TestDoctor_JSONOutputShape verifies that --json emits a JSON envelope.
func TestDoctor_JSONOutputShape(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--json"})
	root.Execute() //nolint:errcheck — exit code is expected; we only care about output shape
	out := buf.String()
	if !strings.Contains(out, `"overall_status"`) {
		t.Errorf("expected JSON output with overall_status, got: %s", out)
	}
}

func TestDoctor_JSONOutputIsMachinePure(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--json"})
	_ = root.Execute()

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("doctor JSON output did not unmarshal: %v\n%s", err, buf.String())
	}
	assertMachineOutputHasNoTUIHints(t, buf.String())
}

// TestDoctor_TextOutputShape verifies that plain-text output contains
// at least one step label.
func TestDoctor_TextOutputShape(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--no-tui", "--non-interactive"})
	root.Execute() //nolint:errcheck
	out := buf.String()
	if !strings.Contains(out, "Docker") {
		t.Errorf("expected text output to mention 'Docker', got: %s", out)
	}
}

func TestDoctor_NonInteractiveOutputStaysPlain(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"doctor", "--non-interactive", "--notui"})
	_ = root.Execute()

	out := buf.String()
	assertMachineOutputHasNoTUIHints(t, out)
	if !strings.Contains(out, "StageServe Doctor") {
		t.Fatalf("doctor fallback missing surface header:\n%s", out)
	}
	if !strings.Contains(out, "Needs fixing") && !strings.Contains(out, "Checks passed") {
		t.Fatalf("doctor fallback missing report focus section:\n%s", out)
	}
}

// TestDoctor_UsesConfigResolvedStateDir verifies doctor honors the shared
// config contract for state-dir resolution via --stack-home.
func TestDoctor_UsesConfigResolvedStateDir(t *testing.T) {
	stackHome := t.TempDir()
	if err := os.MkdirAll(filepath.Join(stackHome, "stacks", "20i"), 0o755); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stackHome, "stacks", "20i", "docker-compose.shared.yml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stackHome, "stacks", "20i", "docker-compose.20i.yml"), []byte("services: {}\n"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--stack-home", stackHome, "doctor", "--json"})
	_ = root.Execute()

	want := filepath.Join(stackHome, ".stageserve-state")
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("expected doctor output to reference config-resolved state dir %q, got: %s", want, buf.String())
	}
}
