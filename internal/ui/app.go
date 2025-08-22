package ui

import (
	"fmt"
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type activeView int

const (
	searchView activeView = iota
	historyView
)

type AppModel struct {
	width, height  int
	styles         Styles
	focus          ports.FocusState
	activeView     activeView
	config         domain.Config
	youtubeService ports.YoutubeService
	playerService  ports.PlayerService
	storageService ports.StorageService
	search         SearchModel
	history        HistoryModel
	player         PlayerModel
}

func InitialModel(ytService ports.YoutubeService, pService ports.PlayerService, sService ports.StorageService, cfg domain.Config) AppModel {
	styles := DefaultStyles()
	return AppModel{
		styles:         styles,
		focus:          ports.GlobalFocus,
		activeView:     searchView,
		config:         cfg,
		youtubeService: ytService,
		playerService:  pService,
		storageService: sService,
		search:         NewSearchModel(ytService, cfg, styles),
		history:        NewHistoryModel(sService, cfg, styles),
		player:         NewPlayerModel(),
	}
}

func (m *AppModel) savePositionAndQuit() tea.Cmd {
	if m.config.Playback.SavePositionOnQuit && (m.player.status == statusPlaying || m.player.status == statusPaused) {
		state, err := m.playerService.GetState()
		if err == nil && state.Position > 0 {
			m.storageService.UpdateHistoryEntryPosition(m.player.song.ID, int(state.Position))
		}
	}
	return tea.Quit
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return ports.TickMsg(t)
	})
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.search.Init(), m.history.Init(), tickCmd())
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ports.ChangeFocusMsg:
		m.focus = msg.NewFocus
		if m.focus == ports.SearchFocus {
			if m.activeView == searchView {
				cmd = m.search.Focus()
			} else {
				cmd = m.history.Focus()
			}
		} else {
			m.search.Blur()
			m.history.Blur()
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case ports.PlaySongMsg:
		m.focus = ports.GlobalFocus
		m.player.SetContent(statusLoading, msg.Song, nil)

		var resumeAt int
		if m.config.Playback.SavePositionOnQuit {
			resumeAt = m.history.GetResumeAt(msg.Song.ID)
		}

		youtubeURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", msg.Song.ID)
		if resumeAt > 0 {
			youtubeURL = fmt.Sprintf("%s&t=%ds", youtubeURL, resumeAt)
		}

		err := m.playerService.Play(youtubeURL)
		if err != nil {
			cmds = append(cmds, func() tea.Msg { return ports.PlayErrorMsg{Err: err} })
		} else {
			cmds = append(cmds, func() tea.Msg { return ports.SongNowPlayingMsg{Song: msg.Song} })
		}

		go m.storageService.AddToHistory(domain.HistoryEntry{Song: msg.Song})

	case ports.SongNowPlayingMsg:
		m.player.SetContent(statusPlaying, msg.Song, nil)

	case ports.PlayErrorMsg:
		m.player.SetContent(statusError, domain.Song{}, msg.Err)

	case ports.TickMsg:
		if m.player.status == statusPlaying || m.player.status == statusPaused {
			state, err := m.playerService.GetState()
			if err == nil {
				cmds = append(cmds, func() tea.Msg { return ports.PlayerStateUpdateMsg{State: state} })
			}
		}
		cmds = append(cmds, tickCmd())

	case ports.PlayerStateUpdateMsg:
		m.player, cmd = m.player.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.focus == ports.GlobalFocus {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, m.savePositionAndQuit()
			case "s":
				m.activeView = searchView
				return m, func() tea.Msg { return ports.ChangeFocusMsg{NewFocus: ports.SearchFocus} }
			case "h":
				m.activeView = historyView
				cmds = append(cmds, m.history.Init())
				cmds = append(cmds, func() tea.Msg { return ports.ChangeFocusMsg{NewFocus: ports.SearchFocus} })
				return m, tea.Batch(cmds...)
			case " ":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.Pause()
				}
			case "right":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.Seek(5)
				}
			case "left":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.Seek(-5)
				}
			case "]":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.ChangeSpeed(0.25)
				}
			case "[":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.ChangeSpeed(-0.25)
				}
			case `\`:
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.ResetSpeed()
				}
			}
		}
	}

	if m.focus == ports.SearchFocus {
		switch m.activeView {
		case searchView:
			m.search, cmd = m.search.Update(msg)
		case historyView:
			m.history, cmd = m.history.Update(msg)
		}
		cmds = append(cmds, cmd)
	}
	m.player, cmd = m.player.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	appWidth := m.width - m.styles.App.GetHorizontalFrameSize()
	appHeight := m.height - m.styles.App.GetVerticalFrameSize()

	footerHeight := 4
	mainPanelHeight := appHeight - footerHeight

	m.search.SetSize(appWidth, mainPanelHeight)
	m.history.SetSize(appWidth, mainPanelHeight)
	m.player.SetSize(appWidth)

	var mainContent, searchFooterContent string
	var internalFocus searchComponentFocus

	switch m.activeView {
	case searchView:
		mainContent, searchFooterContent = m.search.View()
		internalFocus = m.search.GetFocus()
	case historyView:
		mainContent, searchFooterContent = m.history.View()
		internalFocus = m.history.GetFocus()
	}

	playerFooterContent := m.player.View()

	var mainPanelStyle, footerPanelStyle lipgloss.Style
	var footerContent, footerTitle string

	showPlayer := m.player.status != statusIdle
	if m.focus == ports.SearchFocus {
		showPlayer = false
	}

	if showPlayer {
		footerTitle = m.player.ViewTitle()
		footerContent = playerFooterContent
		footerPanelStyle = focusedBorderStyle
		mainPanelStyle = blurredBorderStyle
	} else {
		if m.activeView == searchView {
			footerTitle = "search"
		} else {
			footerTitle = "history"
		}
		footerContent = searchFooterContent
		if m.focus == ports.SearchFocus && internalFocus == inputFocus {
			footerPanelStyle = focusedBorderStyle
			mainPanelStyle = blurredBorderStyle
		} else if m.focus == ports.SearchFocus && internalFocus == listFocus {
			footerPanelStyle = blurredBorderStyle
			mainPanelStyle = focusedBorderStyle
		} else {
			footerPanelStyle = blurredBorderStyle
			mainPanelStyle = blurredBorderStyle
		}
	}

	mainPanel := mainPanelStyle.
		Width(appWidth).
		Height(mainPanelHeight).
		Padding(0, 1).
		Render(mainContent)

	footerPanel := footerPanelStyle.
		Width(appWidth).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(
			lipgloss.Top,
			lipgloss.NewStyle().Bold(true).Render(footerTitle),
			footerContent,
		))

	finalView := lipgloss.JoinVertical(lipgloss.Top,
		mainPanel,
		footerPanel,
	)

	return m.styles.App.Render(finalView)
}
