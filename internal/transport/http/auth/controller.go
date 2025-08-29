package auth

import (
	"demondoof-backend/internal/features/users"
	"demondoof-backend/pkg/middleware"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	userService *users.Service
	authService *Service
	app         *fiber.App
}

func NewController(userService *users.Service) *Controller {
	// Create auth transport service (HTTP logic)
	authTransportService := NewService()

	// Create Fiber app and setup routes
	app := fiber.New()

	ctrl := &Controller{
		userService: userService,
		authService: authTransportService,
		app:         app,
	}

	// Setup routes
	ctrl.app.Post("/register", ctrl.Register)
	ctrl.app.Post("/login", ctrl.Login)

	// Protected routes with database validation
	ctrl.app.Use(middleware.RequireAuth())
	ctrl.app.Get("/me", middleware.RequireAuth(), ctrl.Me)

	return ctrl
}

func (ctrl *Controller) GetApp() *fiber.App {
	return ctrl.app
}

func (ctrl *Controller) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parse request
	if err := ctrl.authService.ParseRequest(c, &req); err != nil {
		return ctrl.authService.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Register user (business logic handles all validation)
	user, err := ctrl.userService.Register(c.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return ctrl.authService.RespondError(c, fiber.StatusBadRequest, "Registration failed")
	}

	// Generate authentication token
	token, err := ctrl.userService.GenerateToken(user.ID.String())
	if err != nil {
		return ctrl.authService.RespondError(c, fiber.StatusInternalServerError, "Authentication failed")
	}

	// Convert response and log
	response := ctrl.authService.ConvertToAuthResponse(user, token)
	ctrl.authService.LogRegistration(response.User.ID, response.User.Email)

	return ctrl.authService.RespondSuccess(c, response)
}

func (ctrl *Controller) Login(c *fiber.Ctx) error {
	var req LoginRequest

	// Parse request
	if err := ctrl.authService.ParseRequest(c, &req); err != nil {
		return ctrl.authService.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Login user (business logic handles all validation)
	user, err := ctrl.userService.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		return ctrl.authService.RespondError(c, fiber.StatusUnauthorized, "Authentication failed")
	}

	// Generate authentication token
	token, err := ctrl.userService.GenerateToken(user.ID.String())
	if err != nil {
		return ctrl.authService.RespondError(c, fiber.StatusInternalServerError, "Authentication failed")
	}

	// Convert response and log
	response := ctrl.authService.ConvertToAuthResponse(user, token)
	ctrl.authService.LogLogin(response.User.ID, response.User.Email)

	return ctrl.authService.RespondSuccess(c, response)
}

func (ctrl *Controller) Me(c *fiber.Ctx) error {
	// Get user from context (already validated by middleware with DB lookup)
	usr, ok := middleware.GetUser(c)
	if !ok || usr == nil {
		return ctrl.authService.RespondError(c, fiber.StatusUnauthorized, "Authentication required")
	}

	// Return DTO from context user (no extra DB query needed)
	return ctrl.authService.RespondSuccess(c, UserDTO{ID: usr.ID.String(), Name: usr.Name, Email: usr.Email})
}
