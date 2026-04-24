// Package tls defines the typed TLS certificate provider contract.
package tls

import "time"

// Bundle locates a generated certificate on disk.
type Bundle struct {
	CertFile string
	KeyFile  string
	Expiry   time.Time
	Hosts    []string
}

// Provider issues / refreshes TLS certificates for a list of hostnames.
//
// The macOS reference implementation wraps `mkcert`. Other platforms may
// return ErrUnsupported.
type Provider interface {
	Available() bool
	Ensure(certFile, keyFile string, hosts []string) (Bundle, error)
}
