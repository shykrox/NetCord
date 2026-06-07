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
	"netcord/backend/api/internal/storage"
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
	objectStore, err := storage.NewMinIOObjectStore(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey)
	if err != nil {
		logger.Error("configure minio", "error", err)
		os.Exit(1)
	}
	if err := objectStore.EnsureBucket(ctx, cfg.MinIOBucketAttachments); err != nil {
		logger.Error("prepare minio attachments bucket", "error", err)
		os.Exit(1)
	}

	authService := service.NewAuthService(userRepo, tokenManager)
	serverService := service.NewServerService(serverRepo)
	fileService := service.NewFileService(serverRepo, objectStore, cfg.MinIOBucketAttachments, cfg.MaxUploadBytes)
	gatewayHub := gateway.NewHub()
	router := httpapi.NewRouter(authService, serverService, fileService, tokenManager, gatewayHub)

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
