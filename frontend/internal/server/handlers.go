package server

import (
	"net/http"

	"github.com/dan-nicholls/danlovesto.run/frontend/internal/ui/pages"
)

type Info struct {
	Uptime string
	Version string
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	// TODO - Parse info properly
	ctx := r.Context()
	info := Info{ Uptime: "1hr30min", Version: "v1.0.0" }
	pages.Home(pages.Info{Version: info.Version, Uptime: info.Uptime}).Render(ctx, w)
}
