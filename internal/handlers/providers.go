package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/JanikSachs/PlayPort/internal/middleware"
	"github.com/JanikSachs/PlayPort/internal/providers/spotify"
	"github.com/JanikSachs/PlayPort/internal/providers/youtubemusic"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

// ProviderHandlers contains provider-specific handlers
type ProviderHandlers struct {
	spotifyProvider      *spotify.SpotifyProvider
	youtubeMusicProvider *youtubemusic.YouTubeMusicProvider
	connectionStore      storage.ConnectionStore
	templates            *template.Template
	spotifyEnabled       bool
	youtubeMusicEnabled  bool
}

// NewProviderHandlers creates new provider handlers
func NewProviderHandlers(spotifyProvider *spotify.SpotifyProvider, youtubeMusicProvider *youtubemusic.YouTubeMusicProvider, connectionStore storage.ConnectionStore, templates *template.Template, spotifyEnabled bool, youtubeMusicEnabled bool) *ProviderHandlers {
	return &ProviderHandlers{
		spotifyProvider:     spotifyProvider,
		youtubeMusicProvider: youtubeMusicProvider,
		connectionStore:     connectionStore,
		templates:           templates,
		spotifyEnabled:      spotifyEnabled,
		youtubeMusicEnabled: youtubeMusicEnabled,
	}
}

// HandleSpotifyPlaylists returns playlists for the Spotify provider
func (h *ProviderHandlers) HandleSpotifyPlaylists(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyEnabled {
		http.Error(w, "Spotify is not configured", http.StatusServiceUnavailable)
		return
	}

	// Check authentication
	if err := h.spotifyProvider.Authenticate(middleware.UserIDFromContext(r.Context())); err != nil {
		log.Printf("Spotify not authenticated: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		if err := h.templates.ExecuteTemplate(w, "spotify-not-connected.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Please connect your Spotify account first", http.StatusUnauthorized)
		}
		return
	}

	// Get playlists
	playlists, err := h.spotifyProvider.GetPlaylists(middleware.UserIDFromContext(r.Context()))
	if err != nil {
		log.Printf("Failed to fetch Spotify playlists: %v", err)
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	// Render playlist list template
	data := map[string]interface{}{
		"Playlists": playlists,
		"Provider":  "Spotify",
	}

	if err := h.templates.ExecuteTemplate(w, "playlist-list.html", data); err != nil {
		log.Printf("Error rendering playlist list: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleYouTubeMusicPlaylists returns playlists for the YouTube Music provider
func (h *ProviderHandlers) HandleYouTubeMusicPlaylists(w http.ResponseWriter, r *http.Request) {
	if !h.youtubeMusicEnabled {
		http.Error(w, "YouTube Music is not configured", http.StatusServiceUnavailable)
		return
	}

	// Check authentication
	if err := h.youtubeMusicProvider.Authenticate(middleware.UserIDFromContext(r.Context())); err != nil {
		log.Printf("YouTube Music not authenticated: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		if err := h.templates.ExecuteTemplate(w, "youtubemusic-not-connected.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Please connect your YouTube Music account first", http.StatusUnauthorized)
		}
		return
	}

	// Get playlists
	playlists, err := h.youtubeMusicProvider.GetPlaylists(middleware.UserIDFromContext(r.Context()))
	if err != nil {
		log.Printf("Failed to fetch YouTube Music playlists: %v", err)
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	// Render playlist list template
	data := map[string]interface{}{
		"Playlists": playlists,
		"Provider":  "YouTube Music",
	}

	if err := h.templates.ExecuteTemplate(w, "playlist-list.html", data); err != nil {
		log.Printf("Error rendering playlist list: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// GetConnectionStatus returns the Spotify connection status
func (h *ProviderHandlers) GetConnectionStatus(userID string) (bool, string) {
	if !h.spotifyEnabled {
		return false, ""
	}

	conn, err := h.connectionStore.Get("spotify", userID)
	if err != nil {
		return false, ""
	}

	if conn.Connected {
		return true, conn.ExternalUserName
	}

	return false, ""
}

// GetYouTubeMusicConnectionStatus returns the YouTube Music connection status
func (h *ProviderHandlers) GetYouTubeMusicConnectionStatus(userID string) (bool, string) {
	if !h.youtubeMusicEnabled {
		return false, ""
	}

	conn, err := h.connectionStore.Get("youtubemusic", userID)
	if err != nil {
		return false, ""
	}

	if conn.Connected {
		return true, conn.ExternalUserName
	}

	return false, ""
}

