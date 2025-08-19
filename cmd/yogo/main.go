package main

import (
	"fmt"
	"os"
	"path/filepath"

	_ "yogo/internal/logger"
	"yogo/internal/services/player"
	"yogo/internal/services/storage"
	"yogo/internal/services/youtube"
	"yogo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	mpvSocketPath := "/tmp/mpvsocket"

	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error crítico: no se pudo encontrar el directorio de configuración: %v\n", err)
		os.Exit(1)
	}
	yogoDir := filepath.Join(configDir, "yogo")
	dbPath := filepath.Join(yogoDir, "yogo.db")

	cookiesPath := os.Getenv("YOGO_COOKIES_PATH")

	youtubeService := youtube.NewYTDLPClient(cookiesPath)
	playerService := player.NewMpvPlayer(mpvSocketPath)
	storageService, err := storage.NewBboltStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error crítico: no se pudo inicializar la base de datos: %v\n", err)
		os.Exit(1)
	}
	defer storageService.Close()

	initialModel := ui.InitialModel(youtubeService, playerService, storageService)

	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "¡Oh no! Hubo un error: %v\n", err)
		os.Exit(1)
	}
}
