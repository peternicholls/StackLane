//go:build darwin

// macOS DNS bootstrap: dnsmasq via Homebrew + /etc/resolver/<suffix> +
// privilege escalation through osascript when the resolver file requires
// admin rights.
package dns

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MacOSProvider implements Provider for darwin.
type MacOSProvider struct{}

// NewProvider returns the macOS provider.
func NewProvider() Provider { return MacOSProvider{} }

func (MacOSProvider) Status(s Settings) Status {
	if !brewAvailable() {
		return Status{Code: CodeBrewMissing, Message: MessageFor(s, CodeBrewMissing)}
	}
	if !dnsmasqInstalled() {
		return Status{Code: CodeDnsmasqMissing, Message: MessageFor(s, CodeDnsmasqMissing)}
	}
	managed, err := dnsmasqManagedFile(s.Suffix)
	if err != nil || !fileExists(managed) {
		return Status{Code: CodeConfigMissing, Message: MessageFor(s, CodeConfigMissing)}
	}
	if !dnsmasqListening(s.Port) {
		return Status{Code: CodeDnsmasqStopped, Message: MessageFor(s, CodeDnsmasqStopped)}
	}
	resolverFile := "/etc/resolver/" + s.Suffix
	if !fileExists(resolverFile) {
		return Status{Code: CodeResolverMissing, Message: MessageFor(s, CodeResolverMissing)}
	}
	contents, _ := os.ReadFile(resolverFile)
	if !strings.Contains(string(contents), "nameserver "+s.IP) || !strings.Contains(string(contents), fmt.Sprintf("port %d", s.Port)) {
		return Status{Code: CodeResolverMismatch, Message: MessageFor(s, CodeResolverMismatch)}
	}
	return Status{Code: CodeReady, Message: MessageFor(s, CodeReady)}
}

func (p MacOSProvider) Bootstrap(s Settings) error {
	if !brewAvailable() {
		return fmt.Errorf("homebrew is required for local DNS bootstrap")
	}
	if !dnsmasqInstalled() {
		return fmt.Errorf("dnsmasq is not installed. Run: brew install dnsmasq")
	}
	if err := WritePreviewFiles(s); err != nil {
		return err
	}
	// Cache brew prefix for reuse across multiple path resolutions.
	prefix, err := brewPrefix()
	if err != nil {
		return err
	}
	managed := dnsmasqManagedFileWithPrefix(s.Suffix, prefix)
	if err := os.MkdirAll(filepath.Dir(managed), 0o755); err != nil {
		return err
	}
	// Strip stale managed configs so duplicate global dnsmasq directives from
	// previous Stacklane / legacy 20i runs do not prevent dnsmasq from starting.
	for _, pattern := range []string{"stacklane-*.conf", "20i-*.conf"} {
		matches, _ := filepath.Glob(filepath.Join(filepath.Dir(managed), pattern))
		for _, m := range matches {
			_ = os.Remove(m)
		}
	}
	if err := copyFile(PreviewConfigPath(s.StateDir, s.Suffix), managed); err != nil {
		return err
	}
	if err := ensureDnsmasqIncludeWithPrefix(prefix); err != nil {
		return err
	}
	if err := exec.Command("brew", "services", "restart", "dnsmasq").Run(); err != nil {
		if err := exec.Command("brew", "services", "start", "dnsmasq").Run(); err != nil {
			return fmt.Errorf("could not start dnsmasq via Homebrew services: %w", err)
		}
	}
	if err := installResolver(s); err != nil {
		return err
	}
	if status := p.Status(s); status.Code != CodeReady {
		return fmt.Errorf("local DNS bootstrap did not reach a ready state (%s)", status.Message)
	}
	return nil
}

// --- helpers ---

func brewAvailable() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func dnsmasqInstalled() bool {
	out, err := exec.Command("brew", "list", "dnsmasq").CombinedOutput()
	if err != nil {
		return false
	}
	return len(out) > 0
}

func dnsmasqListening(port int) bool {
	if _, err := exec.LookPath("lsof"); err != nil {
		return false
	}
	out, err := exec.Command("lsof", "-nP", "-iUDP:"+fmt.Sprint(port)).CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "dnsmasq")
}

func brewPrefix() (string, error) {
	out, err := exec.Command("brew", "--prefix").Output()
	if err != nil {
		return "", err
	}
	prefix := strings.TrimSpace(string(out))
	if prefix == "" {
		return "", fmt.Errorf("brew --prefix returned empty output")
	}
	return prefix, nil
}

func dnsmasqManagedFile(suffix string) (string, error) {
	prefix, err := brewPrefix()
	if err != nil {
		return "", err
	}
	return dnsmasqManagedFileWithPrefix(suffix, prefix), nil
}

// dnsmasqManagedFileWithPrefix returns the path to the dnsmasq-managed config file for the given suffix.
func dnsmasqManagedFileWithPrefix(suffix, prefix string) string {
	return filepath.Join(prefix, "etc", "dnsmasq.d", "stacklane-"+suffix+".conf")
}

func dnsmasqMainConf() (string, error) {
	prefix, err := brewPrefix()
	if err != nil {
		return "", err
	}
	return filepath.Join(prefix, "etc", "dnsmasq.conf"), nil
}

func ensureDnsmasqInclude() error {
	prefix, err := brewPrefix()
	if err != nil {
		return err
	}
	return ensureDnsmasqIncludeWithPrefix(prefix)
}

// ensureDnsmasqIncludeWithPrefix ensures the dnsmasq main config includes the dnsmasq.d directory.
func ensureDnsmasqIncludeWithPrefix(prefix string) error {
	mainConf := filepath.Join(prefix, "etc", "dnsmasq.conf")
	if !fileExists(mainConf) {
		return fmt.Errorf("dnsmasq main config not found: %s", mainConf)
	}
	includeLine := fmt.Sprintf("conf-dir=%s/etc/dnsmasq.d,*.conf", prefix)
	body, err := os.ReadFile(mainConf)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(body), "\n") {
		if strings.TrimSpace(line) == includeLine {
			return nil
		}
	}
	f, err := os.OpenFile(mainConf, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n# Stacklane managed include\n" + includeLine + "\n")
	return err
}

func installResolver(s Settings) error {
	previewResolver := PreviewResolverPath(s.StateDir, s.Suffix)
	resolverFile := "/etc/resolver/" + s.Suffix
	if existing, err := os.ReadFile(resolverFile); err == nil {
		want, _ := os.ReadFile(previewResolver)
		if string(existing) == string(want) {
			return nil
		}
	}
	// Try direct mkdir + copy first (works when /etc/resolver is writable).
	if err := os.MkdirAll(filepath.Dir(resolverFile), 0o755); err == nil {
		if err := copyFile(previewResolver, resolverFile); err == nil {
			return nil
		}
	}
	// Fall back to osascript privilege escalation.
	if _, err := exec.LookPath("osascript"); err == nil {
		shScript := fmt.Sprintf("/bin/mkdir -p %s && /bin/cp %s %s",
			shEscape(filepath.Dir(resolverFile)),
			shEscape(previewResolver),
			shEscape(resolverFile),
		)
		if err := exec.Command("osascript", "-e", `do shell script "`+shScript+`" with administrator privileges`).Run(); err != nil {
			return fmt.Errorf("administrator approval was required to install %s", resolverFile)
		}
		return nil
	}
	return fmt.Errorf("resolver file needs elevated privileges; copy %s to %s", previewResolver, resolverFile)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

// shEscape escapes a string for use in a POSIX shell command by wrapping it in
// single quotes and escaping any embedded single quotes.
func shEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
