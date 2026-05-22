package guidance

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RunShell starts the first guided TUI shell. It renders the existing plan and
// supports navigation/details/quit; action execution is wired in later phases.
func RunShell(plan NextActionPlan, capability TUICapability, input io.Reader, output io.Writer) error {
	model := newShellModel(plan, capability.NoColor)
	program := tea.NewProgram(model, tea.WithInput(input), tea.WithOutput(output))
	_, err := program.Run()
	return err
}

type shellModel struct {
	plan        NextActionPlan
	cursor      int
	showDetails bool
	width       int
	noColor     bool
}

func newShellModel(plan NextActionPlan, noColor bool) shellModel {
	return shellModel{plan: plan, width: 80, noColor: noColor}
}

func (m shellModel) Init() tea.Cmd { return nil }

func (m shellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			m.width = msg.Width
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.plan.DecisionItems)-1 {
				m.cursor++
			}
		case "?":
			m.showDetails = !m.showDetails
		}
	}
	return m, nil
}

func (m shellModel) View() string {
	return renderShellView(m.plan, m.width, m.cursor, m.showDetails, m.noColor)
}

func renderShellView(plan NextActionPlan, width, cursor int, showDetails, noColor bool) string {
	styles := shellStylesFor(noColor)
	var b strings.Builder
	lineWidth := clampInt(width-4, 42, 96)

	fmt.Fprintf(&b, "\n  %s  %s\n", styles.accent.Render("◆"), styles.title.Render("StageServe"))
	fmt.Fprintf(&b, "  %s\n\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	fmt.Fprintf(&b, "  %s\n", styles.verdict.Render(plan.StatusHeader))
	if plan.Summary != "" {
		fmt.Fprintf(&b, "  %s\n", styles.muted.Render(plan.Summary))
	}

	if len(plan.VisibleDefaults) > 0 {
		fmt.Fprintf(&b, "\n%s\n\n", sectionTitle("Key facts", lineWidth, styles))
		for _, item := range plan.VisibleDefaults {
			fmt.Fprintf(&b, "  %s  %s", styles.label.Render(fmt.Sprintf("%-13s", item.Label)), item.Value)
			if item.Note != "" {
				fmt.Fprintf(&b, "  %s", styles.muted.Render(item.Note))
			}
			b.WriteByte('\n')
		}
	}

	if len(plan.DecisionItems) > 0 {
		fmt.Fprintf(&b, "\n%s\n\n", sectionTitle("What you can do", lineWidth, styles))
		for i, item := range plan.DecisionItems {
			marker := " "
			if i == cursor {
				marker = "▶"
			}
			fmt.Fprintf(&b, "  %s %s\n", styles.accent.Render(marker), styles.label.Render(item.Label))
			fmt.Fprintf(&b, "    %s\n\n", styles.muted.Render(item.Description))
		}
	}

	if len(plan.WorkItems) > 0 {
		fmt.Fprintf(&b, "\n%s\n\n", sectionTitle("Next step", lineWidth, styles))
		for _, item := range plan.WorkItems {
			fmt.Fprintf(&b, "  %s\n", styles.label.Render(item.Label))
			if item.Description != "" {
				fmt.Fprintf(&b, "    %s\n", styles.muted.Render(item.Description))
			}
			if item.DirectCommand != "" {
				fmt.Fprintf(&b, "    Direct command: %s\n", styles.command.Render(item.DirectCommand))
			}
		}
	}

	if showDetails || len(plan.DecisionItems) == 0 {
		fmt.Fprintf(&b, "\n%s\n\n", sectionTitle("Details", lineWidth, styles))
		if len(plan.Warnings) == 0 {
			fmt.Fprintf(&b, "  %s\n", styles.muted.Render("No extra warnings for this plan."))
		}
		for _, warning := range plan.Warnings {
			fmt.Fprintf(&b, "  %s\n", warning)
		}
		if len(plan.DirectCommands) > 0 {
			b.WriteByte('\n')
			for _, command := range plan.DirectCommands {
				fmt.Fprintf(&b, "  %s\n", styles.command.Render(command))
			}
		}
	}

	fmt.Fprintf(&b, "\n  %s\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	fmt.Fprintf(&b, "  %s\n\n", styles.footer.Render("↑/↓ inspect • ? details • q quit"))
	return b.String()
}

type shellStyles struct {
	accent  lipgloss.Style
	title   lipgloss.Style
	rule    lipgloss.Style
	verdict lipgloss.Style
	label   lipgloss.Style
	muted   lipgloss.Style
	command lipgloss.Style
	footer  lipgloss.Style
}

func shellStylesFor(noColor bool) shellStyles {
	if noColor {
		return shellStyles{
			accent:  lipgloss.NewStyle(),
			title:   lipgloss.NewStyle(),
			rule:    lipgloss.NewStyle(),
			verdict: lipgloss.NewStyle(),
			label:   lipgloss.NewStyle(),
			muted:   lipgloss.NewStyle(),
			command: lipgloss.NewStyle(),
			footer:  lipgloss.NewStyle(),
		}
	}
	styles := shellStyles{
		accent:  lipgloss.NewStyle(),
		title:   lipgloss.NewStyle().Bold(true),
		rule:    lipgloss.NewStyle(),
		verdict: lipgloss.NewStyle().Bold(true),
		label:   lipgloss.NewStyle().Bold(true),
		muted:   lipgloss.NewStyle(),
		command: lipgloss.NewStyle().Bold(true),
		footer:  lipgloss.NewStyle(),
	}
	styles.accent = styles.accent.Foreground(lipgloss.Color("6"))
	styles.title = styles.title.Foreground(lipgloss.Color("15"))
	styles.rule = styles.rule.Foreground(lipgloss.Color("8"))
	styles.verdict = styles.verdict.Foreground(lipgloss.Color("15"))
	styles.label = styles.label.Foreground(lipgloss.Color("15"))
	styles.muted = styles.muted.Foreground(lipgloss.Color("7"))
	styles.command = styles.command.Foreground(lipgloss.Color("14"))
	styles.footer = styles.footer.Foreground(lipgloss.Color("8"))
	return styles
}

func sectionTitle(title string, width int, styles shellStyles) string {
	fill := maximumInt(0, width-lipgloss.Width(title)-5)
	return "  " + styles.rule.Render("── ") + styles.title.Render(title) + styles.rule.Render(" "+strings.Repeat("─", fill))
}

func clampInt(value, minimum, maximum int) int {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

func maximumInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
