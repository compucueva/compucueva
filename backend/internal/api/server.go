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
	mux.HandleFunc("POST /api/tasks/assign", s.handleAssignUserToTask)
	mux.HandleFunc("GET /api/tasks/{id}/assignees", s.handleGetTaskAssignees)

	mux.HandleFunc("POST /api/users", s.handleCreateUser)
	mux.HandleFunc("GET /api/users/with-tasks", s.handleGetUsersWithTasks)
	mux.HandleFunc("GET /api/users/without-tasks", s.handleGetUsersWithoutTasks)
	mux.HandleFunc("GET /api/users/username/{username}", s.handleGetUserByUsername)
	mux.HandleFunc("GET /api/users/role/{role}", s.handleGetUsersByRole)
	mux.HandleFunc("GET /api/users/{id}", s.handleGetUserByID)

	return mux
}
