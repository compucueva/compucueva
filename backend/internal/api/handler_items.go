package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/compucueva/compucueva/internal/db"
)

func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	var params db.CreateItemParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	item, err := s.db.CreateItem(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}

	rowsAffected, err := s.db.DeleteItem(r.Context(), id)
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

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	items, err := s.db.ListItems(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (s *Server) handleAssignItemToTask(w http.ResponseWriter, r *http.Request) {
	var params db.AssignItemToTaskParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, ErrInvalidJSON, http.StatusBadRequest)
		return
	}

	err := s.db.AssignItemToTask(r.Context(), params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUnassignItemFromTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}
	itemID, err := strconv.ParseInt(r.PathValue("item_id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}

	rowsAffected, err := s.db.UnassignItemFromTask(r.Context(), db.UnassignItemFromTaskParams{
		TaskID: taskID,
		ItemID: itemID,
	})
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

func (s *Server) handleGetTaskItems(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, ErrInvalidID, http.StatusBadRequest)
		return
	}
	items, err := s.db.GetTaskItems(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
