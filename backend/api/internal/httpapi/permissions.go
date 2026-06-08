package httpapi

import (
	"net/http"

	"netcord/backend/api/internal/service"
)

func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}

	var input service.CreateRoleInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	role, err := s.permissionService.CreateRole(r.Context(), userID, serverID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, role)
}

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}

	roles, err := s.permissionService.ListRoles(r.Context(), userID, serverID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, roles)
}

func (s *Server) updateRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	roleID, ok := pathUUID(w, r, "role_id")
	if !ok {
		return
	}

	var input service.UpdateRoleInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	role, err := s.permissionService.UpdateRole(r.Context(), userID, roleID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, role)
}

func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	roleID, ok := pathUUID(w, r, "role_id")
	if !ok {
		return
	}

	if err := s.permissionService.DeleteRole(r.Context(), userID, roleID); err != nil {
		s.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) assignRole(w http.ResponseWriter, r *http.Request) {
	actorID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}
	targetUserID, ok := pathUUID(w, r, "user_id")
	if !ok {
		return
	}
	roleID, ok := pathUUID(w, r, "role_id")
	if !ok {
		return
	}

	if err := s.permissionService.AssignRole(r.Context(), actorID, serverID, targetUserID, roleID); err != nil {
		s.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeRole(w http.ResponseWriter, r *http.Request) {
	actorID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}
	targetUserID, ok := pathUUID(w, r, "user_id")
	if !ok {
		return
	}
	roleID, ok := pathUUID(w, r, "role_id")
	if !ok {
		return
	}

	if err := s.permissionService.RemoveRole(r.Context(), actorID, serverID, targetUserID, roleID); err != nil {
		s.writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	serverID, ok := pathUUID(w, r, "server_id")
	if !ok {
		return
	}

	var input service.CreateInviteInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	invite, err := s.permissionService.CreateInvite(r.Context(), userID, serverID, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, invite)
}

func (s *Server) getInvite(w http.ResponseWriter, r *http.Request) {
	invite, err := s.permissionService.GetInvite(r.Context(), r.PathValue("code"))
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, invite)
}

func (s *Server) joinInvite(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	invite, err := s.permissionService.JoinInvite(r.Context(), userID, r.PathValue("code"))
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, invite)
}
