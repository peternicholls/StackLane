// Composer: every `docker compose` subprocess invocation lives in this file.
// The lifecycle orchestrator obtains a Composer through the interface; the
// SDK transport is owned separately by infra/docker.
package compose

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// CLI runs `docker compose ...` as a subprocess.
type CLI struct {
	// Stdout, Stderr default to os.Stdout / os.Stderr when nil.
	Stdout io.Writer
	Stderr io.Writer
	// Bin overrides the docker binary lookup (test seam).
	Bin string
}

// NewCLI returns a CLI wired to os.Stdout / os.Stderr and the `docker` binary.
func NewCLI() *CLI { return &CLI{} }

func (c *CLI) bin() string {
	if c.Bin != "" {
		return c.Bin
	}
	return "docker"
}

func (c *CLI) stdout() io.Writer {
	if c.Stdout != nil {
		return c.Stdout
	}
	return os.Stdout
}

func (c *CLI) stderr() io.Writer {
	if c.Stderr != nil {
		return c.Stderr
	}
	return os.Stderr
}

// Up runs `docker compose ... up`, with --wait when WaitTimeout > 0.
func (c *CLI) Up(ctx context.Context, opts UpOptions) error {
	args := c.baseArgs(opts.ComposeFile, opts.ProjectName, opts.EnvFile, opts.Profiles)
	args = append(args, "up")
	if opts.Detach {
		args = append(args, "-d")
	} else {
		args = append(args, "-d") // we always want detach for orchestration
	}
	if opts.NoDeps {
		args = append(args, "--no-deps")
	}
	if opts.ForceRecreate {
		args = append(args, "--force-recreate")
	}
	if opts.WaitTimeout > 0 {
		args = append(args, "--wait", "--wait-timeout", strconv.Itoa(int(opts.WaitTimeout.Seconds())))
	}
	args = append(args, opts.Services...)
	return c.run(ctx, opts.ProjectDir, opts.Env, args)
}

func (c *CLI) Down(ctx context.Context, opts DownOptions) error {
	args := c.baseArgs(opts.ComposeFile, opts.ProjectName, opts.EnvFile, nil)
	args = append(args, "down")
	if opts.RemoveVolumes {
		args = append(args, "-v")
	}
	return c.run(ctx, opts.ProjectDir, opts.Env, args)
}

func (c *CLI) Logs(ctx context.Context, opts LogsOptions) error {
	args := c.baseArgs(opts.ComposeFile, opts.ProjectName, opts.EnvFile, nil)
	args = append(args, "logs")
	if opts.Follow {
		args = append(args, "-f")
	}
	if opts.Service != "" {
		args = append(args, opts.Service)
	}
	return c.run(ctx, opts.ProjectDir, opts.Env, args)
}

func (c *CLI) Exec(ctx context.Context, opts ExecOptions) error {
	args := c.baseArgs(opts.ComposeFile, opts.ProjectName, opts.EnvFile, nil)
	args = append(args, "exec", "-T", opts.Service)
	args = append(args, opts.Cmd...)
	return c.run(ctx, opts.ProjectDir, opts.Env, args)
}

func (c *CLI) baseArgs(composeFile, projectName, envFile string, profiles []string) []string {
	args := []string{"compose"}
	if envFile != "" {
		args = append(args, "--env-file", envFile)
	}
	if composeFile != "" {
		args = append(args, "-f", composeFile)
	}
	if projectName != "" {
		args = append(args, "-p", projectName)
	}
	for _, p := range profiles {
		if p != "" {
			args = append(args, "--profile", p)
		}
	}
	return args
}

func (c *CLI) run(ctx context.Context, dir string, env []string, args []string) error {
	cmd := exec.CommandContext(ctx, c.bin(), args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = c.stdout()
	cmd.Stderr = c.stderr()
	cmd.Stdin = nil
	cmd.Env = mergeEnv(os.Environ(), env)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker %s: %w", strings.Join(args, " "), err)
	}
	return nil
}

func mergeEnv(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	out := make([]string, len(base), len(base)+len(extra))
	copy(out, base)
	out = append(out, extra...)
	return out
}
