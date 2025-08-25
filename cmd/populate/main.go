package main

import (
	"fmt"
	"log"

	"os"
	"encoding/json"
	"database/sql"

	"github.com/dan-nicholls/danlovesto.run/internal/api/db"
	"github.com/dan-nicholls/danlovesto.run/pkg/contracts"
	"github.com/dan-nicholls/danlovesto.run/internal/api/service"
	"github.com/spf13/cobra"
)

var (
    dbPath       string
    activitiesFile string
    pbFile      string
    distance    string
	duration 	string
    activityID  int64
)

func init() {
    rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./data/data.db", "path to sqlite database file")

    // db
    rootCmd.AddCommand(dbCmd)

    // activities
    activitiesCmd.Flags().StringVar(&activitiesFile, "file", "", "path to JSON file containing []model.Activity")
    activitiesCmd.MarkFlagRequired("file")
    rootCmd.AddCommand(activitiesCmd)

    // pbs
    pbCmd.Flags().StringVar(&pbFile, "file", "", "path to JSON file containing PB entries (distance and activity_id)")
    pbCmd.Flags().StringVar(&distance, "distance", "", "distance label (e.g. 5km, 1km)")
	pbCmd.Flags().StringVar(&duration, "duration", "", "duration for given section (e.g. 1:30:30)")
    pbCmd.Flags().Int64Var(&activityID, "id", 0, "activity ID to set for PB")
    rootCmd.AddCommand(pbCmd)
}

var rootCmd = &cobra.Command{
    Use:   "populate",
    Short: "Parse and populate data into the database",
}

var dbCmd = &cobra.Command{
    Use:   "db",
    Short: "Initialize database schema and seed data",
    Run: func(cmd *cobra.Command, args []string) {
        d, err := db.NewSqlStore(dbPath)
        if err != nil {
            log.Fatalf("failed to open store: %v", err)
        }
        defer d.Conn.Close()
        if err := d.EnsureSchemas(); err != nil {
            log.Fatalf("failed to ensure schemas: %v", err)
        }
        fmt.Println("Database schema initialized and seeded.")
    },
}

var activitiesCmd = &cobra.Command{
    Use:   "activities",
    Short: "Import activities from a JSON file",
    Run: func(cmd *cobra.Command, args []string) {
        data, err := os.ReadFile(activitiesFile)
        if err != nil {
            log.Fatalf("failed to read file: %v", err)
        }
        var activities []*contracts.Activity
        if err := json.Unmarshal(data, &activities); err != nil {
            log.Fatalf("failed to parse JSON: %v", err)
        }

        d, err := db.NewSqlStore(dbPath)
        if err != nil {
            log.Fatalf("failed to open store: %v", err)
        }
        defer d.Conn.Close()
        if err := d.EnsureSchemas(); err != nil {
            log.Fatalf("failed to ensure schemas: %v", err)
        }

        rs := service.NewRunService(db.NewActivityStore(d), db.NewPBStore(d))
        count := 0
        for _, act := range activities {
            if err := rs.SaveActivity(act); err != nil {
                log.Printf("warning: failed to save activity ID %d: %v", act.ID, err)
                continue
            }
            count++
        }
        fmt.Printf("Imported %d/%d activities.\n", count, len(activities))
    },
}

var pbCmd = &cobra.Command{
    Use:   "pb",
    Short: "Set personal bests from file or flags",
    Run: func(cmd *cobra.Command, args []string) {
        d, err := db.NewSqlStore(dbPath)
        if err != nil {
            log.Fatalf("failed to open store: %v", err)
        }
        defer d.Conn.Close()
		// when using file input mode
        if pbFile != "" {
            data, err := os.ReadFile(pbFile)
            if err != nil {
                log.Fatalf("failed to read PB file: %v", err)
            }
            var entries []struct {
                Distance   string `json:"distance"`
				Duration   string `json:"duration"`
                ActivityID int64  `json:"activity_id"`
            }
            if err := json.Unmarshal(data, &entries); err != nil {
                log.Fatalf("failed to parse PB JSON: %v", err)
            }
            ps := db.NewPBStore(d)
            for _, e := range entries {
				log.Printf("Attempting to store %s - %s - %v", e.Distance, e.Duration, e.ActivityID)
                nid := sql.NullInt64{Int64: e.ActivityID, Valid: true}
                if err := ps.SetPB(e.Distance, e.Duration, nid.Int64); err != nil {
                    log.Printf("warning: failed to set PB %s: %v", e.Distance, err)
                }
            }
            fmt.Printf("Processed %d PB entries.\n", len(entries))
            return
        }

        // when using manual flag mode
        if distance == "" || duration == "" {
            log.Fatal("either --file or --distance and --duration must be provided")
        }
        ps := db.NewPBStore(d)
        nid := sql.NullInt64{Int64: activityID, Valid: true}
		if err := ps.SetPB(distance, duration, nid.Int64); err != nil {
            log.Fatalf("failed to set PB: %v", err)
        }
        fmt.Printf("Set PB for %s to activity %d\n", distance, activityID)
    },
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("danlovesto.run Populator Tool")
	Execute()
}
