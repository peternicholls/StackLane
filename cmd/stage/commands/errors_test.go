package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/peternicholls/stageserve/core/lifecycle"
)

func TestRenderCommandErrorFormatsLifecycleStepError(t *testing.T) {
	err := lifecycle.Wrap("runtime-asset-shared", "demo", errors.New("missing compose file"), "Run stage doctor.")
	out := RenderCommandError(err)

	for _, want := range []string{"StageServe could not complete runtime-asset-shared for demo.", "Problem: missing compose file", "Next step: Run stage doctor."} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "StageServe\nError\n") {
		t.Fatalf("severity header missing:\n%s", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("direct error output contains ANSI escape sequences:\n%s", out)
	}
}
