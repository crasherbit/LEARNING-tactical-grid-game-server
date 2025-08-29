package users

import (
	"context"
	"demondoof-backend/pkg/auth"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service handles user business logic
type Service struct {
	repo      UserRepository
	jwtSecret string
}

// NewService creates a new user service
func NewService(repo UserRepository, jwtSecret string) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, name, email, password string) (*User, error) {
	// Normalize email
	email = strings.ToLower(strings.TrimSpace(email))

	// Validate password
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, email)
	if err != nil && err != ErrUserNotFound {
		return nil, fmt.Errorf("database error: %w", err)
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	// Hash password with bcrypt
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := NewUser(name, email, string(passwordHash))
	if err != nil {
		return nil, err
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user
func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	// Normalize email (same as registration)
	email = strings.ToLower(strings.TrimSpace(email))

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Verify password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// GenerateToken creates a JWT token for a user
func (s *Service) GenerateToken(userID string) (string, error) {
	token, err := auth.GenerateJWT(userID, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}

// GetByID retrieves a user by their ID
func (s *Service) GetByID(ctx context.Context, userID string) (*User, error) {
	// Parse UUID
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Get user from repository
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if err == ErrUserNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}
