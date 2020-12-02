package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	"github.com/SkycoinProject/cx-chains/src/util/logging"

	"github.com/skycoin/cx-tracker/pkg/api"
	"github.com/skycoin/cx-tracker/pkg/store"
)

// memory peers store constants
const (
	memTimeout = time.Minute
	memSize    = 100
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

	peersS := store.NewMemoryPeersStore(memTimeout, memSize)
	go func() {
		t := time.NewTicker(memTimeout/2)
		defer t.Stop()

		for range t.C {
			peersS.GarbageCollect(context.Background())
		}
	}()

	r := api.NewHTTPRouter(specS, peersS)
	log.WithField("addr", addr).WithField("db_file", dbFile).Info("Serving cx-tracker...")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.WithError(err).Fatal("Failed to serve HTTP.")
	}
}
