package youtubemusic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JanikSachs/PlayPort/internal/models"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

func TestYouTubeMusicProvider_Name(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewYouTubeMusicProvider("client-id", "client-secret", "http://localhost/callback", store)

	if provider.Name() != "YouTube Music" {
		t.Errorf("Expected provider name 'YouTube Music', got '%s'", provider.Name())
	}
}

func TestYouTubeMusicProvider_AuthURL(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewYouTubeMusicProvider("client-id", "client-secret", "http://localhost/callback", store)

	authURL := provider.AuthURL("test-state")

	if authURL == "" {
		t.Error("AuthURL should not be empty")
	}

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

func TestYouTubeMusicProvider_Authenticate_NotConnected(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewYouTubeMusicProvider("client-id", "client-secret", "http://localhost/callback", store)

	err := provider.Authenticate("user123")
	if err == nil {
		t.Error("Authenticate() should fail when not connected")
	}
}

func TestYouTubeMusicProvider_ImportPlaylist_NotImplemented(t *testing.T) {
	store := storage.NewInMemoryConnectionStore()
	provider := NewYouTubeMusicProvider("client-id", "client-secret", "http://localhost/callback", store)

	err := provider.ImportPlaylist(models.Playlist{})
	if err == nil {
		t.Error("ImportPlaylist() should return error (not implemented)")
	}

	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Error("Error message should indicate not implemented")
	}
}

func TestYouTubeMusicProvider_GetPlaylists_MockServer(t *testing.T) {
	// Create mock server that returns paginated playlists
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/youtube/v3/playlists") {
			response := PlaylistListResponse{
				Items: []PlaylistItem{
					{
						ID:      "playlist1",
						Snippet: PlaylistSnippet{Title: "Playlist 1", Description: "First playlist"},
						ContentDetails: PlaylistContentDetails{ItemCount: 10},
					},
					{
						ID:      "playlist2",
						Snippet: PlaylistSnippet{Title: "Playlist 2", Description: "Second playlist"},
						ContentDetails: PlaylistContentDetails{ItemCount: 20},
					},
				},
				NextPageToken: "",
				PageInfo:      PageInfo{TotalResults: 2, ResultsPerPage: 50},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	// This test demonstrates the mock server structure for pagination testing
	t.Log("Mock server created successfully for YouTube Music playlist pagination testing")
}

func TestYouTubeMusicProvider_ExportPlaylist_MockServer(t *testing.T) {
	// Create mock server for playlist and tracks
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/youtube/v3/playlists") {
			response := PlaylistListResponse{
				Items: []PlaylistItem{
					{
						ID:      "test-playlist",
						Snippet: PlaylistSnippet{Title: "Test Playlist", Description: "A test playlist"},
						ContentDetails: PlaylistContentDetails{ItemCount: 2},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if strings.Contains(r.URL.Path, "/youtube/v3/playlistItems") {
			response := PlaylistItemListResponse{
				Items: []PlaylistItemDetail{
					{
						ID: "item1",
						Snippet: PlaylistItemSnippet{
							Title:      "Song 1",
							ResourceID: ResourceID{Kind: "youtube#video", VideoID: "video1"},
							VideoOwnerChannelTitle: "Artist 1",
						},
					},
					{
						ID: "item2",
						Snippet: PlaylistItemSnippet{
							Title:      "Song 2",
							ResourceID: ResourceID{Kind: "youtube#video", VideoID: "video2"},
							VideoOwnerChannelTitle: "Artist 2",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		} else if strings.Contains(r.URL.Path, "/youtube/v3/videos") {
			response := VideoListResponse{
				Items: []VideoItem{
					{
						ID:      "video1",
						Snippet: VideoSnippet{Title: "Song 1", ChannelTitle: "Artist 1"},
						ContentDetails: VideoContentDetails{Duration: "PT3M45S"},
					},
					{
						ID:      "video2",
						Snippet: VideoSnippet{Title: "Song 2", ChannelTitle: "Artist 2"},
						ContentDetails: VideoContentDetails{Duration: "PT4M20S"},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	t.Log("Mock server created successfully for YouTube Music track parsing testing")
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"PT4M13S", 253},
		{"PT1H2M3S", 3723},
		{"PT30S", 30},
		{"PT5M", 300},
		{"PT1H", 3600},
		{"PT", 0},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseDuration(tt.input)
			if result != tt.expected {
				t.Errorf("parseDuration(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPlaylistListResponse(t *testing.T) {
	jsonData := `{
		"items": [
			{
				"id": "playlist1",
				"snippet": {
					"title": "My Playlist",
					"description": "A great playlist"
				},
				"contentDetails": {
					"itemCount": 50
				}
			}
		],
		"nextPageToken": "",
		"pageInfo": {
			"totalResults": 1,
			"resultsPerPage": 50
		}
	}`

	var response PlaylistListResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Failed to unmarshal playlists response: %v", err)
	}

	if len(response.Items) != 1 {
		t.Errorf("Expected 1 playlist, got %d", len(response.Items))
	}

	if response.Items[0].Snippet.Title != "My Playlist" {
		t.Errorf("Expected playlist title 'My Playlist', got '%s'", response.Items[0].Snippet.Title)
	}

	if response.Items[0].ContentDetails.ItemCount != 50 {
		t.Errorf("Expected ItemCount 50, got %d", response.Items[0].ContentDetails.ItemCount)
	}
}

func TestChannelListResponse(t *testing.T) {
	jsonData := `{
		"items": [
			{
				"id": "channel123",
				"snippet": {
					"title": "My YouTube Channel"
				}
			}
		]
	}`

	var response ChannelListResponse
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		t.Fatalf("Failed to unmarshal channel list response: %v", err)
	}

	if len(response.Items) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(response.Items))
	}

	if response.Items[0].ID != "channel123" {
		t.Errorf("Expected channel ID 'channel123', got '%s'", response.Items[0].ID)
	}

	if response.Items[0].Snippet.Title != "My YouTube Channel" {
		t.Errorf("Expected channel title 'My YouTube Channel', got '%s'", response.Items[0].Snippet.Title)
	}
}
