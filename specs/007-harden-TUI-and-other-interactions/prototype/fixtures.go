package main

import (
	"fmt"
	"strings"
)

func planFixtures() map[situation]plan {
	baseFooter := prototypeFooter()
	defaults := defaultEditValues()
	return map[situation]plan{
		machineNotReady: {
			Situation:    machineNotReady,
			StatusHeader: "Setting up your computer",
			Context:      prototypeProjectDir,
			Summary:      "StageServe is checking the computer before it looks at the project.",
			WorkItems: []workItem{
				{Label: "Docker Desktop", Status: statusReady, Description: "Docker Desktop is installed."},
				{Label: "StageServe working folder", Status: statusReady, Description: "StageServe can use its working folder."},
				{Label: "Local DNS for .develop", Status: statusNeedsApproval, Description: "StageServe will add a small file so your browser can open addresses like http://pete-site.develop.", EnterAction: "Preview the change and ask for permission to set up local DNS for .develop.", Details: "The real command would explain /etc/resolver/develop only after the user asks for detail."},
				{Label: "Local HTTPS certificates", Status: statusOptional, Description: "The example URL is plain HTTP, so certificates are not blocking first run."},
				{Label: "Network ports 80 and 443", Status: statusPending, Description: "Checked after local DNS."},
			},
			ActiveWorkIndex: 2,
			DetailsTitle:    "Why this computer needs setup",
			Details: []string{
				"StageServe prepares this computer once, then each project can run with a local address.",
				"The main thing a normal user needs to know is the address they will open in the browser.",
				"Advanced implementation names stay behind the advanced view.",
			},
			DirectCommands: []string{"stage setup", "stage doctor"},
			Advanced:       []string{"Advanced checks would show Docker, DNS resolver, mkcert, ports, and working-folder details."},
			Footer:         baseFooter,
		},
		projectMissingConfig: projectSetupPlan(baseFooter, defaults),
		projectReadyToRun:    projectReadyToRunPlan(baseFooter, defaults),
		projectRunning:       projectRunningPlan(baseFooter, defaults),
		projectDown:          projectDownPlan(baseFooter, defaults),
		driftDetected:        driftDetectedPlan(baseFooter, defaults),
		notProject: {
			Situation:    notProject,
			StatusHeader: "This folder is not a StageServe project yet.",
			Context:      prototypeNonProjectDir,
			Summary:      "StageServe can set up this folder, or you can point it at a different folder.",
			Defaults: []defaultValue{
				{Label: "Proposed name", Value: "downloads", Note: "from folder name"},
				{Label: "Domain suffix", Value: ".develop", Note: "machine setting"},
				{Label: "Local URL", Value: "http://downloads.develop", Note: "preview"},
			},
			Decisions: []decisionItem{
				{ID: "init_here", Label: "Set up this folder as a project", Description: "Create project settings here and continue.", DirectCommand: "stage init", Next: projectMissingConfig, ResultTitle: "Project setup preview opened", ResultBody: "Prototype only: StageServe would show the project setup preview."},
				{ID: "pick_folder", Label: "Pick a different folder", Description: "Type a path to look at instead.", ResultTitle: "Different folder", ResultBody: "Prototype only: a path prompt would re-run detection in that folder."},
				moreDecision(),
			},
			DetailsTitle:   "Folder overview",
			Details:        []string{"This prototype stays scoped to one current folder.", "A future version can add a project switcher, but spec 007 does not require it."},
			DirectCommands: []string{"stage init", "stage setup"},
			Advanced:       []string{"Advanced view would explain how StageServe finds project roots."},
			Footer:         baseFooter,
		},
		doctorReportNeedsHelp: {
			Situation:    doctorReportNeedsHelp,
			Surface:      "Doctor",
			StatusHeader: "Not ready - 2 of 7 checks need attention.",
			Context:      "Read-only machine check",
			ReportAttention: []reportItem{
				{
					Label:       "Port 443",
					Description: "Something else on your computer is using port 443.",
					Message:     "StageServe needs elevated permission to identify the process.",
					Command:     "sudo lsof -nP -iTCP:443 -sTCP:LISTEN",
				},
				{
					Label:       "Local DNS resolver",
					Description: "Your computer cannot yet open local project URLs.",
					Message:     "Local DNS is not set up yet.",
					Command:     "stage setup",
				},
			},
			ReportReady: []reportItem{
				{Label: "Docker CLI", Status: statusReady, Message: "docker found at /usr/local/bin/docker"},
				{Label: "Docker Desktop", Status: statusReady, Message: "running"},
				{Label: "State directory", Status: statusReady, Message: "exists"},
				{Label: "Port 80", Status: statusReady, Message: "available"},
				{Label: "mkcert local CA", Status: statusReady, Message: "installed"},
			},
			AssistanceTitle: "Assistance",
			Decisions: []decisionItem{
				{ID: "assist", Label: "Help me fix these", Description: "Walk through each issue one at a time.", Opens: modeAssist},
				{ID: "leave", Label: "Leave it here", Description: "Exit without changing anything.", Quits: true, ResultTitle: "No changes made", ResultBody: "Prototype only: StageServe would exit without changing anything."},
			},
			DetailsTitle:   "Doctor report overview",
			Details:        []string{"StageServe shows the report first so commands remain copy-pasteable.", "Guided help starts only when the user asks for it."},
			DirectCommands: []string{"sudo lsof -nP -iTCP:443 -sTCP:LISTEN", "stage setup", "stage doctor"},
			Advanced:       []string{"Advanced view would include internal check IDs and raw command output."},
			Footer:         baseFooter,
		},
		unknownError: {
			Situation:    unknownError,
			StatusHeader: "StageServe could not safely choose a next step.",
			Context:      prototypeProjectDir,
			Summary:      "StageServe does not want to guess. It can walk through safe recovery steps in order.",
			WorkItems: []workItem{
				{Label: "Step 1: look at this project's current state", Status: statusNext, Description: "Read-only. Nothing on your computer will be changed.", EnterAction: "Run the read-only recovery step."},
				{Label: "Step 2: look at project logs", Status: statusPending, Description: "Read-only."},
				{Label: "Step 3: stop and forget the running record", Status: statusPending, Description: "Requires confirmation before changing StageServe records."},
				{Label: "Step 4: run this project from scratch", Status: statusPending, Description: "Uses existing settings."},
			},
			ActiveWorkIndex: 0,
			Decisions: []decisionItem{
				{ID: "recover", Label: "Run next recovery step", Description: "Run the next least-invasive step.", DirectCommand: "stage doctor", ResultTitle: "Recovery step finished", ResultBody: "Prototype only: StageServe would run a read-only check, then re-plan."},
				{ID: "details", Label: "Show what went wrong", Description: "Read a longer plain-language explanation.", Opens: modeDetails},
				{ID: "stop", Label: "Stop here", Description: "Leave everything as it is and exit this recovery flow.", RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: "Stop recovery for now?", ConfirmBody: []string{"StageServe will leave this project as it is.", "Your files will not be touched.", "Run stage again to come back."}, ResultTitle: "Recovery stopped", ResultBody: "Prototype only: no changes were made."},
				moreDecision(),
			},
			DetailsTitle:   "What went wrong",
			Details:        []string{"StageServe was checking this project and got an answer it could not classify safely.", "The recovery list starts read-only and pauses after each step."},
			DirectCommands: []string{"stage doctor", "stage status", "stage logs"},
			Advanced:       []string{"Advanced view would include raw command output and lower-level error names."},
			Footer:         baseFooter,
		},
	}
}

func defaultEditValues() editValues {
	return editValues{SiteName: "pete-site", WebFolder: "./public_html", Suffix: ".develop"}
}

func prototypeFooter() []string {
	return []string{"? details", "m more...", "q quit"}
}

func moreDecision() decisionItem {
	return decisionItem{ID: "more", Label: "More…", Description: "Show direct commands, plain text output, and advanced detail.", Opens: modeCommands}
}

func newProjectValues(values editValues) projectValues {
	defaults := defaultEditValues()
	siteName := strings.TrimSpace(values.SiteName)
	if siteName == "" {
		siteName = defaults.SiteName
	}
	webFolder := strings.TrimSpace(values.WebFolder)
	if webFolder == "" {
		webFolder = defaults.WebFolder
	}
	suffix := strings.TrimSpace(values.Suffix)
	if suffix == "" {
		suffix = defaults.Suffix
	}
	if !strings.HasPrefix(suffix, ".") {
		suffix = "." + suffix
	}
	host := siteName + suffix
	return projectValues{
		Root:       prototypeProjectDir,
		SiteName:   siteName,
		WebFolder:  webFolder,
		Suffix:     suffix,
		Host:       host,
		URL:        "http://" + host,
		ConfigPath: prototypeProjectDir + "/.env.stageserve",
	}
}

func projectSetupPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	decisions := []decisionItem{
		{ID: "init", Label: "Use these settings", Description: "Write .env.stageserve and continue.", DirectCommand: "stage init", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: "About to write project settings", ConfirmBody: []string{fmt.Sprintf("StageServe will create %s.", current.ConfigPath), fmt.Sprintf("Site name: %s", current.SiteName), fmt.Sprintf("Web folder: %s", current.WebFolder), fmt.Sprintf("Domain suffix: %s", current.Suffix), fmt.Sprintf("Local URL: %s", current.URL), "StageServe will not change any other file."}, Next: projectReadyToRun, ResultTitle: "Project settings created", ResultBody: fmt.Sprintf("Prototype only: StageServe would write %s, then re-detect.", current.ConfigPath)},
		{ID: "edit", Label: "Edit before writing", Description: "Change site name, web folder, or suffix first.", Opens: modeEdit},
		moreDecision(),
	}
	return plan{
		Situation:    projectMissingConfig,
		StatusHeader: "This folder doesn't have StageServe settings yet.",
		Context:      current.Root,
		Summary:      "StageServe will create one file only: .env.stageserve.",
		Defaults: append(commonDefaults(current),
			defaultValue{Label: "Target file", Value: current.ConfigPath, Note: "will be created in this folder"},
			defaultValue{Label: "Stack", Value: "20i", Note: "current supported stack"},
			defaultValue{Label: "Advanced settings", Value: "none yet", Note: "details stay out of first run"}),
		Decisions:      decisions,
		DetailsTitle:   "Project setup overview",
		Details:        []string{fmt.Sprintf("Target file: %s", current.ConfigPath), "The preview shows every value before any write.", "Edits return to the preview. They do not write immediately.", "Advanced settings are summarized instead of forced into first-run setup."},
		DirectCommands: []string{"stage init", "stage init --cli", "stage init --json"},
		Advanced:       []string{"Advanced settings such as PHP, MySQL, timeout, and post-up commands remain in .env.stageserve and direct CLI docs."},
		Footer:         baseFooter,
	}
}

func projectReadyToRunPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	decisions := []decisionItem{
		{ID: "up", Label: "Run this project", Description: "Start the project and open it in your browser.", DirectCommand: "stage up", Next: projectRunning, ResultTitle: "Project started", ResultBody: fmt.Sprintf("Prototype only: StageServe would start the project at %s.", current.URL)},
		{ID: "edit", Label: "Edit project settings", Description: "Change site name, web folder, or suffix before running.", DirectCommand: "stage init", Opens: modeEdit},
		moreDecision(),
	}
	return plan{
		Situation:      projectReadyToRun,
		StatusHeader:   "This project is ready to run.",
		Context:        current.Root,
		Summary:        "StageServe has project settings and can start the local site.",
		Defaults:       commonDefaults(current),
		Decisions:      decisions,
		DetailsTitle:   "Project overview",
		Details:        []string{fmt.Sprintf("Target file: %s", current.ConfigPath), "The project settings file exists.", "StageServe will use the visible defaults unless you edit them first."},
		DirectCommands: []string{"stage up", "stage init --cli", "stage status"},
		Advanced:       []string{"Advanced view would show PHP/MySQL overrides if the project file already had them."},
		Footer:         baseFooter,
	}
}

func projectRunningPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	decisions := []decisionItem{
		{ID: "logs", Label: "View project logs", Description: "Watch what your project is doing right now.", DirectCommand: "stage logs", Opens: modeLogs},
		{ID: "down", Label: "Stop this project", Description: "Free up the local URL and shut down the project.", DirectCommand: "stage down", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: fmt.Sprintf("Stop %s?", current.SiteName), ConfirmBody: []string{fmt.Sprintf("StageServe will stop %s. Your files will not be touched.", current.SiteName), fmt.Sprintf("%s will no longer respond.", current.URL), "You can run it again any time."}, Next: projectDown, ResultTitle: "Project stopped", ResultBody: fmt.Sprintf("Prototype only: StageServe would stop %s and preserve project files.", current.SiteName)},
		moreDecision(),
	}
	return plan{
		Situation:    projectRunning,
		StatusHeader: fmt.Sprintf("%s is running", current.SiteName),
		Context:      current.Root,
		Summary:      "The local site is available. Viewing logs is the safest next step.",
		Defaults: append(commonDefaults(current),
			defaultValue{Label: "Started", Value: "4 minutes ago", Note: "healthy"}),
		Decisions:      decisions,
		DetailsTitle:   "Running project overview",
		Details:        []string{fmt.Sprintf("Local URL: %s", current.URL), fmt.Sprintf("Web folder: %s", current.WebFolder), "Status: healthy", "Default action: view logs."},
		DirectCommands: []string{"stage status", "stage logs", "stage down"},
		Advanced:       []string{"Advanced troubleshooting would show service names, route details, and state paths only here."},
		Footer:         append([]string{"right open in browser"}, baseFooter...),
	}
}

func projectDownPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	decisions := []decisionItem{
		{ID: "up", Label: "Run this project", Description: "Start the project again.", DirectCommand: "stage up", Next: projectRunning, ResultTitle: "Project started", ResultBody: fmt.Sprintf("Prototype only: StageServe would start %s again.", current.SiteName)},
		{ID: "detach", Label: "Remove this project from StageServe", Description: "Stop tracking this project. Your files will not be touched.", DirectCommand: "stage detach", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: false, ConfirmTitle: fmt.Sprintf("Remove %s from StageServe?", current.SiteName), ConfirmBody: []string{fmt.Sprintf("StageServe will forget about %s.", current.SiteName), fmt.Sprintf("%s stays as it is.", current.ConfigPath), "All your project files stay as they are.", fmt.Sprintf("%s will no longer be routed by StageServe.", current.URL)}, Next: notProject, ResultTitle: "Project removed from StageServe", ResultBody: fmt.Sprintf("Prototype only: StageServe would forget %s without deleting files.", current.SiteName)},
		moreDecision(),
	}
	return plan{
		Situation:      projectDown,
		StatusHeader:   fmt.Sprintf("%s is stopped", current.SiteName),
		Context:        current.Root,
		Summary:        "StageServe still knows this project, but it is not running.",
		Defaults:       commonDefaults(current),
		Decisions:      decisions,
		DetailsTitle:   "Stopped project overview",
		Details:        []string{fmt.Sprintf("Local URL: %s", current.URL), "The project has retained settings.", "The safest default is to run it again."},
		DirectCommands: []string{"stage up", "stage detach", "stage status"},
		Advanced:       []string{"Advanced view would show retained records and route state."},
		Footer:         baseFooter,
	}
}

func driftDetectedPlan(baseFooter []string, values editValues) plan {
	current := newProjectValues(values)
	decisions := []decisionItem{
		{ID: "safe", Label: "Use the safe next step", Description: "Treat this project as stopped. Nothing in your folder will be deleted.", Mutates: true, RequiresConfirm: true, ConfirmYesDefault: true, ConfirmTitle: fmt.Sprintf("Reset %s's running record?", current.SiteName), ConfirmBody: []string{fmt.Sprintf("StageServe will forget that %s is running.", current.SiteName), fmt.Sprintf("%s stays as it is.", current.ConfigPath), "All your project files stay as they are.", fmt.Sprintf("After this, you can choose 'Run this project' to start %s again.", current.SiteName)}, Next: projectDown, ResultTitle: "Safe recovery applied", ResultBody: fmt.Sprintf("Prototype only: StageServe would reset %s's running record and re-detect.", current.SiteName)},
		{ID: "retry", Label: "Try to start it again", Description: "Run this project with its current settings.", DirectCommand: "stage up", Next: projectRunning, ResultTitle: "Project started", ResultBody: fmt.Sprintf("Prototype only: StageServe would try stage up again for %s.", current.SiteName)},
		{ID: "details", Label: "Show what does not match", Description: "Read a longer plain-language explanation.", Opens: modeDetails},
		moreDecision(),
	}
	return plan{
		Situation:    driftDetected,
		StatusHeader: fmt.Sprintf("%s does not match what StageServe expects", current.SiteName),
		Context:      current.Root,
		Summary:      fmt.Sprintf("StageServe expected %s to be running, but %s is not responding.", current.SiteName, current.URL),
		Defaults: []defaultValue{
			{Label: "Recorded as running", Value: "yes"},
			{Label: "Project service found", Value: "no"},
			{Label: "Local address connected", Value: "no"},
			{Label: "DNS", Value: current.Host + " resolves"},
		},
		Decisions:      decisions,
		DetailsTitle:   "What does not match",
		Details:        []string{"Project record: StageServe thinks this is running.", fmt.Sprintf("Local URL response: %s is not responding.", current.URL), fmt.Sprintf("DNS for %s: working.", current.Suffix), fmt.Sprintf("Local address link: not connected for %s.", current.Host), "This usually happens after a restart or a stop outside StageServe."},
		DirectCommands: []string{"stage status", "stage doctor", "stage up"},
		Advanced:       []string{"Advanced view would show the project record path and lower-level routing checks."},
		Footer:         baseFooter,
	}
}

func commonDefaults(values projectValues) []defaultValue {
	return []defaultValue{
		{Label: "Site name", Value: values.SiteName, Note: "from folder name"},
		{Label: "Web folder", Value: values.WebFolder, Note: "found here"},
		{Label: "Domain suffix", Value: values.Suffix, Note: "machine setting"},
		{Label: "Local URL", Value: values.URL, Note: "what you'll visit"},
	}
}
