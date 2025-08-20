package ui

import (
	"github.com/charmbracelet/lipgloss"
)

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	highlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Copy().Border(activeTabBorder, true)
)

type TabModel struct {
	Tabs      []string
	ActiveTab int
}

func NewTabModel() TabModel {
	return TabModel{
		Tabs:      []string{"Search", "History"},
		ActiveTab: 0,
	}
}

func (m *TabModel) Next() {
	m.ActiveTab = (m.ActiveTab + 1) % len(m.Tabs)
}

func (m *TabModel) Prev() {
	m.ActiveTab--
	if m.ActiveTab < 0 {
		m.ActiveTab = len(m.Tabs) - 1
	}
}

func (m TabModel) View() string {
	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		if i == m.ActiveTab {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
}
