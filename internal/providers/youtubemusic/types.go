package youtubemusic

// UserChannel represents a YouTube user's channel
type UserChannel struct {
	ID          string `json:"id"`
	DisplayName string `json:"snippet.title"`
}

// ChannelListResponse represents the response from YouTube's channel list endpoint
type ChannelListResponse struct {
	Items []ChannelItem `json:"items"`
}

// ChannelItem represents a YouTube channel
type ChannelItem struct {
	ID      string          `json:"id"`
	Snippet ChannelSnippet  `json:"snippet"`
}

// ChannelSnippet contains channel metadata
type ChannelSnippet struct {
	Title string `json:"title"`
}

// PlaylistListResponse represents the response from YouTube's playlist list endpoint
type PlaylistListResponse struct {
	Items         []PlaylistItem `json:"items"`
	NextPageToken string         `json:"nextPageToken"`
	PageInfo      PageInfo       `json:"pageInfo"`
}

// PlaylistItem represents a YouTube playlist
type PlaylistItem struct {
	ID      string          `json:"id"`
	Snippet PlaylistSnippet `json:"snippet"`
	ContentDetails PlaylistContentDetails `json:"contentDetails"`
}

// PlaylistSnippet contains playlist metadata
type PlaylistSnippet struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// PlaylistContentDetails contains playlist content information
type PlaylistContentDetails struct {
	ItemCount int `json:"itemCount"`
}

// PlaylistItemListResponse represents the response from YouTube's playlistItems endpoint
type PlaylistItemListResponse struct {
	Items         []PlaylistItemDetail `json:"items"`
	NextPageToken string               `json:"nextPageToken"`
	PageInfo      PageInfo             `json:"pageInfo"`
}

// PlaylistItemDetail represents a single item in a YouTube playlist
type PlaylistItemDetail struct {
	ID      string              `json:"id"`
	Snippet PlaylistItemSnippet `json:"snippet"`
}

// PlaylistItemSnippet contains playlist item metadata
type PlaylistItemSnippet struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	ResourceID  ResourceID     `json:"resourceId"`
	VideoOwnerChannelTitle string `json:"videoOwnerChannelTitle"`
}

// ResourceID identifies the resource in a playlist item
type ResourceID struct {
	Kind    string `json:"kind"`
	VideoID string `json:"videoId"`
}

// VideoListResponse represents the response from YouTube's video list endpoint
type VideoListResponse struct {
	Items []VideoItem `json:"items"`
}

// VideoItem represents a YouTube video
type VideoItem struct {
	ID             string          `json:"id"`
	Snippet        VideoSnippet    `json:"snippet"`
	ContentDetails VideoContentDetails `json:"contentDetails"`
}

// VideoSnippet contains video metadata
type VideoSnippet struct {
	Title        string `json:"title"`
	ChannelTitle string `json:"channelTitle"`
}

// VideoContentDetails contains video content information
type VideoContentDetails struct {
	Duration string `json:"duration"` // ISO 8601 duration, e.g. "PT4M13S"
}

// PageInfo contains pagination information
type PageInfo struct {
	TotalResults   int `json:"totalResults"`
	ResultsPerPage int `json:"resultsPerPage"`
}
