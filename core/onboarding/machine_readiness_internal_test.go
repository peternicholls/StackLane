package onboarding

import (
	"fmt"
	"net"
	"syscall"
	"testing"
)

type stubListener struct{}

func (stubListener) Accept() (net.Conn, error) { return nil, fmt.Errorf("not implemented") }
func (stubListener) Close() error              { return nil }
func (stubListener) Addr() net.Addr            { return &net.TCPAddr{} }

func TestCheckPort_PermissionDeniedFallsBackToWildcard(t *testing.T) {
	original := portListen
	defer func() { portListen = original }()

	portListen = func(network, address string) (net.Listener, error) {
		switch address {
		case "127.0.0.1:80":
			return nil, syscall.EACCES
		case "0.0.0.0:80":
			return stubListener{}, nil
		default:
			return nil, fmt.Errorf("unexpected address %s", address)
		}
	}

	r := CheckPort("port.80", 80)
	if r.Status != StatusReady {
		t.Fatalf("want ready, got %s (%s)", r.Status, r.Message)
	}
}

func TestCheckPort_PermissionDeniedThenBusyWildcard(t *testing.T) {
	original := portListen
	defer func() { portListen = original }()

	portListen = func(network, address string) (net.Listener, error) {
		switch address {
		case "127.0.0.1:80":
			return nil, syscall.EACCES
		case "0.0.0.0:80":
			return nil, syscall.EADDRINUSE
		default:
			return nil, fmt.Errorf("unexpected address %s", address)
		}
	}

	r := CheckPort("port.80", 80)
	if r.Status != StatusNeedsAction {
		t.Fatalf("want needs_action, got %s (%s)", r.Status, r.Message)
	}
}
