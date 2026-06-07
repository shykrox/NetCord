package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"netcord/backend/api/internal/auth"
	"netcord/backend/api/internal/config"
	"netcord/backend/api/internal/db"
	"netcord/backend/api/internal/gateway"
	"netcord/backend/api/internal/httpapi"
	"netcord/backend/api/internal/repository"
	"netcord/backend/api/internal/service"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	tokenManager, err := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.TokenTTL)
	if err != nil {
		logger.Error("configure jwt", "error", err)
		os.Exit(1)
	}

	userRepo := repository.NewPostgresUserRepository(pool)
	serverRepo := repository.NewPostgresServerRepository(pool)
	authService := service.NewAuthService(userRepo, tokenManager)
	serverService := service.NewServerService(serverRepo)
	gatewayHub := gateway.NewHub()
	router := httpapi.NewRouter(authService, serverService, tokenManager, gatewayHub)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("starting netcord api", "addr", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
