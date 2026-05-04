package lifecycle

import (
	"errors"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/config"
)

type noopListener struct{}

func (noopListener) Accept() (net.Conn, error) { return nil, errors.New("not implemented") }
func (noopListener) Close() error              { return nil }
func (noopListener) Addr() net.Addr            { return nil }

func TestResolveSharedGatewayPortsDevDefaultsTo8443(t *testing.T) {
	cfg := config.ProjectConfig{
		SiteSuffix: "dev",
		SharedGateway: config.SharedGateway{
			HTTPSPort: 443,
		},
	}

	got := resolveSharedGatewayPorts(cfg)
	if got.SharedGateway.HTTPSPort != 8443 {
		t.Fatalf("https port=%d want 8443", got.SharedGateway.HTTPSPort)
	}
}

func TestResolveSharedGatewayPortsFallsBackWhen443Busy(t *testing.T) {
	originalListen := sharedGatewayListen
	sharedGatewayListen = func(network, address string) (net.Listener, error) {
		switch address {
		case "127.0.0.1:443", "127.0.0.1:8443":
			return nil, errors.New("busy")
		default:
			return noopListener{}, nil
		}
	}
	defer func() { sharedGatewayListen = originalListen }()

	cfg := config.ProjectConfig{
		SiteSuffix: "test",
		SharedGateway: config.SharedGateway{
			HTTPSPort: 443,
		},
	}

	got := resolveSharedGatewayPorts(cfg)
	if got.SharedGateway.HTTPSPort != 8444 {
		t.Fatalf("https port=%d want 8444", got.SharedGateway.HTTPSPort)
	}
}

func TestWriteEnvFileIncludesAppDBSettings(t *testing.T) {
	stateDir := t.TempDir()
	cfg := config.ProjectConfig{
		Slug:               "demo",
		Name:               "demo",
		Dir:                "/tmp/demo",
		Hostname:           "demo.test",
		DocRoot:            "/tmp/demo/public_html",
		ContainerSiteRoot:  "/home/sites/demo",
		ContainerDocRoot:   "/home/sites/demo/public_html",
		ComposeProjectName: "stage-demo",
		PHPVersion:         "8.5",
		StateDir:           stateDir,
		RuntimeNetwork:     "stage-demo-runtime",
		DatabaseVolume:     "stage-demo-db-data",
		WebNetworkAlias:    "stage-demo-web",
		MySQL: config.MySQL{
			Version:      "10.6",
			Database:     "appdb",
			User:         "appuser",
			Password:     "apppass",
			RootPassword: "root",
			Port:         3307,
			PMAPort:      8082,
		},
		SharedGateway: config.SharedGateway{Network: "stage-shared"},
	}

	path, err := writeEnvFile(cfg)
	if err != nil {
		t.Fatalf("writeEnvFile: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env file: %v", err)
	}
	content := string(body)
	for _, want := range []string{
		"DB_HOST=mariadb",
		"DB_PORT=3306",
		"DB_DATABASE=appdb",
		"DB_USERNAME=appuser",
		"DB_PASSWORD=apppass",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("env file missing %q in:\n%s", want, content)
		}
	}
}
