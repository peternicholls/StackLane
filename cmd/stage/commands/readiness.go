package commands

import (
	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/onboarding"
)

func buildMachineReadinessResult(shared *SharedFlags, suffix string) (onboarding.CommandResult, error) {
	cfg, err := loadConfig(shared)
	if err != nil {
		return onboarding.CommandResult{}, err
	}
	checkSuffix := cfg.SiteSuffix
	if suffix != "" {
		checkSuffix = suffix
	}
	return onboarding.BuildResult(machineReadinessSteps(cfg, checkSuffix), nil, nil), nil
}

func buildMachineReadinessResultForConfig(cfg config.ProjectConfig) onboarding.CommandResult {
	return onboarding.BuildResult(machineReadinessSteps(cfg, cfg.SiteSuffix), nil, nil)
}

func machineReadinessSteps(cfg config.ProjectConfig, suffix string) []onboarding.StepResult {
	return []onboarding.StepResult{
		onboarding.CheckDockerBinary(""),
		onboarding.CheckDockerDaemon(),
		onboarding.CheckStateDir(cfg.StateDir),
		onboarding.CheckRequiredFile("stack.shared_file", "Shared runtime file", cfg.SharedFile),
		onboarding.CheckRequiredFile("stack.project_file", "Project runtime file", cfg.StackFile),
		onboarding.CheckPort("port.80", 80),
		onboarding.CheckPort("port.443", 443),
		onboarding.CheckDNS(suffix),
		onboarding.CheckMkcert(),
	}
}
