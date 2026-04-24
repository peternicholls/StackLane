// Package compose owns every `docker compose` subprocess invocation.
//
// This is intentionally separate from infra/docker so the SDK transport and
// the CLI subprocess surface do not tangle. A future iteration that drops the
// compose CLI in favour of native SDK support touches only this package.
package compose

import (
	"context"
	"time"
)

// Composer runs `docker compose` for a single stack file.
type Composer interface {
	// Up runs `docker compose --wait` (where supported), passing through any
	// active profiles and any extra environment.
	Up(ctx context.Context, opts UpOptions) error
	// Down runs `docker compose down`. Idempotent: succeeds when the project
	// is already gone.
	Down(ctx context.Context, opts DownOptions) error
	// Logs streams `docker compose logs -f`. When service is empty, all
	// services in the project stream.
	Logs(ctx context.Context, opts LogsOptions) error
	// Exec runs `docker compose exec -T` for the named service.
	Exec(ctx context.Context, opts ExecOptions) error
}

// UpOptions describes a single compose-up invocation.
type UpOptions struct {
	ProjectDir   string   // working directory for the subprocess
	ComposeFile  string   // -f
	ProjectName  string   // -p
	EnvFile      string   // --env-file (optional)
	Env          []string // extra environment for the subprocess (KEY=VALUE)
	Profiles     []string // --profile <name> ...
	WaitTimeout  time.Duration
	Detach       bool // --detach (default true)
	NoDeps       bool
	ForceRecreate bool
	Services     []string // optional, restricts the up to specific services
}

// DownOptions describes a compose-down invocation.
type DownOptions struct {
	ProjectDir  string
	ComposeFile string
	ProjectName string
	EnvFile     string
	Env         []string
	RemoveVolumes bool
}

// LogsOptions describes a compose-logs invocation.
type LogsOptions struct {
	ProjectDir  string
	ComposeFile string
	ProjectName string
	EnvFile     string
	Env         []string
	Service     string
	Follow      bool
}

// ExecOptions describes a compose-exec invocation.
type ExecOptions struct {
	ProjectDir  string
	ComposeFile string
	ProjectName string
	EnvFile     string
	Env         []string
	Service     string
	Cmd         []string
}
