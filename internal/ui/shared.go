package ui

import (
	"fmt"
	"io"
	"strings"
	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type componentFocus int

const (
	inputFocus componentFocus = iota
	listFocus
)

type listItem interface {
	list.Item
	ID() string
	ToSong() domain.Song
}

type listDataSource interface {
	Fetch(query string) tea.Msg
}

type listAndFilterModel struct {
	title             string
	dataSource        listDataSource
	styles            Styles
	focus             componentFocus
	textInput         textinput.Model
	resultsList       list.Model
	spinner           spinner.Model
	isLoading         bool
	err               error
	fullList          []list.Item
	markedForDeletion map[string]struct{}
}

type itemDelegate struct {
	styles            Styles
	markedForDeletion *map[string]struct{}
}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	listItem, ok := item.(listItem)
	if !ok {
		return
	}

	itemStyle := d.styles.ListNormal
	pointer := "  "
	if index == m.Index() {
		itemStyle = d.styles.ListSelected
		pointer = d.styles.ListPointer.String()
	}

	var lineBuilder strings.Builder
	if _, isMarked := (*d.markedForDeletion)[listItem.ID()]; isMarked {
		lineBuilder.WriteString("x ")
		itemStyle = itemStyle.Strikethrough(true).Faint(true)
	} else {
		lineBuilder.WriteString("")
	}

	lineBuilder.WriteString(listItem.FilterValue())
	line := lineBuilder.String()

	if m.Width() > 0 {
		lineWidth := m.Width() - lipgloss.Width(itemStyle.Render(pointer))
		if len(line) > lineWidth {
			line = line[:lineWidth-3] + "..."
		}
	}
	fmt.Fprint(w, itemStyle.Render(pointer+line))
}

func NewListAndFilterModel(title, placeholder string, source listDataSource, styles Styles) listAndFilterModel {
	m := listAndFilterModel{
		title:             title,
		dataSource:        source,
		styles:            styles,
		focus:             inputFocus,
		markedForDeletion: make(map[string]struct{}),
	}

	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = ""
	ti.PromptStyle = lipgloss.NewStyle()
	ti.TextStyle = lipgloss.NewStyle()
	m.textInput = ti

	delegate := itemDelegate{
		styles:            styles,
		markedForDeletion: &m.markedForDeletion,
	}
	li := list.New([]list.Item{}, delegate, 0, 0)
	li.SetShowTitle(false)
	li.SetShowStatusBar(false)
	li.SetShowPagination(false)
	li.SetShowHelp(false)
	m.resultsList = li

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner
	m.spinner = s

	return m
}

func (m *listAndFilterModel) Init() tea.Cmd {
	m.isLoading = true
	m.markedForDeletion = make(map[string]struct{})
	fetchCmd := func() tea.Msg {
		return m.dataSource.Fetch("")
	}
	return tea.Batch(m.spinner.Tick, fetchCmd)
}

func (m *listAndFilterModel) Focus() tea.Cmd {
	m.focus = inputFocus
	return m.textInput.Focus()
}

func (m *listAndFilterModel) Blur() {
	m.focus = inputFocus
	m.textInput.Blur()
}

func (m *listAndFilterModel) GetFocus() componentFocus { return m.focus }
func (m *listAndFilterModel) SetSize(w, h int) {
	m.textInput.Width = w - 2
	m.resultsList.SetSize(w-2, h-2)
}

func (m *listAndFilterModel) GetItem(id string) list.Item {
	for _, item := range m.fullList {
		if li, ok := item.(listItem); ok && li.ID() == id {
			return item
		}
	}
	return nil
}

func (m listAndFilterModel) Update(msg tea.Msg) (listAndFilterModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ports.SearchResultsMsg:
		m.isLoading = false
		items := make([]list.Item, len(msg.Songs))
		for i, song := range msg.Songs {
			items[i] = searchItem{song: song}
		}
		m.resultsList.SetItems(items)
		return m, nil
	case ports.SearchErrorMsg:
		m.isLoading = false
		m.err = msg.Err
		return m, nil
	case ports.HistoryLoadedMsg:
		m.isLoading = false
		items := make([]list.Item, len(msg.Entries))
		for i, entry := range msg.Entries {
			items[i] = historyItem{entry: entry}
		}
		m.fullList = items
		m.resultsList.SetItems(items)
		return m, nil
	case ports.HistoryErrorMsg:
		m.isLoading = false
		m.err = msg.Err
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
			return m, func() tea.Msg { return ports.ChangeFocusMsg{NewFocus: ports.GlobalFocus} }
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
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

		isSearch := m.title == "search"
		if isSearch {
			if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
				if m.isLoading || m.textInput.Value() == "" {
					return m, nil
				}
				m.isLoading = true
				m.err = nil
				m.resultsList.SetItems([]list.Item{})
				m.focus = listFocus
				m.textInput.Blur()
				cmds = append(cmds, func() tea.Msg {
					return m.dataSource.Fetch(m.textInput.Value())
				})
			}
		} else {
			filterTerm := m.textInput.Value()
			var filteredItems []list.Item
			if filterTerm == "" {
				filteredItems = m.fullList
			} else {
				for _, item := range m.fullList {
					if strings.Contains(strings.ToLower(item.FilterValue()), strings.ToLower(filterTerm)) {
						filteredItems = append(filteredItems, item)
					}
				}
			}
			cmds = append(cmds, m.resultsList.SetItems(filteredItems))
		}

	case listFocus:
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "enter":
				if selectedItem, ok := m.resultsList.SelectedItem().(listItem); ok {
					return m, func() tea.Msg { return ports.PlaySongMsg{Song: selectedItem.ToSong()} }
				}
			case "x":
				if m.title == "history" {
					if selectedItem, ok := m.resultsList.SelectedItem().(listItem); ok {
						songID := selectedItem.ID()
						if _, isMarked := m.markedForDeletion[songID]; isMarked {
							delete(m.markedForDeletion, songID)
						} else {
							m.markedForDeletion[songID] = struct{}{}
						}
						cmd = m.resultsList.SetItems(m.resultsList.Items())
						return m, cmd
					}
				}
			case "d":
				if m.title == "history" && len(m.markedForDeletion) > 0 {
					ids := make([]string, 0, len(m.markedForDeletion))
					for id := range m.markedForDeletion {
						ids = append(ids, id)
					}
					return m, func() tea.Msg { return ports.DeleteFromHistoryMsg{SongIDs: ids} }
				}
			}
		}
		m.resultsList, cmd = m.resultsList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m listAndFilterModel) View() (string, string) {
	var mainView string
	if m.isLoading {
		mainView = m.spinner.View() + " Loading..."
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
