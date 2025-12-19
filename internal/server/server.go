package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/JanikSachs/PlayPort/internal/handlers"
	"github.com/JanikSachs/PlayPort/internal/services"
)

// Server represents the HTTP server
type Server struct {
	addr            string
	mux             *http.ServeMux
	transferService *services.TransferService
	templates       *template.Template
}

// New creates a new server instance
func New(addr string, transferService *services.TransferService) (*Server, error) {
	// Parse templates
	templates, err := template.ParseGlob(filepath.Join("web", "templates", "*.html"))
	if err != nil {
		return nil, err
	}

	s := &Server{
		addr:            addr,
		mux:             http.NewServeMux(),
		transferService: transferService,
		templates:       templates,
	}

	s.setupRoutes()
	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Create handlers
	h := handlers.NewHandlers(s.transferService, s.templates)

	// Static files
	fs := http.FileServer(http.Dir("web/static"))
	s.mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Pages
	s.mux.HandleFunc("/", h.HandleHome)
	s.mux.HandleFunc("/providers", h.HandleProviders)
	s.mux.HandleFunc("/transfer", h.HandleTransfer)

	// HTMX endpoints
	s.mux.HandleFunc("/api/playlists", h.HandleGetPlaylists)
	s.mux.HandleFunc("/api/transfer/start", h.HandleStartTransfer)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Server starting on %s", s.addr)
	return http.ListenAndServe(s.addr, s.mux)
}
