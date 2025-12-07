package providers

import (
	"fmt"
	"time"

	"github.com/JanikSachs/PlayPort/internal/models"
)

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	name         string
	authenticated bool
	playlists    []models.Playlist
}

// NewMockProvider creates a new mock provider with sample data
func NewMockProvider() *MockProvider {
	return &MockProvider{
		name:         "Mock Music",
		authenticated: false,
		playlists: []models.Playlist{
			{
				ID:          "mock-1",
				Name:        "Summer Vibes 2024",
				Description: "Perfect tunes for summer",
				TrackCount:  15,
				Provider:    "Mock Music",
				CreatedAt:   time.Now().AddDate(0, -2, 0),
				UpdatedAt:   time.Now(),
				Tracks: []models.Track{
					{
						ID:       "track-1",
						Title:    "Sunshine Day",
						Artist:   "The Happy Band",
						Album:    "Good Times",
						Duration: 180,
						ISRC:     "MOCK12345001",
					},
					{
						ID:       "track-2",
						Title:    "Beach Walk",
						Artist:   "Ocean Sounds",
						Album:    "Coastal Dreams",
						Duration: 240,
						ISRC:     "MOCK12345002",
					},
					{
						ID:       "track-3",
						Title:    "Summer Breeze",
						Artist:   "Wind Chasers",
						Album:    "Season Collection",
						Duration: 195,
						ISRC:     "MOCK12345003",
					},
				},
			},
			{
				ID:          "mock-2",
				Name:        "Workout Mix",
				Description: "High energy tracks to keep you moving",
				TrackCount:  20,
				Provider:    "Mock Music",
				CreatedAt:   time.Now().AddDate(0, -1, 0),
				UpdatedAt:   time.Now(),
				Tracks: []models.Track{
					{
						ID:       "track-4",
						Title:    "Power Up",
						Artist:   "Energy Squad",
						Album:    "Motivation",
						Duration: 210,
						ISRC:     "MOCK12345004",
					},
					{
						ID:       "track-5",
						Title:    "Push Harder",
						Artist:   "Fitness Beats",
						Album:    "Gym Anthems",
						Duration: 195,
						ISRC:     "MOCK12345005",
					},
				},
			},
			{
				ID:          "mock-3",
				Name:        "Chill Evening",
				Description: "Relaxing music for winding down",
				TrackCount:  12,
				Provider:    "Mock Music",
				CreatedAt:   time.Now().AddDate(0, 0, -15),
				UpdatedAt:   time.Now(),
				Tracks: []models.Track{
					{
						ID:       "track-6",
						Title:    "Moonlight",
						Artist:   "Ambient Dreams",
						Album:    "Night Sky",
						Duration: 300,
						ISRC:     "MOCK12345006",
					},
				},
			},
		},
	}
}

// Name returns the provider's name
func (m *MockProvider) Name() string {
	return m.name
}

// Authenticate simulates authentication
func (m *MockProvider) Authenticate() error {
	m.authenticated = true
	return nil
}

// GetPlaylists returns mock playlists
func (m *MockProvider) GetPlaylists() ([]models.Playlist, error) {
	if !m.authenticated {
		return nil, fmt.Errorf("not authenticated")
	}
	return m.playlists, nil
}

// ExportPlaylist exports a specific playlist by ID
func (m *MockProvider) ExportPlaylist(id string) (models.Playlist, error) {
	if !m.authenticated {
		return models.Playlist{}, fmt.Errorf("not authenticated")
	}

	for _, playlist := range m.playlists {
		if playlist.ID == id {
			return playlist, nil
		}
	}

	return models.Playlist{}, fmt.Errorf("playlist not found: %s", id)
}

// ImportPlaylist simulates importing a playlist
func (m *MockProvider) ImportPlaylist(p models.Playlist) error {
	if !m.authenticated {
		return fmt.Errorf("not authenticated")
	}

	// Simulate adding the playlist with a new ID
	newPlaylist := p
	newPlaylist.ID = fmt.Sprintf("mock-imported-%d", time.Now().Unix())
	newPlaylist.Provider = m.name
	newPlaylist.CreatedAt = time.Now()
	newPlaylist.UpdatedAt = time.Now()

	m.playlists = append(m.playlists, newPlaylist)
	return nil
}
