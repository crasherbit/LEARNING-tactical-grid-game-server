package ws

import (
	"demondoof-backend/internal/server/deps"
	"demondoof-backend/pkg/middleware"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type WebSocketRouter struct {
	app *fiber.App
}

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewWebSocketRouter(_ *deps.Dependencies) *WebSocketRouter {
	app := fiber.New()

	app.Get("/", NewHandler())

	return &WebSocketRouter{app: app}
}

func (r *WebSocketRouter) GetApp() *fiber.App {
	return r.app
}

func NewHandler() fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		// Get user from context (set by auth middleware)
		user, ok := middleware.GetUserFromWebSocket(c)
		if !ok {
			slog.Error("WebSocket handler called without user authentication")
			c.Close()
			return
		}

		slog.Info("WebSocket connection established", "userId", user.ID, "userEmail", user.Email)

		defer func() {
			slog.Info("WebSocket connection closed", "userId", user.ID, "userEmail", user.Email)
			c.Close()
		}()

		for {
			var msg Message
			if err := c.ReadJSON(&msg); err != nil {
				slog.Warn("Error reading WebSocket message", "error", err, "userId", user.ID)
				break
			}

			slog.Debug("Received WebSocket message", "type", msg.Type, "userId", user.ID)

			switch msg.Type {
			case "ping":
				response := Message{
					Type: "pong",
					Data: map[string]interface{}{
						"timestamp": time.Now().Unix(),
						"userId":    user.ID,
						"userEmail": user.Email,
						"userName":  user.Name,
					},
				}
				if err := c.WriteJSON(response); err != nil {
					slog.Warn("Error sending pong", "error", err, "userId", user.ID)
					return
				}
			default:
				// Ignore unknown message types for now
				slog.Debug("Unknown message type", "type", msg.Type, "userId", user.ID)
			}
		}
	})
}
