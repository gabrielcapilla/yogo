package ui

import (
	"fmt"
	"yogo/internal/domain"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PlayerModel struct {
	width, height int
	status        string
	song          domain.Song
	err           error
	styles        Styles
}

func NewPlayerModel(styles Styles) PlayerModel {
	return PlayerModel{status: "Idle", styles: styles}
}

func (m PlayerModel) Init() tea.Cmd                             { return nil }
func (m PlayerModel) Update(msg tea.Msg) (PlayerModel, tea.Cmd) { return m, nil }

func (m *PlayerModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *PlayerModel) SetContent(status string, song domain.Song, err error) {
	m.status = status
	m.song = song
	m.err = err
}

func (m PlayerModel) View() string {
	var content string
	switch m.status {
	case "Idle":
		content = "..."
	case "Reproduciendo":
		sTitle := m.styles.PlayerTitle.Render(m.song.Title)
		sArtist := m.styles.PlayerArtist.Render(m.song.Artists[0])
		content = lipgloss.JoinHorizontal(lipgloss.Left, sTitle, " - ", sArtist)
	case "Error":
		content = m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", m.err))
	default:
		content = m.status + ": " + m.song.Title
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Center, content)
}
