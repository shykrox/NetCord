package httpapi

import (
	"net/http"

	"netcord/backend/api/internal/gateway"
	"netcord/backend/api/internal/service"
)

func (s *Server) createFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input service.CreateFriendRequestInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	request, err := s.socialService.CreateFriendRequest(r.Context(), userID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	if s.gatewayHub != nil {
		s.gatewayHub.BroadcastUser(request.RecipientID, gateway.FriendRequestedEvent(request))
	}

	writeJSON(w, http.StatusCreated, request)
}

func (s *Server) listFriendRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	requests, err := s.socialService.ListFriendRequests(r.Context(), userID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

func (s *Server) acceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	requestID, ok := pathUUID(w, r, "request_id")
	if !ok {
		return
	}

	request, err := s.socialService.AcceptFriendRequest(r.Context(), userID, requestID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	if s.gatewayHub != nil {
		event := gateway.FriendAcceptedEvent(request)
		s.gatewayHub.BroadcastUser(request.RequesterID, event)
		s.gatewayHub.BroadcastUser(request.RecipientID, event)
	}

	writeJSON(w, http.StatusOK, request)
}

func (s *Server) declineFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	requestID, ok := pathUUID(w, r, "request_id")
	if !ok {
		return
	}

	request, err := s.socialService.DeclineFriendRequest(r.Context(), userID, requestID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, request)
}

func (s *Server) listFriends(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	friends, err := s.socialService.ListFriends(r.Context(), userID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, friends)
}

func (s *Server) removeFriend(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	friendID, ok := pathUUID(w, r, "user_id")
	if !ok {
		return
	}

	if err := s.socialService.RemoveFriend(r.Context(), userID, friendID); err != nil {
		s.writeServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createDM(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input service.CreateDMInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	conversation, err := s.socialService.CreateDM(r.Context(), userID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, conversation)
}

func (s *Server) listDMs(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	conversations, err := s.socialService.ListDMs(r.Context(), userID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, conversations)
}

func (s *Server) listDMMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	conversationID, ok := pathUUID(w, r, "conversation_id")
	if !ok {
		return
	}
	limit, ok := queryLimit(w, r)
	if !ok {
		return
	}

	messages, err := s.socialService.ListDMMessages(r.Context(), userID, conversationID, limit)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, messages)
}

func (s *Server) createDMMessage(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	conversationID, ok := pathUUID(w, r, "conversation_id")
	if !ok {
		return
	}

	var input service.CreateDMMessageInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	result, err := s.socialService.CreateDMMessage(r.Context(), userID, conversationID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	if s.gatewayHub != nil {
		s.gatewayHub.BroadcastUsers(result.MemberIDs, gateway.DMMessageCreatedEvent(result.Message))
	}

	writeJSON(w, http.StatusCreated, result.Message)
}
