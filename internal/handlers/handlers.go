package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/JanikSachs/PlayPort/internal/services"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	transferService  *services.TransferService
	templates        *template.Template
	connectionStore  storage.ConnectionStore
	spotifyEnabled   bool
}

// NewHandlers creates a new Handlers instance
func NewHandlers(transferService *services.TransferService, templates *template.Template, connectionStore storage.ConnectionStore, spotifyEnabled bool) *Handlers {
	return &Handlers{
		transferService: transferService,
		templates:       templates,
		connectionStore: connectionStore,
		spotifyEnabled:  spotifyEnabled,
	}
}

// HandleHome renders the home page
func (h *Handlers) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title": "PlayPort - Transfer Your Playlists",
	}

	if err := h.templates.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("Error rendering home template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleProviders renders the providers page
func (h *Handlers) HandleProviders(w http.ResponseWriter, r *http.Request) {
	providers := h.transferService.ListProviders()

	// Check Spotify connection status
	// TODO: Replace hard-coded userID with session-based authentication
	userID := "default" // In production, get from authenticated session
	spotifyConnected := false
	spotifyUserName := ""
	
	if h.spotifyEnabled {
		conn, err := h.connectionStore.Get("spotify", userID)
		if err == nil && conn.Connected {
			spotifyConnected = true
			spotifyUserName = conn.ExternalUserName
		}
	}

	data := map[string]interface{}{
		"Title":            "Available Providers",
		"Providers":        providers,
		"SpotifyEnabled":   h.spotifyEnabled,
		"SpotifyConnected": spotifyConnected,
		"SpotifyUserName":  spotifyUserName,
	}

	if err := h.templates.ExecuteTemplate(w, "providers.html", data); err != nil {
		log.Printf("Error rendering providers template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleTransfer renders the transfer page
func (h *Handlers) HandleTransfer(w http.ResponseWriter, r *http.Request) {
	providers := h.transferService.ListProviders()

	data := map[string]interface{}{
		"Title":     "Transfer Playlists",
		"Providers": providers,
	}

	if err := h.templates.ExecuteTemplate(w, "transfer.html", data); err != nil {
		log.Printf("Error rendering transfer template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleGetPlaylists is an HTMX endpoint that returns playlists for a provider
func (h *Handlers) HandleGetPlaylists(w http.ResponseWriter, r *http.Request) {
	providerName := r.URL.Query().Get("provider")
	if providerName == "" {
		http.Error(w, "Provider parameter required", http.StatusBadRequest)
		return
	}

	provider, err := h.transferService.GetProvider(providerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authenticate
	if err := provider.Authenticate(); err != nil {
		http.Error(w, "Authentication failed", http.StatusInternalServerError)
		return
	}

	// Get playlists
	playlists, err := provider.GetPlaylists()
	if err != nil {
		http.Error(w, "Failed to fetch playlists", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Playlists": playlists,
		"Provider":  providerName,
	}

	if err := h.templates.ExecuteTemplate(w, "playlist-list.html", data); err != nil {
		log.Printf("Error rendering playlist list: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HandleStartTransfer is an HTMX endpoint that initiates a playlist transfer
func (h *Handlers) HandleStartTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	sourceProvider := r.FormValue("source_provider")
	targetProvider := r.FormValue("target_provider")
	playlistID := r.FormValue("playlist_id")

	if sourceProvider == "" || targetProvider == "" || playlistID == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Start transfer (simulated with delay for demo)
	progress := &services.TransferProgress{
		PlaylistID:     playlistID,
		SourceProvider: sourceProvider,
		TargetProvider: targetProvider,
		Status:         "in_progress",
		Progress:       0,
		Message:        "Starting transfer...",
		StartedAt:      time.Now(),
	}

	// Simulate transfer
	go func() {
		time.Sleep(2 * time.Second)
		if err := h.transferService.TransferPlaylist(sourceProvider, targetProvider, playlistID); err != nil {
			log.Printf("Transfer failed: %v", err)
		}
	}()

	// Update progress
	progress.Progress = 50
	progress.Message = "Transferring tracks..."

	time.Sleep(1 * time.Second)

	// Complete
	progress.Progress = 100
	progress.Status = "completed"
	progress.Message = "Transfer complete!"
	now := time.Now()
	progress.CompletedAt = &now

	// Return progress as HTML
	data := map[string]interface{}{
		"Progress": progress,
	}

	if err := h.templates.ExecuteTemplate(w, "transfer-result.html", data); err != nil {
		log.Printf("Error rendering transfer result: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}
