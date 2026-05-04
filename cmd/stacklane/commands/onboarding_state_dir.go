package commands

import "fmt"

func resolveOnboardingStateDir(flags *SharedFlags) (string, error) {
	cfg, err := loadConfig(flags)
	if err != nil {
		return "", fmt.Errorf("resolve onboarding config: %w", err)
	}
	return cfg.StateDir, nil
}