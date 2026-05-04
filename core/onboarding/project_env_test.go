package onboarding_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peternicholls/stageserve/core/onboarding"
)

// TestValidateProjectRoot_RejectsEmpty verifies that an empty project root is rejected.
func TestValidateProjectRoot_RejectsEmpty(t *testing.T) {
	_, err := onboarding.ValidateProjectRoot("")
	if err == nil {
		t.Error("want error for empty project root")
	}
}

// TestValidateProjectRoot_AcceptsExistingDir verifies that an existing directory is accepted.
func TestValidateProjectRoot_AcceptsExistingDir(t *testing.T) {
	dir := t.TempDir()
	got, err := onboarding.ValidateProjectRoot(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dir {
		t.Errorf("want %s, got %s", dir, got)
	}
}

// TestValidateProjectEnv_RejectsDocrootOutsideProjectRoot verifies that a
// docroot path that escapes the project root is rejected.
func TestValidateProjectEnv_RejectsDocrootOutsideProjectRoot(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Dir(dir) // parent of temp dir — outside project root
	err := onboarding.ValidateDocroot(dir, outside)
	if err == nil {
		t.Error("want error for docroot outside project root")
	}
}

// TestValidateProjectEnv_AcceptsRelativeDocroot verifies that a docroot
// relative path that stays within the project root is accepted.
func TestValidateProjectEnv_AcceptsRelativeDocroot(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "public_html")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := onboarding.ValidateDocroot(dir, "public_html"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWriteProjectEnv_SkipsExistingFile verifies that WriteProjectEnv does not
// overwrite an existing .env.stageserve unless force=true.
func TestWriteProjectEnv_SkipsExistingFile(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env.stageserve")
	original := "ORIGINAL=1\n"
	if err := os.WriteFile(envPath, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	action, err := onboarding.WriteProjectEnv(dir, "mysite", "public_html", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != onboarding.InitActionSkipped {
		t.Errorf("want skipped, got %s", action)
	}
	// File should be untouched.
	got, _ := os.ReadFile(envPath)
	if string(got) != original {
		t.Error("file was modified when it should have been preserved")
	}
}

// TestWriteProjectEnv_CreatesNewFile verifies that WriteProjectEnv creates a
// .env.stageserve when none exists.
func TestWriteProjectEnv_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	action, err := onboarding.WriteProjectEnv(dir, "mysite", "public_html", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != onboarding.InitActionCreated {
		t.Errorf("want created, got %s", action)
	}
	if _, err := os.Stat(filepath.Join(dir, ".env.stageserve")); err != nil {
		t.Error("expected .env.stageserve to exist")
	}
}

// TestWriteProjectEnv_OverwritesWithForce verifies that WriteProjectEnv
// overwrites an existing file when force=true.
func TestWriteProjectEnv_OverwritesWithForce(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env.stageserve")
	if err := os.WriteFile(envPath, []byte("OLD=1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	action, err := onboarding.WriteProjectEnv(dir, "newsite", "public_html", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != onboarding.InitActionOverwritten {
		t.Errorf("want overwritten, got %s", action)
	}
}
