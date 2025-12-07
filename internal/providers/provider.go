package providers

import "github.com/JanikSachs/PlayPort/internal/models"

// Provider defines the interface that all music platform providers must implement
type Provider interface {
	// Name returns the provider's name (e.g., "Spotify", "Apple Music")
	Name() string

	// Authenticate authenticates with the provider's API
	Authenticate() error

	// GetPlaylists retrieves all playlists for the authenticated user
	GetPlaylists() ([]models.Playlist, error)

	// ExportPlaylist exports a specific playlist by ID
	ExportPlaylist(id string) (models.Playlist, error)

	// ImportPlaylist imports a playlist into the provider
	ImportPlaylist(p models.Playlist) error
}
