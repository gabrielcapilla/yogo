package ui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	App      lipgloss.Style
	Box      lipgloss.Style
	BoxTitle lipgloss.Style
	Help     lipgloss.Style

	ErrorText lipgloss.Style
	Spinner   lipgloss.Style

	SearchPrompt  lipgloss.Style
	FocusedBorder lipgloss.Style
	BlurredBorder lipgloss.Style
	ListPointer   lipgloss.Style
	ListSelected  lipgloss.Style
	ListNormal    lipgloss.Style
	PlayerTitle   lipgloss.Style
	PlayerArtist  lipgloss.Style
}

func DefaultStyles() Styles {
	s := Styles{}

	s.App = lipgloss.NewStyle().Margin(1, 1)
	s.BoxTitle = lipgloss.NewStyle().Bold(true)
	s.Help = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A0A0A0", Dark: "#626262"})
	s.ErrorText = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	s.Spinner = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	s.SearchPrompt = lipgloss.NewStyle().Bold(true)

	s.BlurredBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true)
	s.FocusedBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(lipgloss.Color("12"))

	s.ListPointer = lipgloss.NewStyle().SetString("> ")
	s.ListSelected = lipgloss.NewStyle().Bold(true)
	s.ListNormal = lipgloss.NewStyle()
	s.PlayerTitle = lipgloss.NewStyle().Bold(true)
	s.PlayerArtist = lipgloss.NewStyle().Faint(true)

	return s
}
