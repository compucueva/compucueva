package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleGetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.ListTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(tasks)
}
