package onboarding_test

import (
	"testing"

	"github.com/peternicholls/stageserve/core/onboarding"
)

// --- ReduceExitCode tests ---

func TestReduceExitCode_AllReady(t *testing.T) {
	steps := []onboarding.StepResult{
		{ID: "a", Status: onboarding.StatusReady},
		{ID: "b", Status: onboarding.StatusReady},
	}
	got := onboarding.ReduceExitCode(steps)
	if got != onboarding.ExitReady {
		t.Fatalf("want ExitReady (0), got %d", got)
	}
}

func TestReduceExitCode_NeedsAction(t *testing.T) {
	steps := []onboarding.StepResult{
		{ID: "a", Status: onboarding.StatusReady},
		{ID: "b", Status: onboarding.StatusNeedsAction},
	}
	got := onboarding.ReduceExitCode(steps)
	if got != onboarding.ExitNeedsAction {
		t.Fatalf("want ExitNeedsAction (1), got %d", got)
	}
}

func TestReduceExitCode_Error(t *testing.T) {
	steps := []onboarding.StepResult{
		{ID: "a", Status: onboarding.StatusNeedsAction},
		{ID: "b", Status: onboarding.StatusError},
	}
	got := onboarding.ReduceExitCode(steps)
	if got != onboarding.ExitError {
		t.Fatalf("want ExitError (2), got %d", got)
	}
}

func TestReduceExitCode_PrefersUnsupportedOS(t *testing.T) {
	steps := []onboarding.StepResult{
		{ID: "a", Status: onboarding.StatusError},
		{ID: "b", Status: onboarding.StatusError, Code: "unsupported-os"},
	}
	got := onboarding.ReduceExitCode(steps)
	if got != onboarding.ExitUnsupportedOS {
		t.Fatalf("want ExitUnsupportedOS (3), got %d", got)
	}
}

// --- DeriveOverallStatus tests ---

func TestOverallStatus_AllReady(t *testing.T) {
	steps := []onboarding.StepResult{
		{Status: onboarding.StatusReady},
	}
	got := onboarding.DeriveOverallStatus(steps)
	if got != onboarding.OverallReady {
		t.Fatalf("want OverallReady, got %s", got)
	}
}

func TestOverallStatus_NeedsActionWithoutError(t *testing.T) {
	steps := []onboarding.StepResult{
		{Status: onboarding.StatusReady},
		{Status: onboarding.StatusNeedsAction},
	}
	got := onboarding.DeriveOverallStatus(steps)
	if got != onboarding.OverallNeedsAction {
		t.Fatalf("want OverallNeedsAction, got %s", got)
	}
}

func TestOverallStatus_ErrorDominates(t *testing.T) {
	steps := []onboarding.StepResult{
		{Status: onboarding.StatusNeedsAction},
		{Status: onboarding.StatusError},
	}
	got := onboarding.DeriveOverallStatus(steps)
	if got != onboarding.OverallError {
		t.Fatalf("want OverallError, got %s", got)
	}
}

// --- BuildResult tests ---

func TestBuildResult_MatchesSteps(t *testing.T) {
	steps := []onboarding.StepResult{
		{ID: "x", Label: "X", Status: onboarding.StatusReady, Message: "ok"},
	}
	result := onboarding.BuildResult(steps, nil, nil)
	if result.OverallStatus != onboarding.OverallReady {
		t.Errorf("want OverallReady, got %s", result.OverallStatus)
	}
	if result.ExitCode != onboarding.ExitReady {
		t.Errorf("want exit 0, got %d", result.ExitCode)
	}
	if len(result.Steps) != 1 {
		t.Errorf("want 1 step, got %d", len(result.Steps))
	}
}

func TestBuildResult_PreservesNextSteps(t *testing.T) {
	ns := []string{"stage up", "stage doctor"}
	result := onboarding.BuildResult(nil, nil, ns)
	if len(result.NextSteps) != 2 {
		t.Errorf("want 2 next_steps, got %d", len(result.NextSteps))
	}
}
