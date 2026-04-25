// Orchestrator wires lifecycle steps over the lower-level interfaces. Up
// follows the documented 11-step flow with rollback at steps 6–9; Down /
// Attach / Detach / Status / Logs delegate to the same interfaces.
package lifecycle

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/peternicholls/stacklane/core/config"
	"github.com/peternicholls/stacklane/core/state"
	"github.com/peternicholls/stacklane/infra/compose"
	"github.com/peternicholls/stacklane/infra/docker"
	"github.com/peternicholls/stacklane/infra/gateway"
	"github.com/peternicholls/stacklane/platform/ports"
)

// Deps bundles the collaborators the orchestrator needs.
type Deps struct {
	Docker  docker.DockerClient
	Compose compose.Composer
	Gateway gateway.GatewayManager
	State   state.StateStore
	Ports   ports.PortAllocator
}

// Orchestrator is the default implementation.
type Orchestrator struct {
	D Deps
}

const sharedGatewayHTTPSFallbackStart = 8443

var sharedGatewayListen = net.Listen

// New returns an orchestrator wired to deps.
func New(d Deps) *Orchestrator { return &Orchestrator{D: d} }

// Up runs the documented 11-step flow.
func (o *Orchestrator) Up(ctx context.Context, cfg config.ProjectConfig) error {
	cfg = resolveSharedGatewayPorts(cfg)

	// Step 1: ensure shared network exists.
	if err := o.ensureSharedNetwork(ctx, cfg); err != nil {
		return Wrap("ensure-shared-network", cfg.Slug, err, "Verify Docker is running and the shared network can be created.")
	}
	// Step 2: allocate ports.
	registry, err := o.D.State.Registry()
	if err != nil {
		return Wrap("registry", cfg.Slug, err, "Inspect the state directory for unreadable JSON files.")
	}
	allocation, err := o.D.Ports.Allocate(ports.Request{
		HostPort:     cfg.Ports.HostPort,
		MySQLPort:    cfg.Ports.MySQLPort,
		PMAPort:      cfg.Ports.PMAPort,
		IsUp:         true,
		OwnSlug:      cfg.Slug,
		ProjectCount: countOtherActive(registry, cfg.Slug),
	}, registry)
	if err != nil {
		return Wrap("allocate-ports", cfg.Slug, fmt.Errorf("%w: %v", ErrPortConflict, err), "Free the conflicting port or pass --mysql-port / --pma-port.")
	}
	cfg.Ports.HostPort = allocation.HostPort
	cfg.Ports.MySQLPort = allocation.MySQLPort
	cfg.Ports.PMAPort = allocation.PMAPort
	cfg.MySQL.Port = allocation.MySQLPort
	cfg.MySQL.PMAPort = allocation.PMAPort

	if err := o.prepareSharedGatewayConfig(cfg, routesFromRegistry(registry), ""); err != nil {
		return Wrap("gateway-config", "", err, "Inspect the gateway config path under the state directory.")
	}

	// Step 4: ensure shared gateway is running.
	if err := o.ensureSharedGateway(ctx, cfg); err != nil {
		return Wrap("shared-gateway", "", err, "Run `docker compose -p stln-shared up -d` and inspect logs.")
	}

	// Step 5: write per-project compose env file (synthesized from cfg).
	envFile, err := writeEnvFile(cfg)
	if err != nil {
		return Wrap("write-env-file", cfg.Slug, err, "Check the project directory is writable.")
	}

	// Step 6: docker compose up --wait.
	composeOpts := compose.UpOptions{
		ProjectDir:  cfg.Dir,
		ComposeFile: cfg.StackFile,
		ProjectName: cfg.ComposeProjectName,
		EnvFile:     envFile,
		Detach:      true,
		WaitTimeout: time.Duration(cfg.WaitTimeoutSecs) * time.Second,
	}
	if err := o.D.Compose.Up(ctx, composeOpts); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("compose-up", cfg.Slug, err, "Check `stacklane logs` for the failing service.")
	}

	// Step 7: wait for healthchecks.
	if err := o.D.Docker.WaitHealthy(ctx, cfg.ComposeProjectName, time.Duration(cfg.WaitTimeoutSecs)*time.Second); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("wait-healthy", cfg.Slug, err, "Inspect container health with `docker ps` then `stacklane logs`.")
	}
	if err := o.runPostUpHook(ctx, cfg); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("post-up-hook", cfg.Slug, err, "Check STACKLANE_POST_UP_COMMAND and verify it succeeds inside the apache container.")
	}

	// Step 8: regenerate gateway config, reload gateway.
	currentRoutes := routesFromRegistry(registry)
	if _, _, err := o.D.Gateway.AddRoute(gateway.Route{
		Hostname:        cfg.Hostname,
		Slug:            cfg.Slug,
		WebNetworkAlias: cfg.WebNetworkAlias,
	}, currentRoutes); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("gateway-config", cfg.Slug, err, "Inspect the gateway config path under the state directory.")
	}
	if err := o.reloadSharedGateway(ctx, cfg); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("gateway-reload", cfg.Slug, err, "Run `docker compose -p stln-shared up -d --force-recreate gateway`.")
	}

	// Step 9: persist state.
	rec := state.Record{
		SchemaVersion:   state.SchemaVersion,
		Project:         cfg,
		AttachmentState: state.StateAttached,
		Runtime:         observedRuntime(ctx, o.D.Docker, cfg),
	}
	if err := o.D.State.Save(rec); err != nil {
		o.rollbackProject(ctx, cfg)
		return Wrap("save-state", cfg.Slug, err, "Inspect permissions on the state directory.")
	}

	// Step 10/11: log success (caller handles human output).
	return nil
}

// Down stops the project, keeps its record, and removes any active route.
func (o *Orchestrator) Down(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	cfg = resolveSharedGatewayPorts(cfg)

	if err := o.stopProject(ctx, cfg, removeVolumes); err != nil {
		return Wrap("compose-down", cfg.Slug, err, "Inspect docker compose output above.")
	}
	rec, err := o.D.State.Load(cfg.Slug)
	if err != nil {
		rec = state.Record{Project: cfg}
	}
	rec.Project = cfg
	rec.AttachmentState = state.StateDown
	if err := o.D.State.Save(rec); err != nil {
		return Wrap("save-state", cfg.Slug, err, "Inspect permissions on the state directory.")
	}
	if err := o.syncSharedGateway(ctx, cfg, ""); err != nil {
		return Wrap("gateway-reload", cfg.Slug, err, "Run `docker compose -p stln-shared up -d gateway`.")
	}
	return nil
}

// DownAll stops every recorded project runtime, removes all state records, and
// clears the shared gateway route set.
func (o *Orchestrator) DownAll(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	cfg = resolveSharedGatewayPorts(cfg)

	registry, err := o.D.State.Registry()
	if err != nil {
		return Wrap("registry", "", err, "Inspect the state directory for unreadable JSON files.")
	}
	for _, row := range registry {
		rec, err := o.D.State.Load(row.Slug)
		if err != nil {
			return Wrap("load-state", row.Slug, err, "Inspect the recorded state for this project.")
		}
		if err := o.stopProject(ctx, rec.Project, removeVolumes); err != nil {
			return Wrap("compose-down", row.Slug, err, "Inspect docker compose output above.")
		}
		if err := o.D.State.Remove(row.Slug); err != nil {
			return Wrap("remove-state", row.Slug, err, "Inspect permissions on the state directory.")
		}
	}
	if err := o.syncSharedGateway(ctx, cfg, ""); err != nil {
		return Wrap("gateway-reload", "", err, "Run `docker compose -p stln-shared up -d gateway`.")
	}
	return nil
}

// Attach updates state + gateway to mark the project routed.
func (o *Orchestrator) Attach(ctx context.Context, cfg config.ProjectConfig) error {
	cfg = resolveSharedGatewayPorts(cfg)

	rec, err := o.D.State.Load(cfg.Slug)
	if err != nil {
		return Wrap("attach", cfg.Slug, err, "Run `stacklane up` first.")
	}
	rec.AttachmentState = state.StateAttached
	if err := o.D.State.Save(rec); err != nil {
		return Wrap("save-state", cfg.Slug, err, "")
	}
	registry, _ := o.D.State.Registry()
	currentRoutes := routesFromRegistry(registry)
	if _, _, err := o.D.Gateway.AddRoute(gateway.Route{
		Hostname:        cfg.Hostname,
		Slug:            cfg.Slug,
		WebNetworkAlias: cfg.WebNetworkAlias,
	}, currentRoutes); err != nil {
		return Wrap("gateway-config", cfg.Slug, err, "")
	}
	return o.reloadSharedGateway(ctx, cfg)
}

// Detach stops the project, removes its state record, and clears its route.
func (o *Orchestrator) Detach(ctx context.Context, cfg config.ProjectConfig) error {
	cfg = resolveSharedGatewayPorts(cfg)

	if err := o.stopProject(ctx, cfg, false); err != nil {
		return Wrap("compose-down", cfg.Slug, err, "Inspect docker compose output above.")
	}
	if err := o.D.State.Remove(cfg.Slug); err != nil {
		return Wrap("remove-state", cfg.Slug, err, "Inspect permissions on the state directory.")
	}
	if err := o.syncSharedGateway(ctx, cfg, ""); err != nil {
		return Wrap("gateway-reload", cfg.Slug, err, "Run `docker compose -p stln-shared up -d gateway`.")
	}
	return nil
}

// --- helpers ---

func (o *Orchestrator) ensureSharedNetwork(ctx context.Context, cfg config.ProjectConfig) error {
	exists, err := o.D.Docker.NetworkExists(ctx, cfg.SharedGateway.Network)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return o.D.Docker.CreateNetwork(ctx, cfg.SharedGateway.Network)
}

func (o *Orchestrator) ensureSharedGateway(ctx context.Context, cfg config.ProjectConfig) error {
	return o.D.Compose.Up(ctx, compose.UpOptions{
		ProjectDir:  cfg.StackHome,
		ComposeFile: cfg.SharedFile,
		ProjectName: cfg.SharedGateway.ComposeProjectName,
		Env:         sharedGatewayEnv(cfg),
		Detach:      true,
		WaitTimeout: time.Duration(cfg.WaitTimeoutSecs) * time.Second,
	})
}

func (o *Orchestrator) reloadSharedGateway(ctx context.Context, cfg config.ProjectConfig) error {
	return o.D.Compose.Up(ctx, compose.UpOptions{
		ProjectDir:    cfg.StackHome,
		ComposeFile:   cfg.SharedFile,
		ProjectName:   cfg.SharedGateway.ComposeProjectName,
		Env:           sharedGatewayEnv(cfg),
		Detach:        true,
		ForceRecreate: true,
		Services:      []string{"gateway"},
	})
}

func (o *Orchestrator) stopProject(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	envFile := envFilePath(cfg)
	return o.D.Compose.Down(ctx, compose.DownOptions{
		ProjectDir:    cfg.Dir,
		ComposeFile:   cfg.StackFile,
		ProjectName:   cfg.ComposeProjectName,
		EnvFile:       envFile,
		RemoveVolumes: removeVolumes,
	})
}

func (o *Orchestrator) syncSharedGateway(ctx context.Context, cfg config.ProjectConfig, preferredSlug string) error {
	registry, err := o.D.State.Registry()
	if err != nil {
		return err
	}
	if err := o.prepareSharedGatewayConfig(cfg, routesFromRegistry(registry), preferredSlug); err != nil {
		return err
	}
	return o.reloadSharedGateway(ctx, cfg)
}

func (o *Orchestrator) prepareSharedGatewayConfig(cfg config.ProjectConfig, routes []gateway.Route, preferredSlug string) error {
	_, _, err := o.D.Gateway.WriteConfig(gateway.RenderInput{
		Routes:        routes,
		PreferredSlug: preferredSlug,
		TLSEnabled:    cfg.SiteSuffix == "dev",
		HTTPSPort:     cfg.SharedGateway.HTTPSPort,
	})
	return err
}

func (o *Orchestrator) runPostUpHook(ctx context.Context, cfg config.ProjectConfig) error {
	if strings.TrimSpace(cfg.PostUpCommand) == "" {
		return nil
	}
	containerID, err := o.findServiceContainer(ctx, cfg.ComposeProjectName, "apache")
	if err != nil {
		return err
	}
	_, err = o.D.Docker.Exec(ctx, docker.ExecOptions{
		ContainerID: containerID,
		Cmd:         []string{"sh", "-lc", cfg.PostUpCommand},
		WorkingDir:  cfg.ContainerSiteRoot,
	})
	return err
}

func (o *Orchestrator) findServiceContainer(ctx context.Context, projectName, service string) (string, error) {
	containers, err := o.D.Docker.ListContainersByLabel(ctx, map[string]string{
		"com.docker.compose.project": projectName,
		"com.docker.compose.service": service,
	})
	if err != nil {
		return "", err
	}
	for _, container := range containers {
		if container.Service == service {
			return container.ID, nil
		}
	}
	return "", fmt.Errorf("%s container not found for compose project %s", service, projectName)
}

func (o *Orchestrator) rollbackProject(ctx context.Context, cfg config.ProjectConfig) {
	envFile := envFilePath(cfg)
	_ = o.D.Compose.Down(ctx, compose.DownOptions{
		ProjectDir:  cfg.Dir,
		ComposeFile: cfg.StackFile,
		ProjectName: cfg.ComposeProjectName,
		EnvFile:     envFile,
	})
}

func envFilePath(cfg config.ProjectConfig) string {
	return filepath.Join(cfg.StateDir, "envfiles", cfg.Slug+".env")
}

func writeEnvFile(cfg config.ProjectConfig) (string, error) {
	path := envFilePath(cfg)
	if err := mkdirAll(filepath.Dir(path)); err != nil {
		return "", err
	}
	body := strings.Join([]string{
		"COMPOSE_PROJECT_NAME=" + cfg.ComposeProjectName,
		"PROJECT_NAME=" + cfg.Name,
		"PROJECT_SLUG=" + cfg.Slug,
		"PROJECT_ROOT=" + cfg.Dir,
		"PROJECT_HOSTNAME=" + cfg.Hostname,
		"PROJECT_DOCROOT=" + cfg.DocRoot,
		"PROJECT_DIR=" + cfg.Dir,
		"DOCROOT=" + cfg.DocRoot,
		"CONTAINER_SITE_ROOT=" + cfg.ContainerSiteRoot,
		"CONTAINER_DOCROOT=" + cfg.ContainerDocRoot,
		"DB_HOST=mariadb",
		"DB_PORT=3306",
		"DB_DATABASE=" + cfg.MySQL.Database,
		"DB_USERNAME=" + cfg.MySQL.User,
		"DB_PASSWORD=" + cfg.MySQL.Password,
		"PHP_VERSION=" + cfg.PHPVersion,
		"MYSQL_VERSION=" + cfg.MySQL.Version,
		"MYSQL_DATABASE=" + cfg.MySQL.Database,
		"MYSQL_USER=" + cfg.MySQL.User,
		"MYSQL_PASSWORD=" + cfg.MySQL.Password,
		"MYSQL_ROOT_PASSWORD=" + cfg.MySQL.RootPassword,
		"MYSQL_PORT=" + intStr(cfg.MySQL.Port),
		"PMA_PORT=" + intStr(cfg.MySQL.PMAPort),
		"WEB_NETWORK_ALIAS=" + cfg.WebNetworkAlias,
		"PROJECT_RUNTIME_NETWORK=" + cfg.RuntimeNetwork,
		"PROJECT_DATABASE_VOLUME=" + cfg.DatabaseVolume,
		"SHARED_GATEWAY_NETWORK=" + cfg.SharedGateway.Network,
	}, "\n") + "\n"
	if err := writeFile(path, []byte(body)); err != nil {
		return "", err
	}
	return path, nil
}

func sharedGatewayEnv(cfg config.ProjectConfig) []string {
	return []string{
		"SHARED_GATEWAY_NETWORK=" + cfg.SharedGateway.Network,
		"SHARED_GATEWAY_HTTP_PORT=" + intStr(cfg.SharedGateway.HTTPPort),
		"SHARED_GATEWAY_HTTPS_PORT=" + intStr(cfg.SharedGateway.HTTPSPort),
		"SHARED_GATEWAY_CONFIG_FILE=" + cfg.SharedGateway.ConfigFile,
	}
}

func resolveSharedGatewayPorts(cfg config.ProjectConfig) config.ProjectConfig {
	if cfg.SharedGateway.HTTPSPort == 0 {
		cfg.SharedGateway.HTTPSPort = 443
	}
	if cfg.SiteSuffix == "dev" && cfg.SharedGateway.HTTPSPort == 443 {
		cfg.SharedGateway.HTTPSPort = sharedGatewayHTTPSFallbackStart
		return cfg
	}
	if cfg.SharedGateway.HTTPSPort != 443 || !sharedGatewayPortInUse(443) {
		return cfg
	}
	if fallback, ok := firstAvailableSharedGatewayPort(sharedGatewayHTTPSFallbackStart); ok {
		cfg.SharedGateway.HTTPSPort = fallback
	}
	return cfg
}

func sharedGatewayPortInUse(port int) bool {
	ln, err := sharedGatewayListen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

func firstAvailableSharedGatewayPort(start int) (int, bool) {
	for port := start; port < 65535; port++ {
		if !sharedGatewayPortInUse(port) {
			return port, true
		}
	}
	return 0, false
}

func intStr(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

func countOtherActive(rows []state.RegistryRow, ownSlug string) int {
	n := 0
	for _, r := range rows {
		if r.Slug == ownSlug {
			continue
		}
		if r.AttachmentState == state.StateAttached {
			n++
		}
	}
	return n
}

func routesFromRegistry(rows []state.RegistryRow) []gateway.Route {
	out := make([]gateway.Route, 0, len(rows))
	for _, r := range rows {
		if r.AttachmentState != state.StateAttached {
			continue
		}
		out = append(out, gateway.Route{
			Hostname:        r.Hostname,
			Slug:            r.Slug,
			WebNetworkAlias: r.WebNetworkAlias,
		})
	}
	return out
}

func observedRuntime(ctx context.Context, dc docker.DockerClient, cfg config.ProjectConfig) state.RuntimeIdentity {
	containers, err := dc.ListContainersByLabel(ctx, map[string]string{"com.docker.compose.project": cfg.ComposeProjectName})
	if err != nil {
		return state.RuntimeIdentity{}
	}
	var rt state.RuntimeIdentity
	names := make([]string, 0, len(containers))
	for _, c := range containers {
		names = append(names, c.Service+"="+c.Status)
		ident := state.ContainerIdentity{ID: c.ID, Name: c.Name, Status: c.Status}
		switch c.Service {
		case "nginx":
			rt.Nginx = ident
		case "apache":
			rt.Apache = ident
		case "mariadb":
			rt.MariaDB = ident
		case "phpmyadmin":
			rt.PhpMyAdmin = ident
		}
	}
	rt.SummaryLine = strings.Join(names, " ")
	return rt
}
