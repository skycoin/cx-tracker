package api

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/skycoin/cx-tracker/pkg/store"
)

// NewHTTPRouter creates a new HTTP router.
func NewHTTPRouter(ss store.SpecStore, ps store.PeersStore) http.Handler {
	log := logging.MustGetLogger("api")

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(SetLoggerMiddleware(log))

	r.HandleFunc("/api/specs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getAllSpecs(ss)(w, r)
			return

		case http.MethodPost:
			postSpec(ss)(w, r)
			return

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	r.HandleFunc("/api/specs/*", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getSpecOfGenesisHash(ss)(w, r)
			return

		case http.MethodDelete:
			deleteSpec(ss)(w, r)
			return

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	r.HandleFunc("/api/peers", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPeersOfChain(ps)(w, r)
			return

		case http.MethodPost:
			postPeers(ps)(w, r)
			return

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	r.HandleFunc("/api/peers/*", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPeer(ps)(w, r)
			return

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	r.HandleFunc("/peerlists/*", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getPeerList(ps)(w, r)
			return

		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	})

	return r
}
