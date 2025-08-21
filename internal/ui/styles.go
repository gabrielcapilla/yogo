package ui

import "github.com/charmbracelet/lipgloss"

var (
	colorMagenta = lipgloss.Color("227")
	colorWhite   = lipgloss.Color("255")

	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colorMagenta)

	blurredBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colorWhite)
)

type Styles struct {
	App lipgloss.Style

	Spinner      lipgloss.Style
	ListPointer  lipgloss.Style
	ListSelected lipgloss.Style
	ListNormal   lipgloss.Style
	ErrorText    lipgloss.Style
}

func DefaultStyles() Styles {
	s := Styles{}

	s.App = lipgloss.NewStyle().
		MarginTop(3).
		MarginLeft(1).
		MarginRight(3)

	s.Spinner = lipgloss.NewStyle().Foreground(colorMagenta)
	s.ListPointer = lipgloss.NewStyle().SetString("> ")
	s.ListSelected = lipgloss.NewStyle().Bold(true).Foreground(colorMagenta)
	s.ListNormal = lipgloss.NewStyle()
	s.ErrorText = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	return s
}
