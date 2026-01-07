package storage

import (
	"testing"
	"time"

	"github.com/JanikSachs/PlayPort/internal/models"
)

func TestConnectionStore_Save(t *testing.T) {
	store := NewInMemoryConnectionStore()

	conn := &models.Connection{
		Provider:       "spotify",
		UserID:         "user123",
		AccessToken:    "access-token",
		RefreshToken:   "refresh-token",
		ExpiresAt:      time.Now().Add(1 * time.Hour),
		Connected:      true,
	}

	err := store.Save(conn)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	if conn.ID == "" {
		t.Error("Save() should set ID")
	}

	if conn.CreatedAt.IsZero() {
		t.Error("Save() should set CreatedAt")
	}

	if conn.UpdatedAt.IsZero() {
		t.Error("Save() should set UpdatedAt")
	}
}

func TestConnectionStore_Save_Validation(t *testing.T) {
	store := NewInMemoryConnectionStore()

	tests := []struct {
		name string
		conn *models.Connection
	}{
		{
			name: "nil connection",
			conn: nil,
		},
		{
			name: "empty provider",
			conn: &models.Connection{UserID: "user123"},
		},
		{
			name: "empty userID",
			conn: &models.Connection{Provider: "spotify"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Save(tt.conn)
			if err == nil {
				t.Error("Save() should return error for invalid connection")
			}
		})
	}
}

func TestConnectionStore_Get(t *testing.T) {
	store := NewInMemoryConnectionStore()

	conn := &models.Connection{
		Provider:     "spotify",
		UserID:       "user123",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		Connected:    true,
	}

	err := store.Save(conn)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	retrieved, err := store.Get("spotify", "user123")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.Provider != conn.Provider {
		t.Errorf("Expected provider %s, got %s", conn.Provider, retrieved.Provider)
	}

	if retrieved.UserID != conn.UserID {
		t.Errorf("Expected userID %s, got %s", conn.UserID, retrieved.UserID)
	}

	if retrieved.AccessToken != conn.AccessToken {
		t.Errorf("Expected access token %s, got %s", conn.AccessToken, retrieved.AccessToken)
	}
}

func TestConnectionStore_Get_NotFound(t *testing.T) {
	store := NewInMemoryConnectionStore()

	_, err := store.Get("spotify", "nonexistent")
	if err == nil {
		t.Error("Get() should return error for non-existent connection")
	}
}

func TestConnectionStore_Update(t *testing.T) {
	store := NewInMemoryConnectionStore()

	conn := &models.Connection{
		Provider:     "spotify",
		UserID:       "user123",
		AccessToken:  "old-token",
		RefreshToken: "refresh-token",
		Connected:    true,
	}

	err := store.Save(conn)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Update the connection
	conn.AccessToken = "new-token"
	err = store.Update(conn)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	// Verify update
	retrieved, err := store.Get("spotify", "user123")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.AccessToken != "new-token" {
		t.Errorf("Expected access token 'new-token', got %s", retrieved.AccessToken)
	}
}

func TestConnectionStore_Update_NotFound(t *testing.T) {
	store := NewInMemoryConnectionStore()

	conn := &models.Connection{
		Provider: "spotify",
		UserID:   "nonexistent",
	}

	err := store.Update(conn)
	if err == nil {
		t.Error("Update() should return error for non-existent connection")
	}
}

func TestConnectionStore_Delete(t *testing.T) {
	store := NewInMemoryConnectionStore()

	conn := &models.Connection{
		Provider:    "spotify",
		UserID:      "user123",
		AccessToken: "access-token",
		Connected:   true,
	}

	err := store.Save(conn)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Delete the connection
	err = store.Delete("spotify", "user123")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deletion
	_, err = store.Get("spotify", "user123")
	if err == nil {
		t.Error("Get() should return error after deletion")
	}
}

func TestConnectionStore_Delete_NotFound(t *testing.T) {
	store := NewInMemoryConnectionStore()

	err := store.Delete("spotify", "nonexistent")
	if err == nil {
		t.Error("Delete() should return error for non-existent connection")
	}
}

func TestConnectionStore_List(t *testing.T) {
	store := NewInMemoryConnectionStore()

	// Add multiple connections for the same user
	conn1 := &models.Connection{
		Provider:    "spotify",
		UserID:      "user123",
		AccessToken: "token1",
		Connected:   true,
	}

	conn2 := &models.Connection{
		Provider:    "apple-music",
		UserID:      "user123",
		AccessToken: "token2",
		Connected:   true,
	}

	conn3 := &models.Connection{
		Provider:    "spotify",
		UserID:      "user456",
		AccessToken: "token3",
		Connected:   true,
	}

	store.Save(conn1)
	store.Save(conn2)
	store.Save(conn3)

	// List connections for user123
	connections, err := store.List("user123")
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(connections) != 2 {
		t.Errorf("Expected 2 connections for user123, got %d", len(connections))
	}

	// List connections for user456
	connections, err = store.List("user456")
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(connections) != 1 {
		t.Errorf("Expected 1 connection for user456, got %d", len(connections))
	}
}
