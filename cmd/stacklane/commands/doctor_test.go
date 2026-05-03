package commands

import (
	"bytes"
	"errors"
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
