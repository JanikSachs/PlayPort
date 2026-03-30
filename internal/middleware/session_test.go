package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JanikSachs/PlayPort/internal/auth"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

func TestSessionMiddleware_NewSession(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)
	userStore := storage.NewInMemoryUserStore()

	mw := SessionMiddleware(sessionStore, userStore)

	var capturedUserID string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if capturedUserID == "" {
		t.Error("Middleware should inject a userID into context")
	}

	// Check that a session cookie was set
	cookies := w.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_token" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil {
		t.Error("Middleware should set a session_token cookie")
	}

	if !sessionCookie.HttpOnly {
		t.Error("session_token cookie should be HttpOnly")
	}
}

func TestSessionMiddleware_ExistingSession(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)
	userStore := storage.NewInMemoryUserStore()

	// Create a user and session manually
	user, err := userStore.Create()
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	token, err := sessionStore.Create(user.ID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	mw := SessionMiddleware(sessionStore, userStore)

	var capturedUserID string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedUserID != user.ID {
		t.Errorf("Expected userID %s, got %s", user.ID, capturedUserID)
	}
}

func TestSessionMiddleware_InvalidCookie_CreatesNewSession(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)
	userStore := storage.NewInMemoryUserStore()

	mw := SessionMiddleware(sessionStore, userStore)

	var capturedUserID string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid-token"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedUserID == "" {
		t.Error("Middleware should create a new session when cookie is invalid")
	}
}

func TestSessionMiddleware_ExpiredSession_CreatesNewSession(t *testing.T) {
	// Use a very short duration so the session expires quickly
	sessionStore := auth.NewInMemorySessionStore(1 * time.Millisecond)
	userStore := storage.NewInMemoryUserStore()

	user, err := userStore.Create()
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	token, err := sessionStore.Create(user.ID)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Wait for the session to expire
	time.Sleep(10 * time.Millisecond)

	mw := SessionMiddleware(sessionStore, userStore)

	var capturedUserID string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedUserID == user.ID {
		t.Error("Middleware should create a new session when the old one has expired")
	}

	if capturedUserID == "" {
		t.Error("Middleware should provide a new userID after expiry")
	}
}

func TestUserIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()
	userID := UserIDFromContext(ctx)
	if userID != "" {
		t.Errorf("Expected empty string, got %q", userID)
	}
}

func TestUserIDFromContext_WithValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), userIDKey, "test-user")
	userID := UserIDFromContext(ctx)
	if userID != "test-user" {
		t.Errorf("Expected 'test-user', got %q", userID)
	}
}

func TestSessionMiddleware_TwoSessionsAreIsolated(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)
	userStore := storage.NewInMemoryUserStore()

	mw := SessionMiddleware(sessionStore, userStore)

	collectUserID := func(cookie *http.Cookie) string {
		var uid string
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid = UserIDFromContext(r.Context())
		}))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return uid
	}

	// First request — no cookie yet
	uid1 := collectUserID(nil)

	// Second request — also no cookie
	uid2 := collectUserID(nil)

	if uid1 == uid2 {
		t.Error("Two independent sessions should produce different userIDs")
	}
}
