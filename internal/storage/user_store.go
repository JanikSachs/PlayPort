package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/JanikSachs/PlayPort/internal/models"
)

// UserStore defines the interface for storing users
type UserStore interface {
	// Create creates a new user with a random ID
	Create() (*models.User, error)

	// Get retrieves a user by ID
	Get(id string) (*models.User, error)

	// CreateWithCredentials creates a new user with a username and password hash
	CreateWithCredentials(username, passwordHash string) (*models.User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(username string) (*models.User, error)
}

// InMemoryUserStore is a thread-safe in-memory user store
type InMemoryUserStore struct {
	mu          sync.RWMutex
	users       map[string]*models.User
	byUsername  map[string]string // username -> user ID
}

// NewInMemoryUserStore creates a new in-memory user store
func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users:      make(map[string]*models.User),
		byUsername: make(map[string]string),
	}
}

// Create creates a new user with a random UUID-style ID
func (s *InMemoryUserStore) Create() (*models.User, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	user := &models.User{
		ID:        hex.EncodeToString(b),
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
	return user, nil
}

// Get retrieves a user by ID
func (s *InMemoryUserStore) Get(id string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	return user, nil
}

// CreateWithCredentials creates a new user with a username and password hash
func (s *InMemoryUserStore) CreateWithCredentials(username, passwordHash string) (*models.User, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate username
	if _, exists := s.byUsername[username]; exists {
		return nil, fmt.Errorf("username already taken: %s", username)
	}

	user := &models.User{
		ID:           hex.EncodeToString(b),
		Username:     username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}

	s.users[user.ID] = user
	s.byUsername[username] = user.ID
	return user, nil
}

// GetByUsername retrieves a user by username
func (s *InMemoryUserStore) GetByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.byUsername[username]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", username)
	}

	return s.users[id], nil
}
