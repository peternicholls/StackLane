// Package dns defines the platform DNS bootstrap contract. Concrete
// implementations are split by build tag (macOS vs the unsupported-platform stub).
package dns

// Code captures the discrete states of the local DNS bootstrap. Mirrors the
// values returned by bash stacklane_dns_status.
type Code string

const (
	CodeReady            Code = "ready"
	CodeUnsupportedOS    Code = "unsupported-os"
	CodeBrewMissing      Code = "brew-missing"
	CodeDnsmasqMissing   Code = "dnsmasq-missing"
	CodeConfigMissing    Code = "dnsmasq-config-missing"
	CodeDnsmasqStopped   Code = "dnsmasq-stopped"
	CodeResolverMissing  Code = "resolver-missing"
	CodeResolverMismatch Code = "resolver-mismatch"
	CodeUnknown          Code = "unknown"
)

// Status describes the bootstrap state in operator-friendly terms.
type Status struct {
	Code    Code
	Message string
}

// Settings carries the resolved DNS configuration for one suffix.
type Settings struct {
	Suffix   string // dnsmasq managed suffix (e.g. "test")
	IP       string // listen-address (default 127.0.0.1)
	Port     int    // listen-port (default 53535)
	Provider string // "dnsmasq"
	StateDir string // managed-file preview / resolver preview live here
}

// Provider bootstraps and inspects the local DNS resolver for a suffix.
type Provider interface {
	Status(s Settings) Status
	Bootstrap(s Settings) error
}
