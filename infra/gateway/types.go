// Package gateway owns the shared nginx gateway: typed Route model, atomic
// config writes, and reload semantics.
//
// Route here replaces the positional pipe-delimited "hostname|alias|slug"
// strings the bash stacklane_gateway_route_lines emitted, and the typed model
// removes the heredoc string-interpolation hazards in stacklane_gateway_block_for_route.
package gateway

// Route describes one hostname → upstream-alias mapping.
type Route struct {
	Hostname        string // public hostname matched by server_name
	WebNetworkAlias string // upstream alias on the shared-gateway network
	Slug            string // owning project slug (for selection / preferred-target logic)
}

// RenderInput is the typed input to the nginx config template.
type RenderInput struct {
	Routes        []Route
	TLSEnabled    bool   // mounts /etc/nginx/certs/{tls.pem,tls-key.pem}
	HTTPSPort     int    // SHARED_GATEWAY_HTTPS_PORT (used in HTTP→HTTPS redirect when != 443)
	PreferredSlug string // optional; surfaced via probe-target selection
}

// HealthState reports the gateway's health snapshot.
type HealthState struct {
	GatewayContainer string
	Reachable        bool
	LastError        string
}

// GatewayManager owns the nginx config file and the reload path.
type GatewayManager interface {
	// WriteConfig regenerates the gateway config from the supplied input. Must
	// be atomic: temp file + rename. Returns the resolved probe target (the
	// upstream alias to wait for) and probe hostname (server_name to wait for)
	// so the caller can sequence post-reload health probes.
	WriteConfig(input RenderInput) (probeTarget string, probeHostname string, err error)
	// AddRoute is a convenience wrapper around WriteConfig with the current
	// route set + the new entry; implementations should call WriteConfig.
	AddRoute(r Route, current []Route) (probeTarget string, probeHostname string, err error)
	// RemoveRoute is a convenience wrapper that drops the named slug.
	RemoveRoute(slug string, current []Route) (probeTarget string, probeHostname string, err error)
	// ConfigPath returns the absolute path to the rendered config file.
	ConfigPath() string
}
