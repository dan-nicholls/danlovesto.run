package main

import (
	stdlog "log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dan-nicholls/danlovesto.run/internal/api/cfg"
	"github.com/dan-nicholls/danlovesto.run/internal/api/db"
	"github.com/dan-nicholls/danlovesto.run/internal/api/log"
	"github.com/dan-nicholls/danlovesto.run/internal/api/service"
	"github.com/dan-nicholls/danlovesto.run/internal/api/server"
)

func main() {
	logger := log.NewLogger(os.Stdout, stdlog.LstdFlags)

	logger.Infof("Starting Run Stats API\n")

	c := cfg.Load("config.json")
	database, err := db.NewSqlStore(c.DatabaseURL)
	if err != nil {
		stdlog.Fatalf("%v", err)
	}

	as := db.NewActivityStore(database)
	ps := db.NewPBStore(database)
	rs := service.NewRunService(as, ps)

	router := http.NewServeMux()
	start := time.Now()
	api.AddRoutes(router, logger, &c, start, rs)

	logger.Infof("Listening on %d...", c.Port)
	http.ListenAndServe(":"+strconv.Itoa(c.Port), router)
}
