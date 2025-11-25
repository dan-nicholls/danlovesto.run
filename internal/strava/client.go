package strava

import (
	"encoding/json"
	"errors"
	"fmt"
	//"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
	"io"
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
	config     StravaConfig
	http       *http.Client
	TokenStore TokenStore
}

type StravaConfig struct {
	ClientID     string
	ClientSecret string
}

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func NewClient(cfg StravaConfig, tokenStore TokenStore) Client {
	return Client{
		config:     cfg,
		http:       &http.Client{},
		TokenStore: tokenStore,
	}
}

// fetchAllActivities paginates /athlete/activities and returns raw Strava JSON entries.
// On 401 it refreshes once and retries; on 429 it waits briefly and retries.
func (c *Client) FetchAllActivities(after, before int64, maxPages int, verbose bool) ([]map[string]any, error) {
	perPage := 200 // Strava max per page
	page := 1
	var all []map[string]any

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
			return nil, err
		}

		// 401 -> refresh and retry once
		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			if verbose {
				fmt.Printf("401 Unauthorized: attempting token refresh\n")
			}
			err := c.refreshAccessToken()
			if err != nil {
				return nil, fmt.Errorf("refresh failed: %w", err)
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

		var items []map[string]any
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

	// TODO - Convert into []contracts.Activity
	return all, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (c *Client) refreshAccessToken() error {
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
	form.Set("refresh_token", tokens.RefreshToken) // ← correct field for refresh

	req, _ := http.NewRequest("POST", oauthTokenURL, strings.NewReader(form.Encode()))
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
	}

	if err := c.TokenStore.Save("strava", newToken); err != nil {
		return fmt.Errorf("failed to store token to store: %w", err)
	}

	fmt.Println("🔐 refreshed access token and updated")
	return nil
}
