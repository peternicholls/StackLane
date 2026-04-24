// State store round-trip tests. Each subtest uses an isolated state directory.
package state

import (
	"testing"

	"github.com/peternicholls/stacklane/core/config"
)

func TestStore_SaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	rec := Record{
		Project: config.ProjectConfig{
			Slug:               "demo",
			Name:               "demo",
			Dir:                "/tmp/demo",
			Hostname:           "demo.test",
			ComposeProjectName: "stacklane-demo",
			MySQL:              config.MySQL{Port: 33060, PMAPort: 8082},
		},
		AttachmentState: StateAttached,
	}
	if err := store.Save(rec); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := store.Load("demo")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Project.Hostname != "demo.test" || got.AttachmentState != StateAttached {
		t.Errorf("round-trip mismatch: %+v", got)
	}
	if got.SchemaVersion != SchemaVersion {
		t.Errorf("schema version not stamped: %d", got.SchemaVersion)
	}
}

func TestStore_RegistryProjection(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	for _, slug := range []string{"alpha", "beta"} {
		rec := Record{
			Project: config.ProjectConfig{
				Slug: slug, Name: slug, Hostname: slug + ".test",
				ComposeProjectName: "stacklane-" + slug,
			},
			AttachmentState: StateAttached,
		}
		if err := store.Save(rec); err != nil {
			t.Fatalf("save %s: %v", slug, err)
		}
	}
	rows, err := store.Registry()
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("registry rows=%d want 2", len(rows))
	}
	if rows[0].Slug != "alpha" || rows[1].Slug != "beta" {
		t.Errorf("registry not slug-sorted: %+v", rows)
	}
}

func TestStore_Remove(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir)
	rec := Record{Project: config.ProjectConfig{Slug: "gone"}}
	_ = store.Save(rec)
	if err := store.Remove("gone"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := store.Load("gone"); err == nil {
		t.Errorf("load after remove should fail")
	}
	if err := store.Remove("never-existed"); err != nil {
		t.Errorf("remove of missing slug should be idempotent: %v", err)
	}
}
