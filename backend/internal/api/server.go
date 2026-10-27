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

	// Tasks
	mux.HandleFunc("GET /api/tasks", s.handleListTasks)
	mux.HandleFunc("POST /api/tasks", s.handleCreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", s.handleGetTaskByID)
	mux.HandleFunc("PUT /api/tasks/{id}", s.handleUpdateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}/status", s.handleUpdateTaskStatus)
	mux.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
	mux.HandleFunc("POST /api/tasks/assign", s.handleAssignUserToTask)
	mux.HandleFunc("DELETE /api/tasks/{id}/assignees/{user_id}", s.handleUnassignUserFromTask)
	mux.HandleFunc("GET /api/tasks/{id}/assignees", s.handleGetTaskAssignees)
	mux.HandleFunc("GET /api/tasks/{id}/items", s.handleGetTaskItems)
	mux.HandleFunc("DELETE /api/tasks/{id}/items/{item_id}", s.handleUnassignItemFromTask)

	// Users & Auth
	mux.HandleFunc("GET /api/me", s.handleGetMe)
	mux.HandleFunc("GET /api/users", s.handleListUsers)
	mux.HandleFunc("POST /api/users", s.handleCreateUser)
	mux.HandleFunc("GET /api/users/with-tasks", s.handleGetUsersWithTasks)
	mux.HandleFunc("GET /api/users/without-tasks", s.handleGetUsersWithoutTasks)
	mux.HandleFunc("GET /api/users/username/{username}", s.handleGetUserByUsername)
	mux.HandleFunc("GET /api/users/role/{role}", s.handleGetUsersByRole)
	mux.HandleFunc("GET /api/users/{id}", s.handleGetUserByID)

	// Items
	mux.HandleFunc("GET /api/items", s.handleListItems)
	mux.HandleFunc("POST /api/items", s.handleCreateItem)
	mux.HandleFunc("DELETE /api/items/{id}", s.handleDeleteItem)
	mux.HandleFunc("POST /api/items/assign", s.handleAssignItemToTask)

	return mux
}
