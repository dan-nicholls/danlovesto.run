package strava

import (
	"bufio"
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	StravaAuthURL  = "https://www.strava.com/oauth/authorize"
	StravaTokenURL = "https://www.strava.com/oauth/token"
)

func (c *Client) runOAuth(ctx context.Context) error {
	if c.config.OAuthMethod == "http" {
		return c.runOAuthHttp(ctx)
	}
	return c.runOAuthCli(ctx)
}

func (c *Client) runOAuthCli(ctx context.Context) error {
	oauthCfg := &oauth2.Config{
		ClientID:     c.config.ClientID,
		ClientSecret: c.config.ClientSecret,
		RedirectURL:  c.config.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  StravaAuthURL,
			TokenURL: StravaTokenURL,
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

func (c *Client) runOAuthHttp(ctx context.Context) error {
	oauthCfg := &oauth2.Config{
		ClientID:     c.config.ClientID,
		ClientSecret: c.config.ClientSecret,
		RedirectURL:  c.config.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  StravaAuthURL,
			TokenURL: StravaTokenURL,
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
