package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func commandHelp(t *testing.T, args ...string) string {
	t.Helper()
	root := NewRoot("test")
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(append(args, "--help"))
	if err := root.Execute(); err != nil {
		t.Fatalf("help %v: %v", args, err)
	}
	return buf.String()
}

func TestOnboardingHelpUsesFinalPlainTextOptOuts(t *testing.T) {
	for _, command := range []string{"setup", "doctor", "init"} {
		help := commandHelp(t, command)
		for _, want := range []string{"--notui", "--cli"} {
			if !strings.Contains(help, want) {
				t.Fatalf("%s help missing %s:\n%s", command, want, help)
			}
		}
		for _, stale := range []string{"--tui", "--no-tui"} {
			if strings.Contains(help, stale) {
				t.Fatalf("%s help should not expose %s:\n%s", command, stale, help)
			}
		}
	}
}

func TestStatusAndLogsHelpExposeDocumentedSelectors(t *testing.T) {
	statusHelp := commandHelp(t, "status")
	if !strings.Contains(statusHelp, "--project") {
		t.Fatalf("status help missing --project:\n%s", statusHelp)
	}

	logsHelp := commandHelp(t, "logs")
	for _, want := range []string{"logs [service]", "--project", "--service"} {
		if !strings.Contains(logsHelp, want) {
			t.Fatalf("logs help missing %s:\n%s", want, logsHelp)
		}
	}
}

func TestResolveLogServiceAcceptsPositionalOrFlag(t *testing.T) {
	tests := []struct {
		name      string
		flagValue string
		args      []string
		want      string
		wantErr   bool
	}{
		{name: "default", want: "nginx"},
		{name: "flag", flagValue: "apache", want: "apache"},
		{name: "positional", args: []string{"apache"}, want: "apache"},
		{name: "same flag and positional", flagValue: "apache", args: []string{"apache"}, want: "apache"},
		{name: "conflict", flagValue: "nginx", args: []string{"apache"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := resolveLogService(test.flagValue, test.args)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveLogService: %v", err)
			}
			if got != test.want {
				t.Fatalf("service=%q want %q", got, test.want)
			}
		})
	}
}

func TestActiveDocsAvoidRemovedCommandPromises(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	activeDocs := []string{
		filepath.Join(repoRoot, "README.md"),
		filepath.Join(repoRoot, "docs", "runtime-contract.md"),
	}

	for _, path := range activeDocs {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(body)
		for _, stale := range []string{"--recheck", "--tui", "--no-tui"} {
			if strings.Contains(content, stale) {
				t.Fatalf("%s still documents removed command promise %s", path, stale)
			}
		}
	}
}
