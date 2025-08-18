package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gabrielcapilla/yogo/internal/domain"
	"github.com/gabrielcapilla/yogo/internal/services/player"
	"github.com/gabrielcapilla/yogo/internal/services/storage"
	"github.com/gabrielcapilla/yogo/internal/services/youtube"
)

func main() {
	log.Println("Starting the walking skeleton of yogo...")

	mpvSocketPath := "/tmp/mpvsocket"

	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("Could not get the user's config directory: %v", err)
	}
	yogoDir := filepath.Join(configDir, "yogo")
	if err := os.MkdirAll(yogoDir, 0755); err != nil {
		log.Fatalf("Could not create the yogo directory: %v", err)
	}
	dbPath := filepath.Join(yogoDir, "yogo.db")

	youtubeService := youtube.NewYTDLPClient()
	playerService := player.NewMpvPlayer(mpvSocketPath)
	storageService, err := storage.NewBboltStore(dbPath)
	if err != nil {
		log.Fatalf("Error initializing the storage service: %v", err)
	}
	defer storageService.Close()

	query := "lofi hip hop radio"
	log.Printf("1. Searching for songs for: '%s'", query)

	songs, err := youtubeService.Search(query)
	if err != nil {
		log.Fatalf("Error in the search: %v", err)
	}
	if len(songs) == 0 {
		log.Fatal("The search returned no results.")
	}

	firstSong := songs[0]
	log.Printf("2. Song found: '%s' (ID: %s)", firstSong.Title, firstSong.ID)

	log.Printf("3. Getting stream URL for the song...")
	streamURL, err := youtubeService.GetStreamURL(firstSong.ID)
	if err != nil {
		log.Fatalf("Error getting the stream URL: %v", err)
	}
	log.Printf("   URL obtained: %s", streamURL)

	log.Println("4. Sending play command to mpv...")
	if err := playerService.Play(streamURL); err != nil {
		log.Fatalf("Error playing: %v", err)
	}
	log.Println("   The music should be playing!")

	log.Println("5. Saving the song to history...")
	historyEntry := domain.HistoryEntry{
		Song:     firstSong,
		PlayedAt: time.Now(),
	}
	if err := storageService.AddToHistory(historyEntry); err != nil {
		log.Fatalf("Error saving to history: %v", err)
	}
	log.Println("   Saved successfully.")

	log.Println("6. Retrieving history...")
	history, err := storageService.GetHistory(5)
	if err != nil {
		log.Fatalf("Error retrieving history: %v", err)
	}

	log.Println("--- Recent History ---")
	for i, entry := range history {
		fmt.Printf("%d: %s (%s)\n", i+1, entry.Song.Title, entry.PlayedAt.Format(time.RFC822))
	}
	log.Println("--------------------------")

	log.Println("The walking skeleton of yogo has completed its execution.")
}
