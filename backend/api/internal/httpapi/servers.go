package httpapi

import (
	"net/http"
	"strconv"

	"netcord/backend/api/internal/gateway"
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

	input, ok := listMessagesInputFromQuery(w, r)
	if !ok {
		return
	}

	messages, err := s.serverService.ListMessages(r.Context(), userID, channelID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

func (s *Server) searchMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	channelID, ok := pathUUID(w, r, "channel_id")
	if !ok {
		return
	}

	limit, ok := queryLimit(w, r)
	if !ok {
		return
	}

	messages, err := s.serverService.SearchMessages(r.Context(), userID, channelID, service.SearchMessagesInput{
		Query: r.URL.Query().Get("q"),
		Limit: limit,
	})
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

	if s.gatewayHub != nil {
		s.gatewayHub.BroadcastMessageCreated(message)
	}
	if s.aiService != nil && s.gatewayHub != nil {
		job, created, err := s.aiService.EnqueueFromMessage(r.Context(), message)
		if err != nil {
			s.writeServiceError(w, err)
			return
		}
		if created {
			s.gatewayHub.Broadcast(message.ServerID, gateway.JobProgressEvent(job))
		}
	}

	writeJSON(w, http.StatusCreated, message)
}

func (s *Server) updateMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	messageID, ok := pathUUID(w, r, "message_id")
	if !ok {
		return
	}

	var input service.UpdateMessageInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	message, err := s.serverService.UpdateMessage(r.Context(), userID, messageID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	if s.gatewayHub != nil {
		s.gatewayHub.BroadcastMessageUpdated(message)
	}

	writeJSON(w, http.StatusOK, message)
}

func (s *Server) deleteMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	messageID, ok := pathUUID(w, r, "message_id")
	if !ok {
		return
	}

	message, err := s.serverService.DeleteMessage(r.Context(), userID, messageID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	if s.gatewayHub != nil {
		s.gatewayHub.BroadcastMessageDeleted(message.ID, message.ServerID, message.ChannelID)
	}

	writeJSON(w, http.StatusOK, message)
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

func listMessagesInputFromQuery(w http.ResponseWriter, r *http.Request) (service.ListMessagesInput, bool) {
	limit, ok := queryLimit(w, r)
	if !ok {
		return service.ListMessagesInput{}, false
	}

	before, ok := optionalQueryUUID(w, r, "before")
	if !ok {
		return service.ListMessagesInput{}, false
	}
	after, ok := optionalQueryUUID(w, r, "after")
	if !ok {
		return service.ListMessagesInput{}, false
	}

	return service.ListMessagesInput{
		Before: before,
		After:  after,
		Limit:  limit,
	}, true
}

func optionalQueryUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return uuid.Nil, true
	}

	id, err := uuid.Parse(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", name+" must be a valid uuid", nil)
		return uuid.Nil, false
	}
	return id, true
}

func queryLimit(w http.ResponseWriter, r *http.Request) (int, bool) {
	value := r.URL.Query().Get("limit")
	if value == "" {
		return 0, true
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", "limit must be a number", nil)
		return 0, false
	}
	return limit, true
}
