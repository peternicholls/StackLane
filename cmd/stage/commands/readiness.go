package commands

import (
	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/onboarding"
)

func buildMachineReadinessResult(shared *SharedFlags, suffix string) (onboarding.CommandResult, error) {
	stateDir, err := resolveOnboardingStateDir(shared)
	if err != nil {
		return onboarding.CommandResult{}, err
	}
	return onboarding.BuildResult(machineReadinessSteps(stateDir, suffix), nil, nil), nil
}

func buildMachineReadinessResultForConfig(cfg config.ProjectConfig) onboarding.CommandResult {
	return onboarding.BuildResult(machineReadinessSteps(cfg.StateDir, cfg.SiteSuffix), nil, nil)
}

func machineReadinessSteps(stateDir, suffix string) []onboarding.StepResult {
	return []onboarding.StepResult{
		onboarding.CheckDockerBinary(""),
		onboarding.CheckDockerDaemon(),
		onboarding.CheckStateDir(stateDir),
		onboarding.CheckPort("port.80", 80),
		onboarding.CheckPort("port.443", 443),
		onboarding.CheckDNS(suffix),
		onboarding.CheckMkcert(),
	}
}
