package httpapi

import (
	"net/http"

	"netcord/backend/api/internal/gateway"
	"netcord/backend/api/internal/models"
	"netcord/backend/api/internal/service"
)

func (s *Server) createAskJob(w http.ResponseWriter, r *http.Request) {
	s.createAIJob(w, r, models.AICommandAsk)
}

func (s *Server) createDrawJob(w http.ResponseWriter, r *http.Request) {
	s.createAIJob(w, r, models.AICommandDraw)
}

func (s *Server) createAIJob(w http.ResponseWriter, r *http.Request, command string) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}

	var input service.CreateAIJobInput
	if err := readJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error(), nil)
		return
	}

	job, err := s.aiService.CreateJob(r.Context(), userID, command, input)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}

	if s.gatewayHub != nil {
		s.gatewayHub.Broadcast(job.ServerID, gateway.JobProgressEvent(job))
	}
	writeJSON(w, http.StatusAccepted, job)
}

func (s *Server) listAIJobs(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	limit, ok := queryLimit(w, r)
	if !ok {
		return
	}
	jobs, err := s.aiService.ListJobs(r.Context(), userID, limit)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

func (s *Server) getAIJob(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(w, r)
	if !ok {
		return
	}
	jobID, ok := pathUUID(w, r, "job_id")
	if !ok {
		return
	}
	job, err := s.aiService.GetJob(r.Context(), userID, jobID)
	if err != nil {
		s.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}
