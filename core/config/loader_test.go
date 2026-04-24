// Tests for the config loader precedence chain. Each subtest builds an
// isolated stack-home + project dir on disk, sets the loader's Env hook to a
// hermetic map, and asserts the materialised ProjectConfig.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func newLoader(t *testing.T, env map[string]string, stackHome string) *Loader {
	t.Helper()
	return &Loader{
		Env: func(k string) (string, bool) {
			v, ok := env[k]
			return v, ok
		},
		StackHomeOverride: stackHome,
	}
}

func TestLoader_DefaultsApplied(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	loader := newLoader(t, nil, stackHome)

	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.SiteSuffix != "test" {
		t.Errorf("default SITE_SUFFIX=%q want test", cfg.SiteSuffix)
	}
	if cfg.PHPVersion != "8.5" {
		t.Errorf("default PHP_VERSION=%q want 8.5", cfg.PHPVersion)
	}
	if cfg.WaitTimeoutSecs != 120 {
		t.Errorf("default WaitTimeoutSecs=%d want 120", cfg.WaitTimeoutSecs)
	}
	if cfg.MySQL.Database != cfg.Slug {
		t.Errorf("MYSQL_DATABASE should default to slug; got %q want %q", cfg.MySQL.Database, cfg.Slug)
	}
	if cfg.MySQL.User != cfg.Slug {
		t.Errorf("MYSQL_USER should default to slug; got %q want %q", cfg.MySQL.User, cfg.Slug)
	}
	if cfg.ContainerSiteRoot != "/home/sites/"+cfg.Slug {
		t.Errorf("ContainerSiteRoot=%q", cfg.ContainerSiteRoot)
	}
	if cfg.WebNetworkAlias != "stacklane-"+cfg.Slug+"-web" {
		t.Errorf("WebNetworkAlias=%q", cfg.WebNetworkAlias)
	}
}

func TestLoader_PrecedenceChain(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	// .env in stack home (lowest of the file layers)
	writeFile(t, filepath.Join(stackHome, ".env"), "PHP_VERSION=8.0\nSITE_SUFFIX=stack-env\n")
	// .stacklane-local in project dir (overrides shell env and .env)
	writeFile(t, filepath.Join(projectDir, ".stacklane-local"), "PHP_VERSION=8.2\nSITE_SUFFIX=local\n")

	env := map[string]string{
		"PHP_VERSION": "8.1",
		"SITE_SUFFIX": "shell-env",
		"PMA_PORT":    "9999",
		"MYSQL_PORT":  "33060",
	}
	loader := newLoader(t, env, stackHome)

	cfg, err := loader.Load(projectDir, CLIFlags{
		PHPVersion: "8.3",
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// CLI flag wins.
	if cfg.PHPVersion != "8.3" {
		t.Errorf("PHP_VERSION precedence: got %q, want 8.3 (CLI)", cfg.PHPVersion)
	}
	// .stacklane-local wins over shell env when no CLI value.
	if cfg.SiteSuffix != "local" {
		t.Errorf("SITE_SUFFIX precedence: got %q, want local (.stacklane-local)", cfg.SiteSuffix)
	}
	// shell env populates ports when no project-local file / CLI value.
	if cfg.MySQL.Port != 33060 {
		t.Errorf("MYSQL_PORT shell env: got %d, want 33060", cfg.MySQL.Port)
	}
	if cfg.MySQL.PMAPort != 9999 {
		t.Errorf("PMA_PORT shell env: got %d, want 9999", cfg.MySQL.PMAPort)
	}
}

func TestLoader_IgnoresLegacyProjectAndStatePaths(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, ".20i-local"), "PHP_VERSION=7.4\nSITE_SUFFIX=legacy\n")
	if err := os.Mkdir(filepath.Join(stackHome, ".20i-state"), 0o755); err != nil {
		t.Fatalf("mkdir legacy state: %v", err)
	}

	loader := newLoader(t, nil, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.PHPVersion != "8.5" {
		t.Errorf("PHP_VERSION=%q, want default 8.5 because .20i-local is ignored", cfg.PHPVersion)
	}
	if cfg.SiteSuffix != "test" {
		t.Errorf("SITE_SUFFIX=%q, want default test because .20i-local is ignored", cfg.SiteSuffix)
	}
	stackHomeResolved, err := filepath.EvalSymlinks(stackHome)
	if err != nil {
		stackHomeResolved = stackHome
	}
	wantStateDir := filepath.Join(stackHomeResolved, ".stacklane-state")
	if cfg.StateDir != wantStateDir {
		t.Errorf("StateDir=%q, want %q", cfg.StateDir, wantStateDir)
	}
}

func TestLoader_WaitTimeoutPrecedence(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	cases := []struct {
		name string
		env  map[string]string
		flag int
		want int
	}{
		{"default", nil, 0, 120},
		{"env wins over default", map[string]string{"STACKLANE_WAIT_TIMEOUT": "30"}, 0, 30},
		{"flag wins over env", map[string]string{"STACKLANE_WAIT_TIMEOUT": "30"}, 60, 60},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loader := newLoader(t, tc.env, stackHome)
			cfg, err := loader.Load(projectDir, CLIFlags{WaitTimeoutSecs: tc.flag})
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if cfg.WaitTimeoutSecs != tc.want {
				t.Errorf("WaitTimeoutSecs=%d want %d", cfg.WaitTimeoutSecs, tc.want)
			}
		})
	}
}

func TestLoader_HostnameDerivation(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := filepath.Join(t.TempDir(), "My Cool Site")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	loader := newLoader(t, nil, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Slug != "my-cool-site" {
		t.Errorf("slug=%q want my-cool-site", cfg.Slug)
	}
	if cfg.Hostname != "my-cool-site.test" {
		t.Errorf("hostname=%q want my-cool-site.test", cfg.Hostname)
	}
}
