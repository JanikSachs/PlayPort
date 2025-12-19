package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/JanikSachs/PlayPort/internal/providers/spotify"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

// ProviderHandlers contains provider-specific handlers
type ProviderHandlers struct {
	spotifyProvider    *spotify.SpotifyProvider
	connectionStore    storage.ConnectionStore
	templates          *template.Template
	spotifyEnabled     bool
}

// NewProviderHandlers creates new provider handlers
func NewProviderHandlers(spotifyProvider *spotify.SpotifyProvider, connectionStore storage.ConnectionStore, templates *template.Template, spotifyEnabled bool) *ProviderHandlers {
	return &ProviderHandlers{
		spotifyProvider: spotifyProvider,
		connectionStore: connectionStore,
		templates:       templates,
		spotifyEnabled:  spotifyEnabled,
	}
}

// HandleSpotifyPlaylists returns playlists for the Spotify provider
func (h *ProviderHandlers) HandleSpotifyPlaylists(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyEnabled {
		http.Error(w, "Spotify is not configured", http.StatusServiceUnavailable)
		return
	}

	// Check authentication
	if err := h.spotifyProvider.Authenticate(); err != nil {
		log.Printf("Spotify not authenticated: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`<div class="notification is-warning">
			<p>Please connect your Spotify account first.</p>
			<a href="/auth/spotify/start" class="button is-primary mt-2">Connect Spotify</a>
		</div>`))
		return
	}

	// Get playlists
	playlists, err := h.spotifyProvider.GetPlaylists()
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
