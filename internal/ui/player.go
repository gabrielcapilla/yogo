package ui

import (
	"fmt"
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PlayerModel struct {
	width, height int
	status        string
	song          domain.Song
	err           error
	styles        Styles
	state         ports.PlayerState
}

func NewPlayerModel(styles Styles) PlayerModel {
	return PlayerModel{status: "Idle", styles: styles}
}

func (m PlayerModel) Init() tea.Cmd { return nil }

func (m PlayerModel) Update(msg tea.Msg) (PlayerModel, tea.Cmd) {
	switch msg := msg.(type) {
	case playerStateUpdateMsg:
		m.state = msg.state
	}
	return m, nil
}

func (m *PlayerModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *PlayerModel) SetContent(status string, song domain.Song, err error) {
	m.status = status
	m.song = song
	m.err = err
	if status != "Playing" {
		m.state = ports.PlayerState{}
	}
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "00:00"
	}
	d := time.Duration(seconds) * time.Second
	return fmt.Sprintf("%02d:%02d", int(d.Minutes())%60, int(d.Seconds())%60)
}

func (m PlayerModel) View() string {
	var content string
	switch m.status {
	case "Idle":
		content = "..."
	case "Playing":
		pos := formatDuration(m.state.Position)
		dur := formatDuration(m.state.Duration)
		timeInfo := fmt.Sprintf("[%s / %s]", pos, dur)

		sTitle := m.styles.PlayerTitle.Render(m.song.Title)
		sArtist := m.styles.PlayerArtist.Render(m.song.Artists[0])
		sTime := m.styles.Help.Render(timeInfo)

		titleBlock := lipgloss.JoinHorizontal(lipgloss.Left, sTitle, " - ", sArtist)

		maxTitleWidth := m.width - lipgloss.Width(sTime) - 1
		if maxTitleWidth < 0 {
			maxTitleWidth = 0
		}
		titleBlock = truncate(titleBlock, maxTitleWidth)

		content = lipgloss.JoinHorizontal(lipgloss.Right, titleBlock, sTime)
	case "Error":
		content = m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", m.err))
	default:
		content = m.status + ": " + truncate(m.song.Title, m.width-len(m.status)-2)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Center, content)
}
