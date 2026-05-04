//go:build !darwin && !linux

package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

// TestSetup_UnsupportedOSExitCode verifies that setup on an unsupported platform
// produces exit_code: 3 and overall_status: error in the JSON output.
func TestSetup_UnsupportedOSExitCode(t *testing.T) {
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"setup", "--json"})
	err := root.Execute()

	// On unsupported OS, we expect a setupExitError with code 3
	var setupErr *setupExitError
	if err == nil {
		t.Fatalf("expected setupExitError with code 3, got nil")
	}
	if !errors.As(err, &setupErr) {
		t.Fatalf("expected setupExitError, got: %T: %v", err, err)
	}
	if setupErr.code != 3 {
		t.Fatalf("expected exit code 3 (unsupported OS), got %d", setupErr.code)
	}

	// Verify JSON output contains the expected structure
	output := buf.String()
	var result struct {
		ExitCode      int    `json:"exit_code"`
		OverallStatus string `json:"overall_status"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("failed to unmarshal JSON output: %v\noutput: %s", err, output)
	}

	if result.ExitCode != 3 {
		t.Errorf("expected exit_code: 3 in JSON, got: %d", result.ExitCode)
	}
	if result.OverallStatus != "error" {
		t.Errorf("expected overall_status: error in JSON, got: %s", result.OverallStatus)
	}
}
