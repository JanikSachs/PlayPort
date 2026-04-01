package providers

import "github.com/JanikSachs/PlayPort/internal/models"

// Provider defines the interface that all music platform providers must implement
type Provider interface {
	// Name returns the provider's name (e.g., "Spotify", "Apple Music")
	Name() string

	// Authenticate checks if the given user has a valid connection to the provider
	Authenticate(userID string) error

	// GetPlaylists retrieves all playlists for the given user
	GetPlaylists(userID string) ([]models.Playlist, error)

	// ExportPlaylist exports a specific playlist by ID for the given user
	ExportPlaylist(userID, id string) (models.Playlist, error)

	// ImportPlaylist imports a playlist into the provider
	ImportPlaylist(p models.Playlist) error
}
