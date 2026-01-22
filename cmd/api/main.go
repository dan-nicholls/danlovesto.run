package main

import (
	"fmt"
	stdlog "log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dan-nicholls/danlovesto.run/internal/api/db"
	"github.com/dan-nicholls/danlovesto.run/internal/api/log"
	"github.com/dan-nicholls/danlovesto.run/internal/api/server"
	"github.com/dan-nicholls/danlovesto.run/internal/api/service"
	"github.com/dan-nicholls/danlovesto.run/internal/buildinfo"
	"github.com/dan-nicholls/danlovesto.run/internal/conf"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Unable to load .env: %v\n", err)
	}
	logger := log.NewLogger(os.Stdout, stdlog.LstdFlags)

	logger.Infof("Starting Run Stats API\n")

	c, err := conf.LoadConfig()
	if err != nil {
		stdlog.Fatalf("%v", err)
	}

	database, err := db.NewSqlStore(c.ApiDatabaseUrl)
	if err != nil {
		stdlog.Fatalf("%v", err)
	}

	as := db.NewActivityStore(database)
	ps := db.NewPBStore(database)
	rs := service.NewRunService(as, ps)

	router := http.NewServeMux()
	start := time.Now()
	v := buildinfo.String()
	api.AddRoutes(router, logger, c, v, start, rs)

	logger.Infof("Listening on %d...", c.ApiPort)
	http.ListenAndServe(":"+strconv.Itoa(c.ApiPort), router)
}
