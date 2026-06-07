package httpapi

import (
	"net/http"

	"netcord/backend/api/internal/middleware"
	"netcord/backend/api/internal/service"

	"github.com/google/uuid"
)

func (s *Server) createServer(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input service.CreateServerInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	server, err := s.serverService.CreateServer(r.Context(), userID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, server)
}

func (s *Server) listServers(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	servers, err := s.serverService.ListServers(r.Context(), userID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, servers)
}

func (s *Server) getServer(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}

	server, err := s.serverService.GetServer(r.Context(), userID, serverID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, server)
}

func (s *Server) createChannel(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}

	var input service.CreateChannelInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	channel, err := s.serverService.CreateChannel(r.Context(), userID, serverID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, channel)
}

func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}

	channels, err := s.serverService.ListChannels(r.Context(), userID, serverID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, channels)
}

func (s *Server) listMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	channelID, ok := pathUUID(w, r, "channel_id")
	if !ok {
		return
	}

	messages, err := s.serverService.ListMessages(r.Context(), userID, channelID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

func (s *Server) createMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	channelID, ok := pathUUID(w, r, "channel_id")
	if !ok {
		return
	}

	var input service.CreateMessageInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	message, err := s.serverService.CreateMessage(r.Context(), userID, channelID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, message)
}

func currentUserID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required", nil)
		return uuid.Nil, false
	}
	return userID, true
}

func pathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", name+" must be a valid uuid", nil)
		return uuid.Nil, false
	}
	return id, true
}
