package server

import (
	"net/http"
	"encoding/json"
	"context"
	"fmt"

	"github.com/dan-nicholls/danlovesto.run/frontend/internal/ui/pages"
)

type Info struct {
	Uptime string `json:"uptime"`
	Version string `json:"version"`
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var info Info
	_ = s.apiGet(ctx, "/info", &info)
	pages.Home(pages.Info{Version: info.Version, Uptime: info.Uptime}).Render(ctx, w)
}

func (s *Server) apiGet(ctx context.Context, path string, v any) error {
	url := fmt.Sprintf("%s%s", s.apiURL, path)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return fmt.Errorf("api %s -> %d", path, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(v)
}
