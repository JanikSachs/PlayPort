package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword() should return a non-empty hash")
	}

	if hash == "testpassword123" {
		t.Error("Hash should not equal plaintext password")
	}
}

func TestHashPassword_Unique(t *testing.T) {
	hash1, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	hash2, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if hash1 == hash2 {
		t.Error("Two hashes of the same password should be different (different salts)")
	}
}

func TestCheckPassword_Success(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if err := CheckPassword(hash, "testpassword123"); err != nil {
		t.Errorf("CheckPassword() should succeed for correct password: %v", err)
	}
}

func TestCheckPassword_Failure(t *testing.T) {
	hash, err := HashPassword("testpassword123")
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	if err := CheckPassword(hash, "wrongpassword"); err == nil {
		t.Error("CheckPassword() should fail for wrong password")
	}
}
