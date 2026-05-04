// Package project contains pure helpers for project identity (slug derivation,
// hostname resolution, document-root canonicalisation).
//
// These helpers are intentionally I/O-free apart from filesystem stat checks on
// the document root.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
var multiDash = regexp.MustCompile(`-{2,}`)

// Slugify lowercases, collapses non-alphanumeric runs into single dashes, trims
// leading/trailing dashes, caps at 63 chars, and falls back to "site" when the
// result is empty.
func Slugify(input string) string {
	v := strings.ToLower(input)
	v = nonAlnum.ReplaceAllString(v, "-")
	v = strings.Trim(v, "-")
	v = multiDash.ReplaceAllString(v, "-")
	if len(v) > 63 {
		v = v[:63]
	}
	v = strings.TrimRight(v, "-")
	if v == "" {
		v = "site"
	}
	return v
}

// AbsDir resolves path to its absolute, symlink-resolved form. If path is
// relative, it is joined with cwd before resolution.
func AbsDir(path string) (string, error) {
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		path = filepath.Join(cwd, path)
	}
	return filepath.EvalSymlinks(path)
}

// AbsPathFromBase joins relative paths under baseDir; absolute paths are
// returned unchanged.
func AbsPathFromBase(baseDir, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(baseDir, path)
}

// ResolveDocRoot reproduces stageserve_resolve_docroot precedence:
//  1. explicit docRoot (CLI / project .env.stageserve / env)
//  2. codeDir (alias)
//  3. <projectDir>/public_html when present
//  4. projectDir itself
//
// It returns the canonical absolute path plus the path relative to projectDir
// ("" when they are the same). Errors when the resolved root does not exist or
// is not under projectDir.
func ResolveDocRoot(projectDir, docRoot, codeDir string) (abs string, rel string, err error) {
	switch {
	case docRoot != "":
		abs = AbsPathFromBase(projectDir, docRoot)
	case codeDir != "":
		abs = AbsPathFromBase(projectDir, codeDir)
	default:
		candidate := filepath.Join(projectDir, "public_html")
		if info, statErr := os.Stat(candidate); statErr == nil && info.IsDir() {
			abs = candidate
		} else {
			abs = projectDir
		}
	}

	info, statErr := os.Stat(abs)
	if statErr != nil || !info.IsDir() {
		return "", "", fmt.Errorf("document root not found: %s", abs)
	}
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return "", "", err
	}

	switch {
	case abs == projectDir:
		rel = ""
	case strings.HasPrefix(abs, projectDir+string(os.PathSeparator)):
		rel = strings.TrimPrefix(abs, projectDir+string(os.PathSeparator))
	default:
		return "", "", fmt.Errorf("document root must live inside the project directory: %s", abs)
	}
	return abs, rel, nil
}

// ResolveHostname reproduces stageserve_resolve_hostname: explicit hostname wins,
// otherwise <slug>.<suffix>. Suffix defaults to "test" and is itself slugified.
func ResolveHostname(slug, siteHostname, siteSuffix string) (hostname, suffix string) {
	suffix = siteSuffix
	if suffix == "" {
		suffix = "test"
	}
	suffix = Slugify(suffix)
	if siteHostname != "" {
		return siteHostname, suffix
	}
	return slug + "." + suffix, suffix
}

var hostnameLabel = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]*[A-Za-z0-9])?$`)

// HostnameValid mirrors stageserve_hostname_valid: requires at least one dot,
// labels of [A-Za-z0-9-], no leading/trailing dash on a label, no consecutive
// dots, no trailing dot.
func HostnameValid(host string) bool {
	if host == "" || !strings.Contains(host, ".") || strings.Contains(host, "..") || strings.HasPrefix(host, ".") || strings.HasSuffix(host, ".") {
		return false
	}
	for _, label := range strings.Split(host, ".") {
		if !hostnameLabel.MatchString(label) {
			return false
		}
	}
	return true
}

var aliasRe = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)

// AliasValid mirrors stageserve_alias_valid (lowercase, digits, internal dashes).
func AliasValid(alias string) bool { return aliasRe.MatchString(alias) }
