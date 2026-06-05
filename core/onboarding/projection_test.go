package onboarding_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/onboarding"
)

func TestTextProjector_RendersActionableReport(t *testing.T) {
	result := onboarding.BuildResult([]onboarding.StepResult{
		{
			ID:      "docker.binary",
			Label:   "Docker CLI",
			Status:  onboarding.StatusReady,
			Message: "docker found at /usr/local/bin/docker",
		},
		{
			ID:          "dns.resolver",
			Label:       "Local DNS resolver",
			Status:      onboarding.StatusNeedsAction,
			Message:     "dnsmasq config missing",
			Remediation: stringPtr("stage setup"),
		},
	}, nil, nil)

	buf := &bytes.Buffer{}
	projector := onboarding.TextProjector{W: buf, Title: "StageServe Doctor", Detailed: true}
	if err := projector.Project(result); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"StageServe Doctor",
		"Not ready — 1 of 2 check needs attention.",
		"Needs fixing",
		"1  Local DNS resolver",
		"Routes *.test domains to your stack",
		"To fix:  stage setup",
		"All clear",
		"✓  Docker CLI",
		"Fix the issues above, then run: stage doctor",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestTUIProjector_RendersReadableSections(t *testing.T) {
	result := onboarding.BuildResult([]onboarding.StepResult{
		{
			ID:      "docker.binary",
			Label:   "Docker CLI",
			Status:  onboarding.StatusReady,
			Message: "docker found at /usr/local/bin/docker",
		},
		{
			ID:          "port.443",
			Label:       "Port 443",
			Status:      onboarding.StatusNeedsAction,
			Message:     "port 443 is already in use",
			Remediation: stringPtr("Find and stop the process using port 443: lsof -nP -iTCP:443 -sTCP:LISTEN"),
		},
	}, nil, nil)

	buf := &bytes.Buffer{}
	projector := onboarding.TUIProjector{W: buf, Title: "StageServe Doctor", Detailed: true}
	if err := projector.Project(result); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"StageServe Doctor",
		"Needs fixing",
		"All clear",
		"Port 443",
		"Docker CLI",
		"Fix the issues above, then run: stage doctor",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestNewProjector_AppliesDetailedOptions(t *testing.T) {
	result := onboarding.BuildResult([]onboarding.StepResult{
		{
			ID:      "docker.binary",
			Label:   "Docker CLI",
			Status:  onboarding.StatusReady,
			Message: "docker found at /usr/local/bin/docker",
		},
	}, nil, nil)

	buf := &bytes.Buffer{}
	projector := onboarding.NewProjector(onboarding.OutputModeText, buf, onboarding.ProjectorOptions{
		Title:    "StageServe Setup",
		Detailed: true,
	})
	if err := projector.Project(result); err != nil {
		t.Fatalf("project failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "StageServe Setup") {
		t.Fatalf("expected projector output to include custom title, got:\n%s", out)
	}
	if !strings.Contains(out, "Checks passed") {
		t.Fatalf("expected detailed projection output, got:\n%s", out)
	}
}

func stringPtr(s string) *string {
	return &s
}
