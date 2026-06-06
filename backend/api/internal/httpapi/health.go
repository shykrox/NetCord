package httpapi

import "net/http"

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Service: "netcord-api",
		Status:  "ok",
	})
}
