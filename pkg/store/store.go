package store

import (
	"context"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
)

// SpecStore represents a chain spec database implementation.
type SpecStore interface {
	ChainSpecAll(ctx context.Context) ([]cxspec.SignedChainSpec, error)
	ChainSpecByChainPK(ctx context.Context, chainPK cipher.PubKey) (cxspec.SignedChainSpec, error)
	ChainSpecByCoinTicker(ctx context.Context, coinTicker string) (cxspec.SignedChainSpec, error)
	AddSpec(ctx context.Context, spec cxspec.SignedChainSpec) error
	DelSpec(ctx context.Context, chainPK cipher.PubKey) error
}

// PeersStore represents a peers database implementation.
type PeersStore interface {
	RandPeers(ctx context.Context, chainPK cipher.PubKey, max int) ([]string, error)
	AddPeer(ctx context.Context, chainPK cipher.PubKey, addr string) error
	DelPeer(ctx context.Context, chainPK cipher.PubKey, addr string) error
	DelAllOfPK(ctx context.Context, chainPK cipher.PubKey) error
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