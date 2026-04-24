// Allocator unit + concurrency tests. The Listen hook lets us simulate "port
// in use" deterministically without binding real sockets.
package ports

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/peternicholls/stacklane/core/state"
)

func freshAllocator(t *testing.T, busy map[int]bool) *Allocator {
	t.Helper()
	return &Allocator{
		StateDir: t.TempDir(),
		Listen: func(network, address string) (net.Listener, error) {
			_, port, _ := net.SplitHostPort(address)
			if busy != nil {
				if v, ok := atoiSafe(port); ok && busy[v] {
					return nil, errors.New("simulated busy")
				}
			}
			return noopListener{}, nil
		},
	}
}

type noopListener struct{}

func (noopListener) Accept() (net.Conn, error) { return nil, errors.New("not implemented") }
func (noopListener) Close() error              { return nil }
func (noopListener) Addr() net.Addr            { return noopAddr{} }

type noopAddr struct{}

func (noopAddr) Network() string { return "tcp" }
func (noopAddr) String() string  { return "127.0.0.1:0" }

type noopConn struct{}

func (noopConn) Read([]byte) (int, error)         { return 0, errors.New("not implemented") }
func (noopConn) Write([]byte) (int, error)        { return 0, errors.New("not implemented") }
func (noopConn) Close() error                     { return nil }
func (noopConn) LocalAddr() net.Addr              { return noopAddr{} }
func (noopConn) RemoteAddr() net.Addr             { return noopAddr{} }
func (noopConn) SetDeadline(time.Time) error      { return nil }
func (noopConn) SetReadDeadline(time.Time) error  { return nil }
func (noopConn) SetWriteDeadline(time.Time) error { return nil }

func atoiSafe(s string) (int, bool) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}

func TestAllocator_AssignsCanonicalPortsForFirstProject(t *testing.T) {
	a := freshAllocator(t, nil)
	out, err := a.Allocate(Request{IsUp: true, OwnSlug: "first", ProjectCount: 0}, nil)
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if out.MySQLPort != 3306 {
		t.Errorf("first project MySQL port=%d want 3306", out.MySQLPort)
	}
	if out.PMAPort != 8081 {
		t.Errorf("first project PMA port=%d want 8081", out.PMAPort)
	}
}

func TestAllocator_AvoidsRegistryReservations(t *testing.T) {
	a := freshAllocator(t, nil)
	registry := []state.RegistryRow{
		{Slug: "other", MySQLPort: 3306, PMAPort: 8081},
	}
	out, err := a.Allocate(Request{IsUp: true, OwnSlug: "second", ProjectCount: 1}, registry)
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if out.MySQLPort == 3306 {
		t.Errorf("second project must not collide with reserved 3306")
	}
	if out.PMAPort == 8081 {
		t.Errorf("second project must not collide with reserved 8081")
	}
}

func TestAllocator_RejectsExplicitConflict(t *testing.T) {
	a := freshAllocator(t, nil)
	registry := []state.RegistryRow{{Slug: "other", MySQLPort: 33060}}
	_, err := a.Allocate(Request{IsUp: true, OwnSlug: "second", MySQLPort: 33060}, registry)
	if err == nil {
		t.Fatalf("expected error for reserved explicit port")
	}
}

func TestAllocator_ConcurrentLockSerialises(t *testing.T) {
	a := freshAllocator(t, nil)
	var wg sync.WaitGroup
	results := make([]Allocation, 4)
	errs := make([]error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = a.Allocate(Request{IsUp: true, OwnSlug: "any"}, nil)
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: %v", i, err)
		}
	}
}
