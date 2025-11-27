package main

import (
	"context"
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/internal/api/db"
	"github.com/dan-nicholls/danlovesto.run/internal/strava"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	baseAPI       = "https://www.strava.com/api/v3"
	oauthTokenURL = "https://www.strava.com/oauth/token"
)

type StravaConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	ListenPort   int
}

type DBConfig struct {
	Path string
}

type Token struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func run(ctx context.Context, client strava.Client) error {
	// 3. Loop every 15 mins
	// TODO - make this into clientCfg
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// Initial Sync
	if err := client.Sync(); err != nil {
		return fmt.Errorf("failed to complete initial sync: %w", err)
	}

	// Sync Loop
	for {
		select {
		case <-ctx.Done():
			fmt.Println("closing stuff here")
			return nil
		case <-ticker.C:
			fmt.Println("do work")
			if err := client.Sync(); err != nil {
				return fmt.Errorf("failed to sync: %w", err)
			}
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Unable to load from .env: %v", err)
		return
	}

	// 1. Load Config
	dbCfg, err := strava.LoadDBConfig()
	if err != nil {
		fmt.Printf("Unable to parse DB config: %v", err)
		return
	}

	cfg, err := strava.LoadStravaConfig()
	if err != nil {
		fmt.Printf("Unable to load config: %v", err)
		return
	}

	// 1. Connect to DB
	conn, err := strava.NewDB(dbCfg)
	if err != nil {
		fmt.Printf("Unable to start DB: %v", err)
		return
	}

	// 2. Get Valid Auth Token from DB
	ts := strava.SQLTokenStore{
		DB: conn,
	}
	ts.EnsureSchemas()

	as := db.NewActivityStore(&db.Store{Conn: conn})

	client := strava.NewClient(cfg, &ts, *as)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, client); err != nil {
		fmt.Printf("Error running app: %v", err)
		os.Exit(1)
	}
}
