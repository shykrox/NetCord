package httpapi

import (
	"net/http"

	"netcord/backend/api/internal/gateway"
	"netcord/backend/api/internal/service"
)

func (s *Server) joinVoice(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input service.JoinVoiceInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	session, err := s.voiceService.Join(r.Context(), userID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	if s.gatewayHub != nil {
		s.gatewayHub.Broadcast(session.ServerID, gateway.VoiceJoinedEvent(session.ServerID, session.ChannelID, userID, session.Room))
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) leaveVoice(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input service.JoinVoiceInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	session, err := s.voiceService.Leave(r.Context(), userID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	if s.gatewayHub != nil {
		s.gatewayHub.Broadcast(session.ServerID, gateway.VoiceLeftEvent(session.ServerID, session.ChannelID, userID, session.Room))
	}
	writeJSON(w, http.StatusOK, session)
}
