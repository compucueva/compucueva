package api

import (
	"net/http"

	"github.com/compucueva/compucueva/internal/db"
)

var (
	ErrInvalidID   = "ID invalido"
	ErrInvalidJSON = "JSON invalido"
	ErrNotFound    = "objeto no encontrado"
)

type Server struct {
	db db.Querier
}

func NewServer(db db.Querier) *Server {
	return &Server{
		db: db,
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTaskByID)
	mux.HandleFunc("PUT /api/tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}/status", s.handleUpdateTaskStatus)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)

	return mux
}
