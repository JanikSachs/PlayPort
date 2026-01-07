package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/JanikSachs/PlayPort/internal/auth"
	"github.com/JanikSachs/PlayPort/internal/providers/spotify"
)

// AuthHandlers contains OAuth authentication handlers
type AuthHandlers struct {
	spotifyProvider *spotify.SpotifyProvider
	stateStore      auth.StateStore
	spotifyEnabled  bool
}

// NewAuthHandlers creates new auth handlers
func NewAuthHandlers(spotifyProvider *spotify.SpotifyProvider, stateStore auth.StateStore, spotifyEnabled bool) *AuthHandlers {
	return &AuthHandlers{
		spotifyProvider: spotifyProvider,
		stateStore:      stateStore,
		spotifyEnabled:  spotifyEnabled,
	}
}

// HandleSpotifyStart redirects to Spotify authorization
func (h *AuthHandlers) HandleSpotifyStart(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyEnabled {
		http.Error(w, "Spotify is not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate state for CSRF protection
	state, err := h.stateStore.Generate()
	if err != nil {
		log.Printf("Failed to generate state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to Spotify authorization
	authURL := h.spotifyProvider.AuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleSpotifyCallback handles the OAuth callback from Spotify
func (h *AuthHandlers) HandleSpotifyCallback(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyEnabled {
		http.Error(w, "Spotify is not configured", http.StatusServiceUnavailable)
		return
	}

	// Validate state
	state := r.URL.Query().Get("state")
	if !h.stateStore.Validate(state) {
		log.Printf("Invalid OAuth state")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Check for error from Spotify
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		log.Printf("Spotify OAuth error: %s", errMsg)
		http.Error(w, fmt.Sprintf("Spotify authorization failed: %s", errMsg), http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := h.spotifyProvider.Exchange(ctx, code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Save connection
	if err := h.spotifyProvider.SaveConnection(ctx, token); err != nil {
		log.Printf("Failed to save connection: %v", err)
		http.Error(w, "Failed to save connection", http.StatusInternalServerError)
		return
	}

	// Redirect to providers page
	http.Redirect(w, r, "/providers", http.StatusFound)
}
