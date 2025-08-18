package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
	"github.com/dan-nicholls/danlovesto.run/frontend/internal/api"
)

type Server struct {
	r *chi.Mux
	api *api.StatsService
}

func New(api *api.StatsService) *Server {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.RequestID, middleware.Recoverer)

	srv := &Server{
		r: r,
		api: api,
	}

	// Setup Routes
	// TODO - Setup static assets
	// TODO - Setup /health
	r.Get("/", srv.handleHome)

	return srv
}

func (s *Server) Router() http.Handler { return s.r }
