package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/JanikSachs/PlayPort/internal/auth"
	"github.com/JanikSachs/PlayPort/internal/handlers"
	"github.com/JanikSachs/PlayPort/internal/middleware"
	"github.com/JanikSachs/PlayPort/internal/providers/spotify"
	"github.com/JanikSachs/PlayPort/internal/providers/youtubemusic"
	"github.com/JanikSachs/PlayPort/internal/services"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	addr                 string
	mux                  *http.ServeMux
	transferService      *services.TransferService
	templates            *template.Template
	spotifyProvider      *spotify.SpotifyProvider
	youtubeMusicProvider *youtubemusic.YouTubeMusicProvider
	connectionStore      storage.ConnectionStore
	userStore            storage.UserStore
	stateStore           auth.StateStore
	sessionStore         auth.SessionStore
	spotifyEnabled       bool
	youtubeMusicEnabled  bool
}

// New creates a new server instance
func New(addr string, transferService *services.TransferService, spotifyProvider *spotify.SpotifyProvider, youtubeMusicProvider *youtubemusic.YouTubeMusicProvider, connectionStore storage.ConnectionStore, userStore storage.UserStore, stateStore auth.StateStore, sessionStore auth.SessionStore, spotifyEnabled bool, youtubeMusicEnabled bool) (*Server, error) {
	// Parse templates
	templates, err := template.ParseGlob(filepath.Join("web", "templates", "*.html"))
	if err != nil {
		return nil, err
	}

	s := &Server{
		addr:                addr,
		mux:                 http.NewServeMux(),
		transferService:     transferService,
		templates:           templates,
		spotifyProvider:     spotifyProvider,
		youtubeMusicProvider: youtubeMusicProvider,
		connectionStore:     connectionStore,
		userStore:           userStore,
		stateStore:          stateStore,
		sessionStore:        sessionStore,
		spotifyEnabled:      spotifyEnabled,
		youtubeMusicEnabled: youtubeMusicEnabled,
	}

	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Create handlers
	h := handlers.NewHandlers(s.transferService, s.templates, s.connectionStore, s.spotifyEnabled, s.youtubeMusicEnabled)
	authHandlers := handlers.NewAuthHandlers(s.spotifyProvider, s.youtubeMusicProvider, s.stateStore, s.spotifyEnabled, s.youtubeMusicEnabled)
	providerHandlers := handlers.NewProviderHandlers(s.spotifyProvider, s.youtubeMusicProvider, s.connectionStore, s.templates, s.spotifyEnabled, s.youtubeMusicEnabled)

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Pages
	s.mux.HandleFunc("/", h.HandleHome)
	s.mux.HandleFunc("/providers", h.HandleProviders)
	s.mux.HandleFunc("/transfer", h.HandleTransfer)

	// OAuth routes - Spotify
	s.mux.HandleFunc("/auth/spotify/start", authHandlers.HandleSpotifyStart)
	s.mux.HandleFunc("/auth/spotify/callback", authHandlers.HandleSpotifyCallback)

	// OAuth routes - YouTube Music
	s.mux.HandleFunc("/auth/youtubemusic/start", authHandlers.HandleYouTubeMusicStart)
	s.mux.HandleFunc("/auth/youtubemusic/callback", authHandlers.HandleYouTubeMusicCallback)

	// Provider-specific endpoints
	s.mux.HandleFunc("/providers/spotify/playlists", providerHandlers.HandleSpotifyPlaylists)
	s.mux.HandleFunc("/providers/youtubemusic/playlists", providerHandlers.HandleYouTubeMusicPlaylists)

	// HTMX endpoints
	s.mux.HandleFunc("/api/playlists", h.HandleGetPlaylists)
	s.mux.HandleFunc("/api/transfer/start", h.HandleStartTransfer)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Server starting on %s", s.addr)
	sessionMW := middleware.SessionMiddleware(s.sessionStore, s.userStore)
	return http.ListenAndServe(s.addr, sessionMW(s.mux))
}
