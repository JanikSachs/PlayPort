package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

const defaultSessionDuration = 24 * time.Hour

// SessionStore manages session tokens
type SessionStore interface {
	// Create creates a new session for a user and returns the token
	Create(userID string) (token string, err error)

	// Get retrieves the userID for a session token
	Get(token string) (userID string, err error)

	// Delete removes a session
	Delete(token string) error
}

type session struct {
	userID    string
	expiresAt time.Time
}

// InMemorySessionStore is a thread-safe in-memory session store
type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*session
	duration time.Duration
}

// NewInMemorySessionStore creates a new in-memory session store with an optional duration.
// If duration is zero or negative, the default of 24 hours is used.
func NewInMemorySessionStore(duration time.Duration) *InMemorySessionStore {
	if duration <= 0 {
		duration = defaultSessionDuration
	}

	store := &InMemorySessionStore{
		sessions: make(map[string]*session),
		duration: duration,
	}

	go store.cleanup()

	return store
}

// Create creates a new session for a user and returns the session token
func (s *InMemorySessionStore) Create(userID string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[token] = &session{
		userID:    userID,
		expiresAt: time.Now().Add(s.duration),
	}

	return token, nil
}

// Get retrieves the userID for a session token
func (s *InMemorySessionStore) Get(token string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sess, exists := s.sessions[token]
	if !exists {
		return "", fmt.Errorf("session not found")
	}

	if time.Now().After(sess.expiresAt) {
		return "", fmt.Errorf("session expired")
	}

	return sess.userID, nil
}

// Delete removes a session
func (s *InMemorySessionStore) Delete(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[token]; !exists {
		return fmt.Errorf("session not found")
	}

	delete(s.sessions, token)
	return nil
}

// cleanup periodically removes expired sessions
func (s *InMemorySessionStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, sess := range s.sessions {
			if now.After(sess.expiresAt) {
				delete(s.sessions, token)
			}
		}
		s.mu.Unlock()
	}
}
