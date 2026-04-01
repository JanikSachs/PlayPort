package storage

import (
	"testing"
)

func TestUserStore_Create(t *testing.T) {
	store := NewInMemoryUserStore()

	user, err := store.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if user.ID == "" {
		t.Error("Create() should set a non-empty ID")
	}

	if user.CreatedAt.IsZero() {
		t.Error("Create() should set CreatedAt")
	}

	// IDs should be unique
	user2, err := store.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if user.ID == user2.ID {
		t.Error("Create() should generate unique IDs")
	}
}

func TestUserStore_Get(t *testing.T) {
	store := NewInMemoryUserStore()

	user, err := store.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	retrieved, err := store.Get(user.ID)
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, retrieved.ID)
	}

	if retrieved.CreatedAt != user.CreatedAt {
		t.Error("Retrieved user should have same CreatedAt")
	}
}

func TestUserStore_Get_NotFound(t *testing.T) {
	store := NewInMemoryUserStore()

	_, err := store.Get("nonexistent")
	if err == nil {
		t.Error("Get() should return error for non-existent user")
	}
}

func TestUserStore_CreateWithCredentials(t *testing.T) {
	store := NewInMemoryUserStore()

	user, err := store.CreateWithCredentials("testuser", "hashedpw")
	if err != nil {
		t.Fatalf("CreateWithCredentials() failed: %v", err)
	}

	if user.ID == "" {
		t.Error("CreateWithCredentials() should set a non-empty ID")
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.PasswordHash != "hashedpw" {
		t.Errorf("Expected password hash 'hashedpw', got '%s'", user.PasswordHash)
	}

	if user.CreatedAt.IsZero() {
		t.Error("CreateWithCredentials() should set CreatedAt")
	}
}

func TestUserStore_CreateWithCredentials_DuplicateUsername(t *testing.T) {
	store := NewInMemoryUserStore()

	_, err := store.CreateWithCredentials("testuser", "hashedpw")
	if err != nil {
		t.Fatalf("CreateWithCredentials() failed: %v", err)
	}

	_, err = store.CreateWithCredentials("testuser", "hashedpw2")
	if err == nil {
		t.Error("CreateWithCredentials() should fail for duplicate username")
	}
}

func TestUserStore_GetByUsername(t *testing.T) {
	store := NewInMemoryUserStore()

	created, err := store.CreateWithCredentials("testuser", "hashedpw")
	if err != nil {
		t.Fatalf("CreateWithCredentials() failed: %v", err)
	}

	retrieved, err := store.GetByUsername("testuser")
	if err != nil {
		t.Fatalf("GetByUsername() failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}

	if retrieved.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", retrieved.Username)
	}
}

func TestUserStore_GetByUsername_NotFound(t *testing.T) {
	store := NewInMemoryUserStore()

	_, err := store.GetByUsername("nonexistent")
	if err == nil {
		t.Error("GetByUsername() should return error for non-existent username")
	}
}
