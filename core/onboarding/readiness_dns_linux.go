//go:build linux

package onboarding

func checkDNS(suffix string) StepResult {
	return StepResult{
		ID:      "dns.resolver",
		Label:   "Local DNS resolver",
		Status:  StatusError,
		Message: "Local DNS bootstrap is not yet automated on Linux. Add a dnsmasq entry manually.",
		Code:    "unsupported-os",
	}
}

func checkMkcert() StepResult {
	return StepResult{
		ID:      "mkcert.binary",
		Label:   "mkcert local CA",
		Status:  StatusError,
		Message: "mkcert setup is not yet automated on Linux. Install mkcert and run 'mkcert -install' manually.",
		Code:    "unsupported-os",
	}
}
