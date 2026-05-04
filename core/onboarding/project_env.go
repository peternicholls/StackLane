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

// WriteProjectEnv writes a starter .env.stageserve in projectRoot.
// If the file already exists and force is false, it returns InitActionSkipped.
// Returns the action taken or an error.
func WriteProjectEnv(projectRoot, siteName, docroot string, force bool) (InitAction, error) {
	path := filepath.Join(projectRoot, projectEnvFile)

	_, err := os.Stat(path)
	exists := err == nil
	if exists && !force {
		return InitActionSkipped, nil
	}

	body := renderEnv(siteName, docroot)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return "", fmt.Errorf("write %s: %w", projectEnvFile, err)
	}
	if exists {
		return InitActionOverwritten, nil
	}
	return InitActionCreated, nil
}

func renderEnv(siteName, docroot string) string {
	var b strings.Builder
	b.WriteString("# StageServe project config — created by `stage init`\n")
	b.WriteString("# Keep project-specific overrides here.\n\n")
	b.WriteString("STAGESERVE_STACK=20i\n\n")
	if siteName != "" {
		b.WriteString("SITE_NAME=" + shellDoubleQuote(siteName) + "\n")
	}
	if docroot != "" {
		b.WriteString("DOCROOT=" + shellDoubleQuote(docroot) + "\n")
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
