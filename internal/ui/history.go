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

type historyItem struct{ entry domain.HistoryEntry }

func (i historyItem) FilterValue() string { return i.entry.Song.Title }
func (i historyItem) SongID() string      { return i.entry.Song.ID }
func (i historyItem) ResumeAt() int       { return i.entry.ResumeAt }

type historyItemDelegate struct{ styles Styles }

func (d historyItemDelegate) Height() int                               { return 1 }
func (d historyItemDelegate) Spacing() int                              { return 0 }
func (d historyItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d historyItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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
	line := i.entry.Song.Title
	if m.Width() > 0 {
		lineWidth := m.Width() - lipgloss.Width(style.Render(pointer))
		if len(line) > lineWidth {
			line = line[:lineWidth-3] + "..."
		}
	}
	fmt.Fprint(w, style.Render(pointer+line))
}

type HistoryModel struct {
	storageService ports.StorageService
	config         domain.Config
	styles         Styles
	focus          searchComponentFocus
	textInput      textinput.Model
	resultsList    list.Model
	spinner        spinner.Model
	isLoading      bool
	err            error
	fullHistory    []list.Item
}

func NewHistoryModel(service ports.StorageService, cfg domain.Config, styles Styles) HistoryModel {
	ti := textinput.New()
	ti.Placeholder = "Filter history..."
	ti.Prompt = ""
	ti.PromptStyle = lipgloss.NewStyle()
	ti.TextStyle = lipgloss.NewStyle()

	li := list.New([]list.Item{}, historyItemDelegate{styles: styles}, 0, 0)
	li.SetShowTitle(false)
	li.SetShowStatusBar(false)
	li.SetShowPagination(false)
	li.SetShowHelp(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.Spinner

	return HistoryModel{
		storageService: service,
		config:         cfg,
		styles:         styles,
		focus:          inputFocus,
		textInput:      ti,
		resultsList:    li,
		spinner:        s,
	}
}

func (m *HistoryModel) GetResumeAt(songID string) int {
	for _, item := range m.fullHistory {
		if hi, ok := item.(historyItem); ok {
			if hi.SongID() == songID {
				return hi.ResumeAt()
			}
		}
	}
	return 0
}

func (m *HistoryModel) loadHistory() tea.Msg {
	entries, err := m.storageService.GetHistory(m.config.HistoryLimit)
	if err != nil {
		return ports.HistoryErrorMsg{Err: err}
	}
	return ports.HistoryLoadedMsg{Entries: entries}
}

func (m *HistoryModel) Init() tea.Cmd {
	m.isLoading = true
	return tea.Batch(m.spinner.Tick, m.loadHistory)
}

func (m *HistoryModel) Focus() tea.Cmd {
	m.focus = inputFocus
	return m.textInput.Focus()
}

func (m *HistoryModel) Blur() {
	m.focus = inputFocus
	m.textInput.Blur()
}

func (m *HistoryModel) GetFocus() searchComponentFocus { return m.focus }
func (m *HistoryModel) SetSize(w, h int) {
	m.textInput.Width = w - 2
	m.resultsList.SetSize(w-2, h-2)
}

func (m HistoryModel) Update(msg tea.Msg) (HistoryModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ports.HistoryLoadedMsg:
		m.isLoading = false
		items := make([]list.Item, len(msg.Entries))
		for i, entry := range msg.Entries {
			items[i] = historyItem{entry: entry}
		}
		m.fullHistory = items
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
		filterTerm := m.textInput.Value()
		var filteredItems []list.Item
		if filterTerm == "" {
			filteredItems = m.fullHistory
		} else {
			for _, item := range m.fullHistory {
				if strings.Contains(strings.ToLower(item.FilterValue()), strings.ToLower(filterTerm)) {
					filteredItems = append(filteredItems, item)
				}
			}
		}
		cmds = append(cmds, m.resultsList.SetItems(filteredItems))
		cmds = append(cmds, cmd)

	case listFocus:
		switch key := msg.(type) {
		case tea.KeyMsg:
			if key.String() == "enter" {
				if selectedItem, ok := m.resultsList.SelectedItem().(historyItem); ok {
					return m, func() tea.Msg { return ports.PlaySongMsg{Song: selectedItem.entry.Song} }
				}
			}
		}
		m.resultsList, cmd = m.resultsList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m HistoryModel) View() (string, string) {
	var mainView string
	if m.isLoading {
		mainView = m.spinner.View() + " Loading history..."
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
