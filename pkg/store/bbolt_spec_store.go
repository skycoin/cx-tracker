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

type BboltSpecStore struct {
	db *bbolt.DB
}

func NewBboltSpecStore(filename string) (*BboltSpecStore, error) {
	db, err := openBboltDB(filename)
	if err != nil {
		return nil, err
	}

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

func (s *BboltSpecStore) ChainSpecAll(ctx context.Context) ([]cxspec.ChainSpec, error) {
	var out []cxspec.ChainSpec

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			n := objectCount(tx, specBucket)
			out = make([]cxspec.ChainSpec, 0, n)

			eachFunc := func(k, v []byte) error {
				if len(v) < bboltSpecPrefixLen {
					return fmt.Errorf("expected value to be >= %d: %w",
						bboltSpecPrefixLen, ErrBboltInvalidValue)
				}

				var spec cxspec.ChainSpec
				if err := json.Unmarshal(v[bboltSpecPrefixLen:], &spec); err != nil {
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

func (s *BboltSpecStore) ChainSpecByChainPK(ctx context.Context, chainPK cipher.PubKey) (cxspec.ChainSpec, error) {
	var out cxspec.ChainSpec

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			return bboltChainSpecByPK(tx, chainPK, &out)
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return cxspec.ChainSpec{}, err
	}

	return out, nil
}



func (s *BboltSpecStore) ChainSpecByCoinTicker(ctx context.Context, coinTicker string) (cxspec.ChainSpec, error) {
	var out cxspec.ChainSpec

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
		return cxspec.ChainSpec{}, err
	}

	return out, nil
}

func (s *BboltSpecStore) AddSpec(ctx context.Context, spec cxspec.ChainSpec) error {
	b, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("failed to encode chain spec: %w", err)
	}
	chainPK := spec.ProcessedChainPubKey()

	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			var (
				specV   = tx.Bucket(specBucket).Get(chainPK[:])
				tickerV = tx.Bucket(specByTickerBucket).Get([]byte(spec.CoinTicker))
			)
			if specV != nil || tickerV != nil {
				return errors.New("attempted to add chain spec with reused chain pk or ticker")
			}
			return tx.Bucket(specBucket).Put([]byte(spec.ChainPubKey), b)

		})
	}

	return doAsync(ctx, action)
}

func (s *BboltSpecStore) DelSpec(ctx context.Context, chainPK cipher.PubKey) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			var spec cxspec.ChainSpec
			if err := bboltChainSpecByPK(tx, chainPK, &spec); err != nil {
				return err
			}

			err1 := tx.Bucket(specBucket).Delete(chainPK[:])
			err2 := tx.Bucket(specByTickerBucket).Delete([]byte(spec.CoinTicker))

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

func bboltChainSpecByPK(tx *bbolt.Tx, pk cipher.PubKey, spec *cxspec.ChainSpec) error {
	v := tx.Bucket(specBucket).Get(pk[:])
	if v == nil {
		return ErrBboltObjectNotExist
	}
	if len(v) < bboltSpecPrefixLen {
		return fmt.Errorf("expected value to be >= %d: %w",
			bboltSpecPrefixLen, ErrBboltInvalidValue)
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

