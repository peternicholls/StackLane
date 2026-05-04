//go:build !darwin && !linux

package onboarding

func checkDNS(suffix string) StepResult {
	return StepResult{
		ID:      "dns.resolver",
		Label:   "Local DNS resolver",
		Status:  StatusError,
		Message: "Local DNS bootstrap is not supported on this platform",
		Code:    "unsupported-os",
	}
}

func checkMkcert() StepResult {
	return StepResult{
		ID:      "mkcert.binary",
		Label:   "mkcert local CA",
		Status:  StatusError,
		Message: "mkcert setup is not supported on this platform",
		Code:    "unsupported-os",
	}
}
