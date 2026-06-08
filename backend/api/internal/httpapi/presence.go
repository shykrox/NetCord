package httpapi

import (
	"net/http"

	"netcord/backend/api/internal/service"

	"github.com/google/uuid"
)

func (s *Server) updatePresence(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	if s.presenceService == nil {
		writeError(w, http.StatusInternalServerError, "presence_unavailable", "presence service is unavailable", nil)
		return
	}

	var input service.UpdatePresenceInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	presence, err := s.presenceService.UpdatePresence(r.Context(), userID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	if s.gatewayHub != nil && s.serverService != nil {
		servers, err := s.serverService.ListServers(r.Context(), userID)
		if err == nil {
			serverIDs := make([]uuid.UUID, 0, len(servers.Servers))
			for _, server := range servers.Servers {
				serverIDs = append(serverIDs, server.ID)
			}
			s.gatewayHub.BroadcastPresence(serverIDs, presence)
		}
	}

	writeJSON(w, http.StatusOK, presence)
}
