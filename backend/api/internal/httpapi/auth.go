package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"netcord/backend/api/internal/middleware"
	"netcord/backend/api/internal/service"
)

const maxJSONBodySize = 1 << 20

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	result, err := s.authService.Register(r.Context(), input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	result, err := s.authService.Login(r.Context(), input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required", nil)
		return
	}

	user, err := s.authService.GetMe(r.Context(), userID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func readJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()

	decoder := json.NewDecoder(io.LimitReader(r.Body, maxJSONBodySize))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("request body is required")
		}
		return errors.New("request body must be valid json")
	}

	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single json object")
	}

	return nil
}
