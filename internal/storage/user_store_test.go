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
