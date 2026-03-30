package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JanikSachs/PlayPort/internal/auth"
)

func TestSessionMiddleware_RedirectsUnauthenticated(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)

	mw := SessionMiddleware(sessionStore)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/login" {
		t.Errorf("Expected redirect to /login, got %s", location)
	}
}

func TestSessionMiddleware_AllowsPublicPaths(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)

	mw := SessionMiddleware(sessionStore)

	publicPaths := []string{"/login", "/register", "/static/css/custom.css"}

	for _, path := range publicPaths {
		t.Run(path, func(t *testing.T) {
			var called bool
			handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if !called {
				t.Errorf("Handler should be called for public path %s", path)
			}

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", path, w.Code)
			}
		})
	}
}

func TestSessionMiddleware_ExistingSession(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)

	token, err := sessionStore.Create("user123")
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	mw := SessionMiddleware(sessionStore)

	var capturedUserID string
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if capturedUserID != "user123" {
		t.Errorf("Expected userID 'user123', got %s", capturedUserID)
	}
}

func TestSessionMiddleware_InvalidCookie_Redirects(t *testing.T) {
	sessionStore := auth.NewInMemorySessionStore(0)

	mw := SessionMiddleware(sessionStore)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid-token"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected redirect status 302, got %d", w.Code)
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
