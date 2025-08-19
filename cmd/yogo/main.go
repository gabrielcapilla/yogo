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

	"github.com/gofrs/flock"

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

	configDir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: could not find user config directory: %v\n", err)
		os.Exit(1)
	}
	yogoDir := filepath.Join(configDir, "yogo")
	// Ensure the directory exists before creating the lock file.
	if err := os.MkdirAll(yogoDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: could not create config directory: %v\n", err)
		os.Exit(1)
	}

	lockPath := filepath.Join(yogoDir, "yogo.lock")
	fileLock := flock.New(lockPath)
	locked, err := fileLock.TryLock()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: could not acquire file lock: %v\n", err)
		os.Exit(1)
	}

	if !locked {
		fmt.Fprintln(os.Stderr, "Error: Another instance of yogo is already running.")
		os.Exit(1)
	}
	defer fileLock.Unlock()

	configService := config.NewViperConfigService()
	cfg, err := configService.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: could not load config: %v\n", err)
		os.Exit(1)
	}

	mpvSocketPath := "/tmp/yogo-mpvsocket"
	dbPath := filepath.Join(yogoDir, "yogo.db")

	youtubeService := youtube.NewYTDLPClient(cfg.CookiesPath)
	playerService := player.NewMpvPlayer(mpvSocketPath)
	storageService, err := storage.NewBboltStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Critical error: could not initialize database: %v\n", err)
		os.Exit(1)
	}
	defer storageService.Close()
	defer playerService.Close()

	initialModel := ui.InitialModel(youtubeService, playerService, storageService, cfg)

	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oh no! There was an error: %v\n", err)
		os.Exit(1)
	}
}
