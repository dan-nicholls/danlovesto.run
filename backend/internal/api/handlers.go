package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/cfg"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/log"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
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
			return
		}
	})
}

func handlePersonalBests(logger log.Logger, rs *service.RunService) http.Handler {
	type Response struct {
		OneKm *model.Activity `json:"1km"`
		FiveKm *model.Activity `json:"5km"`
		TenKm *model.Activity `json:"10km"`
		HalfMarathon *model.Activity `json:"Half-Marathon"`
		Marathon *model.Activity `json:"Marathon"`
	}

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		logger.Infof("%s - %s - Handling Personal Bests Endpoint", r.Method, r.URL.Path)

		pbs, err  := rs.ListDetailedPBs()
		if err != nil {
			logger.Errorf("Failed to fetch PBs: %w", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		res := Response{}
		for _, pb := range pbs {
			logger.Infof("%v", pb) 
			switch pb.Distance {
			case "1km":
				logger.Infof("1km pb found: %v", pb.Activity)
				res.OneKm = pb.Activity
			case "5km":
				logger.Infof("5km pb found: %v", pb.Activity)
				res.FiveKm = pb.Activity
			case "10km":
				logger.Infof("10km pb found: %v", pb.Activity)
				res.TenKm = pb.Activity
			case "Half-Marathon":
				logger.Infof("half pb found: %v", pb.Activity)
				res.HalfMarathon = pb.Activity
			case "Mararthon":
				logger.Infof("marathon pb found: %v", pb.Activity)
				res.Marathon = pb.Activity
			default:
				logger.Errorf("Invalid PB Found: %s", pb.Distance)
			}
		}
		
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
