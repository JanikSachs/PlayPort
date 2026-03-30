package models

import "time"

// User represents an application user
type User struct {
	ID        string
	CreatedAt time.Time
}
