package main

import (
	"log"

	"github.com/JanikSachs/PlayPort/internal/auth"
	"github.com/JanikSachs/PlayPort/internal/config"
	"github.com/JanikSachs/PlayPort/internal/providers"
	"github.com/JanikSachs/PlayPort/internal/providers/spotify"
	"github.com/JanikSachs/PlayPort/internal/server"
	"github.com/JanikSachs/PlayPort/internal/services"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate Spotify configuration
	spotifyEnabled, err := cfg.ValidateSpotify()
	if err != nil {
		log.Fatalf("Spotify configuration error: %v", err)
	}

	if spotifyEnabled {
		log.Println("Spotify integration enabled")
	} else {
		log.Println("Spotify integration disabled (environment variables not set)")
	}

	// Create storage
	connectionStore := storage.NewInMemoryConnectionStore()
	stateStore := auth.NewInMemoryStateStore()

	// Create transfer service
	transferService := services.NewTransferService()

	// Register mock provider
	mockProvider := providers.NewMockProvider()
	transferService.RegisterProvider(mockProvider)

	// Create Spotify provider if enabled
	var spotifyProvider *spotify.SpotifyProvider
	if spotifyEnabled {
		spotifyProvider = spotify.NewSpotifyProvider(
			cfg.SpotifyClientID,
			cfg.SpotifyClientSecret,
			cfg.SpotifyRedirectURL,
			connectionStore,
		)
		transferService.RegisterProvider(spotifyProvider)
	}

	// Create and start server
	srv, err := server.New(cfg.ServerAddr, transferService, spotifyProvider, connectionStore, stateStore, spotifyEnabled)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Printf("Starting PlayPort server on http://localhost%s", cfg.ServerAddr)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
