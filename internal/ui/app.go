package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	width, height int
	styles        Styles
}

func InitialModel() Model {
	return Model{
		styles: DefaultStyles(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	content := "Press 'q' to exit."

	centeredContent := lipgloss.Place(
		m.width-m.styles.App.GetHorizontalFrameSize(),
		m.height-m.styles.App.GetVerticalFrameSize(),
		lipgloss.Center,
		lipgloss.Center,
		content,
	)

	return m.styles.App.Render(centeredContent)
}
