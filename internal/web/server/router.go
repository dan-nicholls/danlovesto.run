package server

import (
	"net/http"

	"github.com/dan-nicholls/danlovesto.run/internal/web/apiclient"
	"github.com/dan-nicholls/danlovesto.run/internal/web"
)

type Server struct {
	r *http.ServeMux
	api *api.StatsService
	mapToken string
}

func New(api *api.StatsService, token string) *Server {
	r := http.NewServeMux()

	srv := &Server{
		r: r,
		api: api,
		mapToken: token,
	}

	// Setup Routes
	// TODO - Setup /health
	r.Handle("/static/", http.StripPrefix("/static/",  web.StaticFileHandler()))
	r.Handle("/", srv.handleHome())

	return srv
}

func (s *Server) Router() http.Handler { return s.r }
