package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/compucueva/compucueva/internal/db"
)

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var params db.CreateUserParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	task, err := s.db.CreateUser(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleGetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}
	user, err := s.db.GetUserByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, ErrNotFound, http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleGetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := s.db.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, ErrNotFound, http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleGetUsersByRole(w http.ResponseWriter, r *http.Request) {
	role := r.PathValue("role")
	users, err := s.db.GetUsersByRole(r.Context(), &role)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *Server) handleGetUsersWithTasks(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.GetUsersWithTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *Server) handleGetUsersWithoutTasks(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.GetUsersWithoutTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *Server) handleGetTaskAssignees(w http.ResponseWriter, r *http.Request) {
	task, err := strconv.ParseInt("id", 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}
	users, err := s.db.GetTaskAssignees(r.Context(), task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
