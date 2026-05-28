package guidance

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/peternicholls/stageserve/core/project"
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

type ActionHandler func(GuidedAction) (ActionResult, error)

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
	utility       *UtilitySurface
	message       string
	editing       bool
	editCursor    int
	editDraft     projectSettingsDraft
	hasEditDraft  bool
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
		if m.editing {
			return m.updateEditing(msg)
		}
		if m.utility != nil {
			return m.updateUtility(msg)
		}
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
			if action.ID == "edit_config" {
				return m.startEditing(), nil
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
	return renderShellViewState(m.plan, m.width, m.cursor, m.showDetails, m.noColor, m.message, m.confirming, m.editing, m.editCursor, m.editDraft, m.utility)
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
		if action.ID == "stop_here" {
			return m, tea.Quit
		}
		m.confirming = false
		return m.runAction(action)
	}
	return m, nil
}

func (m shellModel) updateUtility(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	if message.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.utility != nil && (m.utility.DismissOnAnyKey || message.String() == "q" || message.String() == "esc" || message.String() == "enter") {
		m.utility = nil
		return m, nil
	}
	return m, nil
}

func (m shellModel) runAction(action GuidedAction) (tea.Model, tea.Cmd) {
	if m.actionHandler == nil {
		m.message = "This action is not available in guided mode yet."
		return m, nil
	}
	action = m.actionWithDraft(action)
	result, err := m.actionHandler(action)
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	m.plan = result.Plan
	m.cursor = 0
	m.showDetails = false
	m.editing = false
	m.hasEditDraft = false
	m.utility = result.Utility
	m.message = result.Message
	return m, nil
}

func (m shellModel) startEditing() shellModel {
	m.editing = true
	m.confirming = false
	m.showDetails = false
	m.message = ""
	m.editCursor = 0
	m.editDraft = m.activeProjectSettingsDraft()
	return m
}

func (m shellModel) updateEditing(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.editing = false
		m.message = "No changes made."
		return m, nil
	case "tab", "down":
		m.editCursor = nextProjectSettingsField(m.editCursor)
		return m, nil
	case "shift+tab", "up":
		m.editCursor = previousProjectSettingsField(m.editCursor)
		return m, nil
	case "enter":
		return m.applyEditDraft(), nil
	case "backspace", "ctrl+h":
		m.editDraft.deleteLast(m.editCursor)
		return m, nil
	case "ctrl+u":
		m.editDraft.setValue(m.editCursor, "")
		return m, nil
	}
	if len(message.Runes) > 0 {
		m.editDraft.appendText(m.editCursor, string(message.Runes))
	}
	return m, nil
}

func (m shellModel) applyEditDraft() shellModel {
	m.editDraft = m.editDraft.normalized()
	m.plan = applyProjectSettingsDraft(m.plan, m.editDraft)
	m.hasEditDraft = true
	m.editing = false
	m.cursor = indexOfAction(m.plan.DecisionItems, "init")
	m.message = "Settings preview updated. Nothing has been written yet."
	return m
}

func (m shellModel) actionWithDraft(action GuidedAction) GuidedAction {
	if action.ID != "init" && action.ID != "init_here" {
		return action
	}
	draft := m.activeProjectSettingsDraft().normalized()
	action.Inputs = map[string]string{
		"site_name":   draft.SiteName,
		"docroot":     docRootInputValue(draft.WebFolder),
		"site_suffix": draft.SiteSuffix,
	}
	return action
}

func (m shellModel) activeProjectSettingsDraft() projectSettingsDraft {
	if m.hasEditDraft {
		return m.editDraft
	}
	return projectSettingsDraftFromPlan(m.plan)
}

const projectSettingsFieldCount = 3

type projectSettingsDraft struct {
	SiteName   string
	WebFolder  string
	SiteSuffix string
}

func projectSettingsDraftFromPlan(plan NextActionPlan) projectSettingsDraft {
	draft := projectSettingsDraft{SiteSuffix: "test"}
	for _, item := range plan.VisibleDefaults {
		switch item.Label {
		case "Project", "Site name":
			draft.SiteName = item.Value
		case "Web folder":
			draft.WebFolder = item.Value
		case "Domain suffix":
			draft.SiteSuffix = strings.TrimPrefix(item.Value, ".")
		}
	}
	return draft.normalized()
}

func (draft projectSettingsDraft) normalized() projectSettingsDraft {
	draft.SiteName = strings.TrimSpace(draft.SiteName)
	if draft.SiteName == "" {
		draft.SiteName = "site"
	}
	draft.WebFolder = strings.TrimSpace(draft.WebFolder)
	if draft.WebFolder == "" {
		draft.WebFolder = "."
	}
	draft.SiteSuffix = normalizeDraftSuffix(draft.SiteSuffix)
	return draft
}

func (draft projectSettingsDraft) localURL() string {
	draft = draft.normalized()
	scheme := "http"
	if draft.SiteSuffix == "dev" {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s.%s", scheme, project.Slugify(draft.SiteName), draft.SiteSuffix)
}

func (draft *projectSettingsDraft) appendText(field int, value string) {
	draft.setValue(field, draft.value(field)+value)
}

func (draft *projectSettingsDraft) deleteLast(field int) {
	value := []rune(draft.value(field))
	if len(value) == 0 {
		return
	}
	draft.setValue(field, string(value[:len(value)-1]))
}

func (draft *projectSettingsDraft) setValue(field int, value string) {
	switch field {
	case 0:
		draft.SiteName = value
	case 1:
		draft.WebFolder = value
	case 2:
		draft.SiteSuffix = value
	}
}

func (draft projectSettingsDraft) value(field int) string {
	switch field {
	case 0:
		return draft.SiteName
	case 1:
		return draft.WebFolder
	case 2:
		return draft.SiteSuffix
	default:
		return ""
	}
}

func normalizeDraftSuffix(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, ".")
	if value == "" {
		return "test"
	}
	return project.Slugify(value)
}

func nextProjectSettingsField(current int) int {
	return (current + 1) % projectSettingsFieldCount
}

func previousProjectSettingsField(current int) int {
	if current <= 0 {
		return projectSettingsFieldCount - 1
	}
	return current - 1
}

func applyProjectSettingsDraft(plan NextActionPlan, draft projectSettingsDraft) NextActionPlan {
	draft = draft.normalized()
	updated := []VisibleDefault{
		{Label: "Site name", Value: draft.SiteName},
		{Label: "Web folder", Value: draft.WebFolder},
		{Label: "Domain suffix", Value: displaySuffix(draft.SiteSuffix)},
		{Label: "Local URL", Value: draft.localURL()},
	}
	for _, item := range plan.VisibleDefaults {
		switch item.Label {
		case "Project", "Site name", "Web folder", "Domain suffix", "Local URL":
			continue
		default:
			updated = append(updated, item)
		}
	}
	plan.VisibleDefaults = updated
	return plan
}

func renderProjectSettingsEditor(builder *strings.Builder, draft projectSettingsDraft, editCursor, lineWidth int, styles shellStyles) {
	draft = draft.normalized()
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle("Project settings", lineWidth, styles))
	fields := []struct {
		label string
		value string
	}{
		{label: "Site name", value: draft.SiteName},
		{label: "Web folder", value: draft.WebFolder},
		{label: "Domain suffix", value: draft.SiteSuffix},
	}
	for fieldIndex, field := range fields {
		marker := " "
		if fieldIndex == editCursor {
			marker = "▶"
		}
		fmt.Fprintf(builder, "  %s %s  %s\n", styles.accent.Render(marker), styles.label.Render(fmt.Sprintf("%-13s", field.label)), field.value)
	}
	fmt.Fprintf(builder, "\n  %s  %s\n", styles.label.Render(fmt.Sprintf("%-13s", "Local URL")), draft.localURL())
	fmt.Fprintf(builder, "  %s\n", styles.muted.Render("Saving updates the preview only; confirm project settings to write the file."))
}

func docRootInputValue(webFolder string) string {
	webFolder = strings.TrimSpace(webFolder)
	if webFolder == "." {
		return ""
	}
	return webFolder
}

func indexOfAction(actions []GuidedAction, id string) int {
	for actionIndex, action := range actions {
		if action.ID == id {
			return actionIndex
		}
	}
	return 0
}

func renderShellView(plan NextActionPlan, width, cursor int, showDetails, noColor bool) string {
	return renderShellViewState(plan, width, cursor, showDetails, noColor, "", false, false, 0, projectSettingsDraft{}, nil)
}

func renderShellViewState(plan NextActionPlan, width, cursor int, showDetails, noColor bool, message string, confirming bool, editing bool, editCursor int, editDraft projectSettingsDraft, utility *UtilitySurface) string {
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
	if utility != nil {
		renderUtilitySurface(&b, *utility, lineWidth, styles)
		fmt.Fprintf(&b, "\n  %s\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
		fmt.Fprintf(&b, "  %s\n\n", styles.footer.Render(utilityFooter(*utility)))
		return b.String()
	}

	if editing {
		renderProjectSettingsEditor(&b, editDraft, editCursor, lineWidth, styles)
	} else {
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
				renderConfirmationBody(&b, action, plan, styles)
			}
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
			fmt.Fprintf(&b, "\n%s\n\n", sectionTitle(workSectionTitle(plan), lineWidth, styles))
			for _, item := range plan.WorkItems {
				marker := workItemMarker(item.Status)
				fmt.Fprintf(&b, "  %s %s", styles.accent.Render(marker), styles.label.Render(item.Label))
				if item.Status != "" {
					fmt.Fprintf(&b, "  %s", styles.muted.Render(item.Status))
				}
				b.WriteByte('\n')
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
	}

	fmt.Fprintf(&b, "\n  %s\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	footer := "↑/↓ inspect • enter choose • ? details • q quit"
	if confirming {
		footer = "enter confirm • n cancel • esc cancel"
	} else if editing {
		footer = "type edit • tab/↑/↓ move • enter save • esc cancel"
	}
	fmt.Fprintf(&b, "  %s\n\n", styles.footer.Render(footer))
	return b.String()
}

func workSectionTitle(plan NextActionPlan) string {
	switch plan.Situation {
	case SituationMachineNotReady:
		return "Setup steps"
	case SituationDriftDetected, SituationUnknownError:
		return "Recovery steps"
	default:
		return "Next step"
	}
}

func workItemMarker(status string) string {
	switch status {
	case "ready":
		return "✓"
	case "next", "needs attention", "error":
		return "▶"
	default:
		return "•"
	}
}

func renderConfirmationBody(builder *strings.Builder, action GuidedAction, plan NextActionPlan, styles shellStyles) {
	builder.WriteByte('\n')
	localURL := visibleDefaultValue(plan, "Local URL")
	switch action.ID {
	case "init", "init_here":
		builder.WriteString("  StageServe will update the settings file shown above.\n")
		builder.WriteString("  It will not start containers or change your application files.\n")
	case "down":
		builder.WriteString("  StageServe will stop this project.\n")
		builder.WriteString("  Your files will not be touched.\n")
		if localURL != "" {
			fmt.Fprintf(builder, "  %s will no longer respond until you run it again.\n", localURL)
		}
	case "detach":
		builder.WriteString("  StageServe will remove this project from StageServe.\n")
		builder.WriteString("  .env.stageserve and your application files will stay as they are.\n")
		if localURL != "" {
			fmt.Fprintf(builder, "  %s will no longer be routed by StageServe.\n", localURL)
		}
	case "stop_here":
		builder.WriteString("  StageServe will leave this project as it is.\n")
		builder.WriteString("  No file or runtime change will be made.\n")
	default:
		fmt.Fprintf(builder, "  %s\n", styles.muted.Render("StageServe will apply the selected change after confirmation."))
	}
}

func visibleDefaultValue(plan NextActionPlan, label string) string {
	for _, item := range plan.VisibleDefaults {
		if item.Label == label {
			return item.Value
		}
	}
	return ""
}

func renderUtilitySurface(builder *strings.Builder, utility UtilitySurface, lineWidth int, styles shellStyles) {
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle(utility.Title, lineWidth, styles))
	for _, line := range strings.Split(strings.TrimSuffix(utility.Body, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			builder.WriteByte('\n')
			continue
		}
		fmt.Fprintf(builder, "  %s\n", line)
	}
}

func utilityFooter(utility UtilitySurface) string {
	if utility.Footer != "" {
		return utility.Footer
	}
	if utility.DismissOnAnyKey {
		return "press any key to return • q quit"
	}
	return "q return • esc return"
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
