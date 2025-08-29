package users

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User domain model
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // Password hash, never return in JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetID implements auth.User interface
func (u *User) GetID() string {
	return u.ID.String()
}

// GetEmail implements auth.User interface
func (u *User) GetEmail() string {
	return u.Email
}

// GetName implements auth.User interface
func (u *User) GetName() string {
	return u.Name
}

// Business rules and validation
var (
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrNameRequired     = errors.New("name is required")
	ErrNameTooShort     = errors.New("name must be at least 3 characters")
	ErrNameTooLong      = errors.New("name must be 20 characters or less")
	ErrEmailRequired    = errors.New("email is required")
	ErrPasswordRequired = errors.New("password is required")
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrNameExists       = errors.New("name already exists")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateUser validates user data according to business rules
func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrNameRequired
	}

	if len(u.Name) < 3 {
		return ErrNameTooShort
	}

	if len(u.Name) > 20 {
		return ErrNameTooLong
	}

	if strings.TrimSpace(u.Email) == "" {
		return ErrEmailRequired
	}

	if !emailRegex.MatchString(u.Email) {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return ErrPasswordRequired
	}

	if len(password) < 6 {
		return ErrPasswordTooShort
	}

	return nil
}

// NewUser creates a new user with validation
func NewUser(name, email, passwordHash string) (*User, error) {
	user := &User{
		ID:        uuid.New(),
		Name:      strings.TrimSpace(name),
		Email:     strings.ToLower(strings.TrimSpace(email)),
		Password:  passwordHash,
		CreatedAt: time.Now(),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}
