package customer

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrGeneratePassword = errors.New("password generation failed")

// Repository represents our entity storage for a Customer.
type Repository interface {
	// Add a new customer to the repository.
	Add(ctx context.Context, c *Customer) error

	// FindByUsername searches the repository for a Customer with the given username.
	// If none is found then nil is returned.
	FindByUsername(ctx context.Context, username string) (*Customer, error)
}

// Customer is the main user of our application.
type Customer struct {
	id           uuid.UUID
	username     string
	passwordHash string
}

// New will create a new Customer and assign a new uuid.
func New(username, password string) (*Customer, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return nil, fmt.Errorf("unable to create customer: %w", ErrGeneratePassword)
	}

	return &Customer{
		id:           uuid.New(),
		username:     username,
		passwordHash: string(bytes),
	}, nil
}

// Restore creates an Customer from existing credentials.
// The password wll not be hashed and an existing id is required.
func Restore(id uuid.UUID, username, hashedPassword string) *Customer {
	return &Customer{
		id:           id,
		username:     username,
		passwordHash: hashedPassword,
	}
}

// ID uniquely identifies a Customer within the application.
func (c *Customer) ID() uuid.UUID {
	return c.id
}

// Username is the email address of the Customer.
func (c *Customer) Username() string {
	return c.username
}

// PasswordHash is the hashed version of the customers password.
func (c *Customer) PasswordHash() string {
	return c.passwordHash
}

// HasPassword returns true if the internal password hash matches the given password.
func (c *Customer) HasPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.passwordHash), []byte(password))

	return err == nil
}
