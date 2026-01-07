package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"

	"github.com/JanikSachs/PlayPort/internal/models"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

const (
	baseURL = "https://api.spotify.com/v1"
)

// SpotifyProvider implements the Provider interface for Spotify
type SpotifyProvider struct {
	config          *oauth2.Config
	connectionStore storage.ConnectionStore
	userID          string // Current user ID - TODO: Replace with session-based user identification
	httpClient      *http.Client
}

// NewSpotifyProvider creates a new Spotify provider
func NewSpotifyProvider(clientID, clientSecret, redirectURL string, connectionStore storage.ConnectionStore) *SpotifyProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"user-read-private",
			"user-read-email",
			"playlist-read-private",
			"playlist-read-collaborative",
		},
		Endpoint: spotify.Endpoint,
	}

	return &SpotifyProvider{
		config:          config,
		connectionStore: connectionStore,
		userID:          "default", // TODO: In production, get from authenticated session
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the provider's name
func (p *SpotifyProvider) Name() string {
	return "Spotify"
}

// Authenticate checks if the user has a valid connection
func (p *SpotifyProvider) Authenticate() error {
	conn, err := p.connectionStore.Get("spotify", p.userID)
	if err != nil {
		return fmt.Errorf("not connected to Spotify: %w", err)
	}

	if !conn.Connected {
		return fmt.Errorf("Spotify connection not active")
	}

	return nil
}

// AuthURL returns the OAuth authorization URL
func (p *SpotifyProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange exchanges an authorization code for a token
func (p *SpotifyProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// SaveConnection saves a connection after OAuth
func (p *SpotifyProvider) SaveConnection(ctx context.Context, token *oauth2.Token) error {
	// Get user profile to get Spotify user ID
	profile, err := p.getUserProfile(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	conn := &models.Connection{
		Provider:         "spotify",
		UserID:           p.userID,
		ExternalUserID:   profile.ID,
		ExternalUserName: profile.DisplayName,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		ExpiresAt:        token.Expiry,
		Scopes:           p.config.Scopes,
		Connected:        true,
	}

	return p.connectionStore.Save(conn)
}

// GetPlaylists retrieves all playlists for the authenticated user
func (p *SpotifyProvider) GetPlaylists() ([]models.Playlist, error) {
	conn, err := p.connectionStore.Get("spotify", p.userID)
	if err != nil {
		return nil, fmt.Errorf("not connected: %w", err)
	}

	token := &oauth2.Token{
		AccessToken:  conn.AccessToken,
		RefreshToken: conn.RefreshToken,
		Expiry:       conn.ExpiresAt,
	}

	ctx := context.Background()
	client := p.config.Client(ctx, token)

	var allPlaylists []models.Playlist
	url := fmt.Sprintf("%s/me/playlists?limit=50", baseURL)

	for url != "" {
		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlists: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("spotify API error: %s - %s", resp.Status, string(body))
		}

		var result PlaylistsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, item := range result.Items {
			playlist := models.Playlist{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				TrackCount:  item.Tracks.Total,
				Provider:    "Spotify",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			allPlaylists = append(allPlaylists, playlist)
		}

		url = result.Next
	}

	// Update token if refreshed
	if err := p.updateTokenIfChanged(token, conn); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}

	return allPlaylists, nil
}

// ExportPlaylist exports a specific playlist by ID
func (p *SpotifyProvider) ExportPlaylist(id string) (models.Playlist, error) {
	conn, err := p.connectionStore.Get("spotify", p.userID)
	if err != nil {
		return models.Playlist{}, fmt.Errorf("not connected: %w", err)
	}

	token := &oauth2.Token{
		AccessToken:  conn.AccessToken,
		RefreshToken: conn.RefreshToken,
		Expiry:       conn.ExpiresAt,
	}

	ctx := context.Background()
	client := p.config.Client(ctx, token)

	// Get playlist details
	playlistURL := fmt.Sprintf("%s/playlists/%s", baseURL, id)
	resp, err := client.Get(playlistURL)
	if err != nil {
		return models.Playlist{}, fmt.Errorf("failed to fetch playlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.Playlist{}, fmt.Errorf("spotify API error: %s - %s", resp.Status, string(body))
	}

	var playlistDetail PlaylistDetail
	if err := json.NewDecoder(resp.Body).Decode(&playlistDetail); err != nil {
		return models.Playlist{}, fmt.Errorf("failed to decode playlist: %w", err)
	}

	playlist := models.Playlist{
		ID:          playlistDetail.ID,
		Name:        playlistDetail.Name,
		Description: playlistDetail.Description,
		TrackCount:  playlistDetail.Tracks.Total,
		Provider:    "Spotify",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Fetch all tracks with pagination
	var allTracks []models.Track
	tracksURL := fmt.Sprintf("%s/playlists/%s/tracks?limit=100", baseURL, id)

	for tracksURL != "" {
		resp, err := client.Get(tracksURL)
		if err != nil {
			return models.Playlist{}, fmt.Errorf("failed to fetch tracks: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return models.Playlist{}, fmt.Errorf("spotify API error: %s - %s", resp.Status, string(body))
		}

		var tracksResponse TracksResponse
		if err := json.NewDecoder(resp.Body).Decode(&tracksResponse); err != nil {
			resp.Body.Close()
			return models.Playlist{}, fmt.Errorf("failed to decode tracks: %w", err)
		}
		resp.Body.Close()

		for _, item := range tracksResponse.Items {
			if item.Track.ID == "" {
				continue // Skip null/deleted tracks
			}

			track := models.Track{
				ID:       item.Track.ID,
				Title:    item.Track.Name,
				Album:    item.Track.Album.Name,
				Duration: item.Track.DurationMS / 1000, // Convert to seconds
			}

			// Get all artist names and join them
			if len(item.Track.Artists) > 0 {
				artistNames := make([]string, len(item.Track.Artists))
				for i, artist := range item.Track.Artists {
					artistNames[i] = artist.Name
				}
				track.Artist = strings.Join(artistNames, ", ")
			}

			// Get ISRC if available
			if item.Track.ExternalIDs.ISRC != "" {
				track.ISRC = item.Track.ExternalIDs.ISRC
			}

			allTracks = append(allTracks, track)
		}

		tracksURL = tracksResponse.Next
	}

	playlist.Tracks = allTracks

	// Update token if refreshed
	if err := p.updateTokenIfChanged(token, conn); err != nil {
		return models.Playlist{}, fmt.Errorf("failed to update token: %w", err)
	}

	return playlist, nil
}

// ImportPlaylist imports a playlist into Spotify (not implemented yet)
func (p *SpotifyProvider) ImportPlaylist(playlist models.Playlist) error {
	return fmt.Errorf("importing to Spotify is not yet implemented")
}

// getUserProfile fetches the Spotify user profile
func (p *SpotifyProvider) getUserProfile(ctx context.Context, token *oauth2.Token) (*UserProfile, error) {
	client := p.config.Client(ctx, token)
	
	resp, err := client.Get(fmt.Sprintf("%s/me", baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spotify API error: %s - %s", resp.Status, string(body))
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode profile: %w", err)
	}

	return &profile, nil
}

// updateTokenIfChanged updates the stored token if it was refreshed
func (p *SpotifyProvider) updateTokenIfChanged(newToken *oauth2.Token, conn *models.Connection) error {
	if newToken.AccessToken != conn.AccessToken || newToken.Expiry != conn.ExpiresAt {
		conn.AccessToken = newToken.AccessToken
		if newToken.RefreshToken != "" {
			conn.RefreshToken = newToken.RefreshToken
		}
		conn.ExpiresAt = newToken.Expiry
		return p.connectionStore.Update(conn)
	}
	return nil
}
