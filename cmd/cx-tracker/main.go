package main

import (
	"flag"
	"net/http"

	"github.com/SkycoinProject/cx-chains/src/util/logging"

	"github.com/skycoin/cx-tracker/pkg/api"
	"github.com/skycoin/cx-tracker/pkg/store"
)

var (
	addr = ":8080" // serve address
	dbFile = "./cx_tracker.db" // database file path
)

func init() {
	flag.StringVar(&addr, "addr", addr, "HTTP `ADDRESS` to serve on")
	flag.StringVar(&dbFile, "db", dbFile, "database `FILEPATH`")
}

func main() {
	flag.Parse()
	log := logging.MustGetLogger("main")

	db, err := store.OpenBboltDB(dbFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to open bbolt db.")
	}

	specS, err := store.NewBboltSpecStore(db)
	if err != nil {
		log.WithError(err).Fatal("Failed to init spec store.")
	}

	r := api.NewHTTPRouter(specS)
	log.WithField("addr", addr).WithField("db_file", dbFile).Info("Serving cx-tracker...")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.WithError(err).Fatal("Failed to serve HTTP.")
	}
}
