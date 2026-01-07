package models

import "time"

// Track represents a single song in a playlist
type Track struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Album       string    `json:"album"`
	Duration    int       `json:"duration"` // in seconds
	ISRC        string    `json:"isrc"`     // International Standard Recording Code
	ReleaseDate time.Time `json:"release_date"`
}

// Playlist represents a music playlist from any platform
type Playlist struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tracks      []Track   `json:"tracks"`
	TrackCount  int       `json:"track_count"`
	Provider    string    `json:"provider"` // e.g., "Spotify", "Apple Music"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Connection represents a user's connection to a music platform
type Connection struct {
	ID               string    `json:"id"`
	Provider         string    `json:"provider"`           // e.g., "spotify"
	UserID           string    `json:"user_id"`            // Local app user ID
	ExternalUserID   string    `json:"external_user_id"`   // Provider's user ID
	ExternalUserName string    `json:"external_user_name"` // Provider's display name
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	Scopes           []string  `json:"scopes"`     // OAuth scopes granted
	Connected        bool      `json:"connected"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
