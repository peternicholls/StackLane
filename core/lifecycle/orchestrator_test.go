// Lifecycle orchestrator tests using the in-memory mocks. The full Up flow
// is exercised end-to-end: ensure-network, allocate, ensure shared gateway,
// compose up, wait healthy, gateway add+reload, save state.
package lifecycle_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/peternicholls/stacklane/core/config"
	"github.com/peternicholls/stacklane/core/lifecycle"
	"github.com/peternicholls/stacklane/core/state"
	"github.com/peternicholls/stacklane/infra/docker"
	"github.com/peternicholls/stacklane/internal/mocks"
	"github.com/peternicholls/stacklane/platform/ports"
)

func newCfg(t *testing.T) config.ProjectConfig {
	t.Helper()
	stateDir := t.TempDir()
	dir := t.TempDir()
	stack := t.TempDir()
	return config.ProjectConfig{
		Slug:               "demo",
		Name:               "demo",
		Dir:                dir,
		StackHome:          stack,
		StateDir:           stateDir,
		StackFile:          stack + "/docker-compose.yml",
		SharedFile:         stack + "/docker-compose.shared.yml",
		Hostname:           "demo.test",
		ComposeProjectName: "stln-demo",
		WebNetworkAlias:    "stln-demo-web",
		ContainerSiteRoot:  "/home/sites/demo", ContainerDocRoot: "/home/sites/demo",
		PHPVersion:      "8.5",
		WaitTimeoutSecs: 5,
		MySQL: config.MySQL{
			Version: "10.6", Database: "demo", User: "demo", Password: "demo", RootPassword: "root",
		},
		SharedGateway: config.SharedGateway{
			Network:            "stacklane-shared",
			ComposeProjectName: "stacklane-shared",
			HTTPPort:           80,
			HTTPSPort:          443,
			ConfigFile:         stateDir + "/shared/gateway.conf",
		},
	}
}

func TestOrchestrator_UpHappyPath(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	dc.Containers = []docker.Container{{
		ID: "c1", Name: "stln-demo-nginx", Service: "nginx", Status: "running",
		Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "nginx"},
	}}
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	if err := orch.Up(context.Background(), cfg); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if !dc.Networks["stacklane-shared"] {
		t.Errorf("shared network was not ensured")
	}
	if len(composer.UpCalls) < 2 {
		t.Errorf("expected gateway-up + project-up + reload, got %d up calls", len(composer.UpCalls))
	}
	if len(composer.UpCalls) > 0 {
		got := composer.UpCalls[0].Env
		if !slices.Contains(got, "SHARED_GATEWAY_CONFIG_FILE="+cfg.SharedGateway.ConfigFile) {
			t.Fatalf("shared gateway config env missing from first compose up: %+v", got)
		}
		if !slices.Contains(got, "SHARED_GATEWAY_HTTP_PORT=80") {
			t.Fatalf("shared gateway http port env missing from first compose up: %+v", got)
		}
	}
	rec, err := st.Load("demo")
	if err != nil {
		t.Fatalf("state not persisted: %v", err)
	}
	if rec.AttachmentState != state.StateAttached {
		t.Errorf("expected attached, got %s", rec.AttachmentState)
	}
	if len(gw.Routes) == 0 {
		t.Errorf("gateway not updated")
	}
}

func TestOrchestrator_UpRollbackOnHealthFail(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	dc.WaitErr = errors.New("simulated unhealthy")
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatalf("expected health failure error")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok {
		t.Fatalf("error not StepError: %v", err)
	}
	if se.Step != "wait-healthy" {
		t.Errorf("step=%q want wait-healthy", se.Step)
	}
	if len(composer.DownCalls) == 0 {
		t.Errorf("rollback did not invoke compose down")
	}
	if _, err := st.Load("demo"); err == nil {
		t.Errorf("state should NOT be saved on rollback")
	}
}

func TestOrchestrator_UpRunsPostUpHook(t *testing.T) {
	cfg := newCfg(t)
	cfg.PostUpCommand = "php artisan migrate --force --no-interaction"
	dc := mocks.NewDocker()
	dc.Containers = []docker.Container{
		{
			ID: "nginx-1", Name: "stln-demo-nginx", Service: "nginx", Status: "running",
			Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "nginx"},
		},
		{
			ID: "apache-1", Name: "stln-demo-apache", Service: "apache", Status: "running",
			Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "apache"},
		},
	}
	componer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: componer, Gateway: gw, State: st, Ports: pa,
	})

	if err := orch.Up(context.Background(), cfg); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if len(dc.ExecCalls) != 1 {
		t.Fatalf("exec calls=%d want 1", len(dc.ExecCalls))
	}
	if dc.ExecCalls[0].ContainerID != "apache-1" {
		t.Fatalf("exec container=%q want apache-1", dc.ExecCalls[0].ContainerID)
	}
	if !slices.Equal(dc.ExecCalls[0].Cmd, []string{"sh", "-lc", cfg.PostUpCommand}) {
		t.Fatalf("exec cmd=%v", dc.ExecCalls[0].Cmd)
	}
	if dc.ExecCalls[0].WorkingDir != cfg.ContainerSiteRoot {
		t.Fatalf("working dir=%q want %q", dc.ExecCalls[0].WorkingDir, cfg.ContainerSiteRoot)
	}
}

func TestOrchestrator_UpRollbackOnPostUpHookFailure(t *testing.T) {
	cfg := newCfg(t)
	cfg.PostUpCommand = "php artisan migrate --force --no-interaction"
	dc := mocks.NewDocker()
	dc.ExecErr = errors.New("hook failed")
	dc.Containers = []docker.Container{
		{
			ID: "nginx-1", Name: "stln-demo-nginx", Service: "nginx", Status: "running",
			Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "nginx"},
		},
		{
			ID: "apache-1", Name: "stln-demo-apache", Service: "apache", Status: "running",
			Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "apache"},
		},
	}
	componer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: componer, Gateway: gw, State: st, Ports: pa,
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected post-up hook failure")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok {
		t.Fatalf("error not StepError: %v", err)
	}
	if se.Step != "post-up-hook" {
		t.Fatalf("step=%q want post-up-hook", se.Step)
	}
	if len(componer.DownCalls) == 0 {
		t.Fatal("rollback did not invoke compose down")
	}
	if _, err := st.Load("demo"); err == nil {
		t.Fatal("state should NOT be saved on hook rollback")
	}
}

func TestOrchestrator_UpPortConflictBeforeDocker(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	pa.Err = errors.New("simulated reservation conflict")

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatalf("expected port conflict error")
	}
	se, _ := lifecycle.AsStepError(err)
	if se == nil || se.Step != "allocate-ports" {
		t.Errorf("step should be allocate-ports; got %+v", se)
	}
	if len(composer.UpCalls) != 0 {
		t.Errorf("compose up should not run on port failure")
	}
}

func TestOrchestrator_DownSavesDownStateAndReloadsGateway(t *testing.T) {
	cfg := newCfg(t)
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	_ = st.Save(state.Record{Project: cfg, AttachmentState: state.StateAttached})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: mocks.NewDocker(), Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	if err := orch.Down(context.Background(), cfg, false); err != nil {
		t.Fatalf("Down: %v", err)
	}
	if len(composer.DownCalls) != 1 {
		t.Fatalf("down calls=%d want 1", len(composer.DownCalls))
	}
	if len(composer.UpCalls) != 1 || !composer.UpCalls[0].ForceRecreate {
		t.Fatalf("gateway reload not requested with force recreate: %+v", composer.UpCalls)
	}
	rec, err := st.Load(cfg.Slug)
	if err != nil {
		t.Fatalf("load after down: %v", err)
	}
	if rec.AttachmentState != state.StateDown {
		t.Fatalf("attachment state=%s want down", rec.AttachmentState)
	}
	if len(gw.Routes) != 0 {
		t.Fatalf("gateway routes should be cleared, got %+v", gw.Routes)
	}
}

func TestOrchestrator_DetachRemovesStateAndReloadsGateway(t *testing.T) {
	cfg := newCfg(t)
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	_ = st.Save(state.Record{Project: cfg, AttachmentState: state.StateAttached})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: mocks.NewDocker(), Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	if err := orch.Detach(context.Background(), cfg); err != nil {
		t.Fatalf("Detach: %v", err)
	}
	if len(composer.DownCalls) != 1 {
		t.Fatalf("down calls=%d want 1", len(composer.DownCalls))
	}
	if len(composer.UpCalls) != 1 || !composer.UpCalls[0].ForceRecreate {
		t.Fatalf("gateway reload not requested with force recreate: %+v", composer.UpCalls)
	}
	if _, err := st.Load(cfg.Slug); !errors.Is(err, state.ErrNotFound) {
		t.Fatalf("state should be removed, got %v", err)
	}
	if len(gw.Routes) != 0 {
		t.Fatalf("gateway routes should be cleared, got %+v", gw.Routes)
	}
}

func TestOrchestrator_DownAllStopsEveryRecordedProject(t *testing.T) {
	cfg := newCfg(t)
	other := cfg
	other.Slug = "beta"
	other.Name = "beta"
	other.Hostname = "beta.test"
	other.ComposeProjectName = "stln-beta"
	other.WebNetworkAlias = "stln-beta-web"

	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	_ = st.Save(state.Record{Project: cfg, AttachmentState: state.StateAttached})
	_ = st.Save(state.Record{Project: other, AttachmentState: state.StateAttached})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: mocks.NewDocker(), Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	if err := orch.DownAll(context.Background(), cfg, true); err != nil {
		t.Fatalf("DownAll: %v", err)
	}
	if len(composer.DownCalls) != 2 {
		t.Fatalf("down calls=%d want 2", len(composer.DownCalls))
	}
	for _, call := range composer.DownCalls {
		if !call.RemoveVolumes {
			t.Fatalf("expected remove volumes on every down call: %+v", composer.DownCalls)
		}
	}
	if len(composer.UpCalls) != 1 || !composer.UpCalls[0].ForceRecreate {
		t.Fatalf("gateway reload not requested with force recreate: %+v", composer.UpCalls)
	}
	if _, err := st.Load(cfg.Slug); !errors.Is(err, state.ErrNotFound) {
		t.Fatalf("demo state should be removed, got %v", err)
	}
	if _, err := st.Load(other.Slug); !errors.Is(err, state.ErrNotFound) {
		t.Fatalf("beta state should be removed, got %v", err)
	}
	if len(gw.Routes) != 0 {
		t.Fatalf("gateway routes should be cleared, got %+v", gw.Routes)
	}
}

// TestOrchestrator_PostUpHookFailure_RollbackIsolation proves that a failed
// bootstrap hook in one project does not mutate another attached project's
// state record, recorded routes, or registry projection. This covers the
// post-readiness, pre-state-persist failure window for US2 / FR-006.
func TestOrchestrator_PostUpHookFailure_RollbackIsolation(t *testing.T) {
	cfg := newCfg(t)
	cfg.PostUpCommand = "exit 1"

	other := cfg
	other.Slug = "beta"
	other.Name = "beta"
	other.Hostname = "beta.test"
	other.ComposeProjectName = "stln-beta"
	other.WebNetworkAlias = "stln-beta-web"

	dc := mocks.NewDocker()
	dc.ExecErr = errors.New("hook failed")
	dc.Containers = []docker.Container{
		{
			ID: "apache-1", Name: "stln-demo-apache", Service: "apache", Status: "running",
			Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "apache"},
		},
	}
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	if err := st.Save(state.Record{Project: other, AttachmentState: state.StateAttached}); err != nil {
		t.Fatalf("seed other state: %v", err)
	}
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected post-up hook failure")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok || se.Step != "post-up-hook" {
		t.Fatalf("expected post-up-hook StepError, got %+v", err)
	}

	if _, err := st.Load(cfg.Slug); err == nil {
		t.Fatalf("rolled-back project %q should not be persisted in state", cfg.Slug)
	}

	rec, err := st.Load(other.Slug)
	if err != nil {
		t.Fatalf("isolated project %q lost from state: %v", other.Slug, err)
	}
	if rec.AttachmentState != state.StateAttached {
		t.Fatalf("isolated project %q state=%s want attached", other.Slug, rec.AttachmentState)
	}

	for _, r := range gw.Routes {
		if r.Slug == cfg.Slug {
			t.Fatalf("rolled-back project still has gateway route: %+v", r)
		}
	}
}

// TestOrchestrator_AttachAddsRouteAndMarksAttached covers US3's attach slice:
// the orchestrator marks the project attached, requests a gateway route using
// the stln-<slug>-web alias, and reloads the shared gateway.
func TestOrchestrator_AttachAddsRouteAndMarksAttached(t *testing.T) {
	cfg := newCfg(t)
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	if err := st.Save(state.Record{Project: cfg, AttachmentState: state.StateDown}); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	orch := lifecycle.New(lifecycle.Deps{
		Docker: mocks.NewDocker(), Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	if err := orch.Attach(context.Background(), cfg); err != nil {
		t.Fatalf("Attach: %v", err)
	}
	rec, err := st.Load(cfg.Slug)
	if err != nil {
		t.Fatalf("load after attach: %v", err)
	}
	if rec.AttachmentState != state.StateAttached {
		t.Fatalf("attachment state=%s want attached", rec.AttachmentState)
	}
	var foundAlias string
	for _, r := range gw.Routes {
		if r.Slug == cfg.Slug {
			foundAlias = r.WebNetworkAlias
			break
		}
	}
	if foundAlias == "" {
		t.Fatalf("attach route for %q missing from %+v", cfg.Slug, gw.Routes)
	}
	if foundAlias != "stln-demo-web" {
		t.Fatalf("attach route alias=%q want stln-demo-web", foundAlias)
	}
	if len(composer.UpCalls) != 1 || !composer.UpCalls[0].ForceRecreate {
		t.Fatalf("attach did not request gateway reload with force recreate: %+v", composer.UpCalls)
	}
}
