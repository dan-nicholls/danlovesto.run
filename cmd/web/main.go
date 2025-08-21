package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	
	"github.com/dan-nicholls/danlovesto.run/internal/web/server"
	"github.com/dan-nicholls/danlovesto.run/internal/web/apiclient"
)

func main() {
	fmt.Println("Hello World!")
	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost:3000/api/v1"
	}
	port := os.Getenv("UI_PORT")
	if port == "" {
		port = "3001"
	}

	httpClient := api.NewClient(apiUrl)
	ss := api.NewStatsService(httpClient)
	srv := server.New(ss)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s (API=%s)\n", port, apiUrl)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
