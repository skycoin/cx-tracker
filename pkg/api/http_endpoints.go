package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/sirupsen/logrus"

	"github.com/skycoin/cx-tracker/pkg/store"
)

const (
	patternPK     = "/api/specs/pk:*"
	patternTicker = "/api/specs/ticker:*"
)

// getSpecs returns all chain specs
// URI: /api/specs
// Method: GET
func getSpecs(ss store.SpecStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		specs, err := ss.ChainSpecAll(r.Context())
		if err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		for i, s := range specs {
			if err := s.Verify(); err != nil {
				httpWriteError(log, w, http.StatusInternalServerError,
					fmt.Errorf("failed to verify spec at index %d: %w", i, err))
				return
			}
		}

		httpWriteJson(log, w, r, http.StatusOK, specs)
	}
}

// getSpec returns spec of given pk or ticker
// URI: /api/specs/[pk:<pk>|ticker:<ticker>]
// Method: GET
func getSpec(ss store.SpecStore) http.HandlerFunc {
	specOfPK := func(log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		b := path.Base(r.URL.EscapedPath())
		pkStr := strings.TrimPrefix(b, "pk:")

		pk, err := cipher.PubKeyFromHex(pkStr)
		if err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("provided invalid pk '%s': %w", pkStr, err))
			return
		}

		spec, err := ss.ChainSpecByChainPK(r.Context(), pk)
		if err != nil {
			if errors.Is(err, store.ErrBboltObjectNotExist) {
				httpWriteError(log, w, http.StatusNotFound, err)
				return
			}

			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		if err := spec.Verify(); err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		httpWriteJson(log, w, r, http.StatusOK, spec)
	}

	specOfTicker := func(log logrus.FieldLogger, w http.ResponseWriter, r *http.Request) {
		b := path.Base(r.URL.EscapedPath())
		ticker := strings.TrimPrefix(b, "ticker:")
		ticker = strings.TrimSpace(strings.ToUpper(ticker))

		spec, err := ss.ChainSpecByCoinTicker(r.Context(), ticker)
		if err != nil {
			if errors.Is(err, store.ErrBboltObjectNotExist) {
				httpWriteError(log, w, http.StatusNotFound, err)
				return
			}

			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		if err := spec.Verify(); err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		httpWriteJson(log, w, r, http.StatusOK, spec)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		escPath := r.URL.EscapedPath()

		isPK, err := path.Match(patternPK, escPath)
		if err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}
		if isPK {
			specOfPK(log, w, r)
			return
		}

		isTicker, err := path.Match(patternTicker, escPath)
		if err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}
		if isTicker {
			specOfTicker(log, w, r)
			return
		}

		// handle invalid requests
		httpWriteError(log, w, http.StatusBadRequest,
			fmt.Errorf("request's path does not match '%s' or '%s'", patternPK, patternTicker))
	}
}

// postSpec posts a chain spec
// URI: /api/specs
// Method: POST
func postSpec(ss store.SpecStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		var spec store.SignedSpec
		if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to decode spec: %w", err))
			return
		}

		if err := spec.Verify(); err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to verify spec: %w", err))
			return
		}

		if err := ss.AddSpec(r.Context(), spec); err != nil {
			if errors.Is(err, store.ErrBboltObjectAlreadyExists); err != nil {
				httpWriteError(log, w, http.StatusConflict,
					fmt.Errorf("new spec conflicts with current directory: %w", err))
				return
			}
		}
	}
}

// deleteSpec deletes a chain spec of given pk
// URI: /api/spec/pk:<pk>
// Method: DELETE
func deleteSpec(ss store.SpecStore) http.HandlerFunc {
	deleteOfPK := func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		b := path.Base(r.URL.EscapedPath())
		pkStr := strings.TrimPrefix(b, "pk:")

		pk, err := cipher.PubKeyFromHex(pkStr)
		if err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to verify pk: %w", err))
		}

		if err := ss.DelSpec(r.Context(), pk); err != nil {
			if errors.Is(store.ErrBboltObjectNotExist, err) {
				httpWriteError(log, w, http.StatusNotFound, err)
				return
			}

			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		httpWriteJson(log, w, r, http.StatusOK, true)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		escPath := r.URL.RequestURI()

		isPK, err := path.Match(patternPK, escPath)
		if err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}
		if isPK {
			deleteOfPK(w, r)
			return
		}

		isTicker, err := path.Match(patternTicker, escPath)
		if err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}
		if isTicker {
			// TODO: actually implement this
			httpWriteError(log, w, http.StatusNotImplemented,
				errors.New("endpoint 'DELETE /api/spec/ticker:<ticker>' is not implemented"))
			return
		}

		// handle invalid requests
		httpWriteError(log, w, http.StatusBadRequest,
			fmt.Errorf("request's path does not match '%s' or '%s'", patternPK, patternTicker))
	}
}
