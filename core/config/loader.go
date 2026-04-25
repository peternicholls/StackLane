// Loader implements ConfigLoader. It resolves the configuration precedence
// chain as defined by the active Stacklane config contract:
//
//  1. CLI flags (highest)
//  2. .env.stacklane in the project directory
//  3. shell environment
//  4. .env.stacklane in the stack home (canonical stack-owned defaults file)
//  5. built-in defaults (lowest)
//
// STACKLANE_POST_UP_COMMAND is a special-case bootstrap setting that is only
// honored when set in the project's .env.stacklane file. It is intentionally
// ignored if set via shell environment, stack-home .env.stacklane, or project .env.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/peternicholls/stacklane/core/project"
)

// Loader is the default ConfigLoader implementation.
type Loader struct {
	// Env supplies the shell environment. nil means os.LookupEnv.
	Env func(string) (string, bool)
	// StackHomeOverride forces the stack home (used by tests). When empty the
	// loader uses STACK_HOME or the directory holding docker-compose.shared.yml.
	StackHomeOverride string
}

// NewLoader returns a Loader with the live os environment.
func NewLoader() *Loader { return &Loader{Env: os.LookupEnv} }

// envOrDefault looks up key in the loader environment.
func (l *Loader) envOrDefault(key, fallback string) string {
	get := l.Env
	if get == nil {
		get = os.LookupEnv
	}
	if v, ok := get(key); ok && v != "" {
		return v
	}
	return fallback
}

// loadEnvFile reads a KEY=VALUE file (POSIX-shell quoting) into a map.
// Mirrors stacklane_load_env_file: blank lines and # comments skipped, optional
// leading "export ", surrounding " or ' stripped.
func loadEnvFile(path string) (map[string]string, error) {
	out := map[string]string{}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		value := strings.TrimSpace(line[eq+1:])
		if !validEnvKey(key) {
			continue
		}
		switch {
		case len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"':
			unquoted, unquoteErr := strconv.Unquote(value)
			if unquoteErr != nil {
				value = value[1 : len(value)-1]
			} else {
				value = unquoted
			}
		case len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'':
			value = value[1 : len(value)-1]
		}
		out[key] = value
	}
	return out, sc.Err()
}

func validEnvKey(k string) bool {
	if k == "" {
		return false
	}
	for i, r := range k {
		switch {
		case r == '_':
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

func defaultStateDir(stackHome string) string {
	return filepath.Join(stackHome, ".stacklane-state")
}

func loadProjectStacklaneEnv(projectDir string) (map[string]string, error) {
	return loadEnvFile(filepath.Join(projectDir, ".env.stacklane"))
}

func loadProjectRuntimeEnv(projectDir string) (map[string]string, error) {
	return loadEnvFile(filepath.Join(projectDir, ".env"))
}

// loadStackEnv reads the canonical stack-owned defaults file. .env.stacklane is
// the only supported file; legacy <stackHome>/.stackenv and <stackHome>/.env
// are intentionally NOT consulted (workspace legacy policy: no compat shims).
func loadStackEnv(stackHome string) (map[string]string, error) {
	return loadEnvFile(filepath.Join(stackHome, ".env.stacklane"))
}

func applyProjectRuntimeDBFallback(merged, runtimeEnv map[string]string) {
	defaults := defaults()
	for projectKey, stacklaneKey := range map[string]string{
		"DB_DATABASE": "MYSQL_DATABASE",
		"DB_USERNAME": "MYSQL_USER",
		"DB_PASSWORD": "MYSQL_PASSWORD",
	} {
		value := strings.TrimSpace(runtimeEnv[projectKey])
		if value == "" {
			continue
		}
		if merged[stacklaneKey] == defaults[stacklaneKey] {
			merged[stacklaneKey] = value
			merged["STACKLANE_PROJECT_ENV_"+stacklaneKey] = "1"
		}
	}
}

// resolveStackHome reproduces stacklane_default_stack_home.
func (l *Loader) resolveStackHome() (string, error) {
	if l.StackHomeOverride != "" {
		return project.AbsDir(l.StackHomeOverride)
	}
	if v := l.envOrDefault("STACK_HOME", ""); v != "" {
		return project.AbsDir(v)
	}
	// Walk up from this binary's location to find docker-compose.shared.yml.
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if _, statErr := os.Stat(filepath.Join(dir, "docker-compose.shared.yml")); statErr == nil {
			return dir, nil
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		if _, statErr := os.Stat(filepath.Join(cwd, "docker-compose.shared.yml")); statErr == nil {
			return cwd, nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "docker", "stacklane"), nil
}

// Load implements ConfigLoader.
func (l *Loader) Load(projectDir string, flags CLIFlags) (ProjectConfig, error) {
	cfg := ProjectConfig{}

	// 1. Project dir resolution (CLI wins, otherwise cwd).
	pd := flags.ProjectDir
	if pd == "" {
		pd = projectDir
	}
	if pd == "" {
		var err error
		pd, err = os.Getwd()
		if err != nil {
			return cfg, fmt.Errorf("resolve project dir: %w", err)
		}
	}
	pdAbs, err := project.AbsDir(pd)
	if err != nil {
		return cfg, fmt.Errorf("resolve project dir: %w", err)
	}
	cfg.Dir = pdAbs

	// 2. Stack home + state dir.
	stackHome, err := l.resolveStackHome()
	if err != nil {
		return cfg, fmt.Errorf("resolve stack home: %w", err)
	}
	cfg.StackHome = stackHome
	stateDir := l.envOrDefault("STACK_STATE_DIR", defaultStateDir(stackHome))
	cfg.StateDir = project.AbsPathFromBase(stackHome, stateDir)
	cfg.SharedFile = filepath.Join(stackHome, "docker-compose.shared.yml")

	// 3. Build the precedence-merged map. Lower precedence first; higher
	// precedence overwrites by key. Order: defaults -> stacklane .env ->
	// project runtime .env DB fallback -> shell env -> project .env.stacklane ->
	// CLI flags.
	merged := defaults()

	// .env.stacklane in the stack home applies just above built-in defaults.
	// STACKLANE_POST_UP_COMMAND is excluded from this merge: bootstrap is a
	// project-scoped declaration sourced only from project .env.stacklane.
	if envMap, err := loadStackEnv(stackHome); err == nil {
		for k, v := range envMap {
			if k == "STACKLANE_POST_UP_COMMAND" {
				continue
			}
			merged[k] = v
		}
	}
	// Project runtime .env stays separate from Stacklane config. We only use it
	// as a fallback source for the app's DB identity so the provisioned MariaDB
	// service can match the mounted project.
	if envMap, err := loadProjectRuntimeEnv(pdAbs); err == nil {
		applyProjectRuntimeDBFallback(merged, envMap)
	}
	// shell env: only consider keys we care about, to avoid leaking unrelated env.
	for _, k := range trackedEnvKeys {
		if v, ok := l.lookupEnv(k); ok && v != "" {
			merged[k] = v
		}
	}
	// Project .env.stacklane
	if envMap, err := loadProjectStacklaneEnv(pdAbs); err == nil {
		for k, v := range envMap {
			if projectEnvDisallowedKeys[k] {
				continue
			}
			merged[k] = v
		}
	}
	// CLI flags
	if flags.SiteName != "" {
		merged["SITE_NAME"] = flags.SiteName
	}
	if flags.SiteHostname != "" {
		merged["SITE_HOSTNAME"] = flags.SiteHostname
	}
	if flags.SiteSuffix != "" {
		merged["SITE_SUFFIX"] = flags.SiteSuffix
	}
	if flags.DocRoot != "" {
		merged["DOCROOT"] = flags.DocRoot
	}
	if flags.PHPVersion != "" {
		merged["PHP_VERSION"] = flags.PHPVersion
	}
	if flags.MySQLDatabase != "" {
		merged["MYSQL_DATABASE"] = flags.MySQLDatabase
	}
	if flags.MySQLUser != "" {
		merged["MYSQL_USER"] = flags.MySQLUser
	}
	if flags.MySQLPassword != "" {
		merged["MYSQL_PASSWORD"] = flags.MySQLPassword
	}
	if flags.MySQLPort != "" {
		merged["MYSQL_PORT"] = flags.MySQLPort
	}
	if flags.PMAPort != "" {
		merged["PMA_PORT"] = flags.PMAPort
	}
	if flags.HostPort != "" {
		merged["HOST_PORT"] = flags.HostPort
	}

	// 4. Materialise ProjectConfig from the merged map.
	stackKind := normalizeStackKind(merged["STACKLANE_STACK"])
	if stackKind == "" {
		stackKind = "20i"
	}
	if stackKind != "20i" {
		return cfg, fmt.Errorf("unsupported STACKLANE_STACK %q: only 20i is implemented today", stackKind)
	}
	cfg.StackKind = stackKind
	cfg.StackFile = filepath.Join(stackHome, stackComposeFileName(stackKind))
	cfg.Name = strOr(merged["SITE_NAME"], filepath.Base(pdAbs))
	cfg.Slug = project.Slugify(cfg.Name)
	cfg.SiteSuffix = strOr(merged["SITE_SUFFIX"], "test")
	cfg.Hostname, cfg.SiteSuffix = project.ResolveHostname(cfg.Slug, merged["SITE_HOSTNAME"], cfg.SiteSuffix)

	cfg.ComposeProjectName = strOr(merged["COMPOSE_PROJECT_NAME"], "stln-"+cfg.Slug)
	cfg.WebNetworkAlias = strOr(merged["WEB_NETWORK_ALIAS"], "stln-"+cfg.Slug+"-web")
	cfg.ContainerSiteRoot = "/home/sites/" + cfg.Slug
	cfg.RuntimeNetwork = cfg.ComposeProjectName + "-runtime"
	cfg.DatabaseVolume = cfg.ComposeProjectName + "-db-data"

	docroot, rel, err := project.ResolveDocRoot(pdAbs, merged["DOCROOT"], merged["CODE_DIR"])
	if err != nil {
		return cfg, err
	}
	cfg.DocRoot = docroot
	cfg.DocRootRelative = rel
	if rel == "" {
		cfg.ContainerDocRoot = cfg.ContainerSiteRoot
	} else {
		cfg.ContainerDocRoot = cfg.ContainerSiteRoot + "/" + rel
	}

	cfg.PHPVersion = strOr(merged["PHP_VERSION"], "8.5")

	// MySQL defaults key off the slug.
	cfg.MySQL.Version = strOr(merged["MYSQL_VERSION"], "10.6")
	cfg.MySQL.RootPassword = strOr(merged["MYSQL_ROOT_PASSWORD"], "root")
	cfg.MySQL.Database = strOr(merged["MYSQL_DATABASE"], "devdb")
	if cfg.MySQL.Database == "devdb" && merged["STACKLANE_PROJECT_ENV_MYSQL_DATABASE"] != "1" {
		cfg.MySQL.Database = cfg.Slug
	}
	cfg.MySQL.User = strOr(merged["MYSQL_USER"], "devuser")
	if cfg.MySQL.User == "devuser" && merged["STACKLANE_PROJECT_ENV_MYSQL_USER"] != "1" {
		cfg.MySQL.User = cfg.Slug
	}
	cfg.MySQL.Password = strOr(merged["MYSQL_PASSWORD"], "devpass")

	// Shared gateway settings are runtime-owned, not env-configurable.
	cfg.SharedGateway.Network = "stln-shared"
	cfg.SharedGateway.HTTPPort = 80
	cfg.SharedGateway.HTTPSPort = 443
	cfg.SharedGateway.ComposeProjectName = "stln-shared"
	cfg.SharedGateway.ConfigFile = filepath.Join(cfg.StateDir, "shared", "gateway.conf")

	// Local DNS.
	cfg.LocalDNS.Provider = strOr(merged["LOCAL_DNS_PROVIDER"], "dnsmasq")
	cfg.LocalDNS.IP = strOr(merged["LOCAL_DNS_IP"], "127.0.0.1")
	cfg.LocalDNS.Port = atoiOr(merged["LOCAL_DNS_PORT"], 53535)
	cfg.LocalDNS.Suffix = strOr(merged["LOCAL_DNS_SUFFIX"], cfg.SiteSuffix)

	cfg.Ports.HostPort = atoiOr(merged["HOST_PORT"], 0)
	cfg.Ports.MySQLPort = atoiOr(merged["MYSQL_PORT"], 0)
	cfg.Ports.PMAPort = atoiOr(merged["PMA_PORT"], 0)
	cfg.MySQL.Port = cfg.Ports.MySQLPort
	cfg.MySQL.PMAPort = cfg.Ports.PMAPort

	// Wait timeout: CLI > env > default (FR-009).
	switch {
	case flags.WaitTimeoutSecs > 0:
		cfg.WaitTimeoutSecs = flags.WaitTimeoutSecs
	case merged["STACKLANE_WAIT_TIMEOUT"] != "":
		cfg.WaitTimeoutSecs = atoiOr(merged["STACKLANE_WAIT_TIMEOUT"], 120)
	default:
		cfg.WaitTimeoutSecs = 120
	}
	cfg.PostUpCommand = merged["STACKLANE_POST_UP_COMMAND"]

	return cfg, nil
}

// trackedEnvKeys is the closed set of shell variables ConfigLoader honours.
// STACKLANE_POST_UP_COMMAND is intentionally absent: bootstrap is sourced
// only from project .env.stacklane (FR-016).
var trackedEnvKeys = []string{
	"STACKLANE_STACK",
	"SITE_NAME", "SITE_HOSTNAME", "SITE_SUFFIX", "DOCROOT", "CODE_DIR",
	"PHP_VERSION", "MYSQL_VERSION", "MYSQL_ROOT_PASSWORD",
	"MYSQL_DATABASE", "MYSQL_USER", "MYSQL_PASSWORD", "MYSQL_PORT", "PMA_PORT",
	"HOST_PORT", "COMPOSE_PROJECT_NAME", "WEB_NETWORK_ALIAS",
	"LOCAL_DNS_PROVIDER", "LOCAL_DNS_IP", "LOCAL_DNS_PORT", "LOCAL_DNS_SUFFIX",
	"STACK_HOME", "STACK_STATE_DIR", "STACKLANE_WAIT_TIMEOUT",
}

var projectEnvDisallowedKeys = map[string]bool{}

func (l *Loader) lookupEnv(k string) (string, bool) {
	get := l.Env
	if get == nil {
		get = os.LookupEnv
	}
	return get(k)
}

func defaults() map[string]string {
	return map[string]string{
		"STACKLANE_STACK":     "20i",
		"MYSQL_VERSION":       "10.6",
		"MYSQL_ROOT_PASSWORD": "root",
		"MYSQL_DATABASE":      "devdb",
		"MYSQL_USER":          "devuser",
		"MYSQL_PASSWORD":      "devpass",
		"PHP_VERSION":         "8.5",
		"LOCAL_DNS_PROVIDER":  "dnsmasq",
		"LOCAL_DNS_IP":        "127.0.0.1",
		"LOCAL_DNS_PORT":      "53535",
		"SITE_SUFFIX":         "test",
	}
}

func normalizeStackKind(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func stackComposeFileName(stackKind string) string {
	return "docker-compose." + stackKind + ".yml"
}

func strOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func atoiOr(v string, fallback int) int {
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
