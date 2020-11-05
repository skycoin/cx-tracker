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
	ErrBboltObjectNotExist      = errors.New("bbolt object does not exist")
	ErrBboltObjectAlreadyExists = errors.New("bbolt object already exists")
	ErrBboltInvalidValue        = errors.New("invalid value in bbolt db")
)

const (
	bboltFileMode        = os.FileMode(0600)
	bboltFileOpenTimeout = time.Second * 10
)

// OpenBboltDB opens a bbolt database file.
func OpenBboltDB(filename string) (*bbolt.DB, error) {
	opts := bbolt.DefaultOptions
	opts.Timeout = bboltFileOpenTimeout

	return bbolt.Open(filename, bboltFileMode, opts)
}

var (
	// specBucket is the identifier for the chain spec bucket
	//   key: [33B: chain public key]
	// value: [json encoded chain spec]
	specBucket = []byte("spec")

	// specByTickerBucket relates chain ticker to chain public key
	//   key: [ticker string]
	// value: [33B: chain public key]
	specByTickerBucket = []byte("spec_by_ticker")

	// trustedNodesBucket is the identifier for the trusted nodes bucket
	//   key: [33B: chain public key]
	// value: [json encoded trusted nodes object]
	trustedNodesBucket = []byte("trusted_nodes")

	// clientNodesBucket is the identifier for the client nodes bucket
	//   key: [33B: chain public key]
	// value: [bucket of client node addresses keys]
	clientNodesBucket = []byte("client_nodes")

	// sigBucket is the identifier
	sigBucket = []byte("sig")

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