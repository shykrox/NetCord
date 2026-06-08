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

func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input service.UpdateMeInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	user, err := s.authService.UpdateMe(r.Context(), userID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := pathUUID(w, r, "user_id")
	if !ok {
		return
	}

	user, err := s.authService.GetUser(r.Context(), userID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (s *Server) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	s.uploadProfileImage(w, r, true)
}

func (s *Server) uploadBanner(w http.ResponseWriter, r *http.Request) {
	s.uploadProfileImage(w, r, false)
}

func (s *Server) uploadProfileImage(w http.ResponseWriter, r *http.Request, avatar bool) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	if s.fileService == nil {
		writeError(w, http.StatusInternalServerError, "files_unavailable", "file service is unavailable", nil)
		return
	}

	maxUploadBytes := s.fileService.MaxUploadBytes()
	if maxUploadBytes > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes+multipartOverheadBytes)
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_multipart", "multipart field file is required", nil)
		return
	}
	defer file.Close()

	attachment, err := s.fileService.Upload(r.Context(), userID, service.UploadFileInput{
		OriginalFilename: header.Filename,
		SizeBytes:        header.Size,
		Reader:           file,
	})
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	var user interface{}
	if avatar {
		user, err = s.authService.SetAvatarURL(r.Context(), userID, attachment.DownloadURL)
	} else {
		user, err = s.authService.SetBannerURL(r.Context(), userID, attachment.DownloadURL)
	}
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
