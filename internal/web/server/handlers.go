package server

import (
	"net/http"
	"log"
	"context"
	"fmt"

	"github.com/dan-nicholls/danlovesto.run/internal/web/ui/pages"
)


func (s *Server) handleHome() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		info, err := s.api.Info(ctx)
		if err != nil {
			log.Println("error fetching info: %v", err)
			s.handleError(ctx, w, http.StatusInternalServerError, fmt.Errorf("fetch info: %w", err))		
			return
		}

		summary, err := s.api.Summary(ctx)
		if err != nil {
			s.handleError(ctx, w, http.StatusInternalServerError, fmt.Errorf("fetch summary: %w", err))		
			return
		}
		
		dl, err := s.api.DailyLogs(ctx)
		if err != nil {
			s.handleError(ctx, w, http.StatusInternalServerError, fmt.Errorf("fetch daily logs: %w", err))		
			return
		}
		pages.Home(info, summary, dl).Render(ctx, w)
	})
}

func (s *Server) handleError(ctx context.Context, w http.ResponseWriter, status int, err error) {
	log.Printf("error: %v\n", err)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	pages.Error().Render(ctx, w)
}
