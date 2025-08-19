package main

import (
	"fmt"
	"os"

	_ "yogo/internal/logger"
	"yogo/internal/services/player"
	"yogo/internal/services/youtube"
	"yogo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	mpvSocketPath := "/tmp/mpvsocket"

	youtubeService := youtube.NewYTDLPClient()
	playerService := player.NewMpvPlayer(mpvSocketPath)

	initialModel := ui.InitialModel(youtubeService, playerService)

	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "¡Oh no! Hubo un error: %v\n", err)
		os.Exit(1)
	}
}
