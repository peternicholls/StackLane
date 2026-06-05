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
			action("init", "Set up this directory as a project", "Write .env.stageserve with the values shown here.", "stage init", true),
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
			action("open_browser", "Open in browser", "Open "+ctx.LocalURL+" in your default browser.", ctx.LocalURL, false),
			action("logs", "View project logs", "Watch what your project is doing right now.", "stage logs", false),
			action("status", "Check this project's status", "See the latest recorded and live project status.", "stage status", false),
			action("down", "Stop this project", "Stop the project after confirmation.", "stage down", true),
		}
		plan.DirectCommands = []string{"stage logs", "stage status", "stage down"}
	case SituationProjectDown:
		if ctx.Runtime.Checked && ctx.Runtime.Running {
			plan.StatusHeader = "This project isn't added to StageServe right now."
			plan.Summary = "Your project is already running. StageServe can add the local URL back without restarting it."
			plan.DecisionItems = []GuidedAction{
				action("attach", "Add this project to StageServe", "Add the local URL back without restarting the project.", "stage attach", true),
				action("status", "Check this project's status", "See the latest recorded and live project status.", "stage status", false),
				action("down", "Stop this project", "Stop the project after confirmation.", "stage down", true),
			}
			plan.DirectCommands = []string{"stage attach", "stage status", "stage down"}
			break
		}
		plan.StatusHeader = "This project is stopped."
		plan.DecisionItems = []GuidedAction{
			action("up", "Run this project again", "Start the project and open it in your browser.", "stage up", true),
			action("status", "Check this project's status", "See the latest recorded and live project status.", "stage status", false),
			action("detach", "Remove this project from StageServe", "Stop tracking this project. Your files will not be touched.", "stage detach", true),
		}
		plan.DirectCommands = []string{"stage up", "stage status", "stage detach"}
	case SituationDriftDetected:
		plan.StatusHeader = "This project doesn't match what StageServe expects."
		plan.Summary = "StageServe can walk through the safest checks first, then you can decide what to do next."
		plan.WorkItems = recoveryWorkItems()
		plan.DecisionItems = []GuidedAction{
			action("doctor", "Step 1: run diagnostics", "Read-only. Check machine and runtime readiness before changing anything.", "stage doctor", false),
			action("status", "Step 2: check this project's status", "Read-only. Nothing on your computer will be changed.", "stage status", false),
			action("logs", "Step 3: look at the latest project log", "Read-only. This shows the latest log output for the project.", "stage logs", false),
			action("up", "Step 4: try running this project again", "Restart with the current settings.", "stage up", true),
			action("down", "Step 5: stop this project first", "Stop the project, then try again.", "stage down", true),
		}
		plan.DirectCommands = []string{"stage doctor", "stage status", "stage logs", "stage down", "stage up"}
	case SituationUnknownError:
		plan.StatusHeader = "StageServe couldn't safely choose a next step."
		plan.Summary = "Here is what StageServe can try, in order of risk."
		plan.WorkItems = recoveryWorkItems()
		plan.DecisionItems = []GuidedAction{
			action("doctor", "Step 1: run diagnostics", "Read-only. Check machine and runtime readiness before changing anything.", "stage doctor", false),
			action("status", "Step 2: check this project's status", "Read-only. Nothing on your computer will be changed.", "stage status", false),
			action("logs", "Step 3: look at the latest project log", "Read-only. This shows the latest log output for the project.", "stage logs", false),
			action("up", "Step 4: try running this project again", "Restart with the current settings.", "stage up", true),
			action("down", "Step 5: stop this project first, then retry", "Stop the project, then try again.", "stage down", true),
		}
		plan.DirectCommands = []string{"stage doctor", "stage status", "stage logs", "stage down", "stage up"}
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
	add("Site name", ctx.SiteName, "")
	add("Web folder", ctx.WebFolder, "")
	add("Domain suffix", displaySuffix(ctx.SiteSuffix), "")
	add("Local URL", ctx.LocalURL, "")
	add("Status", displayProjectStatus(ctx), "")
	if ctx.ProjectEnvPreview != nil {
		add("Settings file", ctx.ProjectEnvPreview.Path, "will be created after confirmation")
	} else {
		add("Settings file", ctx.ProjectEnvPath, "")
	}
	add("Stack", ctx.StackID, "")
	return defaults
}

func displayProjectStatus(ctx GuidedContext) string {
	switch {
	case ctx.ProjectState != nil && ctx.ProjectState.AttachmentState == state.StateAttached:
		return "running"
	case ctx.ProjectState != nil && ctx.ProjectState.AttachmentState == state.StateDown:
		if ctx.Runtime.Checked && ctx.Runtime.Running {
			return "running without local routing"
		}
		return "stopped"
	case ctx.ProjectEnvExists:
		return "not running yet"
	default:
		return ""
	}
}

func recoveryWorkItems() []WorkItem {
	return []WorkItem{
		{Label: "Step 1: run diagnostics", Status: "read-only", Description: "Checks machine and runtime readiness before changing anything.", DirectCommand: "stage doctor"},
		{Label: "Step 2: check this project's status", Status: "read-only", Description: "Shows what StageServe currently knows about this project.", DirectCommand: "stage status"},
		{Label: "Step 3: look at the latest project log", Status: "read-only", Description: "Shows the latest log output without changing anything.", DirectCommand: "stage logs"},
		{Label: "Step 4: stop this project", Status: "confirmed change", Description: "Stops the project and frees up its local URL after confirmation.", DirectCommand: "stage down"},
		{Label: "Step 5: run this project again", Status: "uses current settings", Description: "Starts the project again with the settings shown above.", DirectCommand: "stage up"},
	}
}

func displaySuffix(suffix string) string {
	if suffix == "" {
		return ""
	}
	if suffix[0] == '.' {
		return suffix
	}
	return "." + suffix
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
