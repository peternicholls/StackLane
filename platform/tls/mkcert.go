// mkcert-backed TLS provider. Wraps the mkcert subprocess to issue a single
// PEM bundle covering every hostname on the supplied list.
package tls

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Mkcert is the macOS reference TLS provider.
type Mkcert struct{}

// NewMkcert returns an mkcert-backed Provider.
func NewMkcert() *Mkcert { return &Mkcert{} }

// Available reports whether the mkcert binary is on PATH.
func (Mkcert) Available() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

// Ensure renders or refreshes a cert bundle for hosts at certFile/keyFile.
func (m Mkcert) Ensure(certFile, keyFile string, hosts []string) (Bundle, error) {
	if len(hosts) == 0 {
		return Bundle{}, errors.New("tls: hosts list is empty")
	}
	if !m.Available() {
		return Bundle{}, ErrUnsupported
	}
	if err := os.MkdirAll(filepath.Dir(certFile), 0o755); err != nil {
		return Bundle{}, err
	}
	args := []string{"-cert-file", certFile, "-key-file", keyFile}
	args = append(args, hosts...)
	if err := exec.Command("mkcert", args...).Run(); err != nil {
		return Bundle{}, fmt.Errorf("mkcert: %w", err)
	}
	return Bundle{
		CertFile: certFile,
		KeyFile:  keyFile,
		Hosts:    append([]string(nil), hosts...),
		Expiry:   time.Now().Add(365 * 24 * time.Hour), // mkcert default validity is ~825 days; this is a conservative lower bound.
	}, nil
}

// ErrUnsupported is returned when no TLS implementation is available on the
// running platform.
var ErrUnsupported = errors.New("tls: mkcert not available on this platform")
