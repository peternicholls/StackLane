// Lifecycle orchestrator tests using the in-memory mocks. The full Up flow
// is exercised end-to-end: ensure-network, allocate, ensure shared gateway,
// compose up, wait healthy, gateway add+reload, save state.
package lifecycle_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/lifecycle"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/docker"
	"github.com/peternicholls/stageserve/internal/mocks"
	"github.com/peternicholls/stageserve/platform/ports"
	stls "github.com/peternicholls/stageserve/platform/tls"
)

type tlsProviderStub struct {
	certFile string
	keyFile  string
	hosts    []string
	calls    int
}

func (p *tlsProviderStub) Available() bool { return true }

func (p *tlsProviderStub) Ensure(certFile, keyFile string, hosts []string) (stls.Bundle, error) {
	p.calls++
	p.certFile = certFile
	p.keyFile = keyFile
	p.hosts = append([]string(nil), hosts...)
	return stls.Bundle{CertFile: certFile, KeyFile: keyFile, Hosts: append([]string(nil), hosts...)}, nil
}

func newCfg(t *testing.T) config.ProjectConfig {
	t.Helper()
	stateDir := t.TempDir()
	dir := t.TempDir()
	stack := t.TempDir()
	stackAssets := filepath.Join(stack, "stacks", "20i")
	if err := os.MkdirAll(stackAssets, 0o755); err != nil {
		t.Fatalf("mkdir stack assets: %v", err)
	}
	for _, name := range []string{"docker-compose.20i.yml", "docker-compose.shared.yml"} {
		if err := os.WriteFile(filepath.Join(stackAssets, name), []byte("services: {}\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return config.ProjectConfig{
		Slug:               "demo",
		Name:               "demo",
		Dir:                dir,
		StackHome:          stack,
		StateDir:           stateDir,
		StackFile:          stack + "/stacks/20i/docker-compose.20i.yml",
		SharedFile:         stack + "/stacks/20i/docker-compose.shared.yml",
		Hostname:           "demo.test",
		ComposeProjectName: "stage-demo",
		WebNetworkAlias:    "stage-demo-web",
		ContainerSiteRoot:  "/home/sites/demo", ContainerDocRoot: "/home/sites/demo",
		PHPVersion:      "8.5",
		WaitTimeoutSecs: 5,
		MySQL: config.MySQL{
			Version: "10.6", Database: "demo", User: "demo", Password: "demo", RootPassword: "root",
		},
		SharedGateway: config.SharedGateway{
			Network:            "stage-shared",
			ComposeProjectName: "stage-shared",
			HTTPPort:           80,
			HTTPSPort:          443,
			ConfigFile:         stateDir + "/shared/gateway.conf",
		},
	}
}

func TestOrchestrator_UpFailsBeforeDockerWhenRuntimeAssetMissing(t *testing.T) {
	cfg := newCfg(t)
	if err := os.Remove(cfg.SharedFile); err != nil {
		t.Fatalf("remove shared file: %v", err)
	}
	dc := mocks.NewDocker()
	composer := mocks.NewComposer()
	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: mocks.NewGateway(), State: mocks.NewState(), Ports: mocks.NewPorts(ports.Allocation{}),
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected missing runtime asset error")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok {
		t.Fatalf("error not StepError: %v", err)
	}
	if se.Step != "runtime-asset-shared" {
		t.Fatalf("step=%q want runtime-asset-shared", se.Step)
	}
	if !strings.Contains(se.Remedy, "Reinstall StageServe") || !strings.Contains(se.Remedy, cfg.SharedFile) {
		t.Fatalf("remedy=%q", se.Remedy)
	}
	if dc.Networks["stage-shared"] || len(composer.UpCalls) != 0 {
		t.Fatalf("runtime preflight should run before Docker/compose side effects")
	}
}

func TestOrchestrator_UpFailsBeforeDockerWhenProjectRuntimeAssetMissing(t *testing.T) {
	cfg := newCfg(t)
	if err := os.Remove(cfg.StackFile); err != nil {
		t.Fatalf("remove project file: %v", err)
	}
	composer := mocks.NewComposer()
	orch := lifecycle.New(lifecycle.Deps{
		Docker: mocks.NewDocker(), Compose: composer, Gateway: mocks.NewGateway(), State: mocks.NewState(), Ports: mocks.NewPorts(ports.Allocation{}),
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected missing project runtime asset error")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok || se.Step != "runtime-asset-project" {
		t.Fatalf("step error=%+v ok=%v", se, ok)
	}
	if !strings.Contains(se.Remedy, cfg.StackFile) {
		t.Fatalf("remedy=%q", se.Remedy)
	}
	if len(composer.UpCalls) != 0 {
		t.Fatalf("compose up should not run before project runtime asset preflight")
	}
}

func TestOrchestrator_UpHappyPath(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	dc.Containers = []docker.Container{{
		ID: "c1", Name: "stage-demo-nginx", Service: "nginx", Status: "running",
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
	if !dc.Networks["stage-shared"] {
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

func TestOrchestrator_RestartServiceOnlyCallsProjectCompose(t *testing.T) {
	cfg := newCfg(t)
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	st.Records[cfg.Slug] = state.Record{Project: cfg, AttachmentState: state.StateAttached}
	orch := lifecycle.New(lifecycle.Deps{
		Docker: mocks.NewDocker(), Compose: composer, Gateway: gw, State: st, Ports: mocks.NewPorts(ports.Allocation{}),
	})

	if err := orch.RestartService(context.Background(), cfg, "apache"); err != nil {
		t.Fatalf("RestartService: %v", err)
	}
	if len(composer.RestartCalls) != 1 {
		t.Fatalf("restart calls=%d want 1", len(composer.RestartCalls))
	}
	call := composer.RestartCalls[0]
	if call.ProjectDir != cfg.Dir || call.ComposeFile != cfg.StackFile || call.ProjectName != cfg.ComposeProjectName || call.Service != "apache" {
		t.Fatalf("restart call=%+v", call)
	}
	if len(gw.Routes) != 0 {
		t.Fatalf("restart should not change gateway routes: %+v", gw.Routes)
	}
	if _, err := st.Load(cfg.Slug); err != nil {
		t.Fatalf("restart should not remove state record: %v", err)
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
			ID: "nginx-1", Name: "stage-demo-nginx", Service: "nginx", Status: "running",
			Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "nginx"},
		},
		{
			ID: "apache-1", Name: "stage-demo-apache", Service: "apache", Status: "running",
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

func TestOrchestrator_UpPassesDebugProfileToProjectCompose(t *testing.T) {
	cfg := newCfg(t)
	cfg.Profile = "debug"
	dc := mocks.NewDocker()
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
	if len(composer.UpCalls) < 2 {
		t.Fatalf("compose up calls=%d want at least 2", len(composer.UpCalls))
	}
	if len(composer.UpCalls[0].Profiles) != 0 {
		t.Fatalf("shared gateway should not receive project profiles: %+v", composer.UpCalls[0].Profiles)
	}
	if !slices.Contains(composer.UpCalls[1].Profiles, "debug") {
		t.Fatalf("project compose up profiles=%+v want debug", composer.UpCalls[1].Profiles)
	}
}

func TestOrchestrator_UpGeneratesDevTLSAndMountsCerts(t *testing.T) {
	cfg := newCfg(t)
	cfg.SiteSuffix = "dev"
	cfg.Hostname = "demo.dev"
	cfg.SharedGateway.HTTPSPort = 443
	tlsProvider := &tlsProviderStub{}
	dc := mocks.NewDocker()
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa, TLS: tlsProvider,
	})

	if err := orch.Up(context.Background(), cfg); err != nil {
		t.Fatalf("Up: %v", err)
	}
	if tlsProvider.calls == 0 {
		t.Fatal("expected TLS provider to be called")
	}
	certsDir := filepath.Join(cfg.StateDir, "shared", "certs")
	if tlsProvider.certFile != filepath.Join(certsDir, "tls.pem") {
		t.Fatalf("cert file=%q want %q", tlsProvider.certFile, filepath.Join(certsDir, "tls.pem"))
	}
	if tlsProvider.keyFile != filepath.Join(certsDir, "tls-key.pem") {
		t.Fatalf("key file=%q want %q", tlsProvider.keyFile, filepath.Join(certsDir, "tls-key.pem"))
	}
	if !slices.Contains(tlsProvider.hosts, "demo.dev") {
		t.Fatalf("TLS hosts=%+v want demo.dev", tlsProvider.hosts)
	}
	if len(composer.UpCalls) == 0 || !slices.Contains(composer.UpCalls[0].Env, "SHARED_GATEWAY_CERTS_DIR="+certsDir) {
		t.Fatalf("shared gateway env missing certs dir %q: %+v", certsDir, composer.UpCalls)
	}
	if !gw.LastInput.TLSEnabled {
		t.Fatalf("gateway config should be rendered with TLS enabled: %+v", gw.LastInput)
	}
	if gw.LastInput.HTTPSPort != 8443 {
		t.Fatalf("gateway HTTPS port=%d want 8443", gw.LastInput.HTTPSPort)
	}
}

func TestOrchestrator_UpRollbackOnPostUpHookFailure(t *testing.T) {
	cfg := newCfg(t)
	cfg.PostUpCommand = "php artisan migrate --force --no-interaction"
	dc := mocks.NewDocker()
	dc.ExecErr = errors.New("hook failed")
	dc.Containers = []docker.Container{
		{
			ID: "nginx-1", Name: "stage-demo-nginx", Service: "nginx", Status: "running",
			Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "nginx"},
		},
		{
			ID: "apache-1", Name: "stage-demo-apache", Service: "apache", Status: "running",
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

func TestOrchestrator_UpRollbackOnGatewayReloadFailureRemovesRoute(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	composer := mocks.NewComposer()
	composer.UpErr = errors.New("gateway reload failed")
	composer.UpErrOnCall = 3
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected gateway reload failure")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok || se.Step != "gateway-reload" {
		t.Fatalf("expected gateway-reload StepError, got %+v", err)
	}
	if len(composer.DownCalls) == 0 {
		t.Fatal("rollback did not invoke compose down")
	}
	for _, route := range gw.Routes {
		if route.Slug == cfg.Slug {
			t.Fatalf("rolled-back project still has gateway route: %+v", gw.Routes)
		}
	}
	if _, err := st.Load(cfg.Slug); err == nil {
		t.Fatal("state should NOT be saved on gateway reload failure")
	}
}

func TestOrchestrator_UpRollbackOnSaveFailureRemovesRoute(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	st.SaveErr = errors.New("save failed")
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected save failure")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok || se.Step != "save-state" {
		t.Fatalf("expected save-state StepError, got %+v", err)
	}
	if len(composer.DownCalls) == 0 {
		t.Fatal("rollback did not invoke compose down")
	}
	for _, route := range gw.Routes {
		if route.Slug == cfg.Slug {
			t.Fatalf("rolled-back project still has gateway route: %+v", gw.Routes)
		}
	}
	if _, err := st.Load(cfg.Slug); err == nil {
		t.Fatal("state should NOT be saved after save failure")
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

func TestOrchestrator_UpSharedGatewayFailureIncludesSharedComposeFile(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	composer := mocks.NewComposer()
	composer.UpErr = errors.New("shared gateway unavailable")
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{MySQLPort: 3306, PMAPort: 8081})

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
	})

	err := orch.Up(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected shared gateway failure")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok {
		t.Fatalf("error not StepError: %v", err)
	}
	if se.Step != "shared-gateway" {
		t.Fatalf("step=%q want shared-gateway", se.Step)
	}
	if !strings.Contains(se.Remedy, cfg.SharedFile) {
		t.Fatalf("remedy=%q missing shared compose path %q", se.Remedy, cfg.SharedFile)
	}
	if !strings.Contains(se.Remedy, "STACK_HOME") {
		t.Fatalf("remedy=%q missing STACK_HOME guidance", se.Remedy)
	}
}

func TestOrchestrator_DownMarksProjectStoppedAndRemovesEnvFile(t *testing.T) {
	cfg := newCfg(t)
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	_ = st.Save(state.Record{Project: cfg, AttachmentState: state.StateAttached})
	if err := os.MkdirAll(filepath.Join(cfg.StateDir, "envfiles"), 0o755); err != nil {
		t.Fatalf("mkdir envfiles: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfg.StateDir, "envfiles", cfg.Slug+".env"), []byte("PROJECT_SLUG=demo\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

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
		t.Fatalf("state should be retained as down, got %v", err)
	}
	if rec.AttachmentState != state.StateDown {
		t.Fatalf("attachment state=%s want %s", rec.AttachmentState, state.StateDown)
	}
	if _, err := os.Stat(filepath.Join(cfg.StateDir, "envfiles", cfg.Slug+".env")); !os.IsNotExist(err) {
		t.Fatalf("env file should be removed, got %v", err)
	}
	if len(gw.Routes) != 0 {
		t.Fatalf("gateway routes should be cleared, got %+v", gw.Routes)
	}
}

func TestOrchestrator_AttachFailsOnRegistryReadError(t *testing.T) {
	cfg := newCfg(t)
	st := mocks.NewState()
	if err := st.Save(state.Record{Project: cfg, AttachmentState: state.StateDown}); err != nil {
		t.Fatalf("save state: %v", err)
	}
	st.RegistryErr = errors.New("registry unreadable")
	dc := mocks.NewDocker()
	dc.Containers = []docker.Container{{
		ID: "web-1", Service: "apache", Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName},
	}}
	composer := mocks.NewComposer()
	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: mocks.NewGateway(), State: st, Ports: mocks.NewPorts(ports.Allocation{}),
	})

	err := orch.Attach(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected registry error")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok || se.Step != "registry" {
		t.Fatalf("step error=%+v ok=%v", se, ok)
	}
	if !strings.Contains(se.Remedy, "stage attach") {
		t.Fatalf("remedy=%q", se.Remedy)
	}
	if len(composer.UpCalls) != 0 {
		t.Fatalf("gateway reload should not run after registry failure")
	}
}

func TestOrchestrator_DownAllReportsPartialFailures(t *testing.T) {
	cfg := newCfg(t)
	other := newCfg(t)
	other.Slug = "other"
	other.Name = "other"
	other.ComposeProjectName = "stage-other"
	st := mocks.NewState()
	if err := st.Save(state.Record{Project: cfg, AttachmentState: state.StateAttached}); err != nil {
		t.Fatalf("save state: %v", err)
	}
	if err := st.Save(state.Record{Project: other, AttachmentState: state.StateAttached}); err != nil {
		t.Fatalf("save other state: %v", err)
	}
	composer := mocks.NewComposer()
	composer.DownErr = errors.New("compose refused")
	orch := lifecycle.New(lifecycle.Deps{
		Docker: mocks.NewDocker(), Compose: composer, Gateway: mocks.NewGateway(), State: st, Ports: mocks.NewPorts(ports.Allocation{}),
	})

	err := orch.DownAll(context.Background(), cfg, false)
	if err == nil {
		t.Fatal("expected partial failure")
	}
	se, ok := lifecycle.AsStepError(err)
	if !ok || se.Step != "down-all" {
		t.Fatalf("step error=%+v ok=%v", se, ok)
	}
	if !strings.Contains(se.Error(), "compose-down") || !strings.Contains(se.Remedy, "stage status --all") {
		t.Fatalf("error/remedy missing detail: %v", se)
	}
	if len(composer.DownCalls) != 2 {
		t.Fatalf("down calls=%d want 2", len(composer.DownCalls))
	}
}

func TestOrchestrator_DownWithoutEnvFileStillRunsComposeDown(t *testing.T) {
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
		t.Fatalf("Down without env file: %v", err)
	}
	if len(composer.DownCalls) != 1 {
		t.Fatalf("down calls=%d want 1", len(composer.DownCalls))
	}
	if composer.DownCalls[0].EnvFile != "" {
		t.Fatalf("expected compose down without --env-file, got envfile=%q", composer.DownCalls[0].EnvFile)
	}
}

func TestOrchestrator_DetachRemovesStateAndReloadsGateway(t *testing.T) {
	cfg := newCfg(t)
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	_ = st.Save(state.Record{Project: cfg, AttachmentState: state.StateAttached})
	if err := os.MkdirAll(filepath.Join(cfg.StateDir, "envfiles"), 0o755); err != nil {
		t.Fatalf("mkdir envfiles: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cfg.StateDir, "envfiles", cfg.Slug+".env"), []byte("PROJECT_SLUG=demo\n"), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

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
	if _, err := os.Stat(filepath.Join(cfg.StateDir, "envfiles", cfg.Slug+".env")); !os.IsNotExist(err) {
		t.Fatalf("env file should be removed, got %v", err)
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
	other.ComposeProjectName = "stage-beta"
	other.WebNetworkAlias = "stage-beta-web"

	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	_ = st.Save(state.Record{Project: cfg, AttachmentState: state.StateAttached})
	_ = st.Save(state.Record{Project: other, AttachmentState: state.StateAttached})
	if err := os.MkdirAll(filepath.Join(cfg.StateDir, "envfiles"), 0o755); err != nil {
		t.Fatalf("mkdir envfiles: %v", err)
	}
	for _, project := range []config.ProjectConfig{cfg, other} {
		if err := os.WriteFile(filepath.Join(project.StateDir, "envfiles", project.Slug+".env"), []byte("PROJECT_SLUG="+project.Slug+"\n"), 0o644); err != nil {
			t.Fatalf("write env file for %s: %v", project.Slug, err)
		}
	}

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
	if rec, err := st.Load(cfg.Slug); err != nil || rec.AttachmentState != state.StateDown {
		t.Fatalf("demo state should be retained as down, got rec=%+v err=%v", rec, err)
	}
	if rec, err := st.Load(other.Slug); err != nil || rec.AttachmentState != state.StateDown {
		t.Fatalf("beta state should be retained as down, got rec=%+v err=%v", rec, err)
	}
	for _, project := range []config.ProjectConfig{cfg, other} {
		if _, err := os.Stat(filepath.Join(project.StateDir, "envfiles", project.Slug+".env")); !os.IsNotExist(err) {
			t.Fatalf("env file for %s should be removed, got %v", project.Slug, err)
		}
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
	other.ComposeProjectName = "stage-beta"
	other.WebNetworkAlias = "stage-beta-web"

	dc := mocks.NewDocker()
	dc.ExecErr = errors.New("hook failed")
	dc.Containers = []docker.Container{
		{
			ID: "apache-1", Name: "stage-demo-apache", Service: "apache", Status: "running",
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

// TestOrchestrator_AttachBootstrapsWhenRuntimeIsNotRunning covers the
// documented attach-or-bootstrap path: when no project containers are running,
// attach falls back to the full Up flow.
func TestOrchestrator_AttachBootstrapsWhenRuntimeIsNotRunning(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	if err := st.Save(state.Record{Project: cfg, AttachmentState: state.StateDown}); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
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
	if foundAlias != "stage-demo-web" {
		t.Fatalf("attach route alias=%q want stage-demo-web", foundAlias)
	}
	if len(composer.UpCalls) < 3 {
		t.Fatalf("attach should bootstrap via Up, got compose up calls %+v", composer.UpCalls)
	}
	if composer.UpCalls[1].ProjectName != cfg.ComposeProjectName {
		t.Fatalf("project compose up was not invoked during attach bootstrap: %+v", composer.UpCalls)
	}
	if len(composer.DownCalls) != 0 {
		t.Fatalf("attach bootstrap should not roll back on success: %+v", composer.DownCalls)
	}
	if len(dc.WaitCalled) != 1 || dc.WaitCalled[0] != cfg.ComposeProjectName {
		t.Fatalf("attach bootstrap did not wait for project health: %+v", dc.WaitCalled)
	}
	if !dc.Networks[cfg.SharedGateway.Network] {
		t.Fatalf("attach bootstrap did not ensure shared network %q", cfg.SharedGateway.Network)
	}
	if len(composer.UpCalls) < 3 || !composer.UpCalls[2].ForceRecreate {
		t.Fatalf("attach bootstrap did not reload the gateway: %+v", composer.UpCalls)
	}
	if len(gw.Routes) == 0 {
		t.Fatalf("attach bootstrap did not populate gateway routes")
	}
	if len(composer.UpCalls) > 0 {
		got := composer.UpCalls[0].Env
		if !slices.Contains(got, "SHARED_GATEWAY_CONFIG_FILE="+cfg.SharedGateway.ConfigFile) {
			t.Fatalf("shared gateway config env missing from attach bootstrap: %+v", got)
		}
	}
	if rec.AttachmentState != state.StateAttached {
		t.Fatalf("attachment state=%s want attached", rec.AttachmentState)
	}
	if len(gw.Routes) == 0 {
		t.Fatalf("attach bootstrap did not add a gateway route")
	}
	if foundAlias != "stage-demo-web" {
		t.Fatalf("attach route alias=%q want stage-demo-web", foundAlias)
	}
}

// TestOrchestrator_AttachAddsRouteAndMarksAttached keeps the route-only attach
// behavior for already-running project containers.
func TestOrchestrator_AttachAddsRouteAndMarksAttached(t *testing.T) {
	cfg := newCfg(t)
	dc := mocks.NewDocker()
	dc.Containers = []docker.Container{{
		ID: "nginx-1", Name: "stage-demo-nginx", Service: "nginx", Status: "running",
		Labels: map[string]string{"com.docker.compose.project": cfg.ComposeProjectName, "com.docker.compose.service": "nginx"},
	}}
	composer := mocks.NewComposer()
	gw := mocks.NewGateway()
	st := mocks.NewState()
	pa := mocks.NewPorts(ports.Allocation{})
	if err := st.Save(state.Record{Project: cfg, AttachmentState: state.StateDown}); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	orch := lifecycle.New(lifecycle.Deps{
		Docker: dc, Compose: composer, Gateway: gw, State: st, Ports: pa,
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
	if foundAlias != "stage-demo-web" {
		t.Fatalf("attach route alias=%q want stage-demo-web", foundAlias)
	}
	if len(composer.UpCalls) != 1 || !composer.UpCalls[0].ForceRecreate {
		t.Fatalf("attach did not request gateway reload with force recreate: %+v", composer.UpCalls)
	}
}
