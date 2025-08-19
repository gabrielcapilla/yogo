package ui

import (
	"fmt"
	"io"
	"yogo/internal/domain"
	"yogo/internal/ports"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type historyItem struct {
	entry domain.HistoryEntry
}

func (i historyItem) FilterValue() string { return i.entry.Song.Title }

type historyDelegate struct{ styles Styles }

func (d historyDelegate) Height() int                               { return 1 }
func (d historyDelegate) Spacing() int                              { return 0 }
func (d historyDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d historyDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(historyItem)
	if !ok {
		return
	}
	style := d.styles.ListNormal
	pointer := "  "
	if index == m.Index() {
		style = d.styles.ListSelected
		pointer = d.styles.ListPointer.String()
	}
	line := fmt.Sprintf("%s - %s", i.entry.Song.Title, i.entry.PlayedAt.Format("2006-01-02 15:04"))
	truncatedLine := truncate(line, m.Width()-2)
	fmt.Fprint(w, style.Render(pointer+truncatedLine))
}

type HistoryModel struct {
	width, height  int
	storageService ports.StorageService
	resultsList    list.Model
	spinner        spinner.Model
	isLoading      bool
	err            error
	styles         Styles
}

func NewHistoryModel(service ports.StorageService, styles Styles) HistoryModel {
	li := list.New([]list.Item{}, historyDelegate{styles: styles}, 0, 0)
	li.Title = "Playback History"
	li.Styles.Title = styles.BoxTitle
	li.SetShowStatusBar(false)
	li.SetShowPagination(false)
	li.SetShowHelp(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	return HistoryModel{
		storageService: service,
		resultsList:    li,
		spinner:        s,
		styles:         styles,
		isLoading:      true,
	}
}

func (m *HistoryModel) Init() tea.Cmd { return m.spinner.Tick }

func (m *HistoryModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.resultsList.SetSize(w, h)
}

func (m HistoryModel) Update(msg tea.Msg) (HistoryModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case HistoryLoadedMsg:
		m.isLoading = false
		items := make([]list.Item, len(msg.Entries))
		for i, entry := range msg.Entries {
			items[i] = historyItem{entry: entry}
		}
		m.resultsList.SetItems(items)
		return m, nil

	case HistoryErrorMsg:
		m.isLoading = false
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return changeFocusMsg{newFocus: globalFocus} }
		case "enter":
			if selectedItem, ok := m.resultsList.SelectedItem().(historyItem); ok {
				return m, func() tea.Msg {
					return playSongMsg{song: selectedItem.entry.Song}
				}
			}
		}
	}

	if m.isLoading {
		m.spinner, cmd = m.spinner.Update(msg)
	} else {
		m.resultsList, cmd = m.resultsList.Update(msg)
	}

	return m, cmd
}

func (m HistoryModel) View() string {
	if m.isLoading {
		return m.spinner.View() + " Loading history..."
	}
	if m.err != nil {
		return m.styles.ErrorText.Render(fmt.Sprintf("Error: %v", m.err))
	}
	return m.resultsList.View()
}
