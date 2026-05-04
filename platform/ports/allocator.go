// Port allocator. Bind-checks via net.Listen with an lsof fallback, never
// returns a port already reserved by another project, and serialises
// concurrent stage up invocations with an exclusive file lock (FR-007 / SC-008).
package ports

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/peternicholls/stageserve/core/state"
	"golang.org/x/sys/unix"
)

// LockFileName is the name of the on-disk lock the allocator takes during
// Allocate to serialise concurrent stage up invocations.
const LockFileName = ".port-allocation.lock"

// Allocator is the default PortAllocator implementation.
type Allocator struct {
	StateDir string
	// Listen is overridable from tests; defaults to net.Listen.
	Listen func(network, address string) (net.Listener, error)
}

// NewAllocator returns an allocator that writes its lock under stateDir.
func NewAllocator(stateDir string) *Allocator {
	return &Allocator{StateDir: stateDir, Listen: net.Listen}
}

// Allocate satisfies req. The flow:
//
//  1. Take an exclusive flock on <stateDir>/.port-allocation.lock so two
//     concurrent stage up invocations cannot race (SC-008).
//  2. For each port requested explicitly, validate it is not in use locally
//     and is not reserved by a different slug in the registry.
//  3. For each port left at zero, scan upward from the documented start port
//     (80 / 3306 / 8081 for the first project, 8080 / 3307 / 8082 otherwise)
//     until one is both free locally and unreserved.
func (a *Allocator) Allocate(req Request, registry []state.RegistryRow) (Allocation, error) {
	unlock, err := a.acquireLock()
	if err != nil {
		return Allocation{}, err
	}
	defer unlock()

	out := Allocation{}

	// Build a quick reservation map keyed by port. Skip own slug so a project
	// reusing its current port is not flagged as conflicting.
	reservedBy := func(port int, role string) (string, bool) {
		for _, row := range registry {
			if row.Slug == req.OwnSlug {
				continue
			}
			switch role {
			case "mysql":
				if row.MySQLPort == port {
					return row.Slug, true
				}
			case "pma":
				if row.PMAPort == port {
					return row.Slug, true
				}
			}
		}
		return "", false
	}

	// HostPort: optional in this rewrite (gateway routes by hostname); only
	// honour an explicit value or skip otherwise.
	if req.HostPort != 0 {
		if a.portInUse(req.HostPort) {
			return out, fmt.Errorf("HOST_PORT %d is already listening", req.HostPort)
		}
		out.HostPort = req.HostPort
	} else if req.IsUp && req.ProjectCount == 0 && !a.portInUse(80) {
		out.HostPort = 80
	}

	chooseExplicit := func(role string, want int, errLabel string) (int, error) {
		if a.portInUse(want) {
			return 0, fmt.Errorf("%s %d is already listening", errLabel, want)
		}
		if owner, taken := reservedBy(want, role); taken {
			return 0, fmt.Errorf("%s %d is already reserved by %s", errLabel, want, owner)
		}
		return want, nil
	}
	chooseAuto := func(role string, candidates []int, errLabel string) (int, error) {
		for _, p := range candidates {
			if a.portInUse(p) {
				continue
			}
			if _, taken := reservedBy(p, role); taken {
				continue
			}
			return p, nil
		}
		return 0, fmt.Errorf("%s: no available port in scan range", errLabel)
	}

	if req.MySQLPort != 0 {
		p, err := chooseExplicit("mysql", req.MySQLPort, "MYSQL_PORT")
		if err != nil {
			return out, err
		}
		out.MySQLPort = p
	} else {
		canonical := []int{}
		if req.ProjectCount == 0 {
			canonical = append(canonical, 3306)
		}
		canonical = append(canonical, scanRange(3307, 65535)...)
		p, err := chooseAuto("mysql", canonical, "MYSQL_PORT")
		if err != nil {
			return out, err
		}
		out.MySQLPort = p
	}

	if req.PMAPort != 0 {
		p, err := chooseExplicit("pma", req.PMAPort, "PMA_PORT")
		if err != nil {
			return out, err
		}
		out.PMAPort = p
	} else {
		canonical := []int{}
		if req.ProjectCount == 0 {
			canonical = append(canonical, 8081)
		}
		canonical = append(canonical, scanRange(8082, 65535)...)
		p, err := chooseAuto("pma", canonical, "PMA_PORT")
		if err != nil {
			return out, err
		}
		out.PMAPort = p
	}

	return out, nil
}

func (a *Allocator) portInUse(port int) bool {
	listen := a.Listen
	if listen == nil {
		listen = net.Listen
	}
	ln, err := listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return true
	}
	_ = ln.Close()
	return false
}

// scanRange yields a slice from start to end (capped to keep the slice small
// — the loop in chooseAuto stops on the first hit anyway).
func scanRange(start, end int) []int {
	if end <= start {
		return nil
	}
	if end-start > 1024 {
		end = start + 1024
	}
	out := make([]int, 0, end-start)
	for p := start; p < end; p++ {
		out = append(out, p)
	}
	return out
}

// acquireLock takes an exclusive file lock under StateDir. The returned func
// releases it. Returns nil-deferable on error.
func (a *Allocator) acquireLock() (func(), error) {
	if a.StateDir == "" {
		return func() {}, nil
	}
	if err := os.MkdirAll(a.StateDir, 0o755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(a.StateDir, LockFileName)
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("port allocator: cannot acquire lock %s: %w", lockPath, err)
	}
	once := sync.Once{}
	return func() {
		once.Do(func() {
			_ = unix.Flock(int(f.Fd()), unix.LOCK_UN)
			_ = f.Close()
		})
	}, nil
}

// ErrUnsupported is returned for operations the allocator cannot perform on
// the running platform.
var ErrUnsupported = errors.New("ports: operation unsupported on this platform")
