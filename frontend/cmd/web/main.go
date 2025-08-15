package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	
	"github.com/dan-nicholls/danlovesto.run/frontend/internal/server"
)

func main() {
	fmt.Println("Hello World!")
	api := os.Getenv("API_URL")
	if api == "" {
		api = "http://localhost:3000/api/v1"
	}
	port := os.Getenv("UI_PORT")
	if port == "" {
		port = "3001"
	}
	srv := server.New(api)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s (API=%s)\n", port, api)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
