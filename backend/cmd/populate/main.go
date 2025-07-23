package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"encoding/json"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
)

func main() {
	fmt.Println("Strava Activity Populator")

	filePath := flag.String("file", "", "Path to required file (required)")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Please specify the run activity file with --file")
	}

	data, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatal("Error parsing file: %v", err)
	}

	fmt.Printf("✅ Read file: %s (%d bytes)\n", *filePath, len(data))
	
	// Parse JSON data
	var jsonData []model.Activity

	if err := json.Unmarshal(data, &jsonData); err != nil {
		panic(err)
	}
	fmt.Printf("%d activities parsed\n", len(jsonData))  

	// TODO - Store in DB
}
