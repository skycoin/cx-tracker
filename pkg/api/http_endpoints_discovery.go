package api

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"

	"github.com/SkycoinProject/cx/cxgo/cxspec"
	"github.com/skycoin/dmsg/cipher"

	"github.com/skycoin/cx-tracker/pkg/store"
)

// getPeer returns peer of given public key
// URI: /api/peers/<public-key>
// Method: GET
func getPeer(ps store.PeersStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		pkStr := path.Base(r.URL.EscapedPath())

		var pk cipher.PubKey
		if err := pk.Set(pkStr); err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to decode pk '%s': %w", pkStr, err))
			return
		}

		entry, err := ps.Entry(r.Context(), cipher.PubKey(pk))
		if err != nil {
			httpWriteError(log, w, http.StatusNotFound, err)
			return
		}

		if err := entry.Verify(); err != nil {
			httpWriteError(log, w, http.StatusInternalServerError, err)
			return
		}

		httpWriteJson(log, w, r, http.StatusOK, entry)
	}
}

// getPeersOfChain returns peers of a given chain hash
// URI: /api/peers?chain=<chain-hash>
// Method: GET
func getPeersOfChain(ps store.PeersStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		maxStr := r.URL.Query().Get("max")
		if maxStr == "" {
			maxStr = "10"
		}
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("invalid query value '%s' for 'max': %w", maxStr, err))
			return
		}

		hashStrs, ok := r.URL.Query()["chain"]
		if !ok {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("query key 'chain' expects atleast one argument"))
			return
		}

		hashs := make([]cipher.SHA256, len(hashStrs))
		for i, hashStr := range hashStrs {
			b, err := hex.DecodeString(hashStr)
			if err != nil {
				httpWriteError(log, w, http.StatusBadRequest,
					fmt.Errorf("failed to decode chain hash[%d] '%s': %w", i, hashStr, err))
				return
			}
			if copy(hashs[i][:], b) != len(cipher.SHA256{}) {
				httpWriteError(log, w, http.StatusBadRequest,
					fmt.Errorf("chain hash[%d] '%s' is of wrong length", i, hashStr))
				return
			}
		}

		var out []cxspec.CXChainAddresses

		for _, h := range hashs {
			peers, err := ps.RandPeersOfChain(r.Context(), h, max)
			if err != nil {
				log.WithError(err).WithField("chain_hash", h).Info("no peers found")
				continue
			}

			out = append(out, peers...)
		}

		httpWriteJson(log, w, r, http.StatusOK, out)
	}
}

// postPeers posts a peer entry
// URI: /api/peers
// Method: POST
func postPeers(ps store.PeersStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := httpLogger(r)

		var entry cxspec.SignedPeerEntry
		if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to decode entry: %w", err))
			return
		}

		if err := entry.Verify(); err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to verify entry: %w", err))
			return
		}

		if err := ps.UpdateEntry(r.Context(), entry); err != nil {
			httpWriteError(log, w, http.StatusBadRequest,
				fmt.Errorf("failed to update entry: %w", err))
			return
		}

		httpWriteJson(log, w, r, http.StatusOK, true)
	}
}