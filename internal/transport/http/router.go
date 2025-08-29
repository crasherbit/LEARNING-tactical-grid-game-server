package http

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"demondoof-backend/internal/server/deps"
	authController "demondoof-backend/internal/transport/http/auth"
)

type HttpRouter struct {
	app  *fiber.App
	deps *deps.Dependencies
}

func NewHttpRouter(deps *deps.Dependencies) *HttpRouter {
	app := fiber.New()

	router := &HttpRouter{
		app:  app,
		deps: deps,
	}

	// Health endpoints ₍^. .^₎⟆
	app.Get("/health", router.healthCheck)

	// API versioning
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Create auth controller with injected user service
	authCtrl := authController.NewController(deps.UserService)

	// Mount routes (Fiber Mount pattern)
	v1.Mount("/auth", authCtrl.GetApp())

	return router
}

func (r *HttpRouter) GetApp() *fiber.App {
	return r.app
}

func (r *HttpRouter) healthCheck(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	if err := r.deps.Pool.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":     "unavailable",
			"database":   "disconnected",
			"time":       time.Now().Format(time.RFC3339),
			"error":      "database ping failed",
			"serverName": "DemonDoof Backend",
			"version":    "v1.0.0",
		})
	}
	return c.JSON(fiber.Map{
		"status":     "ready",
		"database":   "connected",
		"time":       time.Now().Format(time.RFC3339),
		"serverName": "DemonDoof Backend",
		"version":    "v1.0.0",
	})
}
