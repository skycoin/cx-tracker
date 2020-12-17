package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/skycoin/cx-chains/src/cx/cxspec"
	"github.com/skycoin/skycoin/src/cipher"
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

		if _, err := tx.CreateBucketIfNotExists(countBucket); err != nil {
			return err
		}
		return nil
	}

	if err := db.Update(updateFunc); err != nil {
		return nil, err
	}

	s := &BboltSpecStore{db: db}
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
func (s *BboltSpecStore) ChainSpec(ctx context.Context, hash cipher.SHA256) (cxspec.SignedChainSpec, error) {
	var out cxspec.SignedChainSpec

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			return bboltChainSpecByGenesisHash(tx, hash, &out)
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

	genBlock, err := spec.Spec.GenerateGenesisBlock()
	if err != nil {
		return err
	}
	hash := genBlock.HashHeader()

	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			specV := tx.Bucket(specBucket).Get(hash[:])
			if specV != nil {
				return errors.New("attempted to add chain spec with reused chain genesis hash")
			}

			return tx.Bucket(specBucket).Put(hash[:], b)
		})
	}

	return doAsync(ctx, action)
}

// DelSpec implements SpecStore.
func (s *BboltSpecStore) DelSpec(ctx context.Context, hash cipher.SHA256) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			var spec cxspec.SignedChainSpec
			if err := bboltChainSpecByGenesisHash(tx, hash, &spec); err != nil {
				return err
			}

			return tx.Bucket(specBucket).Delete(hash[:])
		})
	}

	return doAsync(ctx, action)
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func bboltChainSpecByGenesisHash(tx *bbolt.Tx, hash cipher.SHA256, spec *cxspec.SignedChainSpec) error {
	v := tx.Bucket(specBucket).Get(hash[:])
	if v == nil {
		return ErrBboltObjectNotExist
	}

	return json.Unmarshal(v, spec)
}
