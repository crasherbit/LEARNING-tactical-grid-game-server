package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"demondoof-backend/internal/server"
	"demondoof-backend/pkg/config"
	"demondoof-backend/pkg/db"
	"demondoof-backend/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup logger (unified initialization)
	log := logger.New(logger.ParseLevel(cfg.LogLevel))
	log.SetDefault()

	slog.Info("Configuration loaded", "port", cfg.Port, "logLevel", cfg.LogLevel)

	// Connect to database
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create server
	srv := server.New(pool, &cfg.Config)

	// Start server in a goroutine
	go func() {
		slog.Info("Server starting", "port", cfg.Port)
		if err := srv.Start(); err != nil {
			slog.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server shutting down...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.GetApp().ShutdownWithContext(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	} else {
		slog.Info("Server exited gracefully")
	}
}
