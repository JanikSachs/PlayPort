package spotify

// UserProfile represents a Spotify user profile
type UserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

// PlaylistsResponse represents the response from Spotify's playlist list endpoint
type PlaylistsResponse struct {
	Items []PlaylistItem `json:"items"`
	Next  string         `json:"next"`
	Total int            `json:"total"`
}

// PlaylistItem represents a simplified playlist from the list
type PlaylistItem struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tracks      TracksInfo   `json:"tracks"`
}

// TracksInfo contains track count information
type TracksInfo struct {
	Total int `json:"total"`
}

// PlaylistDetail represents detailed playlist information
type PlaylistDetail struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tracks      TracksInfo   `json:"tracks"`
}

// TracksResponse represents the response from Spotify's tracks endpoint
type TracksResponse struct {
	Items []TrackItem `json:"items"`
	Next  string      `json:"next"`
	Total int         `json:"total"`
}

// TrackItem represents a track item in a playlist
type TrackItem struct {
	Track TrackDetail `json:"track"`
}

// TrackDetail represents detailed track information
type TrackDetail struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	DurationMS  int           `json:"duration_ms"`
	Album       AlbumInfo     `json:"album"`
	Artists     []ArtistInfo  `json:"artists"`
	ExternalIDs ExternalIDs   `json:"external_ids"`
}

// AlbumInfo represents album information
type AlbumInfo struct {
	Name string `json:"name"`
}

// ArtistInfo represents artist information
type ArtistInfo struct {
	Name string `json:"name"`
}

// ExternalIDs represents external IDs like ISRC
type ExternalIDs struct {
	ISRC string `json:"isrc"`
}
