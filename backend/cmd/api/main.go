package main

import (
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/cfg"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/db"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/log"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/service"
)

func encode[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func handleHealthCheck(logger log.Logger) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Health Endpoint", r.Method, r.URL.Path)
		err := encode(w, r, http.StatusOK, map[string]string{
			"status": "ok",
		})
		if err != nil {
			logger.Errorf("Failed to encode health response: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
}

func handleInfo(logger log.Logger, cfg *cfg.Config, start time.Time) http.Handler {
	type response struct {
		Version string `json:"version"`	
		Uptime string `json:"uptime"`
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Info Endpoint", r.Method, r.URL.Path)
		uptime := time.Since(start).Truncate(time.Second).String()
		res := response{
			Version: cfg.Version,
			Uptime: uptime,
		}

		if err := encode(w, r, http.StatusOK, res); err != nil {
			logger.Errorf("Failed to encode info response: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
}

func handleStatSummary(logger log.Logger, rs *service.RunService) http.Handler {
	type response struct {
		TotalRuns int `json:"total_runs"`
		TotalDistance float64 `json:"total_distance"`
		TotalHours int `json:"total_hours"`
		TotalClimbed int `json:"total_climbed"`
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Stat Summary Endpoint", r.Method, r.URL.Path)

		totalRuns, _ := rs.TotalRuns()
		totalDistance , _ := rs.TotalDistance()
		totalHours, _ := rs.TotalHours()
		totalClimbed, _ := rs.TotalClimbed()
		res := response{
			TotalRuns: totalRuns,
			TotalDistance: totalDistance,
			TotalHours: totalHours,
			TotalClimbed: totalClimbed,
		}

		if err := encode(w, r, http.StatusOK, res); err != nil {
			logger.Errorf("Failed to encode stats summary response: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
}

func handleUnderConstruction() http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		http.Error(w, "under construction", http.StatusServiceUnavailable)
	})
}

func addRoutes(mux *http.ServeMux, log log.Logger, cfg *cfg.Config, startTime time.Time, rs *service.RunService) {
	mux.Handle("/api/v1/health", handleHealthCheck(log)) 
	mux.Handle("/api/v1/info", handleInfo(log, cfg, startTime)) 
	mux.Handle("/api/v1/runs", handleUnderConstruction()) 
	mux.Handle("/api/v1/runs/latest", handleUnderConstruction()) 
	mux.Handle("/api/v1/stats/summary", handleStatSummary(log, rs)) 
}

func NewRouter(db *db.Store, cfg cfg.Config) *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}

func main() {
	logger := log.NewLogger(os.Stdout, stdlog.LstdFlags)

	fmt.Println("Starting Run Stats API")

	c := cfg.Load("config.json")
	database, err := db.NewSqlStore(c.DatabaseURL)
	if err != nil {
		stdlog.Fatalf("%w", err)
	}

	as := db.NewActivityStore(database)
	ps := db.NewPBStore(database)
	rs := service.NewRunService(as, ps)

	router := NewRouter(database, c)
	start := time.Now()
	addRoutes(router, logger, &c, start, rs)

	logger.Infof("Listening on %d...", c.Port)
	http.ListenAndServe(":"+strconv.Itoa(c.Port), router)
}
