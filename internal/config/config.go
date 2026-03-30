package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	ServerAddr string

	// Spotify OAuth configuration
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURL  string

	// YouTube Music OAuth configuration
	YouTubeMusicClientID     string
	YouTubeMusicClientSecret string
	YouTubeMusicRedirectURL  string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		ServerAddr:                  getEnv("SERVER_ADDR", ":8080"),
		SpotifyClientID:             os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret:         os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURL:          os.Getenv("SPOTIFY_REDIRECT_URL"),
		YouTubeMusicClientID:        os.Getenv("YOUTUBE_MUSIC_CLIENT_ID"),
		YouTubeMusicClientSecret:    os.Getenv("YOUTUBE_MUSIC_CLIENT_SECRET"),
		YouTubeMusicRedirectURL:     os.Getenv("YOUTUBE_MUSIC_REDIRECT_URL"),
	}

	return cfg, nil
}

// ValidateSpotify validates Spotify configuration
// Returns true if Spotify is configured, false if not configured, error if partially configured
func (c *Config) ValidateSpotify() (bool, error) {
	hasClientID := c.SpotifyClientID != ""
	hasClientSecret := c.SpotifyClientSecret != ""
	hasRedirectURL := c.SpotifyRedirectURL != ""

	// If none are set, Spotify is simply not configured
	if !hasClientID && !hasClientSecret && !hasRedirectURL {
		return false, nil
	}

	// If some but not all are set, this is an error
	if !hasClientID {
		return false, fmt.Errorf("SPOTIFY_CLIENT_ID is required when Spotify is configured")
	}
	if !hasClientSecret {
		return false, fmt.Errorf("SPOTIFY_CLIENT_SECRET is required when Spotify is configured")
	}
	if !hasRedirectURL {
		return false, fmt.Errorf("SPOTIFY_REDIRECT_URL is required when Spotify is configured")
	}

	return true, nil
}

// ValidateYouTubeMusic validates YouTube Music configuration
// Returns true if YouTube Music is configured, false if not configured, error if partially configured
func (c *Config) ValidateYouTubeMusic() (bool, error) {
	hasClientID := c.YouTubeMusicClientID != ""
	hasClientSecret := c.YouTubeMusicClientSecret != ""
	hasRedirectURL := c.YouTubeMusicRedirectURL != ""

	// If none are set, YouTube Music is simply not configured
	if !hasClientID && !hasClientSecret && !hasRedirectURL {
		return false, nil
	}

	// If some but not all are set, this is an error
	if !hasClientID {
		return false, fmt.Errorf("YOUTUBE_MUSIC_CLIENT_ID is required when YouTube Music is configured")
	}
	if !hasClientSecret {
		return false, fmt.Errorf("YOUTUBE_MUSIC_CLIENT_SECRET is required when YouTube Music is configured")
	}
	if !hasRedirectURL {
		return false, fmt.Errorf("YOUTUBE_MUSIC_REDIRECT_URL is required when YouTube Music is configured")
	}

	return true, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
