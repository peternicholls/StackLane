package onboarding

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InitAction describes what WriteProjectEnv did.
type InitAction string

const (
	InitActionCreated     InitAction = "created"
	InitActionOverwritten InitAction = "overwritten"
	InitActionSkipped     InitAction = "skipped"
)

const projectEnvFile = ".env.stageserve"

// ProjectEnvPreview is the no-mutation plan for a project-local StageServe
// settings file.
type ProjectEnvPreview struct {
	Path        string
	ProjectRoot string
	SiteName    string
	DocRoot     string
	SiteSuffix  string
	Body        string
	Exists      bool
}

// ProjectEnvSettings contains the user-owned values StageServe writes to the
// project-local settings file.
type ProjectEnvSettings struct {
	SiteName   string
	DocRoot    string
	SiteSuffix string
}

// ValidateProjectRoot verifies that root is a non-empty, existing directory
// and returns its clean absolute path.
func ValidateProjectRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("project root must not be empty")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("project root %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("project root %q is not a directory", abs)
	}
	return abs, nil
}

// ValidateDocroot checks that docroot is a subdirectory of projectRoot.
// Note: Existence is not validated; use other mechanisms if creation or existence checking is needed.
func ValidateDocroot(projectRoot, docroot string) error {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root: %w", err)
	}
	if !filepath.IsAbs(docroot) {
		docroot = filepath.Join(absRoot, docroot)
	}
	absDoc, err := filepath.Abs(docroot)
	if err != nil {
		return fmt.Errorf("resolve docroot: %w", err)
	}
	// Ensure absDoc starts with absRoot + separator.
	rel, err := filepath.Rel(absRoot, absDoc)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("docroot %q must be inside project root %q", absDoc, absRoot)
	}
	return nil
}

// PreviewProjectEnv validates the target and returns the exact file content
// WriteProjectEnv would use, without writing anything.
func PreviewProjectEnv(projectRoot, siteName, docroot string) (ProjectEnvPreview, error) {
	return PreviewProjectEnvWithSettings(projectRoot, ProjectEnvSettings{SiteName: siteName, DocRoot: docroot})
}

// PreviewProjectEnvWithSettings validates the target and returns the exact file
// content WriteProjectEnvWithSettings would use, without writing anything.
func PreviewProjectEnvWithSettings(projectRoot string, settings ProjectEnvSettings) (ProjectEnvPreview, error) {
	root, err := ValidateProjectRoot(projectRoot)
	if err != nil {
		return ProjectEnvPreview{}, err
	}
	if settings.DocRoot != "" {
		if err := ValidateDocroot(root, settings.DocRoot); err != nil {
			return ProjectEnvPreview{}, err
		}
	}
	path := filepath.Join(root, projectEnvFile)
	_, statErr := os.Stat(path)
	if statErr != nil && !os.IsNotExist(statErr) {
		return ProjectEnvPreview{}, fmt.Errorf("check %s: %w", projectEnvFile, statErr)
	}
	return ProjectEnvPreview{
		Path:        path,
		ProjectRoot: root,
		SiteName:    settings.SiteName,
		DocRoot:     settings.DocRoot,
		SiteSuffix:  settings.SiteSuffix,
		Body:        renderEnv(settings),
		Exists:      statErr == nil,
	}, nil
}

// WriteProjectEnv writes a starter .env.stageserve in projectRoot.
// If the file already exists and force is false, it returns InitActionSkipped.
// Returns the action taken or an error.
func WriteProjectEnv(projectRoot, siteName, docroot string, force bool) (InitAction, error) {
	return WriteProjectEnvWithSettings(projectRoot, ProjectEnvSettings{SiteName: siteName, DocRoot: docroot}, force)
}

// WriteProjectEnvWithSettings writes a starter .env.stageserve in projectRoot
// using the supplied project-local settings.
func WriteProjectEnvWithSettings(projectRoot string, settings ProjectEnvSettings, force bool) (InitAction, error) {
	root, err := ValidateProjectRoot(projectRoot)
	if err != nil {
		return "", err
	}
	if settings.DocRoot != "" {
		if err := ValidateDocroot(root, settings.DocRoot); err != nil {
			return "", err
		}
	}
	path := filepath.Join(root, projectEnvFile)

	_, err = os.Stat(path)
	exists := err == nil
	if exists && !force {
		return InitActionSkipped, nil
	}

	body := renderEnv(settings)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return "", fmt.Errorf("write %s: %w", projectEnvFile, err)
	}
	if exists {
		return InitActionOverwritten, nil
	}
	return InitActionCreated, nil
}

func renderEnv(settings ProjectEnvSettings) string {
	var b strings.Builder
	b.WriteString("# StageServe project config — created by `stage init`\n")
	b.WriteString("# Keep project-specific overrides here.\n\n")
	b.WriteString("STAGESERVE_STACK=20i\n\n")
	if settings.SiteName != "" {
		b.WriteString("SITE_NAME=")
		b.WriteString(shellDoubleQuote(settings.SiteName))
		b.WriteString("\n")
	}
	if settings.SiteSuffix != "" {
		b.WriteString("SITE_SUFFIX=")
		b.WriteString(shellDoubleQuote(settings.SiteSuffix))
		b.WriteString("\n")
	}
	if settings.DocRoot != "" {
		b.WriteString("DOCROOT=")
		b.WriteString(shellDoubleQuote(settings.DocRoot))
		b.WriteString("\n")
	}
	return b.String()
}

func shellDoubleQuote(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`$`, `\$`,
		"`", "\\`",
	)
	return `"` + replacer.Replace(value) + `"`
}
