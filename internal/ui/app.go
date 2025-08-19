package ui

import (
	"fmt"

	"yogo/internal/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState int

const (
	globalView viewState = iota
	searchView
	historyView
)

type AppModel struct {
	state          viewState
	width          int
	height         int
	styles         Styles
	youtubeService ports.YoutubeService
	search         SearchModel
}

func InitialModel(service ports.YoutubeService) AppModel {
	return AppModel{
		state:          globalView,
		styles:         DefaultStyles(),
		youtubeService: service,
		search:         NewSearchModel(service),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.search.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	switch m.state {
	case globalView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			return m.updateGlobal(keyMsg)
		}
	case searchView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
			m.state = globalView
			return m, nil
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		return m, cmd
	case historyView:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
			m.state = globalView
			return m, nil
		}
	}

	return m, nil
}

func (m *AppModel) updateGlobal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "s":
		m.state = searchView
		return m, m.search.textInput.Focus()
	case "h":
		m.state = historyView
		return m, nil
	}
	return m, nil
}

func (m AppModel) View() string {
	if m.width == 0 {
		return "Inicializando..."
	}

	mainContentHeight := m.height - m.styles.TopBar.GetHeight() - m.styles.PlayerBar.GetHeight()

	topBar := m.renderTopBar()
	mainContent := m.renderMainContent(mainContentHeight)
	playerBar := m.renderPlayerBar()

	return lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		mainContent,
		playerBar,
	)
}

func (m AppModel) renderTopBar() string {
	stateStr := "Global"
	switch m.state {
	case searchView:
		stateStr = "Search"
	case historyView:
		stateStr = "History"
	}
	title := fmt.Sprintf("yogo | Current Mode: %s", stateStr)
	return m.styles.TopBar.Render(title)
}

func (m AppModel) renderMainContent(height int) string {
	var content string

	switch m.state {
	case searchView:
		content = m.search.View()
	default:
		helpText := "Navegación: [s]earch | [h]istory | [q]uit"
		content = lipgloss.Place(
			m.width-2, height-2,
			lipgloss.Center, lipgloss.Center,
			helpText,
		)
	}

	mainStyle := m.styles.MainContent.
		Width(m.width - 2).
		Height(height - 2)

	return mainStyle.Render(content)
}

func (m AppModel) renderPlayerBar() string {
	return m.styles.PlayerBar.Render("Player: [Idle]")
}

type Styles struct {
	TopBar      lipgloss.Style
	MainContent lipgloss.Style
	PlayerBar   lipgloss.Style
}

func DefaultStyles() Styles {
	return Styles{
		TopBar: lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230")).
			Padding(0, 1),
		MainContent: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
		PlayerBar: lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1),
	}
}
