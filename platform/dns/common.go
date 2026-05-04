// Cross-platform helpers shared by the per-OS DNS providers.
package dns

import (
	"fmt"
	"os"
	"path/filepath"
)

// PreviewConfigPath returns the dnsmasq config preview path under the state dir.
func PreviewConfigPath(stateDir, suffix string) string {
	return filepath.Join(stateDir, "shared", "dnsmasq-"+suffix+".conf")
}

// PreviewResolverPath returns the resolver preview path under the state dir.
func PreviewResolverPath(stateDir, suffix string) string {
	return filepath.Join(stateDir, "shared", "resolver-"+suffix+".conf")
}

// WritePreviewFiles regenerates the dnsmasq + resolver preview files from
// settings. Idempotent.
func WritePreviewFiles(s Settings) error {
	if err := os.MkdirAll(filepath.Join(s.StateDir, "shared"), 0o755); err != nil {
		return err
	}
	cfg := fmt.Sprintf("port=%d\nlisten-address=%s\nbind-interfaces\naddress=/.%s/%s\n", s.Port, s.IP, s.Suffix, s.IP)
	if err := os.WriteFile(PreviewConfigPath(s.StateDir, s.Suffix), []byte(cfg), 0o644); err != nil {
		return err
	}
	resolver := fmt.Sprintf("nameserver %s\nport %d\n", s.IP, s.Port)
	return os.WriteFile(PreviewResolverPath(s.StateDir, s.Suffix), []byte(resolver), 0o644)
}

// MessageFor returns the operator-friendly description for code (mirrors
// stageserve_dns_status_message).
func MessageFor(s Settings, code Code) string {
	switch code {
	case CodeReady:
		return fmt.Sprintf("ready (%s on %s:%d for .%s)", s.Provider, s.IP, s.Port, s.Suffix)
	case CodeUnsupportedOS:
		return "unsupported-os"
	case CodeBrewMissing:
		return "brew missing"
	case CodeDnsmasqMissing:
		return "dnsmasq not installed"
	case CodeConfigMissing:
		return "dnsmasq config missing"
	case CodeDnsmasqStopped:
		return fmt.Sprintf("dnsmasq not running on %s:%d", s.IP, s.Port)
	case CodeResolverMissing:
		return "resolver file missing"
	case CodeResolverMismatch:
		return fmt.Sprintf("resolver file does not point at %s:%d", s.IP, s.Port)
	default:
		return "unknown"
	}
}
