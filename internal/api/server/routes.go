package api

import (
	"time"
	"net/http"

	"github.com/dan-nicholls/danlovesto.run/internal/api/cfg"
	"github.com/dan-nicholls/danlovesto.run/internal/api/log"
	"github.com/dan-nicholls/danlovesto.run/internal/api/service"
)

func AddRoutes(mux *http.ServeMux, log log.Logger, cfg *cfg.Config, startTime time.Time, rs *service.RunService) {
	mux.Handle("/api/v1/health", handleHealthCheck(log)) 
	mux.Handle("/api/v1/info", handleInfo(log, cfg, startTime)) 
	mux.Handle("/api/v1/runs", handleUnderConstruction()) 
	mux.Handle("/api/v1/runs/latest", handleUnderConstruction()) 
	mux.Handle("/api/v1/runs/personal-bests", handlePersonalBests(log, rs)) 
	mux.Handle("/api/v1/runs/longest", handleUnderConstruction()) 
	mux.Handle("/api/v1/stats/summary", handleStatSummary(log, rs)) 
	mux.Handle("/api/v1/stats/heatmap", handleStatsHeatmap(log, rs))
}
