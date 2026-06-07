package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"netcord/backend/api/internal/auth"
)

type authErrorResponse struct {
	Error authError `json:"error"`
}

type authError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RequireAuth(tokens *auth.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				writeUnauthorized(w)
				return
			}

			tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if tokenString == "" {
				writeUnauthorized(w)
				return
			}

			if tokens == nil {
				writeUnauthorized(w)
				return
			}

			userID, err := tokens.Validate(tokenString)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(authErrorResponse{
		Error: authError{
			Code:    "unauthorized",
			Message: "valid bearer token required",
		},
	})
}
