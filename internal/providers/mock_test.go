package providers

import (
	"testing"

	"github.com/JanikSachs/PlayPort/internal/models"
)

func TestMockProvider_Name(t *testing.T) {
	provider := NewMockProvider()
	name := provider.Name()
	
	if name == "" {
		t.Error("Provider name should not be empty")
	}
	
	if name != "Mock Music" {
		t.Errorf("Expected provider name 'Mock Music', got '%s'", name)
	}
}

func TestMockProvider_Authenticate(t *testing.T) {
	provider := NewMockProvider()
	
	err := provider.Authenticate()
	if err != nil {
		t.Errorf("Authenticate() should not return error, got: %v", err)
	}
}

func TestMockProvider_GetPlaylists(t *testing.T) {
	provider := NewMockProvider()
	
	// Should fail without authentication
	_, err := provider.GetPlaylists()
	if err == nil {
		t.Error("GetPlaylists() should fail without authentication")
	}
	
	// Authenticate first
	if err := provider.Authenticate(); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	
	// Should succeed after authentication
	playlists, err := provider.GetPlaylists()
	if err != nil {
		t.Errorf("GetPlaylists() returned error: %v", err)
	}
	
	if len(playlists) == 0 {
		t.Error("GetPlaylists() should return at least one playlist")
	}
	
	// Verify playlist structure
	for _, playlist := range playlists {
		if playlist.ID == "" {
			t.Error("Playlist ID should not be empty")
		}
		if playlist.Name == "" {
			t.Error("Playlist Name should not be empty")
		}
		if playlist.Provider != provider.Name() {
			t.Errorf("Playlist provider should be '%s', got '%s'", provider.Name(), playlist.Provider)
		}
	}
}

func TestMockProvider_ExportPlaylist(t *testing.T) {
	provider := NewMockProvider()
	
	// Should fail without authentication
	_, err := provider.ExportPlaylist("mock-1")
	if err == nil {
		t.Error("ExportPlaylist() should fail without authentication")
	}
	
	// Authenticate first
	if err := provider.Authenticate(); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	
	// Should succeed with valid ID
	playlist, err := provider.ExportPlaylist("mock-1")
	if err != nil {
		t.Errorf("ExportPlaylist() returned error: %v", err)
	}
	
	if playlist.ID != "mock-1" {
		t.Errorf("Expected playlist ID 'mock-1', got '%s'", playlist.ID)
	}
	
	if len(playlist.Tracks) == 0 {
		t.Error("Exported playlist should have tracks")
	}
	
	// Should fail with invalid ID
	_, err = provider.ExportPlaylist("invalid-id")
	if err == nil {
		t.Error("ExportPlaylist() should fail with invalid ID")
	}
}

func TestMockProvider_ImportPlaylist(t *testing.T) {
	provider := NewMockProvider()
	
	testPlaylist := models.Playlist{
		ID:          "test-import",
		Name:        "Test Import Playlist",
		Description: "A test playlist for import",
		TrackCount:  2,
		Tracks: []models.Track{
			{
				ID:     "test-track-1",
				Title:  "Test Song 1",
				Artist: "Test Artist",
			},
			{
				ID:     "test-track-2",
				Title:  "Test Song 2",
				Artist: "Test Artist",
			},
		},
	}
	
	// Should fail without authentication
	err := provider.ImportPlaylist(testPlaylist)
	if err == nil {
		t.Error("ImportPlaylist() should fail without authentication")
	}
	
	// Authenticate first
	if err := provider.Authenticate(); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}
	
	// Get initial playlist count
	initialPlaylists, _ := provider.GetPlaylists()
	initialCount := len(initialPlaylists)
	
	// Should succeed after authentication
	err = provider.ImportPlaylist(testPlaylist)
	if err != nil {
		t.Errorf("ImportPlaylist() returned error: %v", err)
	}
	
	// Verify playlist was added
	playlists, _ := provider.GetPlaylists()
	if len(playlists) != initialCount+1 {
		t.Errorf("Expected %d playlists after import, got %d", initialCount+1, len(playlists))
	}
	
	// Verify the imported playlist has the correct provider
	lastPlaylist := playlists[len(playlists)-1]
	if lastPlaylist.Provider != provider.Name() {
		t.Errorf("Imported playlist provider should be '%s', got '%s'", provider.Name(), lastPlaylist.Provider)
	}
}
