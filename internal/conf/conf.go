package conf

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Web/UI Configs
	WebApiUrl   string
	WebPort     int
	WebMapToken string

	// API Configs
	ApiPort        int
	ApiDatabaseUrl string

	// Webhook Configs
}

func (conf *Config) setDefaults() {
	// Web Defaults
	conf.WebApiUrl = "http://localhost:3000/api/v1"
	conf.WebPort = 3001
	conf.WebMapToken = ""

	// API Defaults
	conf.ApiPort = 3000
	conf.ApiDatabaseUrl = "./data/data.db"
}

func (conf *Config) Validate() error {
	if conf.WebApiUrl != "" &&
		!strings.HasPrefix(conf.WebApiUrl, "http://") &&
		!strings.HasPrefix(conf.WebApiUrl, "https://") {
		return fmt.Errorf("Invalid API URL, must be a valid URL")
	}

	if conf.WebPort <= 0 || conf.WebPort > 65535 {
		return fmt.Errorf("UI_PORT must be within 0 > UI_PORT >= 65535")
	}

	if conf.WebMapToken == "" {
		// TODO - Make print warning
		log.Println("No API_MAP_TOKEN is set. Map data will not be available")
	}

	if conf.ApiPort <= 0 || conf.WebPort > 65535 {
		return fmt.Errorf("API_PORT must be within 0 > API_PORT >= 65535")
	}

	fi, err := os.Stat(conf.ApiDatabaseUrl)
	if err != nil || fi.IsDir() {
		fmt.Println(conf.ApiDatabaseUrl)
		return fmt.Errorf("API_DATABASE_URL must point to a valid db location")
	}

	return nil
}

func (conf *Config) setValues() error {
	// TODO - Load from file
	// After Load Envs
	aStr := os.Getenv("WEB_API_URL")
	if aStr != "" {
		conf.WebApiUrl = aStr
	}
	pStr := os.Getenv("WEB_PORT")
	if pStr != "" {
		p, _ := strconv.Atoi(pStr)
		conf.WebPort = p
	}

	mStr := os.Getenv("WEB_MAP_TOKEN")
	if mStr != "" {
		conf.WebMapToken = mStr
	}

	pStr = os.Getenv("API_PORT")
	if pStr != "" {
		p, _ := strconv.Atoi(pStr)
		conf.ApiPort = p
	}

	dStr := os.Getenv("API_DB_URL")
	if dStr != "" {
		conf.ApiDatabaseUrl = dStr
	}

	return nil
}

func LoadConfig() (*Config, error) {
	conf := &Config{}

	conf.setDefaults()

	err := conf.setValues()
	if err != nil {
		return nil, err
	}
	err = conf.Validate()
	if err != nil {
		return nil, err
	}

	return conf, nil
}
