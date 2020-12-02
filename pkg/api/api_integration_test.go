package api

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/skycoin/cx-tracker/pkg/store"
)

func TestNewHTTPRouter(t *testing.T) {
	tempFilename := filepath.Join(os.TempDir(), fmt.Sprintf("TestNewHTTPRouter_%d.db", time.Now().Unix()))

	db, err := store.OpenBboltDB(tempFilename)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
		assert.NoError(t, os.Remove(tempFilename))
	}()

	ss, err := store.NewBboltSpecStore(db)
	require.NoError(t, err)

	httpS := httptest.NewServer(NewHTTPRouter(ss))
	defer httpS.Close()

	httpC := cxspec.NewCXTrackerClient(logrus.New(), httpS.Client(), httpS.URL)

	// Test 'single_spec' tests registration and deletion of chain specs one at
	// a time.
	const singleSpecRounds = 100
	for i := 0; i < singleSpecRounds; i++ {
		i := i

		t.Run("single_spec", func(t *testing.T) {
			spec, _ := randSpec(t, i)
			block, err := spec.Spec.GenerateGenesisBlock()
			require.NoError(t, err)

			// post spec
			require.NoError(t, httpC.PostSpec(context.TODO(), spec))

			// get spec by hash
			spec2, err := httpC.SpecByGenesisHash(context.TODO(), block.HashHeader())
			require.NoError(t, err)
			require.Equal(t, spec.Sig, spec2.Sig)

			// get all specs
			allSpecs, err := httpC.AllSpecs(context.TODO())
			require.NoError(t, err)
			require.Len(t, allSpecs, 1)
			require.Equal(t, spec.Sig, allSpecs[0].Sig)

			// delete spec
			require.NoError(t, httpC.DelSpec(context.TODO(), block.HashHeader()))

			// get all specs
			allSpecs, err = httpC.AllSpecs(context.TODO())
			require.NoError(t, err)
			require.Len(t, allSpecs, 0)
		})
	}

	// TODO @evanlinjin: We need more tests.
	// - Multiple specification registration/destruction.
	// - Invalid spec registration (test security checks; e.g. duplicates).
}

// randSpec generates a new spec of coin name 'coin%d' and ticker name 'COIN%d'
// given the int 'i'.
// A signed chain spec is returned alongside it's chain secret key.
func randSpec(t *testing.T, i int) (cxspec.SignedChainSpec, cipher.SecKey) {
	pk, sk := cipher.GenerateKeyPair()

	coin := fmt.Sprintf("coin%d", i)
	ticker := fmt.Sprintf("COIN%d", i)
	addr := cipher.AddressFromPubKey(pk)

	spec, err := cxspec.New(coin, ticker, sk, addr, nil)
	require.NoError(t, err)

	signedSpec, err := cxspec.MakeSignedChainSpec(*spec, sk)
	require.NoError(t, err)

	return signedSpec, sk
}
