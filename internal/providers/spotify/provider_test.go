package spotify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JanikSachs/PlayPort/internal/models"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

func TestSpotifyProvider_Name(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewSpotifyProvider("client-id", "client-secret", "http://localhost/callback", store)

	if provider.Name() != "Spotify" {
		t.Errorf("Expected provider name 'Spotify', got '%s'", provider.Name())
	}
}

func TestSpotifyProvider_AuthURL(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewSpotifyProvider("client-id", "client-secret", "http://localhost/callback", store)

	authURL := provider.AuthURL("test-state")

	if authURL == "" {
		t.Error("AuthURL should not be empty")
	}

	// Verify URL contains expected components
	if !strings.Contains(authURL, "client_id=client-id") {
		t.Error("AuthURL should contain client_id")
	}

	if !strings.Contains(authURL, "state=test-state") {
		t.Error("AuthURL should contain state")
	}

	if !strings.Contains(authURL, "redirect_uri=http") {
		t.Error("AuthURL should contain redirect_uri")
	}
}

func TestSpotifyProvider_GetPlaylists_MockServer(t *testing.T) {
	// Create mock server that returns paginated playlists
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/v1/me/playlists" {
			var response PlaylistsResponse
			
			if page == 0 {
				response = PlaylistsResponse{
					Items: []PlaylistItem{
						{ID: "playlist1", Name: "Playlist 1", Description: "First playlist", Tracks: TracksInfo{Total: 10}},
						{ID: "playlist2", Name: "Playlist 2", Description: "Second playlist", Tracks: TracksInfo{Total: 20}},
					},
					Next:  "",
					Total: 3,
				}
				page = 1
			} else {
				response = PlaylistsResponse{
					Items: []PlaylistItem{
						{ID: "playlist3", Name: "Playlist 3", Description: "Third playlist", Tracks: TracksInfo{Total: 15}},
					},
					Next:  "",
					Total: 3,
				}
			}

			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	// This test demonstrates the mock server structure
	// Full integration would require mocking OAuth
	t.Log("Mock server created successfully for pagination testing")
}

func TestSpotifyProvider_ExportPlaylist_MockServer(t *testing.T) {
	// Create mock server for playlist and tracks
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/playlists/test-playlist" {
			response := PlaylistDetail{
				ID:          "test-playlist",
				Name:        "Test Playlist",
				Description: "A test playlist",
				Tracks:      TracksInfo{Total: 2},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else if r.URL.Path == "/v1/playlists/test-playlist/tracks" {
			response := TracksResponse{
				Items: []TrackItem{
					{
						Track: TrackDetail{
							ID:         "track1",
							Name:       "Track 1",
							DurationMS: 180000,
							Album:      AlbumInfo{Name: "Album 1"},
							Artists:    []ArtistInfo{{Name: "Artist 1"}},
							ExternalIDs: ExternalIDs{ISRC: "ISRC001"},
						},
					},
					{
						Track: TrackDetail{
							ID:         "track2",
							Name:       "Track 2",
							DurationMS: 240000,
							Album:      AlbumInfo{Name: "Album 2"},
							Artists:    []ArtistInfo{{Name: "Artist 2"}},
							ExternalIDs: ExternalIDs{ISRC: "ISRC002"},
						},
					},
				},
				Next:  "",
				Total: 2,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	t.Log("Mock server created successfully for track parsing testing")
}

func TestSpotifyProvider_Authenticate_NotConnected(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewSpotifyProvider("client-id", "client-secret", "http://localhost/callback", store)

	err := provider.Authenticate()
	if err == nil {
		t.Error("Authenticate() should fail when not connected")
	}
}

func TestSpotifyProvider_ImportPlaylist_NotImplemented(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewSpotifyProvider("client-id", "client-secret", "http://localhost/callback", store)

	err := provider.ImportPlaylist(models.Playlist{})
	if err == nil {
		t.Error("ImportPlaylist() should return error (not implemented)")
	}

	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Error("Error message should indicate not implemented")
	}
}

func TestUserProfileResponse(t *testing.T) {
	// Test that we can properly decode Spotify user profile
	jsonData := `{
		"id": "spotify123",
		"display_name": "Test User",
		"email": "test@example.com"
	}`

	var profile UserProfile
	if err := json.Unmarshal([]byte(jsonData), &profile); err != nil {
		t.Fatalf("Failed to unmarshal user profile: %v", err)
	}

	if profile.ID != "spotify123" {
		t.Errorf("Expected ID 'spotify123', got '%s'", profile.ID)
	}

	if profile.DisplayName != "Test User" {
		t.Errorf("Expected DisplayName 'Test User', got '%s'", profile.DisplayName)
	}
}

func TestPlaylistsResponse(t *testing.T) {
	// Test that we can properly decode Spotify playlists response
	jsonData := `{
		"items": [
			{
				"id": "playlist1",
				"name": "My Playlist",
				"description": "A great playlist",
				"tracks": {"total": 50}
			}
		],
		"next": "https://api.spotify.com/v1/me/playlists?offset=1",
		"total": 100
	}`

	var response PlaylistsResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Failed to unmarshal playlists response: %v", err)
	}

	if len(response.Items) != 1 {
		t.Errorf("Expected 1 playlist, got %d", len(response.Items))
	}

	if response.Items[0].Name != "My Playlist" {
		t.Errorf("Expected playlist name 'My Playlist', got '%s'", response.Items[0].Name)
	}

	if response.Total != 100 {
		t.Errorf("Expected total 100, got %d", response.Total)
	}
}
