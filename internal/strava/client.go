package strava

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/internal/api/db"
	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
	"golang.org/x/oauth2"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	baseAPI       = "https://www.strava.com/api/v3"
	oauthTokenURL = "https://www.strava.com/oauth/token"
)

type Client struct {
	config        StravaConfig
	http          *http.Client
	TokenStore    TokenStore
	ActivityStore db.ActivityStore
}

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func NewClient(cfg StravaConfig, tokenStore TokenStore, activityStore db.ActivityStore) Client {
	return Client{
		config: cfg,
		http: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   time.Second,
				ResponseHeaderTimeout: 3 * time.Second,
			},
		},
		TokenStore:    tokenStore,
		ActivityStore: activityStore,
	}
}

func (c *Client) Sync(ctx context.Context) error {
	fmt.Println("Starting Sync...")

	if err := ctx.Err(); err != nil {
		return err
	}

	last, err := c.ActivityStore.LatestActivityStart()
	if err != nil {
		return fmt.Errorf("unable to fetch last activity start: %e", err)
	}
	fmt.Printf("Last Activity: %s (UNIX: %d)\n", last.Local().Format("2006-01-02 15:04:05"), last.Unix())

	acts, err := c.FetchAllActivities(ctx, last.Unix(), 0, 0, false)
	if err != nil {
		return fmt.Errorf("Error fetching acts: %w", err)
	}

	for i := range acts {
		act := acts[i]
		// fmt.Printf("Act[%d]: %+v\n", i, act)
		id, err := c.ActivityStore.CreateActivity(&act)
		if err != nil {
			fmt.Printf("unable to store fetched activity %d: %e\n", act.ID, err)
			continue
		}
		fmt.Printf("added activity %d to store\n", id)
	}

	fmt.Println("Sync Complete.")
	return nil
}

func isExpired(exp int64) bool {
	t := time.Unix(exp, 0)
	return time.Now().After(t)
}

func (c *Client) ensureValidToken(ctx context.Context) error {
	token, err := c.TokenStore.Load("strava")
	if err != nil {
		return fmt.Errorf("unable to load token from store: %w", err)
	}

	if token.RefreshToken == "" {
		return c.runOAuth(ctx)
	}

	if token.AccessToken == "" || isExpired(token.ExpiresAt) {
		return c.refreshAccessToken(ctx)
	}

	return nil
}

func (c *Client) FetchAllActivities(ctx context.Context, after, before int64, maxPages int, verbose bool) ([]contracts.Activity, error) {
	perPage := 200
	page := 1
	var all []contracts.Activity

	// Ensure Valid Token
	if err := c.ensureValidToken(ctx); err != nil {
		return all, fmt.Errorf("unable to ensure a valid token: %w", err)
	}
	tokens, err := c.TokenStore.Load("strava")
	if err != nil {
		return all, fmt.Errorf("unable to load token from store: %w", err)
	}

	for {
		if maxPages > 0 && page > maxPages {
			if verbose {
				fmt.Printf("hit max-pages=%d, stopping\n", maxPages)
			}
			break
		}

		u, _ := url.Parse(baseAPI + "/athlete/activities")
		q := u.Query()
		q.Set("page", fmt.Sprintf("%d", page))
		q.Set("per_page", fmt.Sprintf("%d", perPage))
		if after > 0 {
			q.Set("after", fmt.Sprintf("%d", after))
		}
		if before > 0 {
			q.Set("before", fmt.Sprintf("%d", before))
		}
		u.RawQuery = q.Encode()

		if verbose {
			fmt.Printf("GET %s\n", u.String())
		}

		req, _ := http.NewRequest("GET", u.String(), nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed fetching strava activities: %w", err)
		}

		// 401 -> refresh and retry once
		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			if verbose {
				fmt.Printf("401 Unauthorized: attempting token refresh\n")
			}
			err := c.refreshAccessToken(ctx)
			if err != nil {
				return nil, fmt.Errorf("refresh failed: %w", err)
			}

			tokens, err := c.TokenStore.Load("strava")
			if err != nil {
				return all, fmt.Errorf("unable to load token after refresh: %w", err)
			}

			req, _ = http.NewRequest("GET", u.String(), nil)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
			resp, err = c.http.Do(req)
			if err != nil {
				return nil, err
			}
		}

		// Basic rate-limit backoff. Limited to 100req / 15min
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			if verbose {
				fmt.Printf("429 Rate limited: sleeping 60s\n")
			}
			time.Sleep(60 * time.Second)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		var items []contracts.Activity
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if len(items) == 0 {
			if verbose {
				fmt.Printf("page %d returned 0 items — done\n", page)
			}
			break
		}

		all = append(all, items...)
		if verbose {
			fmt.Printf("page %d: got %d (total %d)\n", page, len(items), len(all))
		}
		page++
	}

	return all, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (c *Client) refreshAccessToken(ctx context.Context) error {
	tokens, err := c.TokenStore.Load("strava")
	if err != nil {
		return fmt.Errorf("unable to fetch from token store: %w", err)
	}
	if tokens.RefreshToken == "" {
		return errors.New("REFRESH_TOKEN missing")
	}

	form := url.Values{}
	form.Set("client_id", c.config.ClientID)
	form.Set("client_secret", c.config.ClientSecret)
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", tokens.RefreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, oauthTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("refresh status %d: %s", resp.StatusCode, string(b))
	}

	var tok TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return err
	}
	if tok.AccessToken == "" || tok.RefreshToken == "" {
		return fmt.Errorf("invalid token response from Strava")
	}

	newToken := Token{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		ExpiresAt:    tok.ExpiresAt,
	}

	if err := c.TokenStore.Save("strava", newToken); err != nil {
		return fmt.Errorf("failed to store token to store: %w", err)
	}

	fmt.Println("🔐 refreshed access token and updated")
	return nil
}

func (c *Client) runOAuth(ctx context.Context) error {
	oauthCfg := &oauth2.Config{
		ClientID:     c.config.ClientID,
		ClientSecret: c.config.ClientSecret,
		RedirectURL:  c.config.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.strava.com/oauth/authorize",
			TokenURL: "https://www.strava.com/oauth/token",
		},
		Scopes: []string{"read,activity:read_all"},
	}

	mux := http.NewServeMux()
	done := make(chan error, 1)

	// TODO - update to crypto/rand for better randomness
	state := fmt.Sprintf("st-%d", time.Now().UnixNano())

	mux.HandleFunc("/auth/strava", func(w http.ResponseWriter, r *http.Request) {
		url := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusFound)
	})

	mux.HandleFunc("/oauth/strava/callback", func(w http.ResponseWriter, r *http.Request) {
		// validate state
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			done <- fmt.Errorf("invalid state in callback")
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			done <- fmt.Errorf("missing code in callback")
			return
		}

		token, err := oauthCfg.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			done <- fmt.Errorf("token exchange failed: %w", err)
			return
		}

		// Save tokens into the token store
		env := Token{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry.Unix(),
		}

		if err := c.TokenStore.Save("strava", env); err != nil {
			http.Error(w, "failed to save token", http.StatusInternalServerError)
			done <- fmt.Errorf("failed to save token: %w", err)
			return
		}

		fmt.Fprintln(w, "Strava authorization complete. You can close this window.")
		done <- nil
	})

	srv := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			done <- fmt.Errorf("http server error: %w", err)
		}
	}()

	fmt.Println("Visit http://localhost:8080/auth/strava in browser to authorize Strava.")
	fmt.Println("Waiting for authorization...")

	var err error
	select {
	case err = <-done:
	case <-ctx.Done():
		err = ctx.Err()
	}

	fmt.Println("Authorization compelete.")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)

	return err
}
