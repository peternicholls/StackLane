package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-isatty"
)

func main() {
	var listSections bool
	var noTUI bool
	var plain bool
	var sectionID string

	flag.BoolVar(&listSections, "list-sections", false, "print style guide section IDs")
	flag.BoolVar(&noTUI, "notui", false, "render the guide as plain terminal text")
	flag.BoolVar(&plain, "plain", false, "alias for --notui")
	flag.StringVar(&sectionID, "section", "", "open a specific section ID")
	flag.Parse()

	styleGuide := buildGuide()
	if listSections {
		printSections(os.Stdout, styleGuide)
		return
	}

	if noTUI || plain || !isInteractive(os.Stdin, os.Stdout) {
		renderPlain(os.Stdout, styleGuide, sectionID)
		return
	}

	program := tea.NewProgram(newGuideModel(styleGuide, sectionID), tea.WithAltScreen())
	if _, runErr := program.Run(); runErr != nil {
		fmt.Fprintf(os.Stderr, "style guide failed: %v\n", runErr)
		os.Exit(1)
	}
}

func isInteractive(stdin *os.File, stdout *os.File) bool {
	return isatty.IsTerminal(stdin.Fd()) && isatty.IsTerminal(stdout.Fd())
}

func printSections(writer io.Writer, styleGuide guide) {
	for _, section := range styleGuide.Sections {
		fmt.Fprintln(writer, section.ID)
	}
}

func renderPlain(writer io.Writer, styleGuide guide, sectionID string) {
	sections := styleGuide.Sections
	if sectionID != "" {
		if section, ok := styleGuide.sectionByID(sectionID); ok {
			sections = []guideSection{section}
		}
	}

	for sectionIndex, section := range sections {
		if sectionIndex > 0 {
			fmt.Fprintln(writer)
			fmt.Fprintln(writer, strings.Repeat("=", 72))
			fmt.Fprintln(writer)
		}
		fmt.Fprintln(writer, section.Title)
		fmt.Fprintln(writer, strings.Repeat("-", len(section.Title)))
		fmt.Fprintln(writer)
		plainText := stripGuideANSI(section.Render(78))
		fmt.Fprintln(writer, strings.TrimSpace(plainText))
	}
}

type guideModel struct {
	styleGuide guide
	selected   int
	viewport   viewport.Model
	width      int
	height     int
	ready      bool
	showHelp   bool
}

func newGuideModel(styleGuide guide, sectionID string) guideModel {
	selectedIndex := styleGuide.indexForID(sectionID)
	contentViewport := viewport.New(80, 24)
	guideModel := guideModel{
		styleGuide: styleGuide,
		selected:   selectedIndex,
		viewport:   contentViewport,
	}
	guideModel.refreshContent()
	return guideModel
}

func (guideModel guideModel) Init() tea.Cmd {
	return nil
}

func (guideModel guideModel) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMessage := message.(type) {
	case tea.WindowSizeMsg:
		guideModel.width = typedMessage.Width
		guideModel.height = typedMessage.Height
		guideModel.ready = true
		guideModel.resizeViewport()
		guideModel.refreshContent()
		return guideModel, nil
	case tea.KeyMsg:
		switch typedMessage.String() {
		case "ctrl+c", "q":
			return guideModel, tea.Quit
		case "?":
			guideModel.showHelp = !guideModel.showHelp
			return guideModel, nil
		case "esc":
			if guideModel.showHelp {
				guideModel.showHelp = false
				return guideModel, nil
			}
			return guideModel, tea.Quit
		case "up", "shift+tab", "[":
			guideModel.previousSection()
			return guideModel, nil
		case "down", "tab", "]":
			guideModel.nextSection()
			return guideModel, nil
		case "home":
			guideModel.selected = 0
			guideModel.refreshContent()
			return guideModel, nil
		case "end":
			guideModel.selected = len(guideModel.styleGuide.Sections) - 1
			guideModel.refreshContent()
			return guideModel, nil
		case "enter":
			guideModel.viewport.GotoTop()
			return guideModel, nil
		}
	}

	updatedViewport, command := guideModel.viewport.Update(message)
	guideModel.viewport = updatedViewport
	return guideModel, command
}

func (guideModel guideModel) View() string {
	if !guideModel.ready {
		return "StageServe terminal style guide\n"
	}

	header := renderAppHeader(guideModel.width)
	footer := renderFooter(guideModel.width)
	bodyHeight := maximum(6, guideModel.height-lipglossHeight(header)-lipglossHeight(footer)-2)
	navWidth := navigationWidth(guideModel.width)
	contentWidth := maximum(42, guideModel.width-navWidth-3)
	guideModel.viewport.Width = contentWidth
	guideModel.viewport.Height = bodyHeight

	navigation := renderNavigation(guideModel.styleGuide, guideModel.selected, navWidth, bodyHeight)
	content := contentPaneStyle.Width(contentWidth).Height(bodyHeight).Render(guideModel.viewport.View())
	body := joinHorizontal(navigation, content)

	if guideModel.showHelp {
		help := renderHelpSheet(guideModel.width)
		return strings.Join([]string{header, body, help, footer}, "\n")
	}

	return strings.Join([]string{header, body, footer}, "\n")
}

func (guideModel *guideModel) resizeViewport() {
	navWidth := navigationWidth(guideModel.width)
	contentWidth := maximum(42, guideModel.width-navWidth-3)
	headerHeight := 4
	footerHeight := 2
	bodyHeight := maximum(6, guideModel.height-headerHeight-footerHeight-2)
	guideModel.viewport.Width = contentWidth
	guideModel.viewport.Height = bodyHeight
}

func (guideModel *guideModel) refreshContent() {
	if len(guideModel.styleGuide.Sections) == 0 {
		guideModel.viewport.SetContent("No style guide sections are available.")
		return
	}
	section := guideModel.styleGuide.Sections[guideModel.selected]
	guideModel.viewport.SetContent(section.Render(maximum(42, guideModel.viewport.Width)))
	guideModel.viewport.GotoTop()
}

func (guideModel *guideModel) previousSection() {
	if guideModel.selected == 0 {
		guideModel.selected = len(guideModel.styleGuide.Sections) - 1
	} else {
		guideModel.selected--
	}
	guideModel.refreshContent()
}

func (guideModel *guideModel) nextSection() {
	guideModel.selected = (guideModel.selected + 1) % len(guideModel.styleGuide.Sections)
	guideModel.refreshContent()
}
