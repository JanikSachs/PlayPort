package handlers

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/JanikSachs/PlayPort/internal/auth"
	"github.com/JanikSachs/PlayPort/internal/middleware"
	"github.com/JanikSachs/PlayPort/internal/providers/spotify"
	"github.com/JanikSachs/PlayPort/internal/providers/youtubemusic"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// AuthHandlers contains OAuth authentication handlers
type AuthHandlers struct {
	spotifyProvider      *spotify.SpotifyProvider
	youtubeMusicProvider *youtubemusic.YouTubeMusicProvider
	stateStore           auth.StateStore
	userStore            storage.UserStore
	sessionStore         auth.SessionStore
	templates            *template.Template
	spotifyEnabled       bool
	youtubeMusicEnabled  bool
}

// NewAuthHandlers creates new auth handlers
func NewAuthHandlers(spotifyProvider *spotify.SpotifyProvider, youtubeMusicProvider *youtubemusic.YouTubeMusicProvider, stateStore auth.StateStore, userStore storage.UserStore, sessionStore auth.SessionStore, templates *template.Template, spotifyEnabled bool, youtubeMusicEnabled bool) *AuthHandlers {
	return &AuthHandlers{
		spotifyProvider:      spotifyProvider,
		youtubeMusicProvider: youtubeMusicProvider,
		stateStore:           stateStore,
		userStore:            userStore,
		sessionStore:         sessionStore,
		templates:            templates,
		spotifyEnabled:       spotifyEnabled,
		youtubeMusicEnabled:  youtubeMusicEnabled,
	}
}

// HandleLoginPage renders the login page
func (h *AuthHandlers) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	// Redirect to home if already authenticated
	if middleware.UserIDFromContext(r.Context()) != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	csrfToken, err := h.stateStore.Generate()
	if err != nil {
		log.Printf("Failed to generate CSRF token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"CSRFToken": csrfToken,
	}

	if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
		log.Printf("Error rendering login template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleLogin processes login form submission
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	csrfToken := r.FormValue("csrf_token")
	if !h.stateStore.Validate(csrfToken) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	user, err := h.userStore.GetByUsername(username)
	if err != nil {
		h.renderLoginError(w, "Invalid username or password")
		return
	}

	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		h.renderLoginError(w, "Invalid username or password")
		return
	}

	// Create session
	token, err := h.sessionStore.Create(user.ID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandlers) renderLoginError(w http.ResponseWriter, errMsg string) {
	csrfToken, _ := h.stateStore.Generate()
	data := map[string]interface{}{
		"Error":     errMsg,
		"CSRFToken": csrfToken,
	}
	if err := h.templates.ExecuteTemplate(w, "login.html", data); err != nil {
		log.Printf("Error rendering login template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleRegisterPage renders the registration page
func (h *AuthHandlers) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	// Redirect to home if already authenticated
	if middleware.UserIDFromContext(r.Context()) != "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	csrfToken, err := h.stateStore.Generate()
	if err != nil {
		log.Printf("Failed to generate CSRF token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"CSRFToken": csrfToken,
		"Errors":    map[string]string{},
		"Username":  "",
	}

	if err := h.templates.ExecuteTemplate(w, "register.html", data); err != nil {
		log.Printf("Error rendering register template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleRegister processes registration form submission
func (h *AuthHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	csrfToken := r.FormValue("csrf_token")
	if !h.stateStore.Validate(csrfToken) {
		http.Error(w, "Invalid CSRF token", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm")

	// Validate input
	errors := make(map[string]string)

	if len(username) < 3 || len(username) > 32 {
		errors["username"] = "Username must be between 3 and 32 characters"
	} else if !usernameRegex.MatchString(username) {
		errors["username"] = "Username may only contain letters, numbers, and underscores"
	}

	if len(password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}

	if password != confirm {
		errors["confirm"] = "Passwords do not match"
	}

	if len(errors) > 0 {
		h.renderRegisterErrors(w, errors, username)
		return
	}

	// Hash password
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.userStore.CreateWithCredentials(username, hash)
	if err != nil {
		errors["username"] = "Username is already taken"
		h.renderRegisterErrors(w, errors, username)
		return
	}

	// Create session
	token, err := h.sessionStore.Create(user.ID)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandlers) renderRegisterErrors(w http.ResponseWriter, errors map[string]string, username string) {
	csrfToken, _ := h.stateStore.Generate()
	data := map[string]interface{}{
		"Errors":    errors,
		"Username":  username,
		"CSRFToken": csrfToken,
	}
	if err := h.templates.ExecuteTemplate(w, "register.html", data); err != nil {
		log.Printf("Error rendering register template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HandleLogout handles logout
func (h *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		_ = h.sessionStore.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusFound)
}

// HandleSpotifyStart redirects to Spotify authorization
func (h *AuthHandlers) HandleSpotifyStart(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyEnabled {
		http.Error(w, "Spotify is not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate state for CSRF protection
	state, err := h.stateStore.Generate()
	if err != nil {
		log.Printf("Failed to generate state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to Spotify authorization
	authURL := h.spotifyProvider.AuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleSpotifyCallback handles the OAuth callback from Spotify
func (h *AuthHandlers) HandleSpotifyCallback(w http.ResponseWriter, r *http.Request) {
	if !h.spotifyEnabled {
		http.Error(w, "Spotify is not configured", http.StatusServiceUnavailable)
		return
	}

	// Validate state
	state := r.URL.Query().Get("state")
	if !h.stateStore.Validate(state) {
		log.Printf("Invalid OAuth state")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Check for error from Spotify
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		log.Printf("Spotify OAuth error: %s", errMsg)
		http.Error(w, fmt.Sprintf("Spotify authorization failed: %s", errMsg), http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := h.spotifyProvider.Exchange(ctx, code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Save connection
	if err := h.spotifyProvider.SaveConnection(ctx, token, middleware.UserIDFromContext(r.Context())); err != nil {
		log.Printf("Failed to save connection: %v", err)
		http.Error(w, "Failed to save connection", http.StatusInternalServerError)
		return
	}

	// Redirect to providers page
	http.Redirect(w, r, "/providers", http.StatusFound)
}

// HandleYouTubeMusicStart redirects to Google authorization for YouTube Music
func (h *AuthHandlers) HandleYouTubeMusicStart(w http.ResponseWriter, r *http.Request) {
	if !h.youtubeMusicEnabled {
		http.Error(w, "YouTube Music is not configured", http.StatusServiceUnavailable)
		return
	}

	// Generate state for CSRF protection
	state, err := h.stateStore.Generate()
	if err != nil {
		log.Printf("Failed to generate state: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Redirect to Google authorization
	authURL := h.youtubeMusicProvider.AuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleYouTubeMusicCallback handles the OAuth callback from Google for YouTube Music
func (h *AuthHandlers) HandleYouTubeMusicCallback(w http.ResponseWriter, r *http.Request) {
	if !h.youtubeMusicEnabled {
		http.Error(w, "YouTube Music is not configured", http.StatusServiceUnavailable)
		return
	}

	// Validate state
	state := r.URL.Query().Get("state")
	if !h.stateStore.Validate(state) {
		log.Printf("Invalid OAuth state")
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Check for error from Google
	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		log.Printf("YouTube Music OAuth error: %s", errMsg)
		http.Error(w, fmt.Sprintf("YouTube Music authorization failed: %s", errMsg), http.StatusBadRequest)
		return
	}

	// Get authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := h.youtubeMusicProvider.Exchange(ctx, code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	// Save connection
	if err := h.youtubeMusicProvider.SaveConnection(ctx, token, middleware.UserIDFromContext(r.Context())); err != nil {
		log.Printf("Failed to save connection: %v", err)
		http.Error(w, "Failed to save connection", http.StatusInternalServerError)
		return
	}

	// Redirect to providers page
	http.Redirect(w, r, "/providers", http.StatusFound)
}

