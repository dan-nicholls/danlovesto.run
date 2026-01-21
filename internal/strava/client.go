package strava

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
	"golang.org/x/oauth2"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	baseAPI       = "https://www.strava.com/api/v3"
	oauthTokenURL = "https://www.strava.com/oauth/token"
)

type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	config     StravaConfig
	http       HTTPDoer
	TokenStore TokenStore
}

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func NewClient(cfg StravaConfig, tokenStore TokenStore) Client {
	return Client{
		config: cfg,
		http: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
			},
		},
		TokenStore: tokenStore,
	}
}

func NewClientWithHTTP(cfg StravaConfig, tokenStore TokenStore, httpDoer HTTPDoer) Client {
	return Client{
		config:     cfg,
		http:       httpDoer,
		TokenStore: tokenStore,
	}
}

func (c *Client) FetchAllActivities(ctx context.Context, after, before int64, maxPages int, verbose bool) ([]contracts.StravaActivity, error) {
	perPage := 200
	page := 1
	var all []contracts.StravaActivity

	// Ensure Valid Token
	if err := c.ensureValidToken(ctx); err != nil {
		return all, fmt.Errorf("unable to ensure a valid token: %w", err)
	}
	tokens, err := c.TokenStore.Load("strava")
	if err != nil {
		return all, fmt.Errorf("unable to load token from store: %w", err)
	}

	for {
		// Check broken ctx
		select {
		case <-ctx.Done():
			return all, ctx.Err()
		default:
		}

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
			fmt.Printf("429 Rate limited: sleeping %.2fs\n", c.config.RateLimitInterval.Seconds())

			select {
			case <-time.After(c.config.RateLimitInterval):
				continue
			case <-ctx.Done():
				return all, ctx.Err()
			}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		var items []contracts.StravaActivity
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decoding activities: %w", err)
		}
		resp.Body.Close()
		// body, err := io.ReadAll(resp.Body)
		// resp.Body.Close()
		// if err != nil {
		// 	return nil, fmt.Errorf("unable to parse body: %w", err)
		// }
		//
		// c.RecordJSON("activities", fmt.Sprintf("page_%d", page), body)
		//
		// var items []contracts.StravaActivity
		// if err := json.Unmarshal(body, &items); err != nil {
		// 	return nil, err
		// }

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

func (c *Client) GetActivityDetails(ctx context.Context, id int64, verbose bool) (contracts.StravaDetailedActivity, error) {
	var finalRes contracts.StravaDetailedActivity

	if err := ctx.Err(); err != nil {
		return finalRes, err
	}

	if id <= 0 {
		return finalRes, fmt.Errorf("invalid id: %d", id)
	}

	// Ensure there is a valid auth token
	if err := c.ensureValidToken(ctx); err != nil {
		return finalRes, fmt.Errorf("failed to ensure valid token: %w", err)
	}
	tokens, err := c.TokenStore.Load("strava")
	if err != nil {
		return finalRes, fmt.Errorf("unable to load token from store: %w", err)
	}

	// Make request a loop in the case of
	for {
		select {
		case <-ctx.Done():
			return finalRes, ctx.Err()
		default:
		}

		// Make requests
		idStr := strconv.FormatInt(id, 10)
		u, _ := url.Parse(baseAPI + "/activities/" + idStr)
		req, _ := http.NewRequest("GET", u.String(), nil)
		req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)

		resp, err := c.http.Do(req)
		if err != nil {
			return finalRes, fmt.Errorf("unable to fetch detailed activity: %w", err)
		}

		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			if verbose {
				fmt.Printf("401 Unauthorized: attempting to refresh token\n")
			}
			err := c.refreshAccessToken(ctx)
			if err != nil {
				return finalRes, fmt.Errorf("refresh failed: %w", err)
			}

			tokens, err = c.TokenStore.Load("strava")
			if err != nil {
				return finalRes, fmt.Errorf("failed to fetch token from store after refresh: %w", err)
			}

			req, _ := http.NewRequest("GET", u.String(), nil)
			req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
			resp, err = c.http.Do(req)
			if err != nil {
				return finalRes, err
			}
		}

		// basic rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			fmt.Printf("429 Rate limited: sleeping %.2fs\n", c.config.RateLimitInterval.Seconds())

			select {
			case <-time.After(c.config.RateLimitInterval):
				continue
			case <-ctx.Done():
				return finalRes, ctx.Err()
			}
		}

		// Unexpected Statuses
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return finalRes, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
		}

		// Start parsing
		body, err := io.ReadAll(resp.Body)

		resp.Body.Close()
		if err != nil {
			return finalRes, fmt.Errorf("failed to parse body: %w", err)
		}

		c.RecordJSON("details", fmt.Sprintf("_%d", id), body)
		if err := json.Unmarshal(body, &finalRes); err != nil {
			return finalRes, fmt.Errorf("failed to marshal detailed activity from body: %w", err)
		}
		return finalRes, nil
	}
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

	state := fmt.Sprintf("st-%d", time.Now().UnixNano())
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)

	fmt.Println("Open the following URL in your browser to authorize:")
	fmt.Println(authURL)
	fmt.Println("")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Paste the full redirect URL here: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Failed to read input: %v\n", err)
	}

	code, err := extractCode(strings.TrimSpace(input))
	if err != nil {
		return fmt.Errorf("Unable to parse code:")
	}

	token, err := oauthCfg.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	if err := c.TokenStore.Save("strava", Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry.Unix(),
	}); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("Strava authentication complete")
	return nil
}

func extractCode(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("code is required")
	}

	if !strings.Contains(input, "code=") {
		return "", fmt.Errorf("input does not ")
	}
	parts := strings.Split(input, "code=")
	if len(parts) < 2 {
		return "", fmt.Errorf("redirect url is invalid")
	}
	tail := parts[1]
	code := strings.Split(tail, "&")[0]
	if code == "" {
		return "", fmt.Errorf("code is required")
	}

	return code, nil
}

func isExpired(exp int64) bool {
	t := time.Unix(exp, 0)
	return time.Now().After(t)
}

func (c *Client) RecordJSON(kind string, key string, body []byte) {
	if os.Getenv("STRAVA_RECORD") != "1" {
		return
	}

	dir := "testdata"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Printf("failed to create recording dir: %v\n", err)
		return
	}

	filename := filepath.Join(dir, fmt.Sprintf("%s_%s.json", kind, key))
	if err := os.WriteFile(filename, body, 0o755); err != nil {
		fmt.Printf("record json write failed: %v\n", err)
	} else {
		fmt.Printf("recorded %s_%x.json\n", kind, key)
	}
}
