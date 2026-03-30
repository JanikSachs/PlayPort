package auth

import (
	"testing"
	"time"
)

func TestSessionStore_Create(t *testing.T) {
	store := NewInMemorySessionStore(0)

	token1, err := store.Create("user1")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if token1 == "" {
		t.Error("Create() should return a non-empty token")
	}

	// Tokens should be unique
	token2, err := store.Create("user1")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if token1 == token2 {
		t.Error("Create() should generate unique tokens")
	}
}

func TestSessionStore_Get(t *testing.T) {
	store := NewInMemorySessionStore(0)

	token, err := store.Create("user123")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	userID, err := store.Get(token)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if userID != "user123" {
		t.Errorf("Expected userID 'user123', got '%s'", userID)
	}
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	store := NewInMemorySessionStore(0)

	_, err := store.Get("invalid-token")
	if err == nil {
		t.Error("Get() should return error for non-existent token")
	}
}

func TestSessionStore_Get_Expired(t *testing.T) {
	store := NewInMemorySessionStore(1 * time.Millisecond)

	token, err := store.Create("user123")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Manually expire the session
	store.mu.Lock()
	store.sessions[token].expiresAt = time.Now().Add(-1 * time.Minute)
	store.mu.Unlock()

	_, err = store.Get(token)
	if err == nil {
		t.Error("Get() should return error for expired session")
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewInMemorySessionStore(0)

	token, err := store.Create("user123")
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	err = store.Delete(token)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	_, err = store.Get(token)
	if err == nil {
		t.Error("Get() should return error after Delete()")
	}
}

func TestSessionStore_Delete_NotFound(t *testing.T) {
	store := NewInMemorySessionStore(0)

	err := store.Delete("nonexistent-token")
	if err == nil {
		t.Error("Delete() should return error for non-existent token")
	}
}
