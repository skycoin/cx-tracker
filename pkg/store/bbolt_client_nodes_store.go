package store

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"go.etcd.io/bbolt"
)

type BboltClientNodesStore struct {
	db *bbolt.DB
}

func NewBboltClientNodesStore(db *bbolt.DB) (*BboltClientNodesStore, error) {
	updateFunc := func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(clientNodesBucket); err != nil {
			return err
		}
		return nil
	}

	if err := db.Update(updateFunc); err != nil {
		return nil, err
	}

	s := &BboltClientNodesStore{ db: db }
	return s, nil
}

func (s *BboltClientNodesStore) ClientNodesRand(ctx context.Context, chainPK cipher.PubKey, max int) ([]string, error) {
	all := make([]string, 0, 100)
	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(clientNodesBucket).Bucket(chainPK[:])

			rangeFunc := func(addr, _ []byte) error {
				all = append(all, string(addr))
				return nil
			}

			return b.ForEach(rangeFunc)
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return nil, err
	}

	if len(all) < max {
		max = len(all)
	}

	outMap := make(map[string]struct{}, max)
	for i := 0; i < max; i++ {
		n := rand.Intn(len(all))
		outMap[all[n]] = struct{}{}
	}

	out := make([]string, 0, max)
	for addr := range outMap {
		out = append(out, addr)
	}

	return out, nil
}

func (s *BboltClientNodesStore) AddClientNodes(ctx context.Context, chainPK cipher.PubKey, addresses []string) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			b, err := tx.Bucket(clientNodesBucket).CreateBucketIfNotExists(chainPK[:])
			if err != nil {
				return fmt.Errorf("failed to find client nodes bucket of pk '%s': %w", chainPK.Hex(), err)
			}

			for i, addr := range addresses {
				if err := b.Put([]byte(addr), []byte{1}); err != nil {
					return fmt.Errorf("failed to put client node address '%d:%s' in chain '%s': %w",
						i, addr, chainPK.Hex(), err)
				}
			}

			return nil
		})
	}

	return doAsync(ctx, action)
}

func (s *BboltClientNodesStore) DelClientNodes(ctx context.Context, chainPK cipher.PubKey, addresses []string) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(clientNodesBucket).Bucket(chainPK[:])
			if b == nil {
				return fmt.Errorf("failed to delete client nodes under chain pk '%s': %w",
					chainPK.Hex(), ErrBboltObjectNotExist)
			}

			for _, addr := range addresses {
				_ = b.Delete([]byte(addr)) //nolint:errcheck
			}
			return nil
		})
	}

	return doAsync(ctx, action)
}

func (s *BboltClientNodesStore) DelAllOfChainPK(ctx context.Context, chainPK cipher.PubKey) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			return tx.Bucket(clientNodesBucket).DeleteBucket(chainPK[:])
		})
	}

	return doAsync(ctx, action)
}
