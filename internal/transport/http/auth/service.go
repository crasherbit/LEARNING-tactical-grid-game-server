package auth

import (
	"log/slog"

	"demondoof-backend/internal/features/users"

	"github.com/gofiber/fiber/v2"
)

// Service handles HTTP transport logic for authentication
type Service struct{}

// NewService creates a new auth transport service
func NewService() *Service {
	return &Service{}
}

// ParseRequest parses and validates the request body
func (s *Service) ParseRequest(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		slog.Warn("Failed to parse request body", "error", err)
		return err
	}
	return nil
}

// ConvertToAuthResponse converts user and token to HTTP DTO
func (s *Service) ConvertToAuthResponse(user *users.User, token string) *AuthResponse {
	return &AuthResponse{
		Token: token,
		User: UserDTO{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		},
	}
}

// RespondSuccess sends a successful JSON response
func (s *Service) RespondSuccess(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(data)
}

// RespondError sends an error JSON response
func (s *Service) RespondError(c *fiber.Ctx, statusCode int, message string) error {
	slog.Warn("HTTP error response", "status", statusCode, "message", message)
	return c.Status(statusCode).JSON(ErrorResponse{
		Error:   "error",
		Message: message,
	})
}

// LogRegistration logs successful user registration
func (s *Service) LogRegistration(userID, email string) {
	slog.Info("User registered successfully", "userId", userID, "email", email)
}

// LogLogin logs successful user login
func (s *Service) LogLogin(userID, email string) {
	slog.Info("User logged in successfully", "userId", userID, "email", email)
}
