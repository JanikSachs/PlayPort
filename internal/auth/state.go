package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"
)

// StateStore manages OAuth state tokens for CSRF protection
type StateStore interface {
	// Generate creates a new state token
	Generate() (string, error)

	// Validate checks if a state token is valid and removes it
	Validate(state string) bool
}

// InMemoryStateStore is a thread-safe in-memory state store
type InMemoryStateStore struct {
	mu     sync.RWMutex
	states map[string]time.Time
}

// NewInMemoryStateStore creates a new in-memory state store
func NewInMemoryStateStore() *InMemoryStateStore {
	store := &InMemoryStateStore{
		states: make(map[string]time.Time),
	}
	
	// Start cleanup goroutine
	go store.cleanup()
	
	return store
}

// Generate creates a new state token
func (s *InMemoryStateStore) Generate() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	
	state := base64.URLEncoding.EncodeToString(b)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Store state with expiration time (10 minutes)
	s.states[state] = time.Now().Add(10 * time.Minute)
	
	return state, nil
}

// Validate checks if a state token is valid and removes it
func (s *InMemoryStateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	expiry, exists := s.states[state]
	if !exists {
		return false
	}
	
	// Remove the state (one-time use)
	delete(s.states, state)
	
	// Check if expired
	return time.Now().Before(expiry)
}

// cleanup periodically removes expired states
func (s *InMemoryStateStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for state, expiry := range s.states {
			if now.After(expiry) {
				delete(s.states, state)
			}
		}
		s.mu.Unlock()
	}
}
