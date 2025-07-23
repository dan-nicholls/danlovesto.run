package cfg

import (
	"os"
	"encoding/json"
	"log"
)

type Config struct {
	Port	int 	`json:"port"`
	DatabaseURL string	`json:"database_url"`
}

func Load(path string) Config {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer f.Close()
	
	var c Config
	decoder := json.NewDecoder(f)
	if err := decoder.Decode(&c); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	
	return c
}
