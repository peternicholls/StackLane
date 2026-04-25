package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peternicholls/stacklane/core/config"
)

func TestEnsureProjectEnvFileCreatesStarterWhenMissing(t *testing.T) {
	projectDir := t.TempDir()
	cfg := config.ProjectConfig{
		Dir:             projectDir,
		Name:            "demo-site",
		Slug:            "demo-site",
		Hostname:        "demo-site.test",
		SiteSuffix:      "test",
		DocRootRelative: "public_html",
		PHPVersion:      "8.5",
		MySQL: config.MySQL{
			Database: "demo-site",
			User:     "demo-site",
			Password: "devpass",
		},
	}
	flags := &SharedFlags{PHPVersion: "8.4", SiteName: "custom-site"}

	if err := ensureProjectEnvFile(cfg, flags); err != nil {
		t.Fatalf("ensure project env: %v", err)
	}
	info, err := os.Stat(filepath.Join(projectDir, projectEnvFileName))
	if err != nil {
		t.Fatalf("stat created env: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("created env mode=%#o want 0600", info.Mode().Perm())
	}

	body, err := os.ReadFile(filepath.Join(projectDir, projectEnvFileName))
	if err != nil {
		t.Fatalf("read created env: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "Created automatically on first `stacklane up` or `stacklane attach`.") {
		t.Fatalf("starter header missing: %s", text)
	}
	if !strings.Contains(text, "STACKLANE_STACK=20i") {
		t.Fatalf("expected explicit stack kind in starter file: %s", text)
	}
	if !strings.Contains(text, "SITE_NAME=custom-site") {
		t.Fatalf("expected explicit SITE_NAME to be persisted: %s", text)
	}
	if !strings.Contains(text, "PHP_VERSION=8.4") {
		t.Fatalf("expected explicit PHP_VERSION to be persisted: %s", text)
	}
	if !strings.Contains(text, "# MYSQL_DATABASE=demo-site") {
		t.Fatalf("expected starter defaults to be present: %s", text)
	}
}

func TestEnsureProjectEnvFileDoesNotOverwriteExistingFile(t *testing.T) {
	projectDir := t.TempDir()
	path := filepath.Join(projectDir, projectEnvFileName)
	if err := os.WriteFile(path, []byte("SITE_NAME=keep-me\n"), 0o644); err != nil {
		t.Fatalf("seed env: %v", err)
	}

	cfg := config.ProjectConfig{Dir: projectDir, Name: "demo-site", Slug: "demo-site"}
	if err := ensureProjectEnvFile(cfg, &SharedFlags{PHPVersion: "8.4"}); err != nil {
		t.Fatalf("ensure project env: %v", err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read env: %v", err)
	}
	if string(body) != "SITE_NAME=keep-me\n" {
		t.Fatalf("existing env file was overwritten: %q", string(body))
	}
}

func TestRenderEnvValueQuotesWhitespace(t *testing.T) {
	got := renderEnvValue("my site")
	if got != `"my site"` {
		t.Fatalf("renderEnvValue=%q want %q", got, `"my site"`)
	}
}

func TestRenderEnvValueEscapesMixedQuotesForShell(t *testing.T) {
	got := renderEnvValue(`It's "his" site`)
	want := `"It's \"his\" site"`
	if got != want {
		t.Fatalf("renderEnvValue=%q want %q", got, want)
	}
}
