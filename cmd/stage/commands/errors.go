package commands

import (
	"fmt"

	"github.com/peternicholls/stageserve/core/lifecycle"
)

func RenderCommandError(err error) string {
	if err == nil {
		return ""
	}
	if step, ok := lifecycle.AsStepError(err); ok {
		project := step.Project
		if project == "" {
			project = "shared runtime"
		}
		out := fmt.Sprintf("StageServe\nError\n\nStageServe could not complete %s for %s.\n\nProblem: %v", step.Step, project, step.Cause)
		if step.Remedy != "" {
			out += "\nNext step: " + step.Remedy
		}
		return out
	}
	return err.Error()
}
