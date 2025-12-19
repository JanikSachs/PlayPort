package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/JanikSachs/PlayPort/internal/models"
)

// ConnectionStore defines the interface for storing provider connections
type ConnectionStore interface {
	// Save stores a connection
	Save(conn *models.Connection) error

	// Get retrieves a connection by provider and user ID
	Get(provider, userID string) (*models.Connection, error)

	// Update updates an existing connection
	Update(conn *models.Connection) error

	// Delete removes a connection
	Delete(provider, userID string) error

	// List returns all connections for a user
	List(userID string) ([]*models.Connection, error)
}

// InMemoryConnectionStore is a thread-safe in-memory connection store
type InMemoryConnectionStore struct {
	mu          sync.RWMutex
	connections map[string]*models.Connection // key: "provider:userID"
}

// NewInMemoryConnectionStore creates a new in-memory connection store
func NewInMemoryConnectionStore() *InMemoryConnectionStore {
	return &InMemoryConnectionStore{
		connections: make(map[string]*models.Connection),
	}
}

// Save stores a connection
func (s *InMemoryConnectionStore) Save(conn *models.Connection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if conn.Provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if conn.UserID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := makeKey(conn.Provider, conn.UserID)
	
	// Set timestamps
	now := time.Now()
	if conn.ID == "" {
		conn.ID = fmt.Sprintf("%s-%d", key, now.Unix())
		conn.CreatedAt = now
	}
	conn.UpdatedAt = now

	s.connections[key] = conn
	return nil
}

// Get retrieves a connection by provider and user ID
func (s *InMemoryConnectionStore) Get(provider, userID string) (*models.Connection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := makeKey(provider, userID)
	conn, exists := s.connections[key]
	if !exists {
		return nil, fmt.Errorf("connection not found for provider %s and user %s", provider, userID)
	}

	return conn, nil
}

// Update updates an existing connection
func (s *InMemoryConnectionStore) Update(conn *models.Connection) error {
	if conn == nil {
		return fmt.Errorf("connection cannot be nil")
	}
	if conn.Provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}
	if conn.UserID == "" {
		return fmt.Errorf("userID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := makeKey(conn.Provider, conn.UserID)
	if _, exists := s.connections[key]; !exists {
		return fmt.Errorf("connection not found for provider %s and user %s", conn.Provider, conn.UserID)
	}

	conn.UpdatedAt = time.Now()
	s.connections[key] = conn
	return nil
}

// Delete removes a connection
func (s *InMemoryConnectionStore) Delete(provider, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := makeKey(provider, userID)
	if _, exists := s.connections[key]; !exists {
		return fmt.Errorf("connection not found for provider %s and user %s", provider, userID)
	}

	delete(s.connections, key)
	return nil
}

// List returns all connections for a user
func (s *InMemoryConnectionStore) List(userID string) ([]*models.Connection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var connections []*models.Connection
	for _, conn := range s.connections {
		if conn.UserID == userID {
			connections = append(connections, conn)
		}
	}

	return connections, nil
}

// makeKey creates a unique key for a connection
func makeKey(provider, userID string) string {
	return fmt.Sprintf("%s:%s", provider, userID)
}
