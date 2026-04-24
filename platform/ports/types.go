// Package ports defines the typed PortAllocator contract.
package ports

import (
	"github.com/peternicholls/stacklane/core/state"
)

// Allocation is the resolved set of ports for a single project.
type Allocation struct {
	HostPort  int
	MySQLPort int
	PMAPort   int
}

// Request describes what the caller wants. A zero in any field means "auto-pick".
type Request struct {
	HostPort     int
	MySQLPort    int
	PMAPort      int
	ProjectCount int    // determines whether canonical defaults (80/3306/8081) are tried
	IsUp         bool   // only stacklane up may claim port 80
	OwnSlug      string // permits a project to keep its previously assigned ports
}

// PortAllocator hands out non-conflicting host ports. Allocate must be
// race-safe across concurrent stacklane up invocations on the same machine
// (FR-007 / SC-008): implementations take an exclusive file lock.
type PortAllocator interface {
	// Allocate returns ports that satisfy req over the supplied registry view
	// and the live socket state of the host. Implementations must return an
	// error (rather than silently shifting) when an explicitly requested port
	// is in use by another project.
	Allocate(req Request, registry []state.RegistryRow) (Allocation, error)
}
