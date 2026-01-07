package auth

import (
	"testing"
	"time"
)

func TestStateStore_Generate(t *testing.T) {
	store := NewInMemoryStateStore()

	state1, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if state1 == "" {
		t.Error("Generated state should not be empty")
	}

	// Generate another state
	state2, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	if state1 == state2 {
		t.Error("Generated states should be unique")
	}
}

func TestStateStore_Validate(t *testing.T) {
	store := NewInMemoryStateStore()

	// Generate a state
	state, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Valid state should pass validation
	if !store.Validate(state) {
		t.Error("Validate() should return true for valid state")
	}

	// State should be removed after validation (one-time use)
	if store.Validate(state) {
		t.Error("Validate() should return false for already used state")
	}

	// Invalid state should fail validation
	if store.Validate("invalid-state") {
		t.Error("Validate() should return false for invalid state")
	}
}

func TestStateStore_Expiration(t *testing.T) {
	store := NewInMemoryStateStore()

	// Generate a state
	state, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	// Manually expire the state
	store.mu.Lock()
	store.states[state] = time.Now().Add(-1 * time.Minute)
	store.mu.Unlock()

	// Expired state should fail validation
	if store.Validate(state) {
		t.Error("Validate() should return false for expired state")
	}
}
