package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	apiURL string
	r *chi.Mux
}


func New(apiUrl string) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.RequestID, middleware.Recoverer)

	srv := &Server{
		apiURL: apiUrl,
		r: r,
	}

	// Setup Routes
	r.Get("/", srv.handleHome)

	return srv
}

func (s *Server) Router() http.Handler { return s.r }
