package main

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"time"

	"context"
	"github.com/dan-nicholls/danlovesto.run/internal/strava"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"net/http"
)

var (
	stravaOut      string
	stravaAfter    int64
	stravaBefore   int64
	stravaMaxPages int
	stravaVerbose  bool
	stravaEnvPath  string
)

func init() {
	// Parse flags
	stravaCmd.Flags().StringVar(&stravaOut, "out", "strava_activities.json", "output file for activities JSON")
	stravaCmd.Flags().Int64Var(&stravaAfter, "after", 0, "epoch seconds filter for activities after this time")
	stravaCmd.Flags().Int64Var(&stravaBefore, "before", 0, "epoch seconds filter for activities before this time")
	stravaCmd.Flags().IntVar(&stravaMaxPages, "max-pages", 0, "optional page limit (0 = no limit)")
	stravaCmd.Flags().BoolVarP(&stravaVerbose, "verbose", "v", false, "verbose logging")
	stravaCmd.Flags().StringVar(&stravaEnvPath, "env", ".env", "path to .env with CLIENT_ID/CLIENT_SECRET/ACCESS_TOKEN/REFRESH_TOKEN")

	rootCmd.AddCommand(stravaCmd)
}

func bootstrapTokensFromEnv(store strava.TokenStore, token strava.Token) {
	if token.AccessToken == "" || token.RefreshToken == "" {
		fmt.Println("INFO - nothing to bootstrap")
		return
	}

	existing, err := store.Load("strava")
	if err == nil && existing.AccessToken == "" && existing.RefreshToken == "" {
		_ = store.Save("strava", token)
		if err == nil {
			fmt.Println("token bootstrap successful")
		}
	}
}

var stravaCmd = &cobra.Command{
	Use:   "strava",
	Short: "Fetch activities from Strava and write them to JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := godotenv.Load()
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		cfg, err := strava.LoadStravaConfig()
		if err != nil {
			return fmt.Errorf("Unable to parse strava config: %w", err)
		}
		dbCfg, err := strava.LoadDBConfig()
		if err != nil {
			return fmt.Errorf("Unable to parse DB config: %w", err)
		}

		// Create Token Store
		conn, err := sql.Open("sqlite", dbCfg.Path)
		if err != nil {
			return fmt.Errorf("failed to open db: %w", err)
		}
		if _, err := conn.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
			return fmt.Errorf("failed to set WAL mode: %w", err)
		}
		ts := strava.SQLTokenStore{
			DB: conn,
		}
		ts.EnsureSchemas()

		if err := ensureStravaToken(context.Background(), cfg, &ts); err != nil {
			return fmt.Errorf("failed to ensure strava token: %w", err)
		}

		// Create Strava Client & run fetchAllActivities
		client := strava.NewClient(cfg, &ts)
		fmt.Printf("Client: %v", client)
		//acts, err := client.FetchAllActivities(stravaAfter, stravaBefore, stravaMaxPages, stravaVerbose)
		// if err != nil {
		// 	return err
		// }

		//fmt.Printf("Acts[%d]: %+v\n", len(acts), acts)
		return nil
	},
}

func ensureStravaToken(ctx context.Context, cfg strava.StravaConfig, ts *strava.SQLTokenStore) error {
	token, err := ts.Load("strava")
	if err != nil {
		return err
	}

	// TODO - add check if expired at has been hit
	if token.AccessToken != "" && token.RefreshToken != "" {
		return nil
	}

	// No tokens present
	fmt.Println("No valid token in db. Starting local OAuth flow...")

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.strava.com/oauth/authorize",
			TokenURL: "https://www.strava.com/oauth/token",
		},
		Scopes: []string{"read,activity:read_all"},
	}

	return runLocalOAuthServer(ctx, oauthCfg, ts)
}

func runLocalOAuthServer(ctx context.Context, oauthCfg *oauth2.Config, ts *strava.SQLTokenStore) error {
	mux := http.NewServeMux()
	done := make(chan error, 1)

	// TODO - crypto/rand
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
		env := strava.Token{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresAt:    token.Expiry.Unix(),
		}

		if err := ts.Save("strava", env); err != nil {
			http.Error(w, "failed to save token", http.StatusInternalServerError)
			done <- fmt.Errorf("failed to save token: %w", err)
			return
		}

		fmt.Fprintln(w, "Strava authorization complete. You can close this window.")
		done <- nil
	})

	srv := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			done <- fmt.Errorf("http server error: %w", err)
		}
	}()

	fmt.Println("Visit http://localhost:8080/auth/strava in browser to authorize Strava.")
	fmt.Println("Waiting for authorization...")

	// Wait for callback/cancellation
	select {
	case err := <-done:
		//shutdown
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
		return err
	case <-ctx.Done():
		_ = srv.Shutdown(context.Background())
		return ctx.Err()
	}
}
