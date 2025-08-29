package deps

import (
	"demondoof-backend/internal/features/users"
	"demondoof-backend/pkg/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Dependencies holds all application-wide initialized services and repos
type Dependencies struct {
	Pool        *pgxpool.Pool
	Cfg         *config.Config
	UserRepo    users.UserRepository
	UserService *users.Service
}

// Bootstrap initializes application dependencies ₍^. .^₎⟆
func Bootstrap(pool *pgxpool.Pool, cfg *config.Config) (*Dependencies, error) {
	// repositories
	userRepo := users.NewRepository(pool)

	// services
	userService := users.NewService(userRepo, cfg.JWTSecret)

	return &Dependencies{
		Pool:        pool,
		Cfg:         cfg,
		UserRepo:    userRepo,
		UserService: userService,
	}, nil
}
