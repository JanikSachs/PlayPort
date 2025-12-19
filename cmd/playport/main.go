package main

import (
	"log"

	"github.com/JanikSachs/PlayPort/internal/providers"
	"github.com/JanikSachs/PlayPort/internal/server"
	"github.com/JanikSachs/PlayPort/internal/services"
)

func main() {
	// Create transfer service
	transferService := services.NewTransferService()

	// Register mock provider
	mockProvider := providers.NewMockProvider()
	transferService.RegisterProvider(mockProvider)

	// You can register additional providers here in the future
	// spotifyProvider := providers.NewSpotifyProvider()
	// transferService.RegisterProvider(spotifyProvider)

	// Create and start server
	srv, err := server.New(":8080", transferService)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Println("Starting PlayPort server on http://localhost:8080")
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
