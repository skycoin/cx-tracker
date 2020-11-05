package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
	"go.etcd.io/bbolt"
)

// BboltTrustedNodesStore implements TrustedNodesStore with a bbolt.DB database.
type BboltTrustedNodesStore struct {
	db *bbolt.DB
}

// NewBboltTrustedNodesStore creates a new BboltTrustedNodesStore with a given database file.
func NewBboltTrustedNodesStore(db *bbolt.DB) (*BboltTrustedNodesStore, error) {
	updateFunc := func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(trustedNodesBucket); err != nil {
			return err
		}
		return nil
	}

	if err := db.Update(updateFunc); err != nil {
		return nil, err
	}

	s := &BboltTrustedNodesStore{ db: db }
	return s, nil
}

// TrustedNodesByChainPK implements TrustedNodesStore.
func (s *BboltTrustedNodesStore) TrustedNodesByChainPK(ctx context.Context, chainPK cipher.PubKey) (cxspec.TrustedNodes, error) {
	var out cxspec.TrustedNodes

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			return bboltTrustedNodesByPK(tx, chainPK, &out)
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return cxspec.TrustedNodes{}, err
	}

	return out, nil
}

// AddTrustedNodes implements TrustedNodesStore.
func (s *BboltTrustedNodesStore) AddTrustedNodes(ctx context.Context, nodes cxspec.TrustedNodes) error {
	pk, err := cipher.PubKeyFromHex(nodes.ChainPubKey)
	if err != nil {
		return fmt.Errorf("invalid chain pk: %w", err)
	}

	j, err := json.Marshal(nodes)
	if err != nil {
		return fmt.Errorf("failed to encode object to json: %w", err)
	}

	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(trustedNodesBucket)

			if b.Get(pk[:]) != nil {
				return ErrBboltObjectAlreadyExists
			}

			return b.Put(pk[:], j)
		})
	}

	return doAsync(ctx, action)
}

// DelTrustedNodes implements TrustedNodesStore.
func (s *BboltTrustedNodesStore) DelTrustedNodes(ctx context.Context, chainPK cipher.PubKey) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			return tx.Bucket(trustedNodesBucket).Delete(chainPK[:])
		})
	}

	return doAsync(ctx, action)
}

/*
	<<< HELPER FUNCTIONS >>>
*/

func bboltTrustedNodesByPK(tx *bbolt.Tx, pk cipher.PubKey, nodes *cxspec.TrustedNodes) error {
	v := tx.Bucket(trustedNodesBucket).Get(pk[:])
	if v == nil {
		return ErrBboltObjectNotExist
	}
	return json.Unmarshal(v, nodes)
}

