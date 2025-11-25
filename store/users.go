package store

import (
	"errors"
	"time"

	"example.com/m/models"
	"golang.org/x/crypto/bcrypt"
)

// In-memory store
var users = map[string]models.User{}

// AddUser — hashes the password and adds user to map
func AddUser(id, email, password string, createdAt time.Time) error {
	if _, exists := users[email]; exists {
		return errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users[email] = models.User{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
		CreatedAt:    createdAt,
	}
	return nil
}

// GetUser — fetches user by email
func GetUser(email string) (models.User, bool) {
	user, ok := users[email]
	return user, ok
}
