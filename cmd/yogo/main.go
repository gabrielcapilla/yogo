package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"yogo/internal/logger"
	"yogo/internal/services/config"
	"yogo/internal/services/player"
	"yogo/internal/services/storage"
	"yogo/internal/services/youtube"
	"yogo/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	logger.Setup(*debug)

	configService := config.NewViperConfigService()
	cfg, err := configService.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	ytService := youtube.NewYoutubeClient(cfg.CookiesPath)

	socketPath := filepath.Join(os.TempDir(), "yogo.sock")
	playerService := player.NewMpvPlayer(socketPath, cfg)

	configDir, _ := os.UserConfigDir()
	dbPath := filepath.Join(configDir, "yogo", "history.db")
	storageService, err := storage.NewBboltStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Initial database initialization failed.: %v\n", err)
		os.Exit(1)
	}

	defer func() {
		if err := playerService.Close(); err != nil {
			logger.Log.Error().Err(err).Msg("Error closing the player service")
		}
		if err := storageService.Close(); err != nil {
			logger.Log.Error().Err(err).Msg("Error closing storage service")
		}
	}()

	p := tea.NewProgram(ui.InitialModel(ytService, playerService, storageService, cfg), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing the program: %v\n", err)
		os.Exit(1)
	}
}
