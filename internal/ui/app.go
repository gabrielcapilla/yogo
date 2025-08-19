package ui

import (
	"yogo/internal/domain"
	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	MIN_WIDTH  = 50
	MIN_HEIGHT = 15
)

type AppModel struct {
	width, height  int
	youtubeService ports.YoutubeService
	playerService  ports.PlayerService
	search         SearchModel
	player         PlayerModel
	styles         Styles
}

func InitialModel(ytService ports.YoutubeService, pService ports.PlayerService) AppModel {
	styles := DefaultStyles()
	return AppModel{
		youtubeService: ytService,
		playerService:  pService,
		search:         NewSearchModel(ytService, styles),
		player:         NewPlayerModel(styles),
		styles:         styles,
	}
}

func (m AppModel) Init() tea.Cmd { return m.search.Init() }

func getStreamURLCmd(service ports.YoutubeService, song domain.Song) tea.Cmd {
	return func() tea.Msg {
		url, err := service.GetStreamURL(song.ID)
		if err != nil {
			return playErrorMsg{err}
		}
		return streamURLFetchedMsg{song: song, url: url}
	}
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case playSongMsg:
		m.player.SetContent("Cargando", msg.song, nil)
		return m, getStreamURLCmd(m.youtubeService, msg.song)
	case streamURLFetchedMsg:
		err := m.playerService.Play(msg.url)
		if err != nil {
			return m, func() tea.Msg { return playErrorMsg{err} }
		}
		return m, func() tea.Msg { return songNowPlayingMsg{song: msg.song} }
	case songNowPlayingMsg:
		m.player.SetContent("Reproduciendo", msg.song, nil)
		return m, nil
	case playErrorMsg:
		m.player.SetContent("Error", domain.Song{}, msg.err)
		return m, nil
	}

	m.search, cmd = m.search.Update(msg)
	cmds = append(cmds, cmd)
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

	m.search.SetSize(mainInnerWidth, mainInnerHeight)
	m.player.SetSize(playerInnerWidth, playerInnerHeight)

	searchContent := m.search.View()

	searchBoxStyle := m.styles.Box

	mainPanel := searchBoxStyle.Width(availableWidth).Height(mainHeight).Render(searchContent)

	playerContent := m.player.View()
	playerPanel := m.styles.Box.Width(availableWidth).Height(playerHeight).Render(playerContent)

	helpView := m.styles.Help.Width(availableWidth).Render("Ayuda: [↑/↓/Tab] navegar | [Enter] seleccionar | [q] salir")

	return m.styles.App.Render(lipgloss.JoinVertical(lipgloss.Top,
		mainPanel,
		playerPanel,
		helpView,
	))
}
