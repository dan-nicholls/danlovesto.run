package main

import (
	//	"flag"
	"fmt"
	"log"

	//	"os"
	//	"encoding/json"
	"database/sql"

	"github.com/dan-nicholls/danlovesto.run/backend/internal/cfg"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/db"
	"github.com/dan-nicholls/danlovesto.run/backend/internal/model"
	//"github.com/dan-nicholls/danlovesto.run/backend/internal/service"
)

func main() {
	fmt.Println("Strava Activity Populator")

//	// Parse Flags
//	filePath := flag.String("file", "", "Path to required file (required)")
//	flag.Parse()
//
//	if *filePath == "" {
//		log.Fatal("Please specify the run activity file with --file")
//	}
//
//	// Read Input File
//	data, err := os.ReadFile(*filePath)
//	if err != nil {
//		log.Fatalf("Error parsing file: %v", err)
//	}
//
//	fmt.Printf("✅ Read file: %s (%d bytes)\n", *filePath, len(data))
//	
//	// Parse JSON data
//	var activities []model.Activity
//
//	if err := json.Unmarshal(data, &activities); err != nil {
//		panic(err)
//	}
//	fmt.Printf("%d activities parsed\n", len(activities))

	// Load Config & DB
	c := cfg.Load("config.json")
	d, err := db.NewSqlStore(c.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
//	activityStore := db.NewActivityStore(d)

	// Store Activities
//	for _, a := range activities {
//		fmt.Printf("Saving %d\n", a.ID)
//		err := activityStore.SaveActivity(&a)
//		if err != nil {
//			fmt.Printf("Error: %v\n", err)
//		}
//	}
	
	//as := db.NewActivityStore(d)
	ps := db.NewPBStore(d)
	//rs := service.NewRunService(as, ps)

	// Set PBs
	pbs := []model.PersonalBest{
		model.PersonalBest{
			Distance: "1km",
			ActivityID: sql.NullInt64{
				Int64: 0,
				Valid: false,
			},	
		},
		model.PersonalBest{
			Distance: "5km",
			ActivityID: sql.NullInt64{
				Int64: 0,
				Valid: false,
			},	
		},
		model.PersonalBest{
			Distance: "10km",
			ActivityID: sql.NullInt64{
				Int64: 0,
				Valid: false,
			},	
		},
		model.PersonalBest{
			Distance: "Half-Marathon",
			ActivityID: sql.NullInt64{
				Int64: 12640876152,
				Valid: true,
			},
		},
	}

	for _, pb := range pbs {
		fmt.Printf("Storing PB %s\n", pb.Distance) 
		if err = ps.SetPB(pb.Distance, pb.ActivityID.Int64); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	storedPBs, err := ps.GetAllPBs()
	if err != nil {
		log.Fatalf("Error fetching PBs: %w", err)
	}
	for _, pb := range storedPBs {
		fmt.Printf("%v\n", pb)
	}
}
