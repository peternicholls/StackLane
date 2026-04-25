// Failure-classification tests for lifecycle StepError. The lifecycle classes
// for spec 004 are: ensure-shared-network, registry, allocate-ports,
// gateway-config, shared-gateway, write-env-file, compose-up, wait-healthy,
// post-up-hook, gateway-reload, save-state, compose-down. Each must surface
// as its own named step so the operator can distinguish bootstrap failure
// from gateway/DNS/readiness failures (US2 / FR-004).
package lifecycle

import (
	"errors"
	"strings"
	"testing"
)

func TestStepError_NamedStepsAreDistinct(t *testing.T) {
	cases := []struct {
		step    string
		project string
	}{
		{"ensure-shared-network", "demo"},
		{"allocate-ports", "demo"},
		{"shared-gateway", ""},
		{"compose-up", "demo"},
		{"wait-healthy", "demo"},
		{"post-up-hook", "demo"},
		{"gateway-config", "demo"},
		{"gateway-reload", "demo"},
		{"save-state", "demo"},
		{"compose-down", "demo"},
	}
	seen := make(map[string]bool, len(cases))
	for _, tc := range cases {
		err := Wrap(tc.step, tc.project, errors.New("boom"), "do something")
		if err == nil {
			t.Fatalf("Wrap returned nil for step %q", tc.step)
		}
		se, ok := AsStepError(err)
		if !ok {
			t.Fatalf("AsStepError(%q) = false", tc.step)
		}
		if se.Step != tc.step {
			t.Fatalf("Step=%q want %q", se.Step, tc.step)
		}
		if seen[tc.step] {
			t.Fatalf("step %q reported twice", tc.step)
		}
		seen[tc.step] = true
	}
}

// TestStepError_PostUpHookIsNotMisreported asserts that bootstrap failure does
// not collapse into gateway, DNS, or readiness step names.
func TestStepError_PostUpHookIsNotMisreported(t *testing.T) {
	err := Wrap("post-up-hook", "demo", errors.New("hook exited 1"), "Check STACKLANE_POST_UP_COMMAND")
	se, _ := AsStepError(err)
	if se == nil {
		t.Fatalf("expected StepError")
	}
	for _, forbidden := range []string{"gateway", "dns", "wait-healthy", "compose-up"} {
		if strings.Contains(strings.ToLower(se.Step), forbidden) {
			t.Fatalf("post-up-hook step contaminated with %q: got %q", forbidden, se.Step)
		}
	}
	rendered := se.Error()
	if !strings.Contains(rendered, "post-up-hook") {
		t.Fatalf("rendered error missing step name: %q", rendered)
	}
	if !strings.Contains(rendered, "demo") {
		t.Fatalf("rendered error missing project slug: %q", rendered)
	}
	if !strings.Contains(rendered, "next:") {
		t.Fatalf("rendered error missing remediation: %q", rendered)
	}
}

func TestStepError_UnwrapPreservesCause(t *testing.T) {
	cause := errors.New("underlying")
	err := Wrap("post-up-hook", "demo", cause, "")
	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is should match wrapped cause")
	}
}
