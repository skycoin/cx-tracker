package store

import (
	"context"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"go.etcd.io/bbolt"
)

type BboltSigStore struct {
	db *bbolt.DB
}

func NewBboltSigStore(db *bbolt.DB) (*BboltSigStore, error) {
	updateFunc := func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(sigBucket); err != nil {
			return err
		}
		return nil
	}

	if err := db.Update(updateFunc); err != nil {
		return nil, err
	}

	s := &BboltSigStore{db: db}
	return s, nil
}

func (s *BboltSigStore) Sig(ctx context.Context, ot ObjectType, pk cipher.PubKey) (cipher.Sig, error) {
	key := append(pk[:], []byte(ot)...)
	var sig cipher.Sig

	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			v := tx.Bucket(sigBucket).Get(key)
			if v == nil {
				return ErrBboltObjectNotExist
			}

			var err error
			sig, err = cipher.NewSig(v)
			return err
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return cipher.Sig{}, nil
	}

	return sig, nil
}

func (s *BboltSigStore) SigList(ctx context.Context, ot ObjectType, pks []cipher.PubKey) ([]cipher.Sig, error) {
	out := make([]cipher.Sig, len(pks))

	action := func() error {
		return s.db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(sigBucket)

			for i, pk := range pks {
				key := append(pk[:], []byte(ot)...)
				copy(out[i][:], b.Get(key))
			}
			return nil
		})
	}

	if err := doAsync(ctx, action); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *BboltSigStore) AddSig(ctx context.Context, ot ObjectType, pk cipher.PubKey, sig cipher.Sig) error {
	key := append(pk[:], []byte(ot)...)

	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket(sigBucket)

			if b.Get(key) != nil {
				return ErrBboltObjectAlreadyExists
			}

			return b.Put(key, sig[:])
		})
	}

	return doAsync(ctx, action)
}

func (s *BboltSigStore) DelSig(ctx context.Context, ot ObjectType, pk cipher.PubKey) error {
	key := append(pk[:], []byte(ot)...)

	action := func() error {
		return s.db.Update(func(tx *bbolt.Tx) error {
			return tx.Bucket(sigBucket).Delete(key)
		})
	}

	return doAsync(ctx, action)
}