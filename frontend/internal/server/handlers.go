package server

import (
	"net/http"
	"log"

	"github.com/dan-nicholls/danlovesto.run/frontend/internal/ui/pages"
)


func (s *Server) handleHome() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		info, err := s.api.Info(ctx)
		if err != nil {
			log.Println("error fetching info: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)		
		}

		summary, err := s.api.Summary(ctx)
		if err != nil {
			log.Println("error fetching summary: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)		
		}
		
		dl, err := s.api.DailyLogs(ctx)
		if err != nil {
			log.Println("error fetching daily logs: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)		
		}
		pages.Home(info, summary, dl).Render(ctx, w)
	})
}
