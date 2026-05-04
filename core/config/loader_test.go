// Tests for the config loader precedence chain. Each subtest builds an
// isolated stack-home + project dir on disk, sets the loader's Env hook to a
// hermetic map, and asserts the materialised ProjectConfig.
package config

import (
	"os"
	"path/filepath"
	"strings"
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
	if cfg.StackKind != "20i" {
		t.Errorf("default STAGESERVE_STACK=%q want 20i", cfg.StackKind)
	}
	if want := filepath.Join(cfg.StackHome, "docker-compose.20i.yml"); cfg.StackFile != want {
		t.Errorf("StackFile=%q want %q", cfg.StackFile, want)
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
	if cfg.WebNetworkAlias != "stage-"+cfg.Slug+"-web" {
		t.Errorf("WebNetworkAlias=%q", cfg.WebNetworkAlias)
	}
	if cfg.ComposeProjectName != "stage-"+cfg.Slug {
		t.Errorf("ComposeProjectName=%q", cfg.ComposeProjectName)
	}
	if cfg.RuntimeNetwork != "stage-"+cfg.Slug+"-runtime" {
		t.Errorf("RuntimeNetwork=%q", cfg.RuntimeNetwork)
	}
	if cfg.DatabaseVolume != "stage-"+cfg.Slug+"-db-data" {
		t.Errorf("DatabaseVolume=%q", cfg.DatabaseVolume)
	}
	// Shared resources now use the same stage- prefix family as project runtimes.
	if cfg.SharedGateway.Network != "stage-shared" {
		t.Errorf("SharedGateway.Network=%q want stage-shared", cfg.SharedGateway.Network)
	}
	if cfg.SharedGateway.ComposeProjectName != "stage-shared" {
		t.Errorf("SharedGateway.ComposeProjectName=%q want stage-shared", cfg.SharedGateway.ComposeProjectName)
	}
}

func TestLoader_PrecedenceChain(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	// .env.stageserve in stack home (lowest of the file layers)
	writeFile(t, filepath.Join(stackHome, ".env.stageserve"), "PHP_VERSION=8.0\nSITE_SUFFIX=stack-env\n")
	// project .env.stageserve overrides shell env and stack defaults.
	writeFile(t, filepath.Join(projectDir, ".env.stageserve"), "PHP_VERSION=8.2\nSITE_SUFFIX=local\n")

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
	// Project .env.stageserve wins over shell env when no CLI value.
	if cfg.SiteSuffix != "local" {
		t.Errorf("SITE_SUFFIX precedence: got %q, want local (project .env.stageserve)", cfg.SiteSuffix)
	}
	// shell env populates ports when no project-local file / CLI value.
	if cfg.MySQL.Port != 33060 {
		t.Errorf("MYSQL_PORT shell env: got %d, want 33060", cfg.MySQL.Port)
	}
	if cfg.MySQL.PMAPort != 9999 {
		t.Errorf("PMA_PORT shell env: got %d, want 9999", cfg.MySQL.PMAPort)
	}
}

func TestLoader_ProjectRuntimeEnvDBFallback(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	writeFile(t, filepath.Join(stackHome, ".env.stageserve"), "SITE_SUFFIX=stack-env\n")
	writeFile(t, filepath.Join(projectDir, ".env"), "DB_HOST=127.0.0.1\nDB_DATABASE=budget_forecaster\nDB_USERNAME=devuser\nDB_PASSWORD=devpass\nSITE_SUFFIX=project-should-be-ignored\n")

	loader := newLoader(t, nil, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.SiteSuffix != "stack-env" {
		t.Fatalf("SITE_SUFFIX=%q want stack-env from stageserve .env", cfg.SiteSuffix)
	}
	if cfg.MySQL.Database != "budget_forecaster" {
		t.Fatalf("MYSQL_DATABASE=%q want budget_forecaster from project .env fallback", cfg.MySQL.Database)
	}
	if cfg.MySQL.User != "devuser" {
		t.Fatalf("MYSQL_USER=%q want devuser from project .env fallback", cfg.MySQL.User)
	}
	if cfg.MySQL.Password != "devpass" {
		t.Fatalf("MYSQL_PASSWORD=%q want devpass from project .env fallback", cfg.MySQL.Password)
	}
	if cfg.Hostname != filepath.Base(projectDir)+".stack-env" {
		t.Fatalf("hostname=%q want %q", cfg.Hostname, filepath.Base(projectDir)+".stack-env")
	}
	if cfg.Slug == "" {
		t.Fatal("slug should not be empty")
	}
	if cfg.Hostname != cfg.Slug+".stack-env" {
		t.Fatalf("hostname=%q want %q", cfg.Hostname, cfg.Slug+".stack-env")
	}
}

// TestLoader_StackHomeEnvIsNotLoaded asserts that the loader does NOT consult
// either of the legacy stack-defaults files: <stackHome>/.stackenv (the prior
// canonical name) and <stackHome>/.env (the legacy fallback). Spec 004
// removes both with no compatibility shim.
func TestLoader_StackHomeEnvIsNotLoaded(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	writeFile(t, filepath.Join(stackHome, ".env"), "SITE_SUFFIX=legacy-stack-env\nPHP_VERSION=7.4\n")
	writeFile(t, filepath.Join(stackHome, ".stackenv"), "SITE_SUFFIX=legacy-stackenv\nPHP_VERSION=7.0\n")

	loader := newLoader(t, nil, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.SiteSuffix != "test" {
		t.Fatalf("SITE_SUFFIX=%q want default test (legacy stack files must not load)", cfg.SiteSuffix)
	}
	if cfg.PHPVersion != "8.5" {
		t.Fatalf("PHP_VERSION=%q want default 8.5 (legacy stack files must not load)", cfg.PHPVersion)
	}
}

// TestLoader_StackHomeOverrideLoadsEnvStageserve confirms the canonical
// .env.stageserve file is loaded from a STACK_HOME override directory.
func TestLoader_StackHomeOverrideLoadsEnvStageserve(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(stackHome, ".env.stageserve"), "STAGESERVE_STACK=20i\nPHP_VERSION=8.2\nSITE_SUFFIX=stack-defaults\n")

	loader := newLoader(t, nil, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.PHPVersion != "8.2" {
		t.Fatalf("PHP_VERSION=%q want 8.2 from .env.stageserve", cfg.PHPVersion)
	}
	if cfg.SiteSuffix != "stack-defaults" {
		t.Fatalf("SITE_SUFFIX=%q want stack-defaults from .env.stageserve", cfg.SiteSuffix)
	}
}

func TestLoader_RejectsUnsupportedStackKind(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, ".env.stageserve"), "STAGESERVE_STACK=laravel\n")

	loader := newLoader(t, nil, stackHome)
	_, err := loader.Load(projectDir, CLIFlags{})
	if err == nil {
		t.Fatal("load succeeded, want unsupported stack error")
	}
	if !strings.Contains(err.Error(), "unsupported STAGESERVE_STACK") {
		t.Fatalf("error=%q want unsupported STAGESERVE_STACK message", err)
	}
}

func TestLoader_RejectsInvalidHostname(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	loader := newLoader(t, nil, stackHome)
	_, err := loader.Load(projectDir, CLIFlags{SiteHostname: "localhost"})
	if err == nil {
		t.Fatal("load succeeded, want invalid hostname error")
	}
	if !strings.Contains(err.Error(), "invalid site hostname") {
		t.Fatalf("error=%q want invalid site hostname message", err)
	}
}

func TestLoadEnvFile_UnquotesGeneratedDoubleQuotedValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env.stageserve")
	writeFile(t, path, "SITE_NAME=\"It's \\\"his\\\" site\"\n")

	envMap, err := loadEnvFile(path)
	if err != nil {
		t.Fatalf("loadEnvFile: %v", err)
	}
	if envMap["SITE_NAME"] != `It's "his" site` {
		t.Fatalf("SITE_NAME=%q want %q", envMap["SITE_NAME"], `It's "his" site`)
	}
}

func TestLoader_SharedGatewaySettingsAreNotLoadedFromEnv(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(stackHome, ".env.stageserve"), "SHARED_GATEWAY_HTTP_PORT=8080\nSHARED_GATEWAY_HTTPS_PORT=8443\nSHARED_GATEWAY_NETWORK=stage-central\nSHARED_GATEWAY_COMPOSE_PROJECT_NAME=stage-central\n")
	writeFile(t, filepath.Join(projectDir, ".env.stageserve"), "SHARED_GATEWAY_HTTP_PORT=18080\nSHARED_GATEWAY_HTTPS_PORT=18443\nSHARED_GATEWAY_NETWORK=stage-project\nSHARED_GATEWAY_COMPOSE_PROJECT_NAME=stage-project\n")
	env := map[string]string{
		"SHARED_GATEWAY_HTTP_PORT":            "28080",
		"SHARED_GATEWAY_HTTPS_PORT":           "28443",
		"SHARED_GATEWAY_NETWORK":              "stage-shell",
		"SHARED_GATEWAY_COMPOSE_PROJECT_NAME": "stage-shell",
	}

	loader := newLoader(t, env, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.SharedGateway.HTTPPort != 80 {
		t.Fatalf("SharedGateway.HTTPPort=%d want fixed default 80", cfg.SharedGateway.HTTPPort)
	}
	if cfg.SharedGateway.HTTPSPort != 443 {
		t.Fatalf("SharedGateway.HTTPSPort=%d want fixed default 443", cfg.SharedGateway.HTTPSPort)
	}
	if cfg.SharedGateway.Network != "stage-shared" {
		t.Fatalf("SharedGateway.Network=%q want stage-shared", cfg.SharedGateway.Network)
	}
	if cfg.SharedGateway.ComposeProjectName != "stage-shared" {
		t.Fatalf("SharedGateway.ComposeProjectName=%q want stage-shared", cfg.SharedGateway.ComposeProjectName)
	}
	if cfg.Hostname != cfg.Slug+".test" {
		t.Fatalf("hostname=%q want project defaults to keep working", cfg.Hostname)
	}
	if cfg.PHPVersion != "8.5" {
		t.Fatalf("PHPVersion=%q want normal project/default precedence unaffected", cfg.PHPVersion)
	}
}

func TestLoader_PostUpHookFromProjectLocalConfig(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()

	writeFile(t, filepath.Join(projectDir, ".env.stageserve"), "STAGESERVE_POST_UP_COMMAND=php artisan migrate --force --no-interaction\n")

	loader := newLoader(t, nil, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.PostUpCommand != "php artisan migrate --force --no-interaction" {
		t.Fatalf("PostUpCommand=%q", cfg.PostUpCommand)
	}
}

// TestLoader_PostUpHookOnlyHonoredFromProjectLocal proves FR-016: the bootstrap
// command must be ignored when set via shell env, .env.stageserve, or project
// .env. Each case below asserts cfg.PostUpCommand stays empty.
func TestLoader_PostUpHookOnlyHonoredFromProjectLocal(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, stackHome, projectDir string) map[string]string
	}{
		{
			name: "shell env is ignored",
			setup: func(t *testing.T, stackHome, projectDir string) map[string]string {
				return map[string]string{"STAGESERVE_POST_UP_COMMAND": "echo from-shell"}
			},
		},
		{
			name: ".env.stageserve is ignored",
			setup: func(t *testing.T, stackHome, projectDir string) map[string]string {
				writeFile(t, filepath.Join(stackHome, ".env.stageserve"), "STAGESERVE_POST_UP_COMMAND=echo from-stack-env\n")
				return nil
			},
		},
		{
			name: "project .env is ignored",
			setup: func(t *testing.T, stackHome, projectDir string) map[string]string {
				writeFile(t, filepath.Join(projectDir, ".env"), "STAGESERVE_POST_UP_COMMAND=echo from-project-env\n")
				return nil
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stackHome := t.TempDir()
			projectDir := t.TempDir()
			env := tc.setup(t, stackHome, projectDir)
			loader := newLoader(t, env, stackHome)
			cfg, err := loader.Load(projectDir, CLIFlags{})
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if cfg.PostUpCommand != "" {
				t.Fatalf("PostUpCommand=%q want empty (only project .env.stageserve is honored)", cfg.PostUpCommand)
			}
		})
	}
}

func TestLoader_IgnoresLegacyStageserveLocalFile(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, ".stageserve-local"), "PHP_VERSION=7.4\nSITE_SUFFIX=legacy\n")

	loader := newLoader(t, nil, stackHome)
	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.PHPVersion != "8.5" {
		t.Errorf("PHP_VERSION=%q, want default 8.5 because .stageserve-local is ignored", cfg.PHPVersion)
	}
	if cfg.SiteSuffix != "test" {
		t.Errorf("SITE_SUFFIX=%q, want default test because .stageserve-local is ignored", cfg.SiteSuffix)
	}
	legacy, err := loadEnvFile(filepath.Join(projectDir, ".stageserve-local"))
	if err != nil {
		t.Fatalf("legacy project file should still be parseable for migration checks: %v", err)
	}
	if legacy["PHP_VERSION"] != "7.4" {
		t.Fatalf("legacy project config parse mismatch: %#v", legacy)
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
	wantStateDir := filepath.Join(stackHomeResolved, ".stageserve-state")
	if cfg.StateDir != wantStateDir {
		t.Errorf("StateDir=%q, want %q", cfg.StateDir, wantStateDir)
	}
}

func TestLoader_StateDirUsesStageserveEnvKey(t *testing.T) {
	stackHome := t.TempDir()
	projectDir := t.TempDir()
	wantStateDir := filepath.Join(t.TempDir(), "custom-state")
	legacyStateDir := filepath.Join(t.TempDir(), "legacy-state")
	loader := newLoader(t, map[string]string{
		"STAGESERVE_STATE_DIR": wantStateDir,
		"STACK_STATE_DIR":      legacyStateDir,
	}, stackHome)

	cfg, err := loader.Load(projectDir, CLIFlags{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.StateDir != wantStateDir {
		t.Fatalf("StateDir=%q, want %q", cfg.StateDir, wantStateDir)
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
		{"env wins over default", map[string]string{"STAGESERVE_WAIT_TIMEOUT": "30"}, 0, 30},
		{"flag wins over env", map[string]string{"STAGESERVE_WAIT_TIMEOUT": "30"}, 60, 60},
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
