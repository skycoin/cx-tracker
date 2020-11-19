package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
	"go.etcd.io/bbolt"
)

// BboltSpecStore implements SpecStore with a bbolt.DB database.
type BboltSpecStore struct {
	db *bbolt.DB
}

// NewBboltSpecStore creates a new BboltSpecStore with a given database file.
func NewBboltSpecStore(db *bbolt.DB) (*BboltSpecStore, error) {
	updateFunc := func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(specBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(specByTickerBucket); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(countBucket); err != nil {
			return err
		}
		return nil
	}

	if err := db.Update(updateFunc); err != nil {
		return nil, err
	}

	s := &BboltSpecStore{ db: db }
	return s, nil
}

// ChainSpecAll implements SpecStore.
func (s *BboltSpecStore) ChainSpecAll(ctx context.Context) ([]cxspec.SignedChainSpec, error) {
	var out []cxspec.SignedChainSpec

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			n := objectCount(tx, specBucket)
			out = make([]cxspec.SignedChainSpec, 0, n)

			eachFunc := func(k, v []byte) error {
				var spec cxspec.SignedChainSpec
				if err := json.Unmarshal(v, &spec); err != nil {
					return err
				}

				out = append(out, spec)
				return nil
			}

			return tx.Bucket(specBucket).ForEach(eachFunc)
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return nil, err
	}

	return out, nil
}

// ChainSpecByChainPK implements SpecStore.
func (s *BboltSpecStore) ChainSpecByChainPK(ctx context.Context, chainPK cipher.PubKey) (cxspec.SignedChainSpec, error) {
	var out cxspec.SignedChainSpec

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			return bboltChainSpecByPK(tx, chainPK, &out)
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return cxspec.SignedChainSpec{}, err
	}

	return out, nil
}

// ChainSpecByCoinTicker implements SpecStore.
func (s *BboltSpecStore) ChainSpecByCoinTicker(ctx context.Context, coinTicker string) (cxspec.SignedChainSpec, error) {
	var out cxspec.SignedChainSpec

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			pk, err := bboltChainPKByTicker(tx, coinTicker)
			if err != nil {
				return err
			}

			return bboltChainSpecByPK(tx, pk, &out)
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return cxspec.SignedChainSpec{}, err
	}

	return out, nil
}

// AddSpec implements SpecStore.
func (s *BboltSpecStore) AddSpec(ctx context.Context, spec cxspec.SignedChainSpec) error {
	b, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to encode chain spec: %w", err)
	}
	chainPK := spec.Spec.ProcessedChainPubKey()

	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			var (
				specV   = tx.Bucket(specBucket).Get(chainPK[:])
				tickerV = tx.Bucket(specByTickerBucket).Get([]byte(spec.Spec.CoinTicker))
			)
			if specV != nil || tickerV != nil {
				return errors.New("attempted to add chain spec with reused chain pk or ticker")
			}
			// TODO: Implement replace check.
			return tx.Bucket(specBucket).Put([]byte(spec.Spec.ChainPubKey), b)
		})
	}

	return doAsync(ctx, action)
}

// DelSpec implements SpecStore.
func (s *BboltSpecStore) DelSpec(ctx context.Context, chainPK cipher.PubKey) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			var spec cxspec.SignedChainSpec
			if err := bboltChainSpecByPK(tx, chainPK, &spec); err != nil {
				return err
			}

			err1 := tx.Bucket(specBucket).Delete(chainPK[:])
			err2 := tx.Bucket(specByTickerBucket).Delete([]byte(spec.Spec.CoinTicker))

			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}

			return nil
		})
	}

	return doAsync(ctx, action)
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func bboltChainSpecByPK(tx *bbolt.Tx, pk cipher.PubKey, spec *cxspec.SignedChainSpec) error {
	v := tx.Bucket(specBucket).Get(pk[:])
	if v == nil {
		return ErrBboltObjectNotExist
	}

	return json.Unmarshal(v, spec)
}

func bboltChainPKByTicker(tx *bbolt.Tx, ticker string) (cipher.PubKey, error) {
	rawPK := tx.Bucket(specByTickerBucket).Get([]byte(ticker))
	if rawPK == nil {
		return cipher.PubKey{}, fmt.Errorf("cannot find object in %s bucket of key %s: %w",
			string(specByTickerBucket), ticker, ErrBboltObjectNotExist)
	}

	var pk cipher.PubKey
	copy(pk[:], rawPK)

	if err := pk.Verify(); err != nil {
		return cipher.PubKey{}, fmt.Errorf("%v: %w", ErrBboltInvalidValue, err)
	}

	return pk, nil
}

