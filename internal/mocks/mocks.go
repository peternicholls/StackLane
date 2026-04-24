// Hand-rolled mocks used by the unit-test suite. Each interface in the
// project gets a corresponding mock so `go test -short ./...` runs without a
// running Docker daemon.
package mocks

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/peternicholls/stacklane/core/state"
	"github.com/peternicholls/stacklane/infra/compose"
	"github.com/peternicholls/stacklane/infra/docker"
	"github.com/peternicholls/stacklane/infra/gateway"
	"github.com/peternicholls/stacklane/platform/ports"
)

// --- DockerClient ---

type Docker struct {
	mu           sync.Mutex
	Networks     map[string]bool
	NetworkErr   error
	Containers   []docker.Container
	WaitErr      error
	WaitCalled   []string
	LogsReader   io.ReadCloser
	LogsErr      error
	ListErr      error
	AvailableErr error
}

func NewDocker() *Docker { return &Docker{Networks: map[string]bool{}} }

func (m *Docker) Available(ctx context.Context) error { return m.AvailableErr }

func (m *Docker) NetworkExists(ctx context.Context, name string) (bool, error) {
	if m.NetworkErr != nil {
		return false, m.NetworkErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Networks[name], nil
}

func (m *Docker) CreateNetwork(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Networks[name] = true
	return nil
}

func (m *Docker) RemoveNetwork(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Networks, name)
	return nil
}

func (m *Docker) ListContainersByLabel(ctx context.Context, labels map[string]string) ([]docker.Container, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	out := make([]docker.Container, 0, len(m.Containers))
	for _, c := range m.Containers {
		match := true
		for k, v := range labels {
			got, ok := c.Labels[k]
			if !ok || (v != "" && got != v) {
				match = false
				break
			}
		}
		if match {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *Docker) WaitHealthy(ctx context.Context, project string, timeout time.Duration) error {
	m.mu.Lock()
	m.WaitCalled = append(m.WaitCalled, project)
	m.mu.Unlock()
	return m.WaitErr
}

func (m *Docker) ContainerLogs(ctx context.Context, id string, follow bool) (docker.LogStream, error) {
	if m.LogsErr != nil {
		return nil, m.LogsErr
	}
	return m.LogsReader, nil
}

// --- Composer ---

type Composer struct {
	mu        sync.Mutex
	UpCalls   []compose.UpOptions
	DownCalls []compose.DownOptions
	UpErr     error
	DownErr   error
	LogsErr   error
	ExecErr   error
}

func NewComposer() *Composer { return &Composer{} }

func (m *Composer) Up(ctx context.Context, opts compose.UpOptions) error {
	m.mu.Lock()
	m.UpCalls = append(m.UpCalls, opts)
	m.mu.Unlock()
	return m.UpErr
}

func (m *Composer) Down(ctx context.Context, opts compose.DownOptions) error {
	m.mu.Lock()
	m.DownCalls = append(m.DownCalls, opts)
	m.mu.Unlock()
	return m.DownErr
}

func (m *Composer) Logs(ctx context.Context, opts compose.LogsOptions) error { return m.LogsErr }

func (m *Composer) Exec(ctx context.Context, opts compose.ExecOptions) error { return m.ExecErr }

// --- GatewayManager ---

type Gateway struct {
	mu       sync.Mutex
	Routes   []gateway.Route
	Probe    string
	Host     string
	WriteErr error
}

func NewGateway() *Gateway { return &Gateway{Probe: "stacklane-no-route", Host: "localhost"} }

func (m *Gateway) ConfigPath() string { return "/tmp/mock-gateway.conf" }

func (m *Gateway) WriteConfig(input gateway.RenderInput) (string, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Routes = append([]gateway.Route(nil), input.Routes...)
	if m.WriteErr != nil {
		return "", "", m.WriteErr
	}
	return m.Probe, m.Host, nil
}

func (m *Gateway) AddRoute(r gateway.Route, current []gateway.Route) (string, string, error) {
	merged := append([]gateway.Route(nil), current...)
	merged = append(merged, r)
	return m.WriteConfig(gateway.RenderInput{Routes: merged, PreferredSlug: r.Slug})
}

func (m *Gateway) RemoveRoute(slug string, current []gateway.Route) (string, string, error) {
	merged := []gateway.Route{}
	for _, r := range current {
		if r.Slug != slug {
			merged = append(merged, r)
		}
	}
	return m.WriteConfig(gateway.RenderInput{Routes: merged})
}

// --- StateStore ---

type State struct {
	mu          sync.Mutex
	Records     map[string]state.Record
	StateDirVal string
}

func NewState() *State {
	return &State{Records: map[string]state.Record{}, StateDirVal: "/tmp/stacklane-test-state"}
}

func (m *State) Save(rec state.Record) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Records[rec.Project.Slug] = rec
	return nil
}

func (m *State) Load(slug string) (state.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.Records[slug]
	if !ok {
		return state.Record{}, state.ErrNotFound
	}
	return rec, nil
}

func (m *State) Remove(slug string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Records, slug)
	return nil
}

func (m *State) Registry() ([]state.RegistryRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := []state.RegistryRow{}
	for _, rec := range m.Records {
		rows = append(rows, state.RegistryRow{
			Slug:            rec.Project.Slug,
			AttachmentState: rec.AttachmentState,
			Name:            rec.Project.Name,
			Hostname:        rec.Project.Hostname,
			ComposeProject:  rec.Project.ComposeProjectName,
			RuntimeNetwork:  rec.Project.RuntimeNetwork,
			DatabaseVolume:  rec.Project.DatabaseVolume,
			WebNetworkAlias: rec.Project.WebNetworkAlias,
			MySQLPort:       rec.Project.MySQL.Port,
			PMAPort:         rec.Project.MySQL.PMAPort,
		})
	}
	return rows, nil
}

func (m *State) StateFileForSelector(selector string) (state.Record, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for slug, rec := range m.Records {
		p := rec.Project
		if selector == slug || selector == p.Name || selector == p.Hostname || selector == p.Dir {
			return rec, "", nil
		}
	}
	return state.Record{}, "", state.ErrNotFound
}

func (m *State) StateDir() string { return m.StateDirVal }

// --- PortAllocator ---

type Ports struct {
	Out ports.Allocation
	Err error
}

func NewPorts(a ports.Allocation) *Ports { return &Ports{Out: a} }

func (m *Ports) Allocate(req ports.Request, registry []state.RegistryRow) (ports.Allocation, error) {
	if m.Err != nil {
		return ports.Allocation{}, m.Err
	}
	out := m.Out
	if out.MySQLPort == 0 && req.MySQLPort != 0 {
		out.MySQLPort = req.MySQLPort
	}
	if out.PMAPort == 0 && req.PMAPort != 0 {
		out.PMAPort = req.PMAPort
	}
	return out, nil
}

// errStub keeps the linter happy if we ever extend the package without using errors.
var _ = errors.New
