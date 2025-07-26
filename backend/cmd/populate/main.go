package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"encoding/json"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/cfg"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/db"
)

func main() {
	fmt.Println("Strava Activity Populator")

	// Parse Flags
	filePath := flag.String("file", "", "Path to required file (required)")
	flag.Parse()

	if *filePath == "" {
		log.Fatal("Please specify the run activity file with --file")
	}

	// Read Input File
	data, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatal("Error parsing file: %v", err)
	}

	fmt.Printf("✅ Read file: %s (%d bytes)\n", *filePath, len(data))
	
	// Parse JSON data
	var activities []model.Activity

	if err := json.Unmarshal(data, &activities); err != nil {
		panic(err)
	}
	fmt.Printf("%d activities parsed\n", len(activities))

	// Load Config & DB
	c := cfg.Load("config.json")
	fmt.Println(c)
	d, err := db.New(c.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	runStore := db.NewRunStore(d)

	// Store Activities
	for _, a := range activities {
		fmt.Printf("Saving %d\n", a.ID)
		err := runStore.SaveActivity(&a)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	log.Println("Fetching all Activities")
	var activityArr []*model.Activity
	activityArr, err = runStore.GetAllActivities()
	if err != nil {
		log.Fatalf("Failed to fetch activities: %v", err)
	}
	for _, a := range activityArr {
		b, _ := json.MarshalIndent(a, "", "  ")
		fmt.Println(string(b))
	}
}
