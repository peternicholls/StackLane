package main

import (
	"fmt"
	"strings"
)

type guide struct {
	Sections []guideSection
}

type guideSection struct {
	ID      string
	Title   string
	Surface string
	Render  func(width int) string
}

func (styleGuide guide) indexForID(sectionID string) int {
	for sectionIndex, section := range styleGuide.Sections {
		if section.ID == sectionID {
			return sectionIndex
		}
	}
	return 0
}

func (styleGuide guide) sectionByID(sectionID string) (guideSection, bool) {
	for _, section := range styleGuide.Sections {
		if section.ID == sectionID {
			return section, true
		}
	}
	return guideSection{}, false
}

func buildGuide() guide {
	return guide{Sections: []guideSection{
		{ID: "start", Title: "Start here", Surface: "Guide", Render: renderStartHere},
		{ID: "intent", Title: "Design intent", Surface: "Intent", Render: renderIntent},
		{ID: "anatomy", Title: "Screen anatomy", Surface: "Anatomy", Render: renderAnatomy},
		{ID: "colour", Title: "Colour system", Surface: "Colour", Render: renderColourSystem},
		{ID: "components", Title: "Component vocabulary", Surface: "Components", Render: renderComponents},
		{ID: "report", Title: "Report surface", Surface: "Doctor", Render: renderReportSurface},
		{ID: "guided", Title: "Guided surface", Surface: "Project", Render: renderGuidedSurface},
		{ID: "assistance", Title: "Report to assistance", Surface: "Assistance", Render: renderAssistance},
		{ID: "separation", Title: "Design and code layers", Surface: "Boundaries", Render: renderSeparation},
		{ID: "review", Title: "Style review", Surface: "Review", Render: renderReview},
	}}
}

func renderStartHere(width int) string {
	return renderPage("StageServe Style Guide", "Guide", "This is the terminal design style guide, running as a terminal interface.", width, func(builder *strings.Builder) {
		renderParagraph(builder, width, "This app is the style guide in the graphic-design sense: a central, viewable reference that explains the identity, demonstrates components, and gives reviewers a shared standard. It is separate from spec prototypes and production command paths.")
		renderRule(builder, width, "How to read it", styleCyan)
		renderBullets(builder, []string{
			"Use ↑/↓ or tab to move through sections.",
			"Use j/k or page keys to scroll a long section.",
			"Read examples as design applications, not frozen screenshots.",
			"Treat Bubble Tea and Lip Gloss as implementation tools for this identity.",
		})
		renderRule(builder, width, "What it replaces", styleNeedsAction)
		renderParagraph(builder, width, "Agents should no longer treat a prototype, renderer helper, or historical spec decision as the style guide. The guide is the visual identity reference; code implements it through tokens, components, screen plans, and actions.")
		renderCallout(builder, "Run", "go run ./docs/design/style-guide-tui")
	})
}

func renderIntent(width int) string {
	return renderPage("StageServe", "Intent", "StageServe should feel like a capable local-development tool that is quietly in control.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Identity traits", styleCyan)
		renderTermRows(builder, []termRow{
			{"Clear", "The first scan tells the user where they are, what state they are in, and what matters."},
			{"Competent", "The tool appears to understand the workflow rather than dumping implementation state."},
			{"Calm", "Failures are direct, useful, and untheatrical."},
			{"Practical", "Every blocker has exact remediation or a safe guided next step."},
			{"Recognisable", "Reports, guided screens, confirmations, and fallbacks feel like one product."},
			{"Evolvable", "Better patterns can replace earlier prototype choices when the design reason is clear."},
		})
		renderRule(builder, width, "Anchor", styleReady)
		renderParagraph(builder, width, "The current stage doctor surface is the production visual seed: report anatomy, semantic colour, section rhythm, and quiet confidence. It is the anchor, not the final ceiling for the guided TUI.")
	})
}

func renderAnatomy(width int) string {
	return renderPage("StageServe", "Anatomy", "Every screen answers context, state, priority, and next action in that order.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Screen grammar", styleCyan)
		renderNumbered(builder, []string{
			"Surface header: StageServe identity plus Doctor, Setup, Project, Recovery, or another current surface.",
			"Verdict line: the first human sentence about the current state.",
			"Key facts: the values StageServe will use before the user commits.",
			"Focus section: Needs fixing, Setup steps, Recovery steps, What you can do, or another dominant section.",
			"Secondary surface: More…, details, advanced troubleshooting, logs, or confirmation.",
			"Footer help: only the keys that matter right now.",
		})
		renderRule(builder, width, "Live anatomy sketch", styleNeedsAction)
		renderScreenHeader(builder, width, "StageServe", "Project")
		fmt.Fprintf(builder, "\n  %s\n", styleWhite.Render("This project is ready to run."))
		renderRule(builder, width, "Key facts", styleReady)
		renderFacts(builder, []factRow{{"Local URL", "http://pete-site.develop", "visible before action"}, {"Web folder", "./public_html", "found here"}})
		renderRule(builder, width, "What you can do", styleReady)
		renderDecisionList(builder, []decisionRow{{"Run this project", "Start the project and open it in your browser."}, {"More…", "Show direct commands, plain text output, and advanced detail."}}, 0)
	})
}

func renderColourSystem(width int) string {
	return renderPage("StageServe", "Colour", "Colour carries semantic meaning only. It is not decoration.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Semantic palette", styleCyan)
		renderPaletteRow(builder, "Success / ready", "2", "styleReady", "ready verdicts, check marks, all-clear sections")
		renderPaletteRow(builder, "Warning / action", "3", "styleNeedsAction", "issue numbers, setup attention, warning section titles")
		renderPaletteRow(builder, "Error", "1", "styleError", "blocking verdicts, unsafe states, error issue numbers")
		renderPaletteRow(builder, "Primary command", "14", "styleBrightCyan", "exact commands and command-like remediation")
		renderPaletteRow(builder, "Supporting accent", "6", "styleCyan", "header glyph, compact arrows, secondary command references")
		renderPaletteRow(builder, "Primary structure", "15", "styleWhite", "titles and key issue labels")
		renderPaletteRow(builder, "Supporting text", "7", "styleMuted", "descriptions and secondary copy")
		renderPaletteRow(builder, "De-emphasised", "8", "styleDim", "dividers, quiet evidence, compact status words")
		renderRule(builder, width, "Rule", styleError)
		renderParagraph(builder, width, "Do not add a new colour because a screen feels plain. Add a colour only when the user needs a distinct semantic state, then document its role here and in the markup spec.")
	})
}

func renderComponents(width int) string {
	return renderPage("StageServe", "Components", "The style guide is built from a small set of reusable terminal components.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Component vocabulary", styleCyan)
		renderTermRows(builder, []termRow{
			{"Surface header", "Establishes StageServe identity and current surface."},
			{"Verdict line", "States the human outcome before detail."},
			{"Key facts", "Shows values StageServe will use before commitment."},
			{"Report section", "Presents evidence, blockers first, with remediation adjacent."},
			{"Work checklist", "Shows tool-owned setup or recovery with one active step."},
			{"Decision list", "Offers real user choices with the safe default first."},
			{"Assistance handoff", "Moves from useful report to optional guided help."},
			{"Confirmation sheet", "Names what changes, what does not, and the target."},
			{"More panel", "Holds direct commands, plain text output, and advanced detail."},
			{"Footer help", "Shows only keys that matter on this screen."},
		})
		renderRule(builder, width, "Mini components", styleReady)
		renderWorkChecklist(builder)
		fmt.Fprintln(builder)
		renderConfirmation(builder)
	})
}

func renderReportSurface(width int) string {
	return renderPage("StageServe Doctor", "Report", "Not ready - 2 of 7 checks need attention.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Needs fixing", styleNeedsAction)
		renderIssue(builder, 1, "Port 443", "Port 443 must be free for the local HTTPS gateway to bind to it.", "Something else on your computer is using port 443.", "sudo lsof -nP -iTCP:443 -sTCP:LISTEN")
		renderIssue(builder, 2, "Local DNS resolver", "Routes local project addresses to this computer.", "Your computer cannot yet open local project URLs.", "stage setup")
		renderRule(builder, width, "All clear", styleReady)
		renderPassedRows(builder, []factRow{{"Docker Desktop", "running", ""}, {"State directory", "exists", ""}, {"mkcert local CA", "installed", ""}})
		fmt.Fprintf(builder, "\n  %s\n", styleDim.Render(strings.Repeat("─", visibleLineWidth(width))))
		fmt.Fprintf(builder, "  Fix the issues above, then run: %s\n", styleBrightCyan.Render("stage doctor"))
	})
}

func renderGuidedSurface(width int) string {
	return renderPage("StageServe", "Project", "This project is ready to run.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Key facts", styleReady)
		renderFacts(builder, []factRow{
			{"Local URL", "http://pete-site.develop", "what you will visit"},
			{"Web folder", "./public_html", "found here"},
			{"Status", "not running yet", "safe to start"},
		})
		renderRule(builder, width, "What you can do", styleReady)
		renderDecisionList(builder, []decisionRow{
			{"Run this project", "Start the project and open it in your browser."},
			{"Edit project settings", "Change site name, web folder, or domain suffix first."},
			{"More…", "Show direct commands, plain text output, and advanced detail."},
		}, 0)
		fmt.Fprintf(builder, "\n  %s\n", styleDim.Render("enter run • ? help • q quit"))
	})
}

func renderAssistance(width int) string {
	return renderPage("StageServe Doctor", "Assistance", "A report may offer guided help only after the report is useful on its own.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Assistance", styleNeedsAction)
		renderParagraph(builder, width, "StageServe can help with the issues above.")
		renderDecisionList(builder, []decisionRow{{"Help me fix these", "Walk through each issue one at a time."}, {"Leave it here", "Exit without changing anything."}}, 0)
		renderRule(builder, width, "One blocker at a time", styleCyan)
		renderScreenHeader(builder, width, "StageServe", "Port 443")
		fmt.Fprintf(builder, "\n  %s\n\n", styleWhite.Render("Something else on your computer is using port 443."))
		renderParagraph(builder, width, "StageServe can check which process owns the port. Your computer may ask for your password because macOS hides this detail by default.")
		renderDecisionList(builder, []decisionRow{{"Check with sudo", "Run a read-only command to identify the process."}, {"Skip this issue", "Leave port 443 unresolved for now."}}, 0)
	})
}

func renderSeparation(width int) string {
	return renderPage("StageServe", "Boundaries", "Design coding and functional implementation meet through a screen-plan contract.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Separation of concerns", styleCyan)
		renderTermRows(builder, []termRow{
			{"Planner/domain", "Owns situations, effective values, blockers, and action availability. Not colours or key bindings."},
			{"Screen plan", "Owns normalized facts, sections, actions, severity, and copy slots. Not lifecycle execution."},
			{"Bubble Tea model", "Owns focus, selected action, modal state, drafts, and in-flight UI state. Not project truth."},
			{"Renderer/style", "Owns components, spacing, glyphs, Lip Gloss tokens, and text/TUI parity. Not command behavior."},
			{"Action layer", "Owns setup, lifecycle, logs, status, and recovery execution. Not visual hierarchy."},
		})
		renderRule(builder, width, "Design rule", styleReady)
		renderParagraph(builder, width, "Bubble Tea and Lip Gloss are implementation tools for the style guide. They should express the identity; they should not decide whether a project is running, out of sync, or ready to recover.")
	})
}

func renderReview(width int) string {
	return renderPage("StageServe", "Review", "Review the identity, then the interaction, then the renderer.", width, func(builder *strings.Builder) {
		renderRule(builder, width, "Checklist", styleReady)
		renderBullets(builder, []string{
			"The screen looks like the same StageServe identity as stage doctor.",
			"The first scan reveals context, verdict, priority, and next action.",
			"Colour, glyph, weight, and spacing all have jobs.",
			"The default action and default values are visible before commitment.",
			"Report surfaces remain useful before assisted flow begins.",
			"Guided surfaces are action-first and safe-by-default.",
			"Advanced implementation detail is secondary.",
			"Plain text output preserves the same truth.",
		})
		renderRule(builder, width, "Do not approve", styleError)
		renderBullets(builder, []string{
			"Decorative colour, banners, novelty glyphs, or ornamental boxes.",
			"Bubble Tea view logic deciding lifecycle or config truth.",
			"Docker, compose, gateway, registry, runtime, attach, or detach as first-level copy without a clearer user phrase first.",
			"A prototype or renderer helper treated as the design authority.",
		})
	})
}
