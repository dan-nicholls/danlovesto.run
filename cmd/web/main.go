package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dan-nicholls/danlovesto.run/internal/web/apiclient"
	"github.com/dan-nicholls/danlovesto.run/internal/web/server"
	"github.com/dan-nicholls/danlovesto.run/internal/conf"
)


func main() {
	fmt.Println("danlovesto.run UI Service")

	conf, err := conf.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	httpClient := api.NewClient(conf.WebApiUrl)
	ss := api.NewStatsService(httpClient)
	srv := server.New(ss, conf.WebMapToken)

	addr := fmt.Sprintf(":%d", conf.WebPort)
	log.Printf("Listening on %d (API=%s)\n", conf.WebPort, conf.WebApiUrl)
	if err := http.ListenAndServe(addr, srv.Router()); err != nil {
		log.Fatal(err)
	}
}
