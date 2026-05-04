// Tests for project-level pure helpers.
package project

import "testing"

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"My Cool Site":       "my-cool-site",
		"  spaces  ":         "spaces",
		"Already-Slug":       "already-slug",
		"!!chars$$":          "chars",
		"":                   "site",
		"stageserve-foo--bar": "stageserve-foo-bar",
	}
	for in, want := range cases {
		got := Slugify(in)
		if got != want {
			t.Errorf("Slugify(%q)=%q want %q", in, got, want)
		}
	}
}

func TestHostnameValid(t *testing.T) {
	good := []string{"alpha.test", "alpha-1.example.com", "a.b.c"}
	bad := []string{"", "no_underscore.test", " spaces.test", "trailing.", ".leading"}
	for _, h := range good {
		if !HostnameValid(h) {
			t.Errorf("expected %q valid", h)
		}
	}
	for _, h := range bad {
		if HostnameValid(h) {
			t.Errorf("expected %q invalid", h)
		}
	}
}

func TestResolveHostname(t *testing.T) {
	host, suffix := ResolveHostname("alpha", "", "test")
	if host != "alpha.test" || suffix != "test" {
		t.Errorf("default: host=%q suffix=%q", host, suffix)
	}
	host, suffix = ResolveHostname("alpha", "custom.local", "test")
	if host != "custom.local" || suffix != "test" {
		t.Errorf("explicit: host=%q suffix=%q", host, suffix)
	}
}
