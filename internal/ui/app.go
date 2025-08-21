package ui

import (
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AppModel struct {
	width, height  int
	styles         Styles
	focus          focusState
	youtubeService ports.YoutubeService
	playerService  ports.PlayerService
	search         SearchModel
	player         PlayerModel
}

func InitialModel(ytService ports.YoutubeService, pService ports.PlayerService) AppModel {
	styles := DefaultStyles()
	return AppModel{
		styles:         styles,
		focus:          globalFocus,
		youtubeService: ytService,
		playerService:  pService,
		search:         NewSearchModel(ytService, styles),
		player:         NewPlayerModel(),
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func getStreamURLCmd(service ports.YoutubeService, song domain.Song) tea.Cmd {
	return func() tea.Msg {
		url, err := service.GetStreamURL(song.ID)
		if err != nil {
			return playErrorMsg{err}
		}
		return streamURLFetchedMsg{song: song, url: url}
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.search.Init(), tickCmd())
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case changeFocusMsg:
		m.focus = msg.newFocus
		if m.focus == searchFocus {
			cmd = m.search.Focus()
		} else {
			m.search.Blur()
		}
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case playSongMsg:
		m.focus = globalFocus
		m.player.SetContent(statusLoading, msg.song, nil)
		cmds = append(cmds, getStreamURLCmd(m.youtubeService, msg.song))

	case streamURLFetchedMsg:
		err := m.playerService.Play(msg.url)
		if err != nil {
			m.player.SetContent(statusError, msg.song, err)
		} else {
			cmds = append(cmds, func() tea.Msg { return songNowPlayingMsg{song: msg.song} })
		}

	case songNowPlayingMsg:
		m.player.SetContent(statusPlaying, msg.song, nil)

	case playErrorMsg:
		m.player.SetContent(statusError, domain.Song{}, msg.err)

	case tickMsg:
		if m.player.status == statusPlaying || m.player.status == statusPaused {
			state, err := m.playerService.GetState()
			if err == nil {
				cmds = append(cmds, func() tea.Msg { return playerStateUpdateMsg{state} })
			}
		}
		cmds = append(cmds, tickCmd())

	case playerStateUpdateMsg:
		m.player, cmd = m.player.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.focus == globalFocus {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "s":
				return m, func() tea.Msg { return changeFocusMsg{newFocus: searchFocus} }
			case " ":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.Pause()
				}
			case "right":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.Seek(10)
				}
			case "left":
				if m.player.status == statusPlaying || m.player.status == statusPaused {
					m.playerService.Seek(-10)
				}
			}
		}
	}

	if m.focus == searchFocus {
		m.search, cmd = m.search.Update(msg)
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
	m.player.SetSize(appWidth)

	mainContent, searchFooterContent := m.search.View()
	playerFooterContent := m.player.View()

	var mainPanelStyle, footerPanelStyle lipgloss.Style
	var footerContent, footerTitle string

	searchInternalFocus := m.search.GetFocus()

	showPlayer := m.player.status != statusIdle
	if m.focus == searchFocus {
		showPlayer = false
	}

	if showPlayer {
		footerContent = playerFooterContent
		footerPanelStyle = focusedBorderStyle
		mainPanelStyle = blurredBorderStyle
	} else {
		footerContent = searchFooterContent
		if m.focus == searchFocus && searchInternalFocus == inputFocus {
			footerPanelStyle = focusedBorderStyle
			mainPanelStyle = blurredBorderStyle
		} else if m.focus == searchFocus && searchInternalFocus == listFocus {
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
