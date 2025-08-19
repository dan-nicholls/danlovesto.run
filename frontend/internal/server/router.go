package server

import (
	"net/http"

	"github.com/dan-nicholls/danlovesto.run/frontend/internal/api"
)

type Server struct {
	r *http.ServeMux
	api *api.StatsService
}

func New(api *api.StatsService) *Server {
	r := http.NewServeMux()

	srv := &Server{
		r: r,
		api: api,
	}

	// Setup Routes
	// TODO - Setup static assets
	// TODO - Setup /health
	r.Handle("/", srv.handleHome())

	return srv
}

func (s *Server) Router() http.Handler { return s.r }
