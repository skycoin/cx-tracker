package store

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"go.etcd.io/bbolt"
)

type BboltClientNodesStore struct {
	timeout time.Duration // peer timeout
	db      *bbolt.DB
}

func NewBboltClientNodesStore(db *bbolt.DB, timeout time.Duration) (*BboltClientNodesStore, error) {
	if timeout < 0 {
		panic(fmt.Sprintf("BboltClientNodesStore: timeout cannot be < 0."))
	}

	updateFunc := func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(clientNodesBucket); err != nil {
			return err
		}
		return nil
	}

	if err := db.Update(updateFunc); err != nil {
		return nil, err
	}

	s := &BboltClientNodesStore{ db: db, timeout: timeout }
	return s, nil
}

func (s *BboltClientNodesStore) RandPeers(ctx context.Context, chainPK cipher.PubKey, max int) ([]string, error) {
	all := make([]string, 0, 100)
	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(clientNodesBucket).Bucket(chainPK[:])

			rangeFunc := func(addr, rawTime []byte) error {
				// If timeout is set, and current time has passed expiration,
				// the peer should be marked for deletion and skipped.
				if s.timeout != 0 && time.Now().After(decodeTime(rawTime).Add(s.timeout)) {
					// TODO: Mark for deletion
					return nil
				}

				// Valid peer.
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

func (s *BboltClientNodesStore) AddPeer(ctx context.Context, chainPK cipher.PubKey, addr string) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			b, err := tx.Bucket(clientNodesBucket).CreateBucketIfNotExists(chainPK[:])
			if err != nil {
				return fmt.Errorf("failed to find client nodes bucket of pk '%s': %w", chainPK.Hex(), err)
			}

			if b.Get([]byte(addr)) != nil {
				return nil
			}

			if err := b.Put([]byte(addr), encodeTime(time.Now())); err != nil {
				return fmt.Errorf("failed to put client node address '%s' in chain '%s': %w",
					addr, chainPK.Hex(), err)
			}

			// increment count bucket.
			countK := append(clientNodesBucket, chainPK[:]...)
			return incrementObjectCount(tx, countK, 1)
		})
	}

	return doAsync(ctx, action)
}

func (s *BboltClientNodesStore) DelPeer(ctx context.Context, chainPK cipher.PubKey, addr string) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(clientNodesBucket).Bucket(chainPK[:])
			if b == nil {
				return fmt.Errorf("failed to delete client nodes under chain pk '%s': %w",
					chainPK.Hex(), ErrBboltObjectNotExist)
			}

			if err := b.Delete([]byte(addr)); err != nil {
				return fmt.Errorf("%v: %w", ErrBboltObjectNotExist, err)
			}

			countK := append(clientNodesBucket, chainPK[:]...)
			return decrementObjectCount(tx, countK, 1)
		})
	}

	return doAsync(ctx, action)
}

func (s *BboltClientNodesStore) DelAllOfPK(ctx context.Context, chainPK cipher.PubKey) error {
	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			if err := tx.Bucket(clientNodesBucket).DeleteBucket(chainPK[:]); err != nil {
				return err
			}

			countK := append(clientNodesBucket, chainPK[:]...)
			return tx.Bucket(countBucket).Delete(countK)
		})
	}

	return doAsync(ctx, action)
}
