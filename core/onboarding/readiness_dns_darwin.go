//go:build darwin

package onboarding

import (
	"fmt"
	"os/exec"

	"github.com/peternicholls/stageserve/platform/dns"
)

func checkDNS(suffix string) StepResult {
	label := "Local DNS resolver"
	if suffix == "" {
		suffix = "test"
	}
	settings := dns.Settings{
		Suffix:   suffix,
		IP:       "127.0.0.1",
		Port:     53535,
		Provider: "dnsmasq",
	}
	provider := dns.NewProvider()
	st := provider.Status(settings)
	if st.Code == dns.CodeReady {
		return StepResult{
			ID:      "dns.resolver",
			Label:   label,
			Status:  StatusReady,
			Message: st.Message,
		}
	}
	rem := remediationPtr("Run: stage setup (will bootstrap dnsmasq + /etc/resolver/" + suffix + ")")
	return StepResult{
		ID:          "dns.resolver",
		Label:       label,
		Status:      StatusNeedsAction,
		Message:     st.Message,
		Remediation: rem,
	}
}

func checkMkcert() StepResult {
	label := "mkcert local CA"
	path, err := exec.LookPath("mkcert")
	if err != nil {
		rem := remediationPtr("Install mkcert: brew install mkcert && mkcert -install")
		return StepResult{
			ID:          "mkcert.binary",
			Label:       label,
			Status:      StatusNeedsAction,
			Message:     "mkcert not found in PATH",
			Remediation: rem,
		}
	}
	return StepResult{
		ID:      "mkcert.binary",
		Label:   label,
		Status:  StatusReady,
		Message: fmt.Sprintf("mkcert found at %s", path),
	}
}
