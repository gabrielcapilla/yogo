package ui

import (
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MIN_WIDTH  = 84
	MIN_HEIGHT = 14
)

type focusState int

const (
	globalFocus focusState = iota
	searchFocus
	historyFocus
)

type AppModel struct {
	focus          focusState
	width, height  int
	youtubeService ports.YoutubeService
	playerService  ports.PlayerService
	storageService ports.StorageService
	config         domain.Config
	tabs           TabModel
	search         SearchModel
	history        HistoryModel
	player         PlayerModel
	styles         Styles
}

func InitialModel(ytService ports.YoutubeService, pService ports.PlayerService, sService ports.StorageService, cfg domain.Config) AppModel {
	styles := DefaultStyles()
	return AppModel{
		focus:          globalFocus,
		youtubeService: ytService,
		playerService:  pService,
		storageService: sService,
		config:         cfg,
		tabs:           NewTabModel(),
		search:         NewSearchModel(ytService, styles),
		history:        NewHistoryModel(sService, styles),
		player:         NewPlayerModel(styles),
		styles:         styles,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.search.Init(), m.history.Init(), tickCmd())
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

func (m AppModel) loadHistoryCmd() tea.Cmd {
	return func() tea.Msg {
		entries, err := m.storageService.GetHistory(m.config.HistoryLimit)
		if err != nil {
			return HistoryErrorMsg{Err: err}
		}
		return HistoryLoadedMsg{Entries: entries}
	}
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
			m.search.Focus()
		} else {
			m.search.Blur()
		}
		return m, nil
	}

	switch m.focus {
	case searchFocus:
		m.search, cmd = m.search.Update(msg)
		cmds = append(cmds, cmd)
	case historyFocus:
		m.history, cmd = m.history.Update(msg)
		cmds = append(cmds, cmd)
	case globalFocus:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "tab", "l":
				m.tabs.Next()
				if m.tabs.ActiveTab == 1 {
					cmds = append(cmds, m.loadHistoryCmd())
				}
			case "shift+tab", "h":
				m.tabs.Prev()
				if m.tabs.ActiveTab == 1 {
					cmds = append(cmds, m.loadHistoryCmd())
				}
			case "enter":
				if m.tabs.ActiveTab == 0 {
					m.focus = searchFocus
					m.search.Focus()
				} else {
					m.focus = historyFocus
				}
			case " ":
				if m.player.status == "Playing" || m.player.status == "Paused" {
					m.playerService.Pause()
				}
			case "left":
				if m.player.status == "Playing" || m.player.status == "Paused" {
					m.playerService.Seek(-5)
				}
			case "right":
				if m.player.status == "Playing" || m.player.status == "Paused" {
					m.playerService.Seek(5)
				}
			}
		}
	}

	switch msg := msg.(type) {
	case tickMsg:
		state, err := m.playerService.GetState()
		if err != nil {
			cmds = append(cmds, func() tea.Msg { return playErrorMsg{err} })
		} else if state.IsPlaying || m.player.status == "Playing" {
			cmds = append(cmds, func() tea.Msg { return playerStateUpdateMsg{state} })
		}
		cmds = append(cmds, tickCmd())

	case playerStateUpdateMsg:
		m.player, cmd = m.player.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		m.player, cmd = m.player.Update(msg)
		cmds = append(cmds, cmd)

	case playSongMsg:
		m.player.SetContent("Loading", msg.song, nil)
		go m.storageService.AddToHistory(domain.HistoryEntry{Song: msg.song, PlayedAt: time.Now()})
		cmds = append(cmds, getStreamURLCmd(m.youtubeService, msg.song))

	case streamURLFetchedMsg:
		err := m.playerService.Play(msg.url)
		if err != nil {
			cmds = append(cmds, func() tea.Msg { return playErrorMsg{err} })
		} else {
			cmds = append(cmds, func() tea.Msg { return songNowPlayingMsg{song: msg.song} })
		}

	case songNowPlayingMsg:
		m.player.SetContent("Playing", msg.song, nil)

	case playErrorMsg:
		m.player.SetContent("Error", domain.Song{}, msg.err)

	case HistoryLoadedMsg, HistoryErrorMsg:
		if m.tabs.ActiveTab == 1 {
			m.history, cmd = m.history.Update(msg)
			cmds = append(cmds, cmd)
		}

	case searchResultsMsg, searchErrorMsg:
		if m.tabs.ActiveTab == 0 {
			m.search, cmd = m.search.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	if m.width < MIN_WIDTH || m.height < MIN_HEIGHT {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Terminal too small")
	}

	availableWidth := m.width - m.styles.App.GetHorizontalFrameSize()
	playerHeight := 4
	helpHeight := 1
	tabsHeight := lipgloss.Height(m.tabs.View())
	mainContentHeight := m.height - playerHeight - helpHeight - tabsHeight - m.styles.App.GetVerticalFrameSize()

	var mainContent string
	switch m.tabs.ActiveTab {
	case 0:
		m.search.SetSize(availableWidth-2, mainContentHeight-2)
		mainContent = m.search.View()
	case 1:
		m.history.SetSize(availableWidth-2, mainContentHeight-2)
		mainContent = m.history.View()
	}

	tabsView := m.tabs.View()
	mainPanel := m.styles.Box.Copy().UnsetBorderTop().Width(availableWidth).Height(mainContentHeight).Render(mainContent)
	mainView := lipgloss.JoinVertical(lipgloss.Left, tabsView, mainPanel)

	m.player.SetSize(availableWidth-2, playerHeight-2)
	playerContent := m.player.View()
	playerPanel := m.styles.Box.Width(availableWidth).Height(playerHeight).Render(playerContent)

	helpView := m.styles.Help.Width(availableWidth).Render("Help: [Tab] switch tab | [Enter] focus | [Space] play/pause | [←/→] seek | [q]uit")

	return m.styles.App.Render(lipgloss.JoinVertical(lipgloss.Top,
		mainView,
		playerPanel,
		helpView,
	))
}
