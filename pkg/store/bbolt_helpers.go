package store

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"time"

	"go.etcd.io/bbolt"
)

var binaryEnc = binary.BigEndian

var (
	ErrBboltNoExist = errors.New("bbolt object does not exist")
	ErrBboltInvalidValue = errors.New("invalid value in bbolt db")
)

const (
	bboltFileMode        = os.FileMode(0600)
	bboltFileOpenTimeout = time.Second * 10

	bboltSpecPrefixLen = 8
)

func openBboltDB(filename string) (*bbolt.DB, error) {
	opts := bbolt.DefaultOptions
	opts.Timeout = bboltFileOpenTimeout

	return bbolt.Open(filename, bboltFileMode, opts)
}

var (
	// specBucket is the identifier for the chain spec bucket
	//   key: [33B: chain public key]
	// value: [8B: uint64 iteration] [json encoded chain spec]
	specBucket = []byte("spec")

	// countBucket contains counts of various objects
	countBucket = []byte("count")
)

func objectCount(tx *bbolt.Tx, bucketName []byte) uint64 {
	b := tx.Bucket(countBucket).Get(bucketName)
	if b == nil {
		return 0
	}
	return binaryEnc.Uint64(b)
}

func doAsync(ctx context.Context, action func() error) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- action()
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}