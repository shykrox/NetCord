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
	authService *service.AuthService
}

func NewRouter(authService *service.AuthService, tokens *auth.TokenManager) http.Handler {
	server := &Server{authService: authService}
	mux := http.NewServeMux()

	mux.HandleFunc("/health", method(http.MethodGet, server.health))
	mux.HandleFunc("/auth/register", method(http.MethodPost, server.register))
	mux.HandleFunc("/auth/login", method(http.MethodPost, server.login))
	mux.Handle("/users/me", methodHandler(http.MethodGet, middleware.RequireAuth(tokens)(http.HandlerFunc(server.me))))

	return mux
}

func method(allowed string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowed {
			w.Header().Set("Allow", allowed)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		handler(w, r)
	}
}

func methodHandler(allowed string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowed {
			w.Header().Set("Allow", allowed)
			writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func (s *Server) writeServiceError(w http.ResponseWriter, err error) {
	var validation *service.ValidationError
	switch {
	case errors.As(err, &validation):
		writeError(w, http.StatusBadRequest, "validation_error", "request validation failed", validation.Fields)
	case errors.Is(err, service.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "invalid email or password", nil)
	case errors.Is(err, repository.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user_not_found", "user not found", nil)
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error", nil)
	}
}
