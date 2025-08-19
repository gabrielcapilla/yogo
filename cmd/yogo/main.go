package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"yogo/internal/logger"
	"yogo/internal/services/player"
	"yogo/internal/services/storage"
	"yogo/internal/services/youtube"
	"yogo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("v", false, "print version and exit")
	debugFlag := flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	logger.Setup(*debugFlag)

	mpvSocketPath := "/tmp/yogo-mpvsocket"

	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: could not find config directory: %v\n", err)
		os.Exit(1)
	}
	yogoDir := filepath.Join(configDir, "yogo")
	dbPath := filepath.Join(yogoDir, "yogo.db")

	cookiesPath := os.Getenv("YOGO_COOKIES_PATH")

	youtubeService := youtube.NewYTDLPClient(cookiesPath)
	playerService := player.NewMpvPlayer(mpvSocketPath)
	storageService, err := storage.NewBboltStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: could not initialize database: %v\n", err)
		os.Exit(1)
	}
	defer storageService.Close()
	defer playerService.Close()

	initialModel := ui.InitialModel(youtubeService, playerService, storageService)

	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oh no! There was an error: %v\n", err)
		os.Exit(1)
	}
}
