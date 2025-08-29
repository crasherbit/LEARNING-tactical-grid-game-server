package server

import (
	"demondoof-backend/pkg/config"
	"demondoof-backend/pkg/middleware"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"

	"demondoof-backend/internal/server/deps"
	httpRouter "demondoof-backend/internal/transport/http"
	wsRouter "demondoof-backend/internal/transport/ws"
)

// Server represents the HTTP server
type Server struct {
	app  *fiber.App
	port int
}

// New creates a new server instance
func New(pool *pgxpool.Pool, cfg *config.Config) *Server {
	// Create main Fiber app
	app := fiber.New(fiber.Config{
		ServerHeader: "DemonDoof Backend",
		AppName:      "DemonDoof v1.0.0",
		BodyLimit:    2 * 1024 * 1024, // 2MB body limit
	})

	// Add global middleware
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format:     "${time} | ${status} | ${latency} | ${ip} | ${method} ${path} | ${error}\n",
		TimeFormat: "02/01/2006 15:04:05",
	}))
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(compress.New())
	app.Use(limiter.New(limiter.Config{
		Max: 100, // 100 requests per minute per IP
	}))
	// Bootstrap application dependencies
	deps, err := deps.Bootstrap(pool, cfg)
	if err != nil {
		slog.Error("Failed to bootstrap dependencies", "error", err)
	}

	// Auth middleware with user lookup using existing UserService.GetByID
	app.Use(middleware.AuthMiddleware(cfg.JWTSecret, *deps.UserService))

	// HTTP routes (public)
	http := httpRouter.NewHttpRouter(deps)
	app.Mount("/", http.GetApp())

	// WebSocket routes (protected)
	ws := wsRouter.NewWebSocketRouter(deps)
	app.Mount("/ws", ws.GetApp())

	return &Server{
		app:  app,
		port: cfg.Port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	slog.Info("Server starting", "address", addr)
	return s.app.Listen(addr)
}

// GetApp returns the Fiber app instance
func (s *Server) GetApp() *fiber.App {
	return s.app
}
