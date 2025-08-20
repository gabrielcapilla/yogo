package ui

import (
	"fmt"
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/progress"
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
	progress      progress.Model
}

func NewPlayerModel(styles Styles) PlayerModel {
	return PlayerModel{
		status:   "Idle",
		styles:   styles,
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m PlayerModel) Init() tea.Cmd { return nil }

func (m PlayerModel) Update(msg tea.Msg) (PlayerModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case playerStateUpdateMsg:
		m.state = msg.state
		if m.state.Duration > 0 {
			cmd = m.progress.SetPercent(m.state.Position / m.state.Duration)
		}
		return m, cmd
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}
	return m, nil
}

func (m *PlayerModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.progress.Width = w - 20
}

func (m *PlayerModel) SetContent(status string, song domain.Song, err error) {
	m.status = status
	m.song = song
	m.err = err
	if status != "Playing" {
		m.state = ports.PlayerState{}
		m.progress.SetPercent(0)
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
	case "Loading":
		content = "Loading: " + truncate(m.song.Title, m.width-len("Loading: ")-2)
	case "Playing":
		posStr := formatDuration(m.state.Position)
		durStr := formatDuration(m.state.Duration)

		songInfo := fmt.Sprintf("%s - %s", m.styles.PlayerTitle.Render(m.song.Title), m.styles.PlayerArtist.Render(m.song.Artists[0]))
		songInfo = truncate(songInfo, m.width)

		progressView := lipgloss.JoinHorizontal(lipgloss.Center,
			m.styles.Help.Render(posStr),
			" ",
			m.progress.View(),
			" ",
			m.styles.Help.Render(durStr),
		)

		content = lipgloss.JoinVertical(lipgloss.Left, songInfo, progressView)

	case "Error":
		content = m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", m.err))
	default:
		content = m.status
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Center, content)
}
