package store

import (
	"context"

	"github.com/skycoin/cx-chains/src/cx/cxspec"
	cipher2 "github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/cipher"
)

// SpecStore represents a chain spec database implementation.
type SpecStore interface {
	ChainSpecAll(ctx context.Context) ([]cxspec.SignedChainSpec, error)
	ChainSpec(ctx context.Context, hash cipher.SHA256) (cxspec.SignedChainSpec, error)
	AddSpec(ctx context.Context, spec cxspec.SignedChainSpec) error
	DelSpec(ctx context.Context, hash cipher.SHA256) error
}

// PeersStore represents a peers database implementation.
type PeersStore interface {
	UpdateEntry(ctx context.Context, entry cxspec.SignedPeerEntry) error
	Entry(ctx context.Context, pk cipher2.PubKey) (cxspec.SignedPeerEntry, error)
	RandPeersOfChain(ctx context.Context, hash cipher2.SHA256, max int) ([]cxspec.CXChainAddresses, error)
	GarbageCollect(ctx context.Context)
}

// DeleteProblematicSpecs removes problematic specs.
// TODO @evanlinjin: Determine if this is still needed.
func DeleteProblematicSpecs(ctx context.Context, ss SpecStore, pks []cipher.PubKey) error {
	action := func() error {
		var pk cipher.PubKey
		pk.Null()
		return nil
	}

	return doAsync(ctx, action)
}
