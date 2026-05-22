package guidance

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RunShell starts the guided TUI shell. It renders the current plan and owns
// navigation, details, confirmation, and optional action dispatch.
func RunShell(plan NextActionPlan, capability TUICapability, input io.Reader, output io.Writer, opts ...ShellOption) error {
	config := shellConfig{}
	for _, opt := range opts {
		opt(&config)
	}
	model := newShellModel(plan, capability.NoColor)
	model.actionHandler = config.actionHandler
	program := tea.NewProgram(model, tea.WithInput(input), tea.WithOutput(output))
	_, err := program.Run()
	return err
}

type ActionHandler func(GuidedAction) (NextActionPlan, string, error)

type ShellOption func(*shellConfig)

type shellConfig struct {
	actionHandler ActionHandler
}

func WithActionHandler(handler ActionHandler) ShellOption {
	return func(config *shellConfig) {
		config.actionHandler = handler
	}
}

type shellModel struct {
	plan          NextActionPlan
	cursor        int
	showDetails   bool
	confirming    bool
	message       string
	width         int
	noColor       bool
	actionHandler ActionHandler
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
		if m.confirming {
			return m.updateConfirmation(msg.String())
		}
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
		case "enter":
			action, ok := m.selectedAction()
			if !ok {
				return m, nil
			}
			if action.RequiresConfirmation {
				m.confirming = true
				m.message = ""
				return m, nil
			}
			return m.runAction(action)
		}
	}
	return m, nil
}

func (m shellModel) View() string {
	return renderShellViewState(m.plan, m.width, m.cursor, m.showDetails, m.noColor, m.message, m.confirming)
}

func (m shellModel) selectedAction() (GuidedAction, bool) {
	if m.cursor < 0 || m.cursor >= len(m.plan.DecisionItems) {
		return GuidedAction{}, false
	}
	return m.plan.DecisionItems[m.cursor], true
}

func (m shellModel) updateConfirmation(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "n":
		m.confirming = false
		m.message = "No changes made."
		return m, nil
	case "enter", "y":
		action, ok := m.selectedAction()
		if !ok {
			m.confirming = false
			return m, nil
		}
		m.confirming = false
		return m.runAction(action)
	}
	return m, nil
}

func (m shellModel) runAction(action GuidedAction) (tea.Model, tea.Cmd) {
	if m.actionHandler == nil {
		m.message = "This action is not available in guided mode yet."
		return m, nil
	}
	nextPlan, message, err := m.actionHandler(action)
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	m.plan = nextPlan
	m.cursor = 0
	m.showDetails = false
	m.message = message
	return m, nil
}

func renderShellView(plan NextActionPlan, width, cursor int, showDetails, noColor bool) string {
	return renderShellViewState(plan, width, cursor, showDetails, noColor, "", false)
}

func renderShellViewState(plan NextActionPlan, width, cursor int, showDetails, noColor bool, message string, confirming bool) string {
	styles := shellStylesFor(noColor)
	var b strings.Builder
	lineWidth := clampInt(width-4, 42, 96)

	fmt.Fprintf(&b, "\n  %s  %s\n", styles.accent.Render("◆"), styles.title.Render("StageServe"))
	fmt.Fprintf(&b, "  %s\n\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	fmt.Fprintf(&b, "  %s\n", styles.verdict.Render(plan.StatusHeader))
	if plan.Summary != "" {
		fmt.Fprintf(&b, "  %s\n", styles.muted.Render(plan.Summary))
	}
	if message != "" {
		fmt.Fprintf(&b, "  %s\n", styles.muted.Render(message))
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

	if confirming {
		fmt.Fprintf(&b, "\n%s\n\n", sectionTitle("Confirm change", lineWidth, styles))
		if action, ok := selectedAction(plan, cursor); ok {
			fmt.Fprintf(&b, "  %s\n", styles.label.Render(action.Label))
			fmt.Fprintf(&b, "    %s\n", styles.muted.Render(action.Description))
		}
		b.WriteString("\n  StageServe will update the settings file shown above.\n")
		b.WriteString("  It will not start containers or change your application files.\n")
	} else if len(plan.DecisionItems) > 0 {
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
	footer := "↑/↓ inspect • enter choose • ? details • q quit"
	if confirming {
		footer = "enter confirm • n cancel • esc cancel"
	}
	fmt.Fprintf(&b, "  %s\n\n", styles.footer.Render(footer))
	return b.String()
}

func selectedAction(plan NextActionPlan, cursor int) (GuidedAction, bool) {
	if cursor < 0 || cursor >= len(plan.DecisionItems) {
		return GuidedAction{}, false
	}
	return plan.DecisionItems[cursor], true
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
