package youtubemusic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/JanikSachs/PlayPort/internal/models"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

const (
	baseURL = "https://www.googleapis.com/youtube/v3"
)

// YouTubeMusicProvider implements the Provider interface for YouTube Music
type YouTubeMusicProvider struct {
	config          *oauth2.Config
	connectionStore storage.ConnectionStore
	userID          string // Current user ID - TODO: Replace with session-based user identification
	httpClient      *http.Client
}

// NewYouTubeMusicProvider creates a new YouTube Music provider
func NewYouTubeMusicProvider(clientID, clientSecret, redirectURL string, connectionStore storage.ConnectionStore) *YouTubeMusicProvider {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube.readonly",
		},
		Endpoint: google.Endpoint,
	}

	return &YouTubeMusicProvider{
		config:          config,
		connectionStore: connectionStore,
		userID:          "default", // TODO: In production, get from authenticated session
		httpClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

// Name returns the provider's name
func (p *YouTubeMusicProvider) Name() string {
	return "YouTube Music"
}

// Authenticate checks if the user has a valid connection
func (p *YouTubeMusicProvider) Authenticate() error {
	conn, err := p.connectionStore.Get("youtubemusic", p.userID)
	if err != nil {
		return fmt.Errorf("not connected to YouTube Music: %w", err)
	}

	if !conn.Connected {
		return fmt.Errorf("YouTube Music connection not active")
	}

	return nil
}

// AuthURL returns the OAuth authorization URL
func (p *YouTubeMusicProvider) AuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// Exchange exchanges an authorization code for a token
func (p *YouTubeMusicProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return p.config.Exchange(ctx, code)
}

// SaveConnection saves a connection after OAuth
func (p *YouTubeMusicProvider) SaveConnection(ctx context.Context, token *oauth2.Token) error {
	// Get user channel to obtain the YouTube channel ID and display name
	channel, err := p.getUserChannel(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to get user channel: %w", err)
	}

	conn := &models.Connection{
		Provider:         "youtubemusic",
		UserID:           p.userID,
		ExternalUserID:   channel.ID,
		ExternalUserName: channel.Snippet.Title,
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		ExpiresAt:        token.Expiry,
		Scopes:           p.config.Scopes,
		Connected:        true,
	}

	return p.connectionStore.Save(conn)
}

// GetPlaylists retrieves all playlists for the authenticated user
func (p *YouTubeMusicProvider) GetPlaylists() ([]models.Playlist, error) {
	conn, err := p.connectionStore.Get("youtubemusic", p.userID)
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
	pageToken := ""

	for {
		url := fmt.Sprintf("%s/playlists?part=snippet,contentDetails&mine=true&maxResults=50", baseURL)
		if pageToken != "" {
			url += "&pageToken=" + pageToken
		}

		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch playlists: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
		}

		var result PlaylistListResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		for _, item := range result.Items {
			playlist := models.Playlist{
				ID:          item.ID,
				Name:        item.Snippet.Title,
				Description: item.Snippet.Description,
				TrackCount:  item.ContentDetails.ItemCount,
				Provider:    "YouTube Music",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			allPlaylists = append(allPlaylists, playlist)
		}

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	// Update token if refreshed
	if err := p.updateTokenIfChanged(token, conn); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}

	return allPlaylists, nil
}

// ExportPlaylist exports a specific playlist by ID
func (p *YouTubeMusicProvider) ExportPlaylist(id string) (models.Playlist, error) {
	conn, err := p.connectionStore.Get("youtubemusic", p.userID)
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
	playlistURL := fmt.Sprintf("%s/playlists?part=snippet,contentDetails&id=%s", baseURL, id)
	resp, err := client.Get(playlistURL)
	if err != nil {
		return models.Playlist{}, fmt.Errorf("failed to fetch playlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return models.Playlist{}, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	var playlistList PlaylistListResponse
	if err := json.NewDecoder(resp.Body).Decode(&playlistList); err != nil {
		return models.Playlist{}, fmt.Errorf("failed to decode playlist: %w", err)
	}

	if len(playlistList.Items) == 0 {
		return models.Playlist{}, fmt.Errorf("playlist not found: %s", id)
	}

	playlistDetail := playlistList.Items[0]
	playlist := models.Playlist{
		ID:          playlistDetail.ID,
		Name:        playlistDetail.Snippet.Title,
		Description: playlistDetail.Snippet.Description,
		TrackCount:  playlistDetail.ContentDetails.ItemCount,
		Provider:    "YouTube Music",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Fetch all playlist items (video IDs) with pagination
	var videoIDs []string
	var itemSnippets []PlaylistItemSnippet
	pageToken := ""

	for {
		itemsURL := fmt.Sprintf("%s/playlistItems?part=snippet&playlistId=%s&maxResults=50", baseURL, id)
		if pageToken != "" {
			itemsURL += "&pageToken=" + pageToken
		}

		resp, err := client.Get(itemsURL)
		if err != nil {
			return models.Playlist{}, fmt.Errorf("failed to fetch playlist items: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return models.Playlist{}, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
		}

		var itemsResponse PlaylistItemListResponse
		if err := json.NewDecoder(resp.Body).Decode(&itemsResponse); err != nil {
			resp.Body.Close()
			return models.Playlist{}, fmt.Errorf("failed to decode playlist items: %w", err)
		}
		resp.Body.Close()

		for _, item := range itemsResponse.Items {
			if item.Snippet.ResourceID.Kind == "youtube#video" && item.Snippet.ResourceID.VideoID != "" {
				videoIDs = append(videoIDs, item.Snippet.ResourceID.VideoID)
				itemSnippets = append(itemSnippets, item.Snippet)
			}
		}

		if itemsResponse.NextPageToken == "" {
			break
		}
		pageToken = itemsResponse.NextPageToken
	}

	// Fetch video details in batches of 50 to get duration
	var allTracks []models.Track
	for i := 0; i < len(videoIDs); i += 50 {
		end := i + 50
		if end > len(videoIDs) {
			end = len(videoIDs)
		}
		batch := videoIDs[i:end]
		snippetBatch := itemSnippets[i:end]

		videosURL := fmt.Sprintf("%s/videos?part=snippet,contentDetails&id=%s", baseURL, strings.Join(batch, ","))
		resp, err := client.Get(videosURL)
		if err != nil {
			return models.Playlist{}, fmt.Errorf("failed to fetch video details: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return models.Playlist{}, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
		}

		var videosResponse VideoListResponse
		if err := json.NewDecoder(resp.Body).Decode(&videosResponse); err != nil {
			resp.Body.Close()
			return models.Playlist{}, fmt.Errorf("failed to decode video details: %w", err)
		}
		resp.Body.Close()

		// Build a map for quick lookup
		videoMap := make(map[string]VideoItem, len(videosResponse.Items))
		for _, v := range videosResponse.Items {
			videoMap[v.ID] = v
		}

		for j, videoID := range batch {
			snippet := snippetBatch[j]
			track := models.Track{
				ID:    videoID,
				Title: snippet.Title,
			}

			if video, ok := videoMap[videoID]; ok {
				track.Artist = video.Snippet.ChannelTitle
				track.Duration = parseDuration(video.ContentDetails.Duration)
			} else {
				track.Artist = snippet.VideoOwnerChannelTitle
			}

			allTracks = append(allTracks, track)
		}
	}

	playlist.Tracks = allTracks

	// Update token if refreshed
	if err := p.updateTokenIfChanged(token, conn); err != nil {
		return models.Playlist{}, fmt.Errorf("failed to update token: %w", err)
	}

	return playlist, nil
}

// ImportPlaylist imports a playlist into YouTube Music (not implemented yet)
func (p *YouTubeMusicProvider) ImportPlaylist(playlist models.Playlist) error {
	return fmt.Errorf("importing to YouTube Music is not yet implemented")
}

// getUserChannel fetches the authenticated user's YouTube channel
func (p *YouTubeMusicProvider) getUserChannel(ctx context.Context, token *oauth2.Token) (*ChannelItem, error) {
	client := p.config.Client(ctx, token)

	resp, err := client.Get(fmt.Sprintf("%s/channels?part=snippet&mine=true", baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user channel: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error: %s - %s", resp.Status, string(body))
	}

	var channelList ChannelListResponse
	if err := json.NewDecoder(resp.Body).Decode(&channelList); err != nil {
		return nil, fmt.Errorf("failed to decode channel list: %w", err)
	}

	if len(channelList.Items) == 0 {
		return nil, fmt.Errorf("no YouTube channel found for authenticated user")
	}

	return &channelList.Items[0], nil
}

// updateTokenIfChanged updates the stored token if it was refreshed
func (p *YouTubeMusicProvider) updateTokenIfChanged(newToken *oauth2.Token, conn *models.Connection) error {
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

// parseDuration parses an ISO 8601 duration string (e.g. "PT4M13S") into seconds
func parseDuration(iso8601 string) int {
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	matches := re.FindStringSubmatch(iso8601)
	if matches == nil {
		return 0
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])

	return hours*3600 + minutes*60 + seconds
}
