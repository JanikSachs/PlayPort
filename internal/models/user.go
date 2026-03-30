package models

import "time"

// User represents an application user
type User struct {
	ID           string
	Username     string
	PasswordHash string // bcrypt hash
	CreatedAt    time.Time
}
