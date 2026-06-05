package guidance

import "github.com/charmbracelet/lipgloss"

type shellStyles struct {
	accent         lipgloss.Style
	focus          lipgloss.Style
	title          lipgloss.Style
	surface        lipgloss.Style
	rule           lipgloss.Style
	verdict        lipgloss.Style
	verdictReady   lipgloss.Style
	verdictWarn    lipgloss.Style
	verdictError   lipgloss.Style
	label          lipgloss.Style
	muted          lipgloss.Style
	command        lipgloss.Style
	footer         lipgloss.Style
	sectionNeutral lipgloss.Style
	sectionAction  lipgloss.Style
	sectionWarn    lipgloss.Style
}

type sectionTone int

const (
	sectionToneNeutral sectionTone = iota
	sectionToneAction
	sectionToneWarning
)

func shellStylesFor(noColor bool) shellStyles {
	if noColor {
		return shellStyles{
			accent:         lipgloss.NewStyle(),
			focus:          lipgloss.NewStyle().Bold(true),
			title:          lipgloss.NewStyle(),
			surface:        lipgloss.NewStyle(),
			rule:           lipgloss.NewStyle(),
			verdict:        lipgloss.NewStyle(),
			verdictReady:   lipgloss.NewStyle(),
			verdictWarn:    lipgloss.NewStyle(),
			verdictError:   lipgloss.NewStyle(),
			label:          lipgloss.NewStyle(),
			muted:          lipgloss.NewStyle(),
			command:        lipgloss.NewStyle(),
			footer:         lipgloss.NewStyle(),
			sectionNeutral: lipgloss.NewStyle(),
			sectionAction:  lipgloss.NewStyle(),
			sectionWarn:    lipgloss.NewStyle(),
		}
	}
	styles := shellStyles{
		accent:         lipgloss.NewStyle(),
		focus:          lipgloss.NewStyle().Bold(true),
		title:          lipgloss.NewStyle().Bold(true),
		surface:        lipgloss.NewStyle().Bold(true),
		rule:           lipgloss.NewStyle(),
		verdict:        lipgloss.NewStyle().Bold(true),
		verdictReady:   lipgloss.NewStyle().Bold(true),
		verdictWarn:    lipgloss.NewStyle().Bold(true),
		verdictError:   lipgloss.NewStyle().Bold(true),
		label:          lipgloss.NewStyle().Bold(true),
		muted:          lipgloss.NewStyle(),
		command:        lipgloss.NewStyle().Bold(true),
		footer:         lipgloss.NewStyle(),
		sectionNeutral: lipgloss.NewStyle().Bold(true),
		sectionAction:  lipgloss.NewStyle().Bold(true),
		sectionWarn:    lipgloss.NewStyle().Bold(true),
	}
	styles.accent = styles.accent.Foreground(lipgloss.Color("6"))
	styles.focus = styles.focus.Foreground(lipgloss.Color("6"))
	styles.title = styles.title.Foreground(lipgloss.Color("15"))
	styles.surface = styles.surface.Foreground(lipgloss.Color("7"))
	styles.rule = styles.rule.Foreground(lipgloss.Color("8"))
	styles.verdict = styles.verdict.Foreground(lipgloss.Color("15"))
	styles.verdictReady = styles.verdictReady.Foreground(lipgloss.Color("2"))
	styles.verdictWarn = styles.verdictWarn.Foreground(lipgloss.Color("3"))
	styles.verdictError = styles.verdictError.Foreground(lipgloss.Color("1"))
	styles.label = styles.label.Foreground(lipgloss.Color("15"))
	styles.muted = styles.muted.Foreground(lipgloss.Color("7"))
	styles.command = styles.command.Foreground(lipgloss.Color("14"))
	styles.footer = styles.footer.Foreground(lipgloss.Color("8"))
	styles.sectionNeutral = styles.sectionNeutral.Foreground(lipgloss.Color("15"))
	styles.sectionAction = styles.sectionAction.Foreground(lipgloss.Color("6"))
	styles.sectionWarn = styles.sectionWarn.Foreground(lipgloss.Color("3"))
	return styles
}

func verdictStyle(plan NextActionPlan, styles shellStyles) lipgloss.Style {
	switch plan.Situation {
	case SituationProjectReadyToRun, SituationProjectRunning:
		return styles.verdictReady
	case SituationMachineNotReady, SituationDriftDetected:
		return styles.verdictWarn
	case SituationUnknownError:
		return styles.verdictError
	default:
		return styles.verdict
	}
}

func severityLabel(situation Situation) string {
	switch situation {
	case SituationProjectReadyToRun, SituationProjectRunning, SituationProjectDown, SituationProjectMissingConf, SituationNotProject:
		return "Ready"
	case SituationMachineNotReady, SituationDriftDetected:
		return "Needs attention"
	case SituationUnknownError:
		return "Error"
	default:
		return "Status"
	}
}
