package commands

import (
	"io"
	"os"

	"github.com/peternicholls/stageserve/core/guidance"
)

func runGuidedTUI(plan guidance.NextActionPlan, capability guidance.TUICapability, output io.Writer) error {
	return guidance.RunShell(plan, capability, os.Stdin, output)
}
