package store

import (
	"context"
	"encoding/json"
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
			v := tx.Bucket(specBucket).Get(chainPK[:])
			if v == nil {
				return ErrBboltNoExist
			}
			if len(v) < bboltSpecPrefixLen {
				return fmt.Errorf("expected value to be >= %d: %w",
					bboltSpecPrefixLen, ErrBboltInvalidValue)
			}

			return json.Unmarshal(v, &out)
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return cxspec.ChainSpec{}, err
	}

	return out, nil
}

