package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"strconv"

	"github.com/dan-nicholls/danlovesto.run/internal/conf"
	"github.com/dan-nicholls/danlovesto.run/internal/api/log"
	"github.com/dan-nicholls/danlovesto.run/internal/api/service"
	"github.com/dan-nicholls/danlovesto.run/internal/api/stats"
	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
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

func handleInfo(logger log.Logger, conf *conf.Config, v string, start time.Time) http.Handler {
	type response struct {
		Version string `json:"version"`	
		Uptime string `json:"uptime"`
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Info Endpoint", r.Method, r.URL.Path)
		uptime := time.Since(start).Truncate(time.Second).String()
		res := response{
			Version: v,
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
		TotalDistance int `json:"total_distance"`
		TotalHours int `json:"total_hours"`
		TotalClimbed int `json:"total_climbed"`
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Stat Summary Endpoint", r.Method, r.URL.Path)

		totalRuns, _ := rs.TotalRuns()
		totalDistanceF , _ := rs.TotalDistance()
		totalDistance := int(totalDistanceF)
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
			return
		}
	})
}

func handlePersonalBests(logger log.Logger, rs *service.RunService) http.Handler {
	type Response struct {
		Data []contracts.DetailedPersonalBest `json:"data"`
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Personal Bests Endpoint", r.Method, r.URL.Path)

		pbs, err  := rs.ListDetailedPBs()
		if err != nil {
			logger.Errorf("Failed to fetch PBs: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}


		res := Response{ Data: pbs }

		if err := encode(w, r, http.StatusOK, res); err != nil {
			logger.Errorf("Failed to encode personal best response: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
}

func handleUnderConstruction() http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		http.Error(w, "under construction", http.StatusServiceUnavailable)
	})
}

func handleStatsHeatmap(logger log.Logger, rs *service.RunService ) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// Parse and validate input parameters
		q := r.URL.Query()

		fromYear, err := parseInt(q, "from_year", 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		toYear, err := parseInt(q, "to_year", 0)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		levels, err := parseInt(q, "levels", 5)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		params := contracts.HeatMapParams{
			FromYear: fromYear,
			ToYear: toYear,
			Unit: "km",
			Scale: "linear",
			Levels: levels,
		}

		// Fetch Activities
		acts, err := rs.ListActivities()
		if err != nil {
			logger.Errorf("Failed to get activities for heatmap: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		heatmap, err := stats.CreateHeatMap(acts, params)
		if err != nil {
			logger.Errorf("Failed to create heatmap: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		err = encode(w, r, http.StatusOK, heatmap)
		if err != nil {
			logger.Errorf("Failed to encode heatmap response: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})
}

func parseInt(q url.Values, key string, def int) (int, error) {
	s := q.Get(key)
	if s == "" {
		return def, nil
	}
	v, e := strconv.Atoi(s)
	if e != nil {
		return 0, fmt.Errorf("%s must be in integer", key)
	}
	return v, nil
}
