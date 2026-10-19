package server

import (
	"encoding/json"
	"net/http"
)


func (s *Server) handleRiotAccount(w http.ResponseWriter, r *http.Request) {
	acc, err := s.riot.AccountByRiotID("europe", r.PathValue("name"), r.PathValue("tag"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(&acc)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}
