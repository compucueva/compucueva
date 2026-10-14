package api

import (
	"net/http"

	"github.com/compucueva/compucueva/internal/db"
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
	return mux
}
