package guidance

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/state"
)

// CollectOptions contains optional seams for checks that are too expensive for
// the first bare-stage render unless a caller explicitly supplies them.
type CollectOptions struct {
	Capability       TUICapability
	MachineReadiness func(context.Context, config.ProjectConfig) MachineReadinessSummary
	RuntimeStatus    func(context.Context, config.ProjectConfig, *state.Record) RuntimeSummary
}

// Collect builds a no-mutation context from an already-resolved config.
func Collect(ctx context.Context, cfg config.ProjectConfig, opts CollectOptions) GuidedContext {
	projectEnvPath := filepath.Join(cfg.Dir, ".env.stageserve")
	context := GuidedContext{
		CWD:             cfg.Dir,
		ProjectRoot:     cfg.Dir,
		ProjectEnvPath:  projectEnvPath,
		ProjectEnvValid: true,
		StackHome:       cfg.StackHome,
		StackID:         cfg.StackKind,
		StateDir:        cfg.StateDir,
		Hostname:        cfg.Hostname,
		LocalURL:        localURL(cfg),
		WebFolder:       webFolder(cfg),
		SiteName:        cfg.Name,
		Capability:      opts.Capability,
	}

	if info, err := os.Stat(cfg.Dir); err != nil || !info.IsDir() {
		context.NotProject = true
		if err != nil {
			context.Warnings = append(context.Warnings, fmt.Sprintf("Project folder is not available: %v", err))
		}
		return context
	}

	if _, err := os.Stat(projectEnvPath); err == nil {
		context.ProjectEnvExists = true
	} else if !os.IsNotExist(err) {
		context.ProjectEnvValid = false
		context.Warnings = append(context.Warnings, fmt.Sprintf("Project settings could not be checked: %v", err))
	}

	record, err := readRecord(cfg.StateDir, cfg.Slug)
	if err == nil {
		context.ProjectState = &record
		if record.Project.Dir != "" && record.Project.Dir != cfg.Dir {
			context.Warnings = append(context.Warnings, "Project settings do not match the recorded project path.")
		}
	} else if !errors.Is(err, state.ErrNotFound) {
		context.Warnings = append(context.Warnings, fmt.Sprintf("Recorded project state could not be read: %v", err))
	}

	if opts.MachineReadiness != nil {
		context.MachineReadiness = opts.MachineReadiness(ctx, cfg)
	}
	if opts.RuntimeStatus != nil {
		context.Runtime = opts.RuntimeStatus(ctx, cfg, context.ProjectState)
	}
	return context
}

// ContextFromError preserves a useful text fallback when config collection
// fails before a ProjectConfig can be built.
func ContextFromError(cwd string, capability TUICapability, err error) GuidedContext {
	return GuidedContext{
		CWD:        cwd,
		Capability: capability,
		Err:        err,
		Warnings:   []string{err.Error()},
	}
}

func readRecord(stateDir, slug string) (state.Record, error) {
	if strings.TrimSpace(stateDir) == "" || strings.TrimSpace(slug) == "" {
		return state.Record{}, state.ErrNotFound
	}
	path := filepath.Join(stateDir, "projects", slug+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state.Record{}, state.ErrNotFound
		}
		return state.Record{}, err
	}
	var record state.Record
	if err := json.Unmarshal(data, &record); err != nil {
		return state.Record{}, err
	}
	return record, nil
}

func localURL(cfg config.ProjectConfig) string {
	scheme := "http"
	port := ""
	if cfg.SiteSuffix == "dev" {
		scheme = "https"
		port = ":8443"
		if cfg.SharedGateway.HTTPSPort != 0 && cfg.SharedGateway.HTTPSPort != 443 {
			port = fmt.Sprintf(":%d", cfg.SharedGateway.HTTPSPort)
		}
	}
	return fmt.Sprintf("%s://%s%s", scheme, cfg.Hostname, port)
}

func webFolder(cfg config.ProjectConfig) string {
	if cfg.DocRootRelative == "" {
		return "."
	}
	return cfg.DocRootRelative
}
