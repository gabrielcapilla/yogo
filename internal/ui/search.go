package ui

import (
	"fmt"
	"io"
	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type searchItem struct{ song domain.Song }

func (i searchItem) FilterValue() string { return i.song.Title }

type itemDelegate struct{ styles Styles }

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(searchItem)
	if !ok {
		return
	}
	style := d.styles.ListNormal
	pointer := "  "
	if index == m.Index() {
		style = d.styles.ListSelected
		pointer = d.styles.ListPointer.String()
	}
	line := i.song.Title
	if m.Width() > 0 {
		lineWidth := m.Width() - lipgloss.Width(style.Render(pointer))
		if len(line) > lineWidth {
			line = line[:lineWidth-3] + "..."
		}
	}
	fmt.Fprint(w, style.Render(pointer+line))
}

type searchComponentFocus int

const (
	inputFocus searchComponentFocus = iota
	listFocus
)

type SearchModel struct {
	youtubeService ports.YoutubeService
	styles         Styles
	focus          searchComponentFocus
	textInput      textinput.Model
	resultsList    list.Model
	spinner        spinner.Model
	isLoading      bool
	err            error
}

func NewSearchModel(service ports.YoutubeService, styles Styles) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search for a song..."
	ti.Prompt = ""
	ti.PromptStyle = lipgloss.NewStyle()
	ti.TextStyle = lipgloss.NewStyle()

	li := list.New([]list.Item{}, itemDelegate{styles: styles}, 0, 0)
	li.SetShowTitle(false)
	li.SetShowStatusBar(false)
	li.SetShowPagination(false)
	li.SetShowHelp(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	return SearchModel{
		youtubeService: service,
		styles:         styles,
		focus:          inputFocus,
		textInput:      ti,
		resultsList:    li,
		spinner:        s,
	}
}

func (m *SearchModel) Init() tea.Cmd { return m.spinner.Tick }
func (m *SearchModel) Focus() tea.Cmd {
	m.focus = inputFocus
	return m.textInput.Focus()
}
func (m *SearchModel) Blur() {
	m.focus = inputFocus
	m.textInput.Blur()
}
func (m *SearchModel) GetFocus() searchComponentFocus { return m.focus }
func (m *SearchModel) SetSize(w, h int) {
	m.textInput.Width = w - 2
	m.resultsList.SetSize(w-2, h-2)
}
func (m *SearchModel) performSearch() tea.Msg {
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
	}

	if m.isLoading {
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return changeFocusMsg{newFocus: globalFocus} }
		case "tab":
			if m.focus == inputFocus {
				m.focus = listFocus
				m.textInput.Blur()
			} else {
				m.focus = inputFocus
				m.textInput.Focus()
			}
			return m, nil
		}
	}

	switch m.focus {
	case inputFocus:
		switch key := msg.(type) {
		case tea.KeyMsg:
			if key.String() == "enter" {
				m.isLoading = true
				m.err = nil
				m.resultsList.SetItems([]list.Item{})
				m.focus = listFocus
				m.textInput.Blur()
				cmds = append(cmds, m.performSearch)
			}
		}
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	case listFocus:
		switch key := msg.(type) {
		case tea.KeyMsg:
			if key.String() == "enter" {
				if selectedItem, ok := m.resultsList.SelectedItem().(searchItem); ok {
					return m, func() tea.Msg { return playSongMsg(selectedItem) }
				}
			}
		}
		m.resultsList, cmd = m.resultsList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SearchModel) View() (string, string) {
	var mainView string
	if m.isLoading {
		mainView = m.spinner.View() + " Searching..."
	} else if m.err != nil {
		mainView = m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", m.err))
	} else {
		mainView = m.resultsList.View()
	}

	footerView := lipgloss.JoinVertical(lipgloss.Left,
		m.textInput.View(),
		"",
	)
	return mainView, footerView
}
