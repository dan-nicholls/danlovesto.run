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

func run(ctx context.Context, client strava.Client) error {
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
			fmt.Println("Run complete. Closing...")
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

	conn, err := db.NewSqlStore(dbCfg.Path)
	if err != nil {
		fmt.Printf("Unable to start DB: %v", err)
		return
	}
	defer conn.Close()

	ts := strava.SQLTokenStore{
		DB: conn,
	}
	as := db.NewActivityStore(conn)

	client := strava.NewClient(cfg, &ts, *as)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, client); err != nil {
		fmt.Printf("Error running app: %v", err)
		os.Exit(1)
	}
}
