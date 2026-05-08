package main

import (
	"fmt"
	"io"
	"strings"
)

func sectionHeader(title string) string {
	const total = 38
	fill := total - 3 - len(title) - 1
	if fill < 2 {
		fill = 2
	}
	return dim("── ") + bold(title) + dim(" "+strings.Repeat("─", fill))
}

func renderScreenHeader(b *strings.Builder, title, surface string) {
	if surface == "" {
		fmt.Fprintf(b, "  %s  %s\n", cyan("◆"), bold(title))
	} else {
		fmt.Fprintf(b, "  %s  %-32s %s\n", cyan("◆"), bold(title), dim(surface))
	}
	fmt.Fprintf(b, "  %s\n", dim(strings.Repeat("─", 38)))
}

func renderVerdict(b *strings.Builder, p plan) {
	if strings.TrimSpace(p.StatusHeader) == "" {
		return
	}
	fmt.Fprintf(b, "\n  %s  %s\n", verdictIcon(p), p.StatusHeader)
}

func renderActionList(b *strings.Builder, title string, decisions []decisionItem, cursor int) {
	if len(decisions) == 0 {
		return
	}
	if title != "" {
		fmt.Fprintf(b, "\n%s\n\n", sectionHeader(title))
	} else {
		fmt.Fprintln(b)
	}
	for i, item := range decisions {
		prefix := " "
		if i == cursor {
			prefix = yellow("▶")
		}
		fmt.Fprintf(b, "  %s %s\n", prefix, item.Label)
		if item.Description != "" {
			fmt.Fprintf(b, "    %s\n", item.Description)
		}
		if i < len(decisions)-1 {
			fmt.Fprintln(b)
		}
	}
}

func renderFooterHelp(b *strings.Builder, text string) {
	if text == "" {
		return
	}
	fmt.Fprintf(b, "\n  %s\n", dim(text))
}

func renderReportSections(b *strings.Builder, attention, ready []reportItem) {
	if len(attention) > 0 {
		fmt.Fprintf(b, "\n%s\n\n", sectionHeader("Needs fixing"))
		for i, item := range attention {
			fmt.Fprintf(b, "\n  %s  %s\n", yellow(fmt.Sprintf("%d", i+1)), bold(item.Label))
			if item.Description != "" {
				fmt.Fprintf(b, "     %s\n", item.Description)
			}
			if item.Message != "" {
				fmt.Fprintf(b, "\n     %s\n", dim(item.Message))
			}
			if item.Command != "" {
				fmt.Fprintf(b, "     %s %s\n", bold("To fix:"), cyan(item.Command))
			}
		}
	}
	if len(ready) > 0 {
		fmt.Fprintf(b, "\n%s\n\n", sectionHeader("All clear"))
		for _, item := range ready {
			fmt.Fprintf(b, "  %s  %-18s %s\n", green("✓"), item.Label, dim(reportReadySummary(item)))
		}
	}
}

func reportReadySummary(item reportItem) string {
	if item.Message != "" {
		return item.Message
	}
	return string(item.Status)
}

func (m model) renderMain() string {
	p := m.currentPlan()
	var b strings.Builder
	renderScreenHeader(&b, "StageServe", surfaceForPlan(p))
	if p.Context != "" {
		fmt.Fprintf(&b, "\n  %s\n", dim(p.Context))
	}
	renderVerdict(&b, p)
	if p.Summary != "" {
		fmt.Fprintf(&b, "\n  %s\n", p.Summary)
	}
	renderReportSections(&b, p.ReportAttention, p.ReportReady)
	renderDefaultFacts(&b, "Key facts", p.Defaults)
	renderWorkPanel(&b, p)
	if p.AssistanceTitle != "" && len(p.Decisions) > 0 {
		fmt.Fprintf(&b, "\n%s\n", sectionHeader(p.AssistanceTitle))
		renderActionList(&b, "", p.Decisions, m.cursor)
	} else {
		renderActionList(&b, "What you can do", p.Decisions, m.cursor)
	}
	if m.resultTitle != "" {
		fmt.Fprintf(&b, "\n%s\n\n  %s\n", sectionHeader("Recent step"), m.resultTitle)
		if m.resultBody != "" {
			fmt.Fprintf(&b, "  %s\n", m.resultBody)
		}
	}
	renderFooterHelp(&b, footerText(p))
	return b.String()
}

func surfaceForPlan(p plan) string {
	if p.Surface != "" {
		return p.Surface
	}
	switch p.Situation {
	case machineNotReady:
		return "Setup"
	case projectMissingConfig, notProject:
		return "Project setup"
	case projectReadyToRun, projectRunning, projectDown:
		return "Project"
	case driftDetected, unknownError:
		return "Recovery"
	case doctorReportNeedsHelp:
		return "Doctor"
	default:
		return ""
	}
}

func footerText(p plan) string {
	parts := []string{"↑/↓ navigate", "enter use highlighted"}
	parts = append(parts, p.Footer...)
	parts = append(parts, "tab next scenario")
	return strings.Join(parts, " • ")
}

func renderDefaultFacts(b *strings.Builder, title string, defaults []defaultValue) {
	if len(defaults) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s\n\n", sectionHeader(title))
	for _, item := range defaults {
		note := ""
		if item.Note != "" {
			note = "  " + dim("("+item.Note+")")
		}
		fmt.Fprintf(b, "  %-16s %-34s%s\n", item.Label, item.Value, note)
	}
}

func renderWorkPanel(b *strings.Builder, p plan) {
	if len(p.WorkItems) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s\n\n", sectionHeader(workSectionTitle(p)))
	for i, item := range p.WorkItems {
		marker := dim("•")
		switch {
		case i == p.ActiveWorkIndex:
			marker = yellow("▶")
		case item.Status == statusReady:
			marker = green("✓")
		}
		fmt.Fprintf(b, "  %s  %-34s %s\n", marker, item.Label, dim(string(item.Status)))
		if i == p.ActiveWorkIndex {
			if item.Description != "" {
				fmt.Fprintf(b, "     %s\n", item.Description)
			}
			if item.EnterAction != "" {
				fmt.Fprintf(b, "     %s %s\n", bold("Enter:"), item.EnterAction)
			}
		}
		if i < len(p.WorkItems)-1 {
			fmt.Fprintln(b)
		}
	}
}

func workSectionTitle(p plan) string {
	switch p.Situation {
	case machineNotReady:
		return "Setup steps"
	case unknownError:
		return "Recovery steps"
	default:
		return "Steps"
	}
}

func (m model) renderConfirm() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold(m.pending.Title))
	for _, line := range m.pending.Body {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	yesPrefix := " "
	noPrefix := " "
	if m.confirmYes {
		yesPrefix = ">"
	} else {
		noPrefix = ">"
	}
	yes := m.pending.YesLabel
	no := m.pending.NoLabel
	if yes == "" {
		yes = "Yes"
	}
	if no == "" {
		no = "No"
	}
	fmt.Fprintf(&b, "\n%s %s    %s %s\n", yesPrefix, yes, noPrefix, no)
	fmt.Fprintf(&b, "\n%s\n", dim("←/→ choose • enter confirm • y yes • n no • esc cancel • q quit"))
	return b.String()
}

func (m model) renderDetails() string {
	p := m.currentPlan()
	var b strings.Builder
	title := p.DetailsTitle
	if title == "" {
		title = "What StageServe knows"
	}
	fmt.Fprintf(&b, "%s\n\n", bold(title))
	if len(p.Details) == 0 {
		fmt.Fprintf(&b, "StageServe has no extra detail for this prototype screen.\n")
	}
	for _, line := range p.Details {
		fmt.Fprintf(&b, "%s\n", line)
	}
	fmt.Fprintf(&b, "\n%s\n", dim("enter/esc/q back"))
	return b.String()
}

func (m model) renderCommands() string {
	p := m.currentPlan()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold("More…"))
	fmt.Fprintf(&b, "%s\n\n", sectionHeader("Show direct commands"))
	for _, cmd := range p.DirectCommands {
		fmt.Fprintf(&b, "  %s\n", cmd)
	}
	fmt.Fprintf(&b, "\n%s\n\n", sectionHeader("Plain text output"))
	fmt.Fprintf(&b, "  go run ./specs/007-harden-TUI-and-other-interactions/prototype --notui --scenario %s\n", p.Situation)
	if target := valueForDefault(p, "Target file"); target != "" {
		fmt.Fprintf(&b, "\n%s\n\n", sectionHeader("Paths"))
		fmt.Fprintf(&b, "  Project settings  %s\n", target)
	}
	fmt.Fprintf(&b, "\n%s\n\n", sectionHeader("Advanced and troubleshooting"))
	if len(p.Advanced) == 0 {
		fmt.Fprintf(&b, "  No advanced detail is needed for this prototype screen.\n")
	} else {
		for _, line := range p.Advanced {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}
	fmt.Fprintf(&b, "\n%s\n", dim("enter/esc/q back"))
	return b.String()
}

func (m model) renderAdvanced() string {
	p := m.currentPlan()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold("Advanced and troubleshooting"))
	if len(p.Advanced) == 0 {
		fmt.Fprintf(&b, "No advanced detail is needed for this prototype screen.\n")
	}
	for _, line := range p.Advanced {
		fmt.Fprintf(&b, "%s\n", line)
	}
	fmt.Fprintf(&b, "\n%s\n", dim("enter/esc/q back"))
	return b.String()
}

func (m model) renderLogs() string {
	p := m.currentPlan()
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold(valueForDefault(p, "Site name")+" logs"))
	fmt.Fprintf(&b, "10:42:13  GET /                  200  12ms\n")
	fmt.Fprintf(&b, "10:42:14  GET /admin             200  21ms\n")
	fmt.Fprintf(&b, "10:42:21  GET /favicon.ico       404  2ms\n")
	fmt.Fprintf(&b, "10:42:26  %s is still running at %s\n", p.Defaults[0].Value, valueForDefault(p, "Local URL"))
	fmt.Fprintf(&b, "\n%s\n", dim("q/esc exit logs"))
	return b.String()
}

func (m model) renderEdit() string {
	values := []defaultValue{
		{Label: "Site name", Value: m.editValues.SiteName, Note: "used in the local URL"},
		{Label: "Web folder", Value: m.editValues.WebFolder, Note: "relative to this project"},
		{Label: "Domain suffix", Value: m.editValues.Suffix, Note: "most people leave this"},
		{Label: "Local URL", Value: "http://" + m.editValues.SiteName + m.editValues.Suffix, Note: "preview only"},
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", bold("Edit project settings"))
	for i, item := range values {
		prefix := " "
		if i == m.editCursor && i < 3 {
			prefix = ">"
		}
		fmt.Fprintf(&b, "%s %-16s %-30s %s\n", prefix, item.Label, item.Value, dim("("+item.Note+")"))
	}
	fmt.Fprintf(&b, "\nThis prototype cycles sample values when you press enter. It does not write files.\n")
	fmt.Fprintf(&b, "\n%s\n", dim("↑/↓ field • enter cycle value • s save to preview • esc discard • q quit"))
	return b.String()
}

func (m model) renderAssist() string {
	var b strings.Builder
	renderScreenHeader(&b, "StageServe", "Port 443")
	renderVerdict(&b, plan{Situation: doctorReportNeedsHelp, StatusHeader: "Something else on your computer is using port 443."})
	fmt.Fprintf(&b, "\n  StageServe can check which process owns the port. Your computer\n")
	fmt.Fprintf(&b, "  will ask for your password because macOS hides this detail by default.\n")
	renderActionList(&b, "What you can do", []decisionItem{
		{Label: "Check with sudo", Description: "Run a read-only command to identify the process."},
		{Label: "Skip this issue", Description: "Leave port 443 unresolved for now."},
	}, 0)
	renderFooterHelp(&b, "enter check • s skip • esc back • q quit")
	return b.String()
}

func valueForDefault(p plan, label string) string {
	for _, item := range p.Defaults {
		if item.Label == label {
			return item.Value
		}
	}
	return ""
}

func renderText(w io.Writer, p plan) {
	surface := surfaceForPlan(p)
	if surface != "" {
		fmt.Fprintf(w, "StageServe %s\n\n", surface)
	} else {
		fmt.Fprintf(w, "StageServe\n\n")
	}
	if p.Context != "" {
		fmt.Fprintf(w, "%s\n", p.Context)
	}
	if p.StatusHeader != "" {
		if p.Context != "" {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "%s\n", p.StatusHeader)
	}
	if p.Summary != "" {
		fmt.Fprintf(w, "\n%s\n", p.Summary)
	}
	if len(p.ReportAttention) > 0 {
		fmt.Fprintf(w, "\nNeeds fixing\n")
		for i, item := range p.ReportAttention {
			fmt.Fprintf(w, "\n%d. %s\n", i+1, item.Label)
			if item.Description != "" {
				fmt.Fprintf(w, "   %s\n", item.Description)
			}
			if item.Message != "" {
				fmt.Fprintf(w, "   %s\n", item.Message)
			}
			if item.Command != "" {
				fmt.Fprintf(w, "   To fix: %s\n", item.Command)
			}
		}
	}
	if len(p.ReportReady) > 0 {
		fmt.Fprintf(w, "\nAll clear\n")
		for _, item := range p.ReportReady {
			fmt.Fprintf(w, "- %s: %s\n", item.Label, reportReadySummary(item))
		}
	}
	if len(p.Defaults) > 0 {
		fmt.Fprintf(w, "\nKey facts\n")
		for _, item := range p.Defaults {
			fmt.Fprintf(w, "  %s: %s", item.Label, item.Value)
			if item.Note != "" {
				fmt.Fprintf(w, " (%s)", item.Note)
			}
			fmt.Fprintln(w)
		}
	}
	if len(p.WorkItems) > 0 {
		fmt.Fprintf(w, "\n%s\n", workSectionTitle(p))
		for i, item := range p.WorkItems {
			marker := "-"
			if i == p.ActiveWorkIndex {
				marker = ">"
			}
			fmt.Fprintf(w, "%s %s: %s\n", marker, item.Label, item.Status)
			if i == p.ActiveWorkIndex {
				fmt.Fprintf(w, "  %s\n", item.Description)
				if item.EnterAction != "" {
					fmt.Fprintf(w, "  Enter: %s\n", item.EnterAction)
				}
			}
		}
	}
	if p.AssistanceTitle != "" && len(p.Decisions) > 0 {
		fmt.Fprintf(w, "\n%s\n", p.AssistanceTitle)
		for i, item := range p.Decisions {
			prefix := "-"
			if i == 0 {
				prefix = ">"
			}
			fmt.Fprintf(w, "\n%s %s\n", prefix, item.Label)
			if item.Description != "" {
				fmt.Fprintf(w, "  %s\n", item.Description)
			}
		}
	} else if len(p.Decisions) > 0 {
		fmt.Fprintf(w, "\nWhat you can do\n")
		for i, item := range p.Decisions {
			prefix := "-"
			if i == 0 {
				prefix = ">"
			}
			fmt.Fprintf(w, "%s %s\n", prefix, item.Label)
			fmt.Fprintln(w)
			if item.Description != "" {
				fmt.Fprintf(w, "  %s\n", item.Description)
			}
			if i < len(p.Decisions)-1 {
				fmt.Fprintln(w)
			}
		}
	}
	if p.StatusHeader != "" && false {
		// no-op placeholder to keep future parity changes localized
	}
	if footer := footerText(p); footer != "" {
		fmt.Fprintf(w, "\n%s\n", footer)
	}
	if false {
		for _, cmd := range p.DirectCommands {
			fmt.Fprint(w, cmd)
		}
	}
}

func verdictIcon(p plan) string {
	switch p.Situation {
	case projectReadyToRun, projectRunning, projectDown:
		return green("✓")
	case doctorReportNeedsHelp:
		return red("✗")
	default:
		return yellow("!")
	}
}

func bold(s string) string {
	return "\033[1m" + s + "\033[0m"
}

func dim(s string) string {
	return "\033[90m" + s + "\033[0m"
}

func cyan(s string) string {
	return "\033[36m" + s + "\033[0m"
}

func green(s string) string {
	return "\033[32m" + s + "\033[0m"
}

func yellow(s string) string {
	return "\033[33m" + s + "\033[0m"
}

func red(s string) string {
	return "\033[31m" + s + "\033[0m"
}
