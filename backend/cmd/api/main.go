package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"time"
	stdlog "log"
	"os"
	"strconv"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/db"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/cfg"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/log"
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

func handleStatSummary(logger log.Logger, as *db.ActivityStore) http.Handler {
	type response struct {
		TotalRuns int `json:"total_runs"`
		TotalDistance float64 `json:"total_distance"`
		TotalHours int `json:"total_hours"`
		TotalClimbed int `json:"total_climbed"`
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Stat Summary Endpoint", r.Method, r.URL.Path)

		totalRuns, _ := as.GetTotalRuns()
		totalDistance , _ := as.GetTotalDistance()
		totalHours, _ := as.GetTotalHours()
		totalClimbed, _ := as.GetTotalClimbed()
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

func addRoutes(mux *http.ServeMux, log log.Logger, cfg *cfg.Config, startTime time.Time, as *db.ActivityStore) {
	mux.Handle("/api/v1/health", handleHealthCheck(log)) 
	mux.Handle("/api/v1/info", handleInfo(log, cfg, startTime)) 
	mux.Handle("/api/v1/runs", handleUnderConstruction()) 
	mux.Handle("/api/v1/runs/latest", handleUnderConstruction()) 
	mux.Handle("/api/v1/stats/summary", handleStatSummary(log, as)) 
}

func NewRouter(db db.Database, cfg cfg.Config) *http.ServeMux {
	mux := http.NewServeMux()
	return mux
}

func main() {
	logger := log.NewLogger(os.Stdout, stdlog.LstdFlags)

	fmt.Println("Starting Run Stats API")

	c := cfg.Load("config.json")
	database, err := db.New(c.DatabaseURL)
	if err != nil {
		stdlog.Fatalf("%w", err)
	}

	as := db.NewActivityStore(database)

	router := NewRouter(database, c)
	start := time.Now()
	addRoutes(router, logger, &c, start, as)

	logger.Infof("Listening on %d...", c.Port)
	http.ListenAndServe(":"+strconv.Itoa(c.Port), router)
}
