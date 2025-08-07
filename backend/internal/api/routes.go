package api

import (
	"time"
	"net/http"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/cfg"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/log"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/service"
)

func AddRoutes(mux *http.ServeMux, log log.Logger, cfg *cfg.Config, startTime time.Time, rs *service.RunService) {
	mux.Handle("/api/v1/health", handleHealthCheck(log)) 
	mux.Handle("/api/v1/info", handleInfo(log, cfg, startTime)) 
	mux.Handle("/api/v1/runs", handleUnderConstruction()) 
	mux.Handle("/api/v1/runs/latest", handleUnderConstruction()) 
	mux.Handle("/api/v1/runs/personal-bests", handlePersonalBests(log, rs)) 
	mux.Handle("/api/v1/runs/longest", handleUnderConstruction()) 
	mux.Handle("/api/v1/stats/summary", handleStatSummary(log, rs)) 
}
