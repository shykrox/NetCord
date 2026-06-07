package httpapi

import (
	"errors"
	"net/http"

	"netcord/backend/api/internal/auth"
	"netcord/backend/api/internal/middleware"
	"netcord/backend/api/internal/repository"
	"netcord/backend/api/internal/service"
)

type Server struct {
	authService   *service.AuthService
	serverService *service.ServerService
}

func NewRouter(authService *service.AuthService, serverService *service.ServerService, tokens *auth.TokenManager) http.Handler {
	server := &Server{authService: authService, serverService: serverService}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("POST /auth/register", server.register)
	mux.HandleFunc("POST /auth/login", server.login)
	mux.Handle("GET /users/me", protected(tokens, server.me))

	mux.Handle("POST /servers", protected(tokens, server.createServer))
	mux.Handle("GET /servers", protected(tokens, server.listServers))
	mux.Handle("GET /servers/{server_id}", protected(tokens, server.getServer))
	mux.Handle("POST /servers/{server_id}/channels", protected(tokens, server.createChannel))
	mux.Handle("GET /servers/{server_id}/channels", protected(tokens, server.listChannels))
	mux.Handle("GET /channels/{channel_id}/messages", protected(tokens, server.listMessages))
	mux.Handle("POST /channels/{channel_id}/messages", protected(tokens, server.createMessage))

	return mux
}

func protected(tokens *auth.TokenManager, handler http.HandlerFunc) http.Handler {
	return middleware.RequireAuth(tokens)(http.HandlerFunc(handler))
}

func (s *Server) writeServiceError(w http.ResponseWriter, err error) {
	var validation *service.ValidationError
	switch {
	case errors.As(err, &validation):
		writeError(w, http.StatusBadRequest, "validation_error", "request validation failed", validation.Fields)
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", "resource not found", nil)
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user_not_found", "user not found", nil)
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
