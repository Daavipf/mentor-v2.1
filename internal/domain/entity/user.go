package entity

import (
	"time"
)

// User represents an application user matching the Users model in schema.prisma.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Never expose password hash in JSON responses
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

