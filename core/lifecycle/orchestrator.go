// Orchestrator wires lifecycle steps over the lower-level interfaces. Up
// follows the documented 11-step flow with rollback at steps 6–9; Down /
// Attach / Detach / Status / Logs delegate to the same interfaces.
package lifecycle

import (
	"context"
	"fmt"
	"path/filepath"
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

// New returns an orchestrator wired to deps.
func New(d Deps) *Orchestrator { return &Orchestrator{D: d} }

// Up runs the documented 11-step flow.
func (o *Orchestrator) Up(ctx context.Context, cfg config.ProjectConfig) error {
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

	// Step 4: ensure shared gateway is running.
	if err := o.ensureSharedGateway(ctx, cfg); err != nil {
		return Wrap("shared-gateway", "", err, "Run `docker compose -p stacklane-shared up -d` and inspect logs.")
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
		return Wrap("gateway-reload", cfg.Slug, err, "Run `docker compose -p stacklane-shared up -d --force-recreate gateway`.")
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

// Down stops the project. AttachmentState is left intact when keepAttached is true
// (so detach + down semantics remain separable).
func (o *Orchestrator) Down(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	envFile := envFilePath(cfg)
	if err := o.D.Compose.Down(ctx, compose.DownOptions{
		ProjectDir:    cfg.Dir,
		ComposeFile:   cfg.StackFile,
		ProjectName:   cfg.ComposeProjectName,
		EnvFile:       envFile,
		RemoveVolumes: removeVolumes,
	}); err != nil {
		return Wrap("compose-down", cfg.Slug, err, "Inspect docker compose output above.")
	}
	if rec, err := o.D.State.Load(cfg.Slug); err == nil {
		rec.AttachmentState = state.StateDown
		_ = o.D.State.Save(rec)
	}
	return nil
}

// Attach updates state + gateway to mark the project routed.
func (o *Orchestrator) Attach(ctx context.Context, cfg config.ProjectConfig) error {
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

// Detach unroutes a project but leaves it running.
func (o *Orchestrator) Detach(ctx context.Context, cfg config.ProjectConfig) error {
	rec, err := o.D.State.Load(cfg.Slug)
	if err != nil {
		return Wrap("detach", cfg.Slug, err, "")
	}
	rec.AttachmentState = state.StateDown
	if err := o.D.State.Save(rec); err != nil {
		return Wrap("save-state", cfg.Slug, err, "")
	}
	registry, _ := o.D.State.Registry()
	currentRoutes := routesFromRegistry(registry)
	if _, _, err := o.D.Gateway.RemoveRoute(cfg.Slug, currentRoutes); err != nil {
		return Wrap("gateway-config", cfg.Slug, err, "")
	}
	return o.reloadSharedGateway(ctx, cfg)
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
		Detach:      true,
		WaitTimeout: time.Duration(cfg.WaitTimeoutSecs) * time.Second,
	})
}

func (o *Orchestrator) reloadSharedGateway(ctx context.Context, cfg config.ProjectConfig) error {
	return o.D.Compose.Up(ctx, compose.UpOptions{
		ProjectDir:    cfg.StackHome,
		ComposeFile:   cfg.SharedFile,
		ProjectName:   cfg.SharedGateway.ComposeProjectName,
		Detach:        true,
		ForceRecreate: true,
		Services:      []string{"gateway"},
	})
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
		"PROJECT_SLUG=" + cfg.Slug,
		"PROJECT_DIR=" + cfg.Dir,
		"DOCROOT=" + cfg.DocRoot,
		"CONTAINER_SITE_ROOT=" + cfg.ContainerSiteRoot,
		"CONTAINER_DOCROOT=" + cfg.ContainerDocRoot,
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
