package middleware

import (
	"demondoof-backend/internal/features/users"
	"demondoof-backend/pkg/auth"
	"log/slog"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var userKeyUser users.User

const userKeyUserWebSocket = "user"

// AuthMiddleware creates a middleware that tries to authenticate but doesn't fail if no token ₍^. .^₎⟆ ₍^. .^₎⟆ ₍^. .^₎⟆
func AuthMiddleware(jwtSecret string, usrService users.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try to get token from Authorization header first
		authHeader := c.Get("Authorization")
		var tokenString string

		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Fallback to cookie
			tokenString = c.Cookies("auth_token")
		}

		if tokenString != "" {
			// Validate token if present
			claims, err := auth.ValidateJWT(tokenString, jwtSecret)
			if err == nil && claims != nil {
				// retrieve user via provided lookup function
				if u, err := usrService.GetByID(c.Context(), claims.Subject); err == nil && u != nil {
					// save with string for websocket
					c.Locals(userKeyUserWebSocket, u)
					// save with typed key for HTTP
					c.Locals(userKeyUser, u)
				}
			}
		}

		return c.Next()
	}
}

// GetUser extracts full user from context
func GetUser(c *fiber.Ctx) (*users.User, bool) {
	user, ok := c.Locals(userKeyUser).(*users.User)
	return user, ok
}

// GetUserFromWebSocket extracts full user from WebSocket connection
func GetUserFromWebSocket(c *websocket.Conn) (*users.User, bool) {
	user, ok := c.Locals(userKeyUserWebSocket).(*users.User)
	return user, ok
}

// RequireAuth middleware that can be used as a guard
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		test, _ := GetUser(c)
		slog.Info("Require -- Authenticating user", "user", test)

		if _, ok := GetUser(c); !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}
		return c.Next()
	}
}
