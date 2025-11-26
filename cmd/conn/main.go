package main

import (
	"fmt"
	"github.com/dan-nicholls/danlovesto.run/internal/strava"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
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

func NewDB() {}

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
		fmt.Println("Unable to load config: %v", err)
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

	client := strava.NewClient(cfg, &ts)

	acts, err := client.FetchAllActivities(0, 0, 0, false)
	if err != nil {
		fmt.Printf("Error fetching acts: %v", err)
		return
	}

	fmt.Printf("Acts: %+v", acts)
	// 3. Loop every 15 mins
	// 3a. Fetch Acitivites between last available activity

	// 3b. Parse into []Activity and store in DB

	// 4. Cancel ctx received -> Close
}
