package strava

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type StravaConfig struct {
	ClientID          string
	ClientSecret      string
	RedirectURL       string
	SyncInterval      time.Duration
	RateLimitInterval time.Duration
	OAuthMethod       string
}

func (conf *StravaConfig) setValues() error {
	conf.ClientID = strings.TrimSpace(os.Getenv("STRAVA_CLIENT_ID"))
	conf.ClientSecret = strings.TrimSpace(os.Getenv("STRAVA_CLIENT_SECRET"))
	conf.RedirectURL = strings.TrimSpace(os.Getenv("STRAVA_REDIRECT_URL"))

	if OAuthMethodStr := strings.ToLower(strings.TrimSpace(os.Getenv("STRAVA_OAUTH_METHOD"))); OAuthMethodStr != "" {
		conf.OAuthMethod = OAuthMethodStr
	}

	if syncIntStr := strings.TrimSpace(os.Getenv("STRAVA_SYNC_INTERVAL")); syncIntStr != "" {
		syncInt, err := strconv.Atoi(syncIntStr)
		if err != nil {
			return fmt.Errorf("failed to parse sync interval")
		}
		conf.SyncInterval = time.Duration(syncInt) * time.Minute
	}

	if rateLimitIntervalStr := strings.TrimSpace(os.Getenv("STRAVA_RATE_LIMIT_INTERVAL")); rateLimitIntervalStr != "" {
		rateLimInt, err := strconv.Atoi(rateLimitIntervalStr)
		if err != nil {
			return fmt.Errorf("failed to parse rate limit interval")
		}
		conf.RateLimitInterval = time.Duration(rateLimInt) * time.Minute
	}

	return nil
}

func (conf *StravaConfig) setDefaults() {
	conf.SyncInterval = 15 * time.Minute
	conf.RateLimitInterval = 1 * time.Minute
	conf.OAuthMethod = "cli"
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
		return fmt.Errorf("STRAVA_SYNC_INTERVAL must not be <= 0")
	}

	if conf.RateLimitInterval <= 0 {
		return fmt.Errorf("STRAVA_RATE_LIMIT_INTERVAL must not be <= 0")
	}

	if conf.OAuthMethod != "cli" && conf.OAuthMethod != "http" {
		return fmt.Errorf("STRAVA_OAUTH_METHOD must be either \"cli\" or \"http\"")
	}

	return nil
}

func LoadStravaConfig() (StravaConfig, error) {
	conf := StravaConfig{}

	conf.setDefaults()

	if err := conf.setValues(); err != nil {
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

func (conf *DBConfig) setDefaults() {
	conf.Path = "./data/data.db"
}

func (conf *DBConfig) setValues() {
	pathStr := strings.TrimSpace(os.Getenv("DB_PATH"))
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

	conf.setDefaults()
	conf.setValues()
	if err := conf.Validate(); err != nil {
		return conf, err
	}
	return conf, nil
}
