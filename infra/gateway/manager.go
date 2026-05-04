// GatewayManager: atomic config writes (temp + rename) and a typed reload
// API. Reload is intentionally NOT here — the lifecycle orchestrator triggers
// `docker compose up --force-recreate gateway` via the Composer when needed.
package gateway

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/peternicholls/stageserve/core/project"
)

// Manager is the default GatewayManager.
type Manager struct {
	configPath string
}

// NewManager binds a manager to a config path. Caller must ensure the
// directory exists; WriteConfig will create it if missing.
func NewManager(configPath string) *Manager { return &Manager{configPath: configPath} }

// ConfigPath returns the absolute config path.
func (m *Manager) ConfigPath() string { return m.configPath }

// WriteConfig regenerates the gateway config from input. Atomic via temp +
// rename. Returns the probe target/hostname pair the caller should wait for.
func (m *Manager) WriteConfig(input RenderInput) (string, string, error) {
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0o755); err != nil {
		return "", "", err
	}
	// Drop invalid routes silently to mirror the bash behaviour.
	cleaned := make([]Route, 0, len(input.Routes))
	for _, r := range input.Routes {
		if project.HostnameValid(r.Hostname) && project.AliasValid(r.WebNetworkAlias) {
			cleaned = append(cleaned, r)
		}
	}
	// Stable order so generated configs are deterministic.
	sort.SliceStable(cleaned, func(i, j int) bool { return cleaned[i].Hostname < cleaned[j].Hostname })
	input.Routes = cleaned

	rendered, err := Render(input)
	if err != nil {
		return "", "", fmt.Errorf("render gateway config: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(m.configPath), filepath.Base(m.configPath)+".tmp.*")
	if err != nil {
		return "", "", err
	}
	if _, err := tmp.WriteString(rendered); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", "", err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", "", err
	}
	if info, err := os.Stat(m.configPath); err == nil && info.IsDir() {
		if err := os.RemoveAll(m.configPath); err != nil {
			os.Remove(tmp.Name())
			return "", "", err
		}
	}
	if err := os.Rename(tmp.Name(), m.configPath); err != nil {
		return "", "", err
	}

	return probeFor(input)
}

// AddRoute writes the union of current and r.
func (m *Manager) AddRoute(r Route, current []Route) (string, string, error) {
	merged := make([]Route, 0, len(current)+1)
	for _, existing := range current {
		if existing.Slug != r.Slug {
			merged = append(merged, existing)
		}
	}
	merged = append(merged, r)
	return m.WriteConfig(RenderInput{Routes: merged, PreferredSlug: r.Slug})
}

// RemoveRoute strips slug from current and writes the result.
func (m *Manager) RemoveRoute(slug string, current []Route) (string, string, error) {
	merged := make([]Route, 0, len(current))
	for _, existing := range current {
		if existing.Slug != slug {
			merged = append(merged, existing)
		}
	}
	return m.WriteConfig(RenderInput{Routes: merged})
}

// probeFor selects the upstream + hostname pair the orchestrator should wait
// for after a reload (mirrors stageserve_write_gateway_config probe selection).
func probeFor(input RenderInput) (target, host string, err error) {
	if len(input.Routes) == 0 {
		return "stageserve-no-route", "localhost", nil
	}
	for _, r := range input.Routes {
		if r.Slug == input.PreferredSlug {
			return r.WebNetworkAlias, r.Hostname, nil
		}
	}
	return input.Routes[0].WebNetworkAlias, input.Routes[0].Hostname, nil
}
