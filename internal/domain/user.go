package domain

import (
	"github.com/google/uuid"
)

type User struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

// NewUser - конструктор для структуры User
func NewUser(email string) *User {
	return &User{
		ID:    uuid.New(),
		Email: email,
	}
}
