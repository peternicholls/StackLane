//go:build !darwin

// Linux/other-OS DNS provider stub. macOS-only behaviour returns the
// unsupported-platform code rather than silently failing (FR-012).
package dns

import "errors"

// LinuxProvider implements Provider for non-darwin platforms.
type LinuxProvider struct{}

// NewProvider returns the unsupported-platform stub.
func NewProvider() Provider { return LinuxProvider{} }

func (LinuxProvider) Status(s Settings) Status {
	return Status{Code: CodeUnsupportedOS, Message: MessageFor(s, CodeUnsupportedOS)}
}

func (LinuxProvider) Bootstrap(Settings) error {
	return errors.New("local DNS bootstrap is currently implemented for macOS only")
}
