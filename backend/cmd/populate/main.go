package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	// TODO - Complete parsing here
}
