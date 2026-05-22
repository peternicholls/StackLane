package guidance

import "github.com/peternicholls/stageserve/core/state"

// Plan chooses the next safe user-facing action from a collected context.
func Plan(ctx GuidedContext) NextActionPlan {
	situation := classify(ctx)
	plan := NextActionPlan{
		Situation:       situation,
		Title:           "StageServe",
		Warnings:        append([]string(nil), ctx.Warnings...),
		VisibleDefaults: visibleDefaults(ctx),
		FooterActions: []GuidedAction{
			footerAction("show_commands", "Show commands", "Show direct commands for this situation."),
			footerAction("advanced", "Advanced troubleshooting", "Show more detail for troubleshooting."),
		},
	}

	switch situation {
	case SituationMachineNotReady:
		plan.StatusHeader = "Your computer isn't ready yet."
		plan.Summary = "Fix the next setup item, then StageServe can plan the project step."
		plan.WorkItems = ctx.MachineReadiness.WorkItems
		if len(plan.WorkItems) == 0 {
			plan.WorkItems = []WorkItem{{Label: ctx.MachineReadiness.NextFixLabel, Status: "needs attention", DirectCommand: commandOr(ctx.MachineReadiness.NextCommand, "stage setup")}}
		}
		plan.DirectCommands = []string{"stage setup"}
	case SituationNotProject:
		plan.StatusHeader = "This folder isn't a StageServe project yet."
		plan.DecisionItems = []GuidedAction{action("init_here", "Set up this folder as a project", "Create project settings for this folder.", "stage init", true)}
		plan.DirectCommands = []string{"stage init"}
	case SituationProjectMissingConf:
		plan.StatusHeader = "This folder doesn't have StageServe settings yet."
		plan.Summary = "StageServe can create a small .env.stageserve file for this project."
		plan.DecisionItems = []GuidedAction{
			action("init", "Create project settings", "Write .env.stageserve with the values shown here.", "stage init", true),
			action("edit_config", "Edit before writing", "Change the project name, web folder, or local address first.", "stage init", false),
		}
		plan.DirectCommands = []string{"stage init"}
	case SituationProjectReadyToRun:
		plan.StatusHeader = "This project is ready to run."
		plan.DecisionItems = []GuidedAction{
			action("up", "Run this project", "Start the project and open it in your browser.", "stage up", true),
			action("edit_config", "Edit project settings", "Change the project name, web folder, or local address first.", ".env.stageserve", false),
		}
		plan.DirectCommands = []string{"stage up", "stage status"}
	case SituationProjectRunning:
		plan.StatusHeader = "This project is running at " + ctx.LocalURL + "."
		plan.DecisionItems = []GuidedAction{
			action("logs", "View project logs", "Watch what your project is doing right now.", "stage logs", false),
			action("down", "Stop this project", "Stop the project after confirmation.", "stage down", true),
		}
		plan.DirectCommands = []string{"stage logs", "stage status", "stage down"}
	case SituationProjectDown:
		plan.StatusHeader = "This project is stopped."
		plan.DecisionItems = []GuidedAction{
			action("up", "Run this project again", "Start the project and open it in your browser.", "stage up", true),
			action("detach", "Remove this project from StageServe", "Forget the stopped project after confirmation.", "stage detach", true),
		}
		plan.DirectCommands = []string{"stage up", "stage detach"}
	case SituationDriftDetected:
		plan.StatusHeader = "This project doesn't match what StageServe expects."
		plan.Summary = "Review the mismatch before starting or changing anything."
		plan.DecisionItems = []GuidedAction{action("diagnose", "Troubleshoot this problem", "Check the project and show the safest next step.", "stage doctor", false)}
		plan.DirectCommands = []string{"stage status", "stage doctor"}
	case SituationUnknownError:
		plan.StatusHeader = "StageServe couldn't safely choose a next step."
		plan.Summary = "Use the recovery commands below, starting with the read-only check."
		plan.WorkItems = []WorkItem{{Label: "Troubleshoot this problem", Status: "next", DirectCommand: "stage doctor"}}
		plan.DirectCommands = []string{"stage doctor", "stage status"}
	}
	return plan
}

func classify(ctx GuidedContext) Situation {
	switch {
	case ctx.Err != nil:
		return SituationUnknownError
	case ctx.MachineReadiness.Checked && ctx.MachineReadiness.Blocked:
		return SituationMachineNotReady
	case ctx.NotProject:
		return SituationNotProject
	case !ctx.ProjectEnvExists:
		return SituationProjectMissingConf
	case len(ctx.Warnings) > 0:
		return SituationDriftDetected
	case ctx.ProjectState != nil && ctx.ProjectState.AttachmentState == state.StateAttached:
		return SituationProjectRunning
	case ctx.ProjectState != nil && ctx.ProjectState.AttachmentState == state.StateDown:
		return SituationProjectDown
	default:
		return SituationProjectReadyToRun
	}
}

func visibleDefaults(ctx GuidedContext) []VisibleDefault {
	defaults := []VisibleDefault{}
	add := func(label, value, note string) {
		if value != "" {
			defaults = append(defaults, VisibleDefault{Label: label, Value: value, Note: note})
		}
	}
	add("Project", ctx.SiteName, "")
	add("Web folder", ctx.WebFolder, "")
	add("Local URL", ctx.LocalURL, "")
	if ctx.ProjectEnvPreview != nil {
		add("Settings file", ctx.ProjectEnvPreview.Path, "will be created after confirmation")
	} else {
		add("Settings file", ctx.ProjectEnvPath, "")
	}
	add("Stack", ctx.StackID, "")
	return defaults
}

func action(id, label, description, directCommand string, mutates bool) GuidedAction {
	return GuidedAction{ID: id, Kind: "choose", Label: label, Description: description, MutatesState: mutates, RequiresConfirmation: mutates, DirectCommand: directCommand}
}

func footerAction(id, label, description string) GuidedAction {
	return GuidedAction{ID: id, Kind: "footer", Label: label, Description: description}
}

func commandOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
