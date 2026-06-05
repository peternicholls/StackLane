package guidance

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
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
	model := newShellModelWithConfig(plan, capability.NoColor, config)
	program := tea.NewProgram(model, tea.WithInput(input), tea.WithOutput(output))
	_, err := program.Run()
	return err
}

type ActionHandler func(GuidedAction) (ActionResult, error)

type ShellOption func(*shellConfig)

type shellConfig struct {
	actionHandler ActionHandler
	startEditing  bool
}

func WithActionHandler(handler ActionHandler) ShellOption {
	return func(config *shellConfig) {
		config.actionHandler = handler
	}
}

func WithInitialEditor() ShellOption {
	return func(config *shellConfig) {
		config.startEditing = true
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
	// async action execution
	loading bool
	spinner spinner.Model
}

// actionCompleteMsg is sent by the async action goroutine when it finishes.
type actionCompleteMsg struct {
	result ActionResult
	err    error
}

func newShellModel(plan NextActionPlan, noColor bool) shellModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	if !noColor {
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	}
	return shellModel{plan: plan, width: 80, noColor: noColor, spinner: s}
}

func newShellModelWithConfig(plan NextActionPlan, noColor bool, config shellConfig) shellModel {
	model := newShellModel(plan, noColor)
	model.actionHandler = config.actionHandler
	if config.startEditing {
		model = model.startEditing()
	}
	return model
}

func (m shellModel) Init() tea.Cmd { return nil }

func (m shellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case actionCompleteMsg:
		m.loading = false
		if msg.err != nil {
			m.message = msg.err.Error()
			return m, nil
		}
		m.plan = msg.result.Plan
		m.cursor = 0
		m.showDetails = false
		m.editing = false
		m.hasEditDraft = false
		m.utility = msg.result.Utility
		m.message = msg.result.Message
		return m, nil
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		if msg.Width > 0 {
			m.width = msg.Width
		}
	case tea.KeyMsg:
		if m.loading {
			// absorb all key input while an action is running
			return m, nil
		}
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
		case "c":
			m.utility = commandUtilitySurface(m.plan)
			m.showDetails = false
			return m, nil
		case "d":
			return m.runFooterAction(GuidedAction{ID: "doctor", Kind: "footer", Label: "Run diagnostics", Description: "Show the current StageServe diagnostics for this project."})
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
			action, ok := m.selectedInteractiveAction()
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
	if m.loading {
		return renderLoadingView(m.plan, m.width, m.noColor, m.spinner)
	}
	return renderShellViewState(m.plan, m.width, m.cursor, m.showDetails, m.noColor, m.message, m.confirming, m.editing, m.editCursor, m.editDraft, m.utility)
}

func (m shellModel) selectedAction() (GuidedAction, bool) {
	if m.cursor < 0 || m.cursor >= len(m.plan.DecisionItems) {
		return GuidedAction{}, false
	}
	return m.plan.DecisionItems[m.cursor], true
}

func (m shellModel) selectedInteractiveAction() (GuidedAction, bool) {
	if action, ok := m.selectedAction(); ok {
		return action, true
	}
	return setupAction(m.plan)
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
	m.loading = true
	m.message = ""
	handler := m.actionHandler
	return m, tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			result, err := handler(action)
			return actionCompleteMsg{result: result, err: err}
		},
	)
}

func (m shellModel) runFooterAction(action GuidedAction) (tea.Model, tea.Cmd) {
	if action.ID == "show_commands" {
		m.utility = commandUtilitySurface(m.plan)
		return m, nil
	}
	return m.runAction(action)
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
	m.cursor = indexOfAction(m.plan.DecisionItems, primaryProjectSettingsActionID(m.plan.DecisionItems))
	m.message = "Settings preview updated. Nothing has been written yet."
	return m
}

func (m shellModel) actionWithDraft(action GuidedAction) GuidedAction {
	if action.ID != "init" && action.ID != "init_here" && action.ID != "overwrite_init" {
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
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle("Project settings", lineWidth, sectionToneNeutral, styles))
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
		labelStyle := styles.label
		valueStyle := lipgloss.NewStyle()
		if fieldIndex == editCursor {
			marker = "▶"
			labelStyle = styles.focus
			valueStyle = styles.focus
		}
		fmt.Fprintf(builder, "  %s %s  %s\n", styles.accent.Render(marker), labelStyle.Render(fmt.Sprintf("%-13s", field.label)), valueStyle.Render(field.value))
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

func primaryProjectSettingsActionID(actions []GuidedAction) string {
	for _, action := range actions {
		if action.ID == "init" || action.ID == "overwrite_init" {
			return action.ID
		}
	}
	return "init"
}

func renderShellView(plan NextActionPlan, width, cursor int, showDetails, noColor bool) string {
	return renderShellViewState(plan, width, cursor, showDetails, noColor, "", false, false, 0, projectSettingsDraft{}, nil)
}

// renderLoadingView shows a spinner screen while an async action is running.
func renderLoadingView(plan NextActionPlan, width int, noColor bool, spin spinner.Model) string {
	styles := shellStylesFor(noColor)
	var b strings.Builder
	lineWidth := clampInt(width-4, 42, 96)
	renderSurfaceHeader(&b, lineWidth, surfaceLabel(plan, 0, false, false, nil), styles)
	fmt.Fprintf(&b, "  %s\n\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	fmt.Fprintf(&b, "  %s %s\n", spin.View(), styles.muted.Render("Working…"))
	fmt.Fprintf(&b, "\n  %s\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	fmt.Fprintf(&b, "  %s\n\n", styles.footer.Render("please wait"))
	return b.String()
}

// renderConfirmationModal renders a bordered confirmation panel for the given action.
func renderConfirmationModal(builder *strings.Builder, action GuidedAction, plan NextActionPlan, lineWidth int, styles shellStyles, noColor bool) {
	var inner strings.Builder
	fmt.Fprintf(&inner, "%s\n", styles.label.Render(action.Label))
	fmt.Fprintf(&inner, "%s\n", styles.muted.Render(action.Description))
	renderConfirmationBody(&inner, action, plan, styles)
	if lineWidth < 58 {
		fmt.Fprintf(&inner, "\n%s", styles.footer.Render("[enter] confirm  [n/esc] cancel"))
	} else {
		fmt.Fprintf(&inner, "\n%s", styles.footer.Render("[enter] confirm  [n] cancel  [esc] cancel"))
	}

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(lineWidth)
	if !noColor {
		panel = panel.BorderForeground(lipgloss.Color(confirmBorderColor(action)))
	}
	fmt.Fprintf(builder, "\n  %s\n", panel.Render(inner.String()))
}

// confirmBorderColor returns the ANSI color code for the confirmation border.
func confirmBorderColor(action GuidedAction) string {
	switch action.ID {
	case "down", "detach":
		return "1" // red — destructive
	case "init", "init_here", "overwrite_init":
		return "6" // cyan — constructive
	default:
		return "8" // dim — neutral
	}
}

func renderShellViewState(plan NextActionPlan, width, cursor int, showDetails, noColor bool, message string, confirming bool, editing bool, editCursor int, editDraft projectSettingsDraft, utility *UtilitySurface) string {
	styles := shellStylesFor(noColor)
	var b strings.Builder
	lineWidth := clampInt(width-4, 42, 96)

	renderSurfaceHeader(&b, lineWidth, surfaceLabel(plan, cursor, editing, confirming, utility), styles)
	fmt.Fprintf(&b, "  %s\n\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	fmt.Fprintf(&b, "  %s\n", verdictStyle(plan, styles).Render(plan.StatusHeader))
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
			renderVisibleDefaultsSection(&b, plan.VisibleDefaults, lineWidth, styles)
		}

		if confirming {
			if action, ok := selectedAction(plan, cursor); ok {
				renderConfirmationModal(&b, action, plan, lineWidth, styles, noColor)
			}
		} else {
			if workItemsFirst(plan) {
				renderWorkSection(&b, plan, lineWidth, styles)
				renderDecisionSection(&b, plan.DecisionItems, cursor, lineWidth, styles)
			} else {
				renderDecisionSection(&b, plan.DecisionItems, cursor, lineWidth, styles)
				renderWorkSection(&b, plan, lineWidth, styles)
			}
		}

		if showDetails || len(plan.DecisionItems) == 0 {
			renderDetailsSection(&b, plan, lineWidth, styles)
		}
	}

	fmt.Fprintf(&b, "\n  %s\n", styles.rule.Render(strings.Repeat("─", lineWidth)))
	footer := footerHint(plan, lineWidth, utility, confirming, editing)
	fmt.Fprintf(&b, "  %s\n\n", styles.footer.Render(footer))
	return b.String()
}

func renderSurfaceHeader(builder *strings.Builder, lineWidth int, surface string, styles shellStyles) {
	left := styles.accent.Render("◆") + "  " + styles.title.Render("StageServe")
	if surface == "" {
		fmt.Fprintf(builder, "\n  %s\n", left)
		return
	}
	gap := maximumInt(2, lineWidth-lipgloss.Width("◆  StageServe")-lipgloss.Width(surface))
	fmt.Fprintf(builder, "\n  %s%s%s\n", left, strings.Repeat(" ", gap), styles.surface.Render(surface))
}

func renderVisibleDefaultsSection(builder *strings.Builder, defaults []VisibleDefault, lineWidth int, styles shellStyles) {
	if len(defaults) == 0 {
		return
	}
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle("Key facts", lineWidth, sectionToneNeutral, styles))
	stacked := lineWidth < 58
	for _, item := range defaults {
		renderVisibleDefaultRow(builder, item, stacked, styles)
	}
}

func renderDecisionSection(builder *strings.Builder, items []GuidedAction, cursor, lineWidth int, styles shellStyles) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle("What you can do", lineWidth, sectionToneAction, styles))
	for itemIndex, item := range items {
		renderDecisionRow(builder, item, itemIndex == cursor, styles)
		if itemIndex < len(items)-1 {
			builder.WriteByte('\n')
		}
	}
}

func renderWorkSection(builder *strings.Builder, plan NextActionPlan, lineWidth int, styles shellStyles) {
	if len(plan.WorkItems) == 0 {
		return
	}
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle(workSectionTitle(plan), lineWidth, workSectionTone(plan), styles))
	highlighted := false
	for _, item := range plan.WorkItems {
		isActive := !highlighted && item.Status != "ready"
		if isActive {
			highlighted = true
		}
		renderWorkItemRow(builder, item, isActive, styles)
	}
}

func renderDetailsSection(builder *strings.Builder, plan NextActionPlan, lineWidth int, styles shellStyles) {
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle("Details", lineWidth, sectionToneNeutral, styles))
	if len(plan.Warnings) == 0 {
		fmt.Fprintf(builder, "  %s\n", styles.muted.Render("No extra warnings for this plan."))
	}
	for _, warning := range plan.Warnings {
		fmt.Fprintf(builder, "  %s\n", warning)
	}
	if len(plan.DirectCommands) > 0 {
		builder.WriteByte('\n')
		for _, command := range plan.DirectCommands {
			fmt.Fprintf(builder, "  %s\n", styles.command.Render(command))
		}
	}
}

func workItemsFirst(plan NextActionPlan) bool {
	switch plan.Situation {
	case SituationMachineNotReady, SituationDriftDetected, SituationUnknownError:
		return len(plan.WorkItems) > 0
	default:
		return false
	}
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
		builder.WriteString("  It will not run the project or change your application files.\n")
	case "overwrite_init":
		builder.WriteString("  StageServe will update the settings file shown above.\n")
		builder.WriteString("  It will not run the project or change your application files.\n")
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
	fmt.Fprintf(builder, "\n%s\n\n", sectionTitle(utility.Title, lineWidth, sectionToneNeutral, styles))
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

func commandUtilitySurface(plan NextActionPlan) *UtilitySurface {
	body := "No direct commands for this screen yet."
	if len(plan.DirectCommands) > 0 {
		body = strings.Join(plan.DirectCommands, "\n")
	}
	return &UtilitySurface{
		Title:           "Direct commands",
		Body:            body,
		DismissOnAnyKey: true,
	}
}

func renderVisibleDefaultRow(builder *strings.Builder, item VisibleDefault, stacked bool, styles shellStyles) {
	if stacked {
		fmt.Fprintf(builder, "  %s\n", styles.label.Render(item.Label))
		fmt.Fprintf(builder, "    %s\n", item.Value)
		if item.Note != "" {
			fmt.Fprintf(builder, "    %s\n", styles.muted.Render(item.Note))
		}
		return
	}
	fmt.Fprintf(builder, "  %s  %s", styles.label.Render(fmt.Sprintf("%-13s", item.Label)), item.Value)
	if item.Note != "" {
		fmt.Fprintf(builder, "  %s", styles.muted.Render(item.Note))
	}
	builder.WriteByte('\n')
}

func renderDecisionRow(builder *strings.Builder, item GuidedAction, selected bool, styles shellStyles) {
	marker := " "
	labelStyle := styles.label
	if selected {
		marker = "▶"
		labelStyle = styles.focus
	}
	fmt.Fprintf(builder, "  %s %s\n", styles.accent.Render(marker), labelStyle.Render(item.Label))
	fmt.Fprintf(builder, "    %s\n", styles.muted.Render(item.Description))
}

func renderWorkItemRow(builder *strings.Builder, item WorkItem, active bool, styles shellStyles) {
	marker := workItemMarker(item.Status)
	labelStyle := styles.label
	if active {
		labelStyle = styles.focus
	}
	fmt.Fprintf(builder, "  %s %s", styles.accent.Render(marker), labelStyle.Render(item.Label))
	if item.Status != "" {
		fmt.Fprintf(builder, "  %s", styles.muted.Render(item.Status))
	}
	builder.WriteByte('\n')
	if item.Description != "" {
		fmt.Fprintf(builder, "    %s\n", styles.muted.Render(item.Description))
	}
	if item.DirectCommand != "" {
		fmt.Fprintf(builder, "    Direct command: %s\n", styles.command.Render(item.DirectCommand))
	}
}

func selectedAction(plan NextActionPlan, cursor int) (GuidedAction, bool) {
	if cursor < 0 || cursor >= len(plan.DecisionItems) {
		return GuidedAction{}, false
	}
	return plan.DecisionItems[cursor], true
}

func setupAction(plan NextActionPlan) (GuidedAction, bool) {
	if plan.Situation != SituationMachineNotReady || len(plan.WorkItems) == 0 {
		return GuidedAction{}, false
	}
	item := plan.WorkItems[0]
	return GuidedAction{
		ID:            "setup",
		Kind:          "work",
		Label:         item.Label,
		Description:   "Review what StageServe still needs before it can run this project.",
		DirectCommand: commandOr(item.DirectCommand, "stage setup"),
	}, true
}

func surfaceLabel(plan NextActionPlan, cursor int, editing, confirming bool, utility *UtilitySurface) string {
	if editing {
		return "Project settings"
	}
	if confirming {
		if action, ok := selectedAction(plan, cursor); ok {
			switch action.ID {
			case "init", "init_here", "overwrite_init":
				return "Project setup"
			}
		}
	}
	if utility != nil {
		switch plan.Situation {
		case SituationMachineNotReady:
			return "Setup"
		case SituationDriftDetected, SituationUnknownError:
			return "Recovery"
		default:
			return "Project"
		}
	}
	switch plan.Situation {
	case SituationMachineNotReady:
		return "Setup"
	case SituationNotProject, SituationProjectMissingConf:
		return "Project setup"
	case SituationDriftDetected, SituationUnknownError:
		return "Recovery"
	default:
		return "Project"
	}
}

func workSectionTone(plan NextActionPlan) sectionTone {
	switch plan.Situation {
	case SituationMachineNotReady, SituationDriftDetected, SituationUnknownError:
		return sectionToneWarning
	default:
		return sectionToneNeutral
	}
}

func footerHint(plan NextActionPlan, lineWidth int, utility *UtilitySurface, confirming, editing bool) string {
	if utility != nil {
		return utilityFooter(*utility)
	}
	if confirming {
		if lineWidth < 58 {
			return "enter confirm • n cancel • esc"
		}
		return "enter confirm • n cancel • esc cancel"
	}
	if editing {
		if lineWidth < 58 {
			return "type edit • tab move • enter save • esc"
		}
		return "type edit • tab/↑/↓ move • enter save • esc cancel"
	}
	if _, ok := setupAction(plan); ok {
		if lineWidth < 58 {
			return "enter review • c commands • d doctor • q"
		}
		return "enter review • c commands • d doctor • q quit"
	}
	if lineWidth < 58 {
		return "↑/↓ move • enter choose • c commands • d doctor • q"
	}
	return "↑/↓ inspect • enter choose • c commands • d doctor • ? details • q quit"
}

func sectionTitle(title string, width int, tone sectionTone, styles shellStyles) string {
	style := styles.sectionNeutral
	switch tone {
	case sectionToneAction:
		style = styles.sectionAction
	case sectionToneWarning:
		style = styles.sectionWarn
	}
	fill := maximumInt(0, width-lipgloss.Width(title)-5)
	return "  " + styles.rule.Render("── ") + style.Render(title) + styles.rule.Render(" "+strings.Repeat("─", fill))
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
