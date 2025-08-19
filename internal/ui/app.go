package ui

import (
	"time"
	"yogo/internal/domain"
	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MIN_WIDTH  = 82
	MIN_HEIGHT = 14
)

type viewState int

const (
	searchView viewState = iota
	historyView
)

type AppModel struct {
	state          viewState
	width, height  int
	youtubeService ports.YoutubeService
	playerService  ports.PlayerService
	storageService ports.StorageService
	search         SearchModel
	history        HistoryModel
	player         PlayerModel
	styles         Styles
}

func InitialModel(ytService ports.YoutubeService, pService ports.PlayerService, sService ports.StorageService) AppModel {
	styles := DefaultStyles()
	return AppModel{
		state:          searchView,
		youtubeService: ytService,
		playerService:  pService,
		storageService: sService,
		search:         NewSearchModel(ytService, styles),
		history:        NewHistoryModel(sService, styles),
		player:         NewPlayerModel(styles),
		styles:         styles,
	}
}

func (m AppModel) Init() tea.Cmd {
	return tea.Batch(m.search.Init(), m.history.Init())
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

func loadHistoryCmd(service ports.StorageService) tea.Cmd {
	return func() tea.Msg {
		entries, err := service.GetHistory(50)
		if err != nil {
			return HistoryErrorMsg{Err: err}
		}
		return HistoryLoadedMsg{Entries: entries}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch m.state {
	case searchView:
		m.search, cmd = m.search.Update(msg)
		cmds = append(cmds, cmd)
	case historyView:
		m.history, cmd = m.history.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "s":
			if m.state != searchView {
				m.state = searchView
			}
		case "h":
			if m.state != historyView {
				m.state = historyView
				cmds = append(cmds, loadHistoryCmd(m.storageService))
			}
		}
	case playSongMsg:
		m.player.SetContent("Cargando", msg.song, nil)
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
		m.player.SetContent("Reproduciendo", msg.song, nil)
	case playErrorMsg:
		m.player.SetContent("Error", domain.Song{}, msg.err)
	case HistoryLoadedMsg, HistoryErrorMsg:
	}

	m.player, cmd = m.player.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() string {
	if m.width < MIN_WIDTH || m.height < MIN_HEIGHT {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Terminal demasiado pequeña")
	}

	availableWidth := m.width - m.styles.App.GetHorizontalFrameSize()
	playerHeight := 3
	helpHeight := 1
	mainHeight := m.height - playerHeight - helpHeight - m.styles.App.GetVerticalFrameSize()
	mainInnerWidth := availableWidth - 2
	playerInnerWidth := availableWidth - 2
	mainInnerHeight := mainHeight - 2
	playerInnerHeight := playerHeight - 2

	var mainContent string
	switch m.state {
	case searchView:
		m.search.SetSize(mainInnerWidth, mainInnerHeight)
		mainContent = m.search.View()
	case historyView:
		m.history.SetSize(mainInnerWidth, mainInnerHeight)
		mainContent = m.history.View()
	}

	mainPanel := m.styles.Box.Width(availableWidth).Height(mainHeight).Render(mainContent)

	m.player.SetSize(playerInnerWidth, playerInnerHeight)
	playerContent := m.player.View()
	playerPanel := m.styles.Box.Width(availableWidth).Height(playerHeight).Render(playerContent)

	helpView := m.styles.Help.Width(availableWidth).Render("Ayuda: [s]earch | [h]istory | [↑/↓] navegar | [Enter] seleccionar | [q] salir")

	return m.styles.App.Render(lipgloss.JoinVertical(lipgloss.Top,
		mainPanel,
		playerPanel,
		helpView,
	))
}
