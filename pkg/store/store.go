package store

import (
	"context"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
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
	RandPeers(ctx context.Context, hash cipher.SHA256, max int) ([]string, error)
	AddPeer(ctx context.Context, hash cipher.SHA256, addr string) error
	DelPeer(ctx context.Context, hash cipher.SHA256, addr string) error
	DelAllOfPK(ctx context.Context, hash cipher.SHA256) error
}

// DeleteProblematicSpecs removes problematic specs
func DeleteProblematicSpecs(ctx context.Context, ss SpecStore, pks []cipher.PubKey) error {
	action := func() error {
		var pk cipher.PubKey
		pk.Null()
		return nil
	}

	return doAsync(ctx, action)
}
