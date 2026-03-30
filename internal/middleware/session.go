package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/JanikSachs/PlayPort/internal/auth"
	"github.com/JanikSachs/PlayPort/internal/storage"
)

type contextKey string

const userIDKey contextKey = "userID"

// SessionMiddleware returns an HTTP middleware that manages anonymous sessions.
// It reads a "session_token" cookie, looks up the session, and injects the userID
// into the request context. If no valid session exists, a new anonymous user and
// session are created automatically.
func SessionMiddleware(sessionStore auth.SessionStore, userStore storage.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var userID string

			cookie, err := r.Cookie("session_token")
			if err == nil {
				uid, err := sessionStore.Get(cookie.Value)
				if err == nil {
					userID = uid
				}
			}

			if userID == "" {
				user, err := userStore.Create()
				if err != nil {
					log.Printf("Failed to create user: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				token, err := sessionStore.Create(user.ID)
				if err != nil {
					log.Printf("Failed to create session: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     "session_token",
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
					Expires:  time.Now().Add(24 * time.Hour),
				})

				userID = user.ID
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext retrieves the userID from the request context.
// Returns an empty string if no userID is present.
func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}
