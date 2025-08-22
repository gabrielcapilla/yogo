package ui

import (
	"fmt"
	"strings"
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type playerStatus int

const (
	statusIdle playerStatus = iota
	statusLoading
	statusPlaying
	statusPaused
	statusError
)

type PlayerModel struct {
	width    int
	status   playerStatus
	song     domain.Song
	err      error
	state    ports.PlayerState
	progress progress.Model
}

func NewPlayerModel() PlayerModel {
	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	return PlayerModel{
		status:   statusIdle,
		progress: prog,
	}
}

func (m *PlayerModel) SetSize(w int) {
	m.width = w
}

func (m *PlayerModel) SetContent(status playerStatus, song domain.Song, err error) {
	m.status = status
	m.song = song
	m.err = err
	if status != statusPlaying && status != statusPaused {
		m.state = ports.PlayerState{}
		m.progress.SetPercent(0)
	}
}

func (m PlayerModel) Update(msg tea.Msg) (PlayerModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case playerStateUpdateMsg:
		m.state = msg.state
		if m.state.IsPlaying {
			m.status = statusPlaying
		} else if m.status == statusPlaying {
			m.status = statusPaused
		}

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

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "00:00"
	}
	d := time.Duration(seconds) * time.Second
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func (m PlayerModel) ViewTitle() string {
	playPauseSymbol := "▶"
	if m.state.IsPlaying {
		playPauseSymbol = "Ⅱ"
	}

	speed := 1.0
	if m.state.Speed > 0 {
		speed = m.state.Speed
	}
	var speedStr string
	if speed == float64(int64(speed)) {
		speedStr = fmt.Sprintf("x%d", int64(speed))
	} else {
		speedStr = fmt.Sprintf("x%.2f", speed)
	}

	controls := fmt.Sprintf("« %s »", playPauseSymbol)

	return fmt.Sprintf("Player | %s | %s", controls, speedStr)
}

func (m PlayerModel) View() string {
	var content string
	switch m.status {
	case statusIdle:
		content = "\n"
	case statusLoading:
		title := "Loading: " + m.song.Title
		if len(title) > m.width-2 {
			title = title[:m.width-5] + "..."
		}
		content = lipgloss.JoinVertical(lipgloss.Left, title, "")
	case statusPlaying, statusPaused:
		artists := ""
		if len(m.song.Artists) > 0 {
			artists = " - " + strings.Join(m.song.Artists, ", ")
		}
		songInfo := m.song.Title + artists
		if len(songInfo) > m.width-2 {
			songInfo = songInfo[:m.width-5] + "..."
		}

		posStr := formatDuration(m.state.Position)
		durStr := formatDuration(m.state.Duration)

		availableWidth := m.width - 2

		timeWidth := lipgloss.Width(posStr) + lipgloss.Width(durStr) + 2
		progressBarWidth := availableWidth - timeWidth
		if progressBarWidth < 1 {
			progressBarWidth = 1
		}
		m.progress.Width = progressBarWidth

		progressView := lipgloss.JoinHorizontal(lipgloss.Center,
			posStr,
			" ",
			m.progress.View(),
			" ",
			durStr,
		)
		content = lipgloss.JoinVertical(lipgloss.Left, songInfo, progressView)
	case statusError:
		errorMsg := fmt.Sprintf("Error: %v", m.err)
		if len(errorMsg) > m.width-2 {
			errorMsg = errorMsg[:m.width-5] + "..."
		}
		content = lipgloss.JoinVertical(lipgloss.Left, errorMsg, "")
	}
	return content
}
