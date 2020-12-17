package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/skycoin/cx-chains/src/cx/cxspec"
	"github.com/skycoin/skycoin/src/cipher"

	"github.com/skycoin/cx-tracker/pkg/store"
)

// getAllSpecs returns all chain specs
// URI: /api/specs
// Method: GET
func getAllSpecs(ss store.SpecStore) http.HandlerFunc {
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

// getSpecOfGenesisHash returns spec of given genesis hash
// URI: /api/specs/<genesis-hash>
// Method: GET
func getSpecOfGenesisHash(ss store.SpecStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		hashStr := path.Base(r.URL.EscapedPath())

		hash, err := cipher.SHA256FromHex(hashStr)
		if err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to decode hash '%s': %w", hashStr, err))
			return
		}

		spec, err := ss.ChainSpec(r.Context(), hash)
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
}

// postSpec posts a chain spec
// URI: /api/specs
// Method: POST
func postSpec(ss store.SpecStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		var spec cxspec.SignedChainSpec
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

			httpWriteError(log, w, http.StatusInternalServerError,
				fmt.Errorf("failed to post spec: %w", err))
			return
		}

		httpWriteJson(log, w, r, http.StatusOK, true)
	}
}

// deleteSpec deletes a chain spec of given pk
// TODO @evanlinjin: We need to sign this.
// URI: /api/spec/<genesis-hash>
// Method: DELETE
func deleteSpec(ss store.SpecStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		hashStr := path.Base(r.URL.EscapedPath())

		hash, err := cipher.SHA256FromHex(hashStr)
		if err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to decode hash '%s': %w", hashStr, err))
		}

		if err := ss.DelSpec(r.Context(), hash); err != nil {
			if errors.Is(store.ErrBboltObjectNotExist, err) {
				httpWriteError(log, w, http.StatusNotFound, err)
				return
			}

			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		httpWriteJson(log, w, r, http.StatusOK, true)
	}
}
