package strava

import (
	"fmt"
	"os"
)

type StravaConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func (conf *StravaConfig) SetValues() {
	conf.ClientID = os.Getenv("STRAVA_CLIENT_ID")
	conf.ClientSecret = os.Getenv("STRAVA_CLIENT_SECRET")
	conf.RedirectURL = os.Getenv("STRAVA_REDIRECT_URL")
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

	return nil
}

func LoadStravaConfig() (StravaConfig, error) {
	conf := StravaConfig{}

	conf.SetValues()
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
