// Lifecycle orchestrator tests using the in-memory mocks. The full Up flow
// is exercised end-to-end: ensure-network, allocate, ensure shared gateway,
// compose up, wait healthy, gateway add+reload, save state.
package lifecycle_test

import (
	"context"
	"errors"
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
		ComposeProjectName: "stacklane-demo",
		WebNetworkAlias:    "stacklane-demo-web",
		ContainerSiteRoot:  "/home/sites/demo",
		ContainerDocRoot:   "/home/sites/demo",
		PHPVersion:         "8.5",
		WaitTimeoutSecs:    5,
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
		ID: "c1", Name: "stacklane-demo-nginx", Service: "nginx", Status: "running",
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
