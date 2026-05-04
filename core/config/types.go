// Package config defines the typed ProjectConfig and the ConfigLoader contract.
//
// ConfigLoader replaces mutable shell-global configuration with a single struct
// returned from Load().
package config

// CLIFlags carries the highest-precedence configuration source: command-line flags
// the operator passed in. Empty fields fall through to the next layer
// (project .env.stageserve -> shell env -> stack .env.stageserve -> defaults).
type CLIFlags struct {
	ProjectDir      string
	SiteName        string
	SiteHostname    string
	SiteSuffix      string
	DocRoot         string
	PHPVersion      string
	MySQLDatabase   string
	MySQLUser       string
	MySQLPassword   string
	MySQLPort       string
	PMAPort         string
	HostPort        string
	WaitTimeoutSecs int // 0 = use default / env
}

// SharedGateway captures the resolved shared-gateway runtime settings. These were
// SHARED_GATEWAY_* globals in the bash implementation.
type SharedGateway struct {
	Network            string // fixed shared network name
	HTTPPort           int    // fixed HTTP host port (80)
	HTTPSPort          int    // fixed/default HTTPS host port (443, or runtime fallback for .dev)
	ComposeProjectName string // fixed shared compose project name
	ConfigFile         string // generated gateway config path
}

// LocalDNS captures the resolved local-DNS settings (LOCAL_DNS_* in bash).
type LocalDNS struct {
	Provider string // LOCAL_DNS_PROVIDER (default: dnsmasq)
	IP       string // LOCAL_DNS_IP (default: 127.0.0.1)
	Port     int    // LOCAL_DNS_PORT (default: 53535)
	Suffix   string // LOCAL_DNS_SUFFIX (derived from SITE_SUFFIX)
}

// MySQL captures resolved per-project DB settings.
type MySQL struct {
	Version      string // MYSQL_VERSION (default 10.6)
	RootPassword string // MYSQL_ROOT_PASSWORD (default root)
	Database     string // MYSQL_DATABASE (default <slug>)
	User         string // MYSQL_USER (default <slug>)
	Password     string // MYSQL_PASSWORD (default devpass)
	Port         int    // MYSQL_PORT
	PMAPort      int    // PMA_PORT
}

// PortAllocation is the typed view of a project's reserved ports. HTTP/HTTPS
// gateway ports live on the shared gateway, not on the per-project allocation.
type PortAllocation struct {
	HostPort  int
	MySQLPort int
	PMAPort   int
}

// ProjectConfig is the resolved view of a single project's settings after the
// precedence chain has been applied. Replaces the loose set of global shell
// values for the current project.
type ProjectConfig struct {
	// Identity
	StackKind          string // STAGESERVE_STACK
	Name               string // PROJECT_NAME
	Slug               string // PROJECT_SLUG
	Dir                string // PROJECT_DIR (absolute)
	Hostname           string // HOSTNAME
	SiteSuffix         string // SITE_SUFFIX
	ComposeProjectName string // COMPOSE_PROJECT_NAME
	WebNetworkAlias    string // WEB_NETWORK_ALIAS

	// Filesystem
	DocRoot           string // DOCROOT (absolute)
	DocRootRelative   string // DOCROOT_RELATIVE
	ContainerSiteRoot string // CONTAINER_SITE_ROOT
	ContainerDocRoot  string // CONTAINER_DOCROOT

	// Stack home / state dir
	StackHome  string // STACK_HOME (absolute)
	StateDir   string // STAGESERVE_STATE_DIR (absolute)
	StackFile  string // STAGESERVE_STACK_FILE (absolute)
	SharedFile string // STAGESERVE_SHARED_STACK_FILE (absolute)

	// Runtime
	PHPVersion      string // PHP_VERSION
	MySQL           MySQL
	Ports           PortAllocation
	RuntimeNetwork  string // PROJECT_RUNTIME_NETWORK
	DatabaseVolume  string // PROJECT_DATABASE_VOLUME
	SharedGateway   SharedGateway
	LocalDNS        LocalDNS
	WaitTimeoutSecs int    // FR-009: default 120
	PostUpCommand   string // STAGESERVE_POST_UP_COMMAND

	// Operator-visible flags
	DryRun  bool
	All     bool
	Profile string // "" or "debug" — opt-in for phpMyAdmin per FR-011

	// Selectors / extras (reserved for status/logs commands)
	ProjectSelector string
}

// ConfigLoader resolves the full precedence chain (CLI flags -> project
// .env.stageserve -> shell env -> stack .env.stageserve -> defaults) and returns
// a populated ProjectConfig.
// STAGESERVE_POST_UP_COMMAND is the one project-scoped exception: it is honored
// only when set in project .env.stageserve.
//
// Implementations must NOT depend on Docker, the network, or any subsystem
// outside the local filesystem.
type ConfigLoader interface {
	Load(projectDir string, flags CLIFlags) (ProjectConfig, error)
}
