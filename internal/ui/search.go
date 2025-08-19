package ui

import (
	"fmt"

	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type searchItem struct {
	song domain.Song
}

func (i searchItem) FilterValue() string { return i.song.Title }

func (i searchItem) Title() string { return i.song.Title }
func (i searchItem) Description() string {
	if len(i.song.Artists) > 0 {
		return fmt.Sprintf("Por: %s", i.song.Artists[0])
	}
	return "Artista desconocido"
}

type SearchModel struct {
	youtubeService ports.YoutubeService
	textInput      textinput.Model
	resultsList    list.Model
	isLoading      bool
	err            error
	width, height  int
}

func NewSearchModel(service ports.YoutubeService) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Lofi hip hop radio..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	li := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	li.Title = "Resultados de la Búsqueda"
	li.SetShowStatusBar(false)
	li.SetFilteringEnabled(false)

	return SearchModel{
		youtubeService: service,
		textInput:      ti,
		resultsList:    li,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) performSearch() tea.Msg {
	songs, err := m.youtubeService.Search(m.textInput.Value())
	if err != nil {
		return searchErrorMsg{err}
	}
	return searchResultsMsg{songs}
}

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resultsList.SetWidth(msg.Width - 2)
		m.resultsList.SetHeight(msg.Height - 10)
		return m, nil

	case searchResultsMsg:
		m.isLoading = false
		items := make([]list.Item, len(msg.songs))
		for i, song := range msg.songs {
			items[i] = searchItem{song: song}
		}
		m.resultsList.SetItems(items)
		return m, nil

	case searchErrorMsg:
		m.isLoading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.textInput.Focused() {
			switch msg.String() {
			case "tab":
				m.textInput.Blur()
			case "enter":
				m.isLoading = true
				m.err = nil
				m.textInput.Blur()
				return m, m.performSearch
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				cmds = append(cmds, cmd)
			}
		} else {
			switch msg.String() {
			case "tab":
				m.textInput.Focus()
			case "enter":
				return m, nil
			default:
				m.resultsList, cmd = m.resultsList.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m SearchModel) View() string {
	searchViewStyle := lipgloss.NewStyle().Padding(1, 2)
	inputView := m.textInput.View()
	var mainView string

	if m.isLoading {
		mainView = "Buscando..."
	} else if m.err != nil {
		mainView = fmt.Sprintf("Error: %v", m.err)
	} else {
		mainView = m.resultsList.View()
	}

	return searchViewStyle.Render(lipgloss.JoinVertical(lipgloss.Left, inputView, mainView))
}
