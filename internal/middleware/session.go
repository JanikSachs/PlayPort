package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/JanikSachs/PlayPort/internal/auth"
)

type contextKey string

const userIDKey contextKey = "userID"

// SessionMiddleware returns an HTTP middleware that enforces authentication.
// It reads a "session_token" cookie, looks up the session, and injects the userID
// into the request context. Unauthenticated requests are redirected to /login,
// except for public paths (login, register, static assets).
func SessionMiddleware(sessionStore auth.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow public paths without authentication
			if isPublicPath(r) {
				// Still try to inject userID if session exists
				cookie, err := r.Cookie("session_token")
				if err == nil {
					uid, err := sessionStore.Get(cookie.Value)
					if err == nil {
						ctx := context.WithValue(r.Context(), userIDKey, uid)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
				next.ServeHTTP(w, r)
				return
			}

			// For protected paths, require authentication
			cookie, err := r.Cookie("session_token")
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			uid, err := sessionStore.Get(cookie.Value)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isPublicPath returns true for paths that don't require authentication
func isPublicPath(r *http.Request) bool {
	path := r.URL.Path

	if path == "/login" || path == "/register" || path == "/logout" {
		return true
	}

	if strings.HasPrefix(path, "/static/") {
		return true
	}

	return false
}

// UserIDFromContext retrieves the userID from the request context.
// Returns an empty string if no userID is present.
func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}
