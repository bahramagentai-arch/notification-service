package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BahramRousta/notification-service/config"
	"github.com/BahramRousta/notification-service/internal/notification/providers/fake"
	"github.com/BahramRousta/notification-service/internal/notification/service"
	pgstore "github.com/BahramRousta/notification-service/internal/notification/store"
	"github.com/BahramRousta/notification-service/internal/notification/transport/httpapi"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		logger.Debug("no .env file found")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("invalid configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx := context.Background()
	logger.Info("connecting to database")
	db, err := pgstore.NewStore(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("running database migrations")
	if err := runMigrations(cfg.DatabaseURL); err != nil {
		logger.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("migrations applied successfully")

	provider := fake.New(logger)
	_ = provider

	var repo pgstore.Repository = db
	svc := service.NewService(repo, logger)
	handler := httpapi.New(svc, logger)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{Addr: cfg.HttpAddr, Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("HTTP server starting", slog.String("addr", cfg.HttpAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		logger.Error("server crashed", slog.String("error", err.Error()))

	case sig := <-quit:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down HTTP server...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced shutdown", slog.String("error", err.Error()))
	}
	logger.Info("closing database connection pool...")
	db.Close()
	logger.Info("shutdown complete")

}

func runMigrations(databaseURL string) error {
	mg, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer mg.Close()

	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}
