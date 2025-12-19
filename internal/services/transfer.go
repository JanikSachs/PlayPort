package services

import (
	"fmt"
	"time"

	"github.com/JanikSachs/PlayPort/internal/providers"
)

// TransferService handles playlist transfers between providers
type TransferService struct {
	providers map[string]providers.Provider
}

// NewTransferService creates a new transfer service
func NewTransferService() *TransferService {
	return &TransferService{
		providers: make(map[string]providers.Provider),
	}
}

// RegisterProvider registers a provider with the service
func (s *TransferService) RegisterProvider(provider providers.Provider) {
	s.providers[provider.Name()] = provider
}

// GetProvider retrieves a provider by name
func (s *TransferService) GetProvider(name string) (providers.Provider, error) {
	provider, ok := s.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// ListProviders returns all registered providers
func (s *TransferService) ListProviders() []string {
	names := make([]string, 0, len(s.providers))
	for name := range s.providers {
		names = append(names, name)
	}
	return names
}

// TransferPlaylist transfers a playlist from source to target provider
func (s *TransferService) TransferPlaylist(sourceProvider, targetProvider, playlistID string) error {
	// Get source provider
	source, err := s.GetProvider(sourceProvider)
	if err != nil {
		return fmt.Errorf("source provider error: %w", err)
	}

	// Get target provider
	target, err := s.GetProvider(targetProvider)
	if err != nil {
		return fmt.Errorf("target provider error: %w", err)
	}

	// Authenticate with source
	if err := source.Authenticate(); err != nil {
		return fmt.Errorf("source authentication failed: %w", err)
	}

	// Authenticate with target
	if err := target.Authenticate(); err != nil {
		return fmt.Errorf("target authentication failed: %w", err)
	}

	// Export playlist from source
	playlist, err := source.ExportPlaylist(playlistID)
	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	// Import playlist to target
	if err := target.ImportPlaylist(playlist); err != nil {
		return fmt.Errorf("import failed: %w", err)
	}

	return nil
}

// TransferProgress represents the status of a playlist transfer
type TransferProgress struct {
	PlaylistID    string    `json:"playlist_id"`
	PlaylistName  string    `json:"playlist_name"`
	SourceProvider string   `json:"source_provider"`
	TargetProvider string   `json:"target_provider"`
	Status        string    `json:"status"` // "pending", "in_progress", "completed", "failed"
	Progress      int       `json:"progress"` // 0-100
	Message       string    `json:"message"`
	StartedAt     time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}
