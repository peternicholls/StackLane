// Gateway template golden tests. Each fixture exercises one of the documented
// shapes (no routes, single route, multi route, TLS toggled). When the bash
// reference is regenerated, copy the output into testdata/ to refresh.
package gateway

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteConfig_ReplacesStaleDirectory(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "shared", "gateway.conf")
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		t.Fatalf("mkdir stale gateway path: %v", err)
	}
	m := NewManager(configPath)

	if _, _, err := m.WriteConfig(RenderInput{Routes: nil, TLSEnabled: false, HTTPSPort: 443}); err != nil {
		t.Fatalf("write config: %v", err)
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("stat config path: %v", err)
	}
	if info.IsDir() {
		t.Fatalf("config path should be a file, got directory")
	}
}

func TestRender_NoRoutes_NoTLS(t *testing.T) {
	out, err := Render(RenderInput{Routes: nil, TLSEnabled: false, HTTPSPort: 443})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := readGolden(t, "no-routes-no-tls.conf")
	if out != want {
		writeActualForReview(t, "no-routes-no-tls.conf", out)
		t.Errorf("output diverged from golden (see -actual file)")
	}
}

func TestRender_SingleRoute_TLS(t *testing.T) {
	in := RenderInput{
		TLSEnabled: true,
		HTTPSPort:  443,
		Routes: []Route{
			{Hostname: "alpha.test", Slug: "alpha", WebNetworkAlias: "stln-alpha-web"},
		},
	}
	out, err := Render(in)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := readGolden(t, "single-route-tls.conf")
	if out != want {
		writeActualForReview(t, "single-route-tls.conf", out)
		t.Errorf("output diverged from golden")
	}
}

func TestRender_MultiRoute_NoTLS(t *testing.T) {
	in := RenderInput{
		TLSEnabled: false,
		HTTPSPort:  443,
		Routes: []Route{
			{Hostname: "alpha.test", Slug: "alpha", WebNetworkAlias: "stln-alpha-web"},
			{Hostname: "beta.test", Slug: "beta", WebNetworkAlias: "stln-beta-web"},
		},
	}
	out, err := Render(in)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	want := readGolden(t, "multi-route-no-tls.conf")
	if out != want {
		writeActualForReview(t, "multi-route-no-tls.conf", out)
		t.Errorf("output diverged from golden")
	}
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	return string(body)
}

func writeActualForReview(t *testing.T, name, body string) {
	t.Helper()
	_ = os.WriteFile(filepath.Join("testdata", name+".actual"), []byte(body), 0o644)
}
