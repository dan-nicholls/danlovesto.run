package strava

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type StravaConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	SyncInterval time.Duration
}

func (conf *StravaConfig) SetValues() error {
	conf.ClientID = os.Getenv("STRAVA_CLIENT_ID")
	conf.ClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
	conf.RedirectURL = os.Getenv("STRAVA_REDIRECT_URL")

	if syncIntStr := os.Getenv("STRAVA_SYNC_INTERVAL"); syncIntStr != "" {
		syncInt, err := strconv.Atoi(syncIntStr)
		if err != nil {
			return fmt.Errorf("failed to parse sync interval: %w", err)
		}
		conf.SyncInterval = time.Duration(syncInt) * time.Minute
	}

	return nil
}

func (conf *StravaConfig) SetDefaults() {
	conf.SyncInterval = 15 * time.Minute
}

func (conf *StravaConfig) Validate() error {
	if conf.ClientID == "" {
		return fmt.Errorf("STRAVA_CLIENT_ID must not be empty")
	}

	if conf.ClientSecret == "" {
		return fmt.Errorf("STRAVA_CLIENT_SECRET must not be empty")
	}

	if conf.RedirectURL == "" {
		return fmt.Errorf("STRAVA_REDIRECT_URL must not be empty")
	}

	if conf.SyncInterval <= 0 {
		return fmt.Errorf("STRAVA_SYNC_INTERVAL must be not be < 0")
	}

	return nil
}

func LoadStravaConfig() (StravaConfig, error) {
	conf := StravaConfig{}

	conf.SetDefaults()

	if err := conf.SetValues(); err != nil {
		return conf, err
	}

	if err := conf.Validate(); err != nil {
		return conf, err
	}
	return conf, nil
}

type DBConfig struct {
	Path string
}

func (conf *DBConfig) SetDefaults() {
	conf.Path = "./data/data.db"
}

func (conf *DBConfig) SetValues() {
	pathStr := os.Getenv("DB_PATH")
	if pathStr != "" {
		conf.Path = pathStr
	}
}

func (conf *DBConfig) Validate() error {
	// TODO - Add any db validation
	return nil
}

func LoadDBConfig() (DBConfig, error) {
	conf := DBConfig{}

	conf.SetDefaults()
	conf.SetValues()
	if err := conf.Validate(); err != nil {
		return conf, err
	}
	return conf, nil
}
