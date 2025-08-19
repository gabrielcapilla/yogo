package ui

import (
	"fmt"
	"io"

	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
)

type searchItem struct {
	song domain.Song
}

func (i searchItem) FilterValue() string { return i.song.Title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(searchItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.song.Title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + lipgloss.JoinHorizontal(lipgloss.Left, s...))
		}
	}

	fmt.Fprint(w, fn(str))
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

	li := list.New([]list.Item{}, itemDelegate{}, 0, 0)
	li.Title = "Resultados de la Búsqueda"
	li.Styles.Title = titleStyle
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
				selectedItem, ok := m.resultsList.SelectedItem().(searchItem)
				if ok {
					return m, func() tea.Msg {
						return playSongMsg(selectedItem)
					}
				}
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
