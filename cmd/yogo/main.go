package main

import (
	"fmt"
	"os"

	"yogo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	initialModel := ui.InitialModel()

	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "¡Oh no! Hubo un error: %v\n", err)
		os.Exit(1)
	}
}
