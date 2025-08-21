package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	App lipgloss.Style
}

func DefaultStyles() Styles {
	s := Styles{}
	s.App = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		Padding(1, 2)
	return s
}
