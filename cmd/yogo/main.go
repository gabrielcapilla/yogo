// cmd/yogo/main.go
package main

import (
	"fmt"
	"os"
	"yogo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(ui.InitialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error ejecutando el programa: %v\n", err)
		os.Exit(1)
	}
}
