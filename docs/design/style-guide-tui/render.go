package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	styleReady       = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleNeedsAction = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleError       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	styleDim         = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleMuted       = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	styleBold        = lipgloss.NewStyle().Bold(true)
	styleCyan        = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	styleBrightCyan  = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	styleWhite       = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)

	appHeaderStyle   = lipgloss.NewStyle().Padding(0, 1)
	navStyle         = lipgloss.NewStyle().PaddingRight(1).Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(lipgloss.Color("8"))
	contentPaneStyle = lipgloss.NewStyle().PaddingLeft(1)
	footerStyle      = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("8"))
	helpStyle        = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))

	ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

type factRow struct {
	Label string
	Value string
	Note  string
}

type termRow struct {
	Term        string
	Description string
}

type decisionRow struct {
	Label       string
	Description string
}

func renderAppHeader(width int) string {
	title := fmt.Sprintf("%s  %s", styleCyan.Render("◆"), styleWhite.Render("StageServe Terminal Style Guide"))
	subtitle := styleDim.Render("graphic-design reference • Bubble Tea/Lip Gloss edition")
	line := styleDim.Render(strings.Repeat("─", maximum(24, width-2)))
	return appHeaderStyle.Width(width).Render(title + "\n" + subtitle + "\n" + line)
}

func renderFooter(width int) string {
	footerText := "↑/↓ section • tab next • j/k scroll • pgup/pgdn scroll • enter top • ? help • q quit"
	return footerStyle.Width(width).Render(footerText)
}

func renderNavigation(styleGuide guide, selected int, width int, height int) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s\n\n", styleWhite.Render("Sections"))
	for sectionIndex, section := range styleGuide.Sections {
		marker := " "
		labelStyle := styleMuted
		if sectionIndex == selected {
			marker = styleNeedsAction.Render("▶")
			labelStyle = styleWhite
		}
		line := truncateVisible(section.Title, maximum(8, width-4))
		fmt.Fprintf(&builder, "%s %s\n", marker, labelStyle.Render(line))
	}

	return navStyle.Width(width).Height(height).Render(builder.String())
}

func renderHelpSheet(width int) string {
	content := strings.Join([]string{
		styleWhite.Render("Guide controls"),
		"",
		"↑/↓ or tab moves between guide sections.",
		"j/k, page up, and page down scroll the current section.",
		"enter returns the current section to the top.",
		"? toggles this help sheet.",
		"q or esc exits.",
		"",
		styleDim.Render("The app is documentation. It does not run Docker, write files, or mutate StageServe state."),
	}, "\n")
	return helpStyle.Width(maximum(36, width-6)).Render(content)
}

func renderPage(title string, surface string, verdict string, width int, renderBody func(builder *strings.Builder)) string {
	var builder strings.Builder
	renderScreenHeader(&builder, width, title, surface)
	if verdict != "" {
		fmt.Fprintf(&builder, "\n  %s\n", styleWhite.Render(verdict))
	}
	if renderBody != nil {
		fmt.Fprintln(&builder)
		renderBody(&builder)
	}
	return builder.String()
}

func renderScreenHeader(builder *strings.Builder, width int, title string, surface string) {
	lineWidth := visibleLineWidth(width)
	if surface == "" {
		fmt.Fprintf(builder, "  %s  %s\n", styleCyan.Render("◆"), styleWhite.Render(title))
	} else {
		availableTitleWidth := maximum(12, lineWidth-lipgloss.Width(surface)-7)
		fmt.Fprintf(builder, "  %s  %-*s %s\n", styleCyan.Render("◆"), availableTitleWidth, styleWhite.Render(title), styleDim.Render(surface))
	}
	fmt.Fprintf(builder, "  %s\n", styleDim.Render(strings.Repeat("─", lineWidth)))
}

func renderRule(builder *strings.Builder, width int, title string, titleStyle lipgloss.Style) {
	lineWidth := visibleLineWidth(width)
	fill := lineWidth - 3 - lipgloss.Width(title) - 1
	if fill < 2 {
		fill = 2
	}
	fmt.Fprintf(builder, "\n%s\n\n", styleDim.Render("── ")+titleStyle.Bold(true).Render(title)+styleDim.Render(" "+strings.Repeat("─", fill)))
}

func renderParagraph(builder *strings.Builder, width int, text string) {
	paragraph := lipgloss.NewStyle().Width(maximum(36, visibleLineWidth(width))).Render(text)
	fmt.Fprintf(builder, "  %s\n", paragraph)
}

func renderBullets(builder *strings.Builder, items []string) {
	for _, item := range items {
		fmt.Fprintf(builder, "  %s  %s\n", styleDim.Render("•"), item)
	}
}

func renderNumbered(builder *strings.Builder, items []string) {
	for itemIndex, item := range items {
		fmt.Fprintf(builder, "  %s  %s\n", styleNeedsAction.Render(fmt.Sprintf("%d", itemIndex+1)), item)
	}
}

func renderTermRows(builder *strings.Builder, rows []termRow) {
	for _, row := range rows {
		fmt.Fprintf(builder, "  %-22s %s\n", styleWhite.Render(row.Term), row.Description)
	}
}

func renderFacts(builder *strings.Builder, rows []factRow) {
	for _, row := range rows {
		note := ""
		if row.Note != "" {
			note = "  " + styleDim.Render("("+row.Note+")")
		}
		fmt.Fprintf(builder, "  %-16s %-34s%s\n", row.Label, row.Value, note)
	}
}

func renderDecisionList(builder *strings.Builder, rows []decisionRow, selected int) {
	for rowIndex, row := range rows {
		marker := " "
		if rowIndex == selected {
			marker = styleNeedsAction.Render("▶")
		}
		fmt.Fprintf(builder, "  %s %s\n", marker, styleWhite.Render(row.Label))
		if row.Description != "" {
			fmt.Fprintf(builder, "    %s\n", row.Description)
		}
		if rowIndex < len(rows)-1 {
			fmt.Fprintln(builder)
		}
	}
}

func renderPaletteRow(builder *strings.Builder, role string, ansiCode string, token string, usage string) {
	swatch := lipgloss.NewStyle().Background(lipgloss.Color(ansiCode)).Foreground(lipgloss.Color("0")).Bold(true).Padding(0, 1).Render("ANSI " + ansiCode)
	fmt.Fprintf(builder, "  %-19s %-14s %-20s %s\n", role, swatch, token, usage)
}

func renderIssue(builder *strings.Builder, number int, label string, description string, message string, command string) {
	fmt.Fprintf(builder, "\n  %s  %s\n", styleNeedsAction.Render(fmt.Sprintf("%d", number)), styleWhite.Render(label))
	fmt.Fprintf(builder, "     %s\n", styleMuted.Render(description))
	fmt.Fprintf(builder, "\n     %s\n", styleDim.Render(message))
	fmt.Fprintf(builder, "     %s  %s\n", styleBold.Render("To fix:"), styleBrightCyan.Render(command))
}

func renderPassedRows(builder *strings.Builder, rows []factRow) {
	for _, row := range rows {
		fmt.Fprintf(builder, "  %s  %-18s %s\n", styleReady.Render("✓"), styleBold.Render(row.Label), styleDim.Render(row.Value))
	}
}

func renderWorkChecklist(builder *strings.Builder) {
	fmt.Fprintf(builder, "  %s  %-34s %s\n", styleReady.Render("✓"), "Docker Desktop", styleDim.Render("ready"))
	fmt.Fprintf(builder, "  %s  %-34s %s\n", styleNeedsAction.Render("▶"), "Local DNS for .develop", styleDim.Render("needs approval"))
	fmt.Fprintf(builder, "     %s\n", "StageServe will add a small file so your browser can open local project URLs.")
	fmt.Fprintf(builder, "     %s %s\n", styleBold.Render("Enter:"), "preview the change and confirm it.")
	fmt.Fprintf(builder, "  %s  %-34s %s\n", styleDim.Render("•"), "Local HTTPS certificates", styleDim.Render("optional"))
}

func renderConfirmation(builder *strings.Builder) {
	fmt.Fprintf(builder, "%s\n\n", styleWhite.Render("Stop pete-site?"))
	fmt.Fprintf(builder, "  StageServe will stop this project.\n")
	fmt.Fprintf(builder, "  Your files will not be touched.\n")
	fmt.Fprintf(builder, "  http://pete-site.develop will no longer respond.\n")
	fmt.Fprintf(builder, "\n  %s %s    %s\n", styleNeedsAction.Render("▶"), styleWhite.Render("No, keep it running"), "Yes, stop it")
}

func renderCallout(builder *strings.Builder, label string, value string) {
	fmt.Fprintf(builder, "\n  %s  %s\n", styleBold.Render(label+":"), styleBrightCyan.Render(value))
}

func visibleLineWidth(width int) int {
	return maximum(38, minimum(72, width-4))
}

func navigationWidth(width int) int {
	if width < 82 {
		return 20
	}
	return 26
}

func joinHorizontal(left string, right string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func lipglossHeight(value string) int {
	return lipgloss.Height(value)
}

func truncateVisible(value string, maxWidth int) string {
	if lipgloss.Width(value) <= maxWidth {
		return value
	}
	if maxWidth <= 1 {
		return ""
	}
	return value[:maximum(0, minimum(len(value), maxWidth-1))] + "…"
}

func stripGuideANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}

func maximum(first int, second int) int {
	if first > second {
		return first
	}
	return second
}

func minimum(first int, second int) int {
	if first < second {
		return first
	}
	return second
}
