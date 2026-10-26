package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/compucueva/compucueva/internal/db"
)

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var params db.CreateTaskParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	task, err := s.db.CreateTask(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleUpdateTaskStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}
	var params db.UpdateTaskStatusParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}
	params.ID = id

	task, err := s.db.UpdateTaskStatus(r.Context(), params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, ErrNotFound, http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}
	var params db.UpdateTaskParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}
	params.ID = id

	task, err := s.db.UpdateTask(r.Context(), params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, ErrNotFound, http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}

	rowsAffected, err := s.db.DeleteTask(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, ErrNotFound, http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := s.db.ListTasks(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (s *Server) handleGetTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}
	task, err := s.db.GetTask(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, ErrNotFound, http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (s *Server) handleAssignUserToTask(w http.ResponseWriter, r *http.Request) {
	var params db.AssignUserToTaskParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	err := s.db.AssignUserToTask(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}
