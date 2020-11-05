package store

import (
	"context"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
)

type ObjectType string

const (
	SpecObjT = ObjectType("spec")
	NodeObjT = ObjectType("node")
)

// SpecStore represents a chain spec database implementation.
type SpecStore interface {
	ChainSpecAll(ctx context.Context) ([]cxspec.ChainSpec, error)
	ChainSpecByChainPK(ctx context.Context, chainPK cipher.PubKey) (cxspec.ChainSpec, error)
	ChainSpecByCoinTicker(ctx context.Context, coinTicker string) (cxspec.ChainSpec, error)
	AddSpec(ctx context.Context, spec cxspec.ChainSpec) error
	DelSpec(ctx context.Context, chainPK cipher.PubKey) error
}

// TrustedNodesStore represents a trusted nodes database implementation.
type TrustedNodesStore interface {
	TrustedNodesByChainPK(ctx context.Context, chainPK cipher.PubKey) (cxspec.TrustedNodes, error)
	AddTrustedNodes(ctx context.Context, nodes cxspec.TrustedNodes) error
	DelTrustedNodes(ctx context.Context, chainPK cipher.PubKey) error
}

// ClientNodesStore represents a client nodes database implementation.
type ClientNodesStore interface {
	ClientNodesRand(ctx context.Context, chainPK cipher.PubKey, max int) ([]string, error)
	AddClientNodes(ctx context.Context, chainPK cipher.PubKey, addresses []string) error
	DelClientNodes(ctx context.Context, chainPK cipher.PubKey, addresses []string) error
	DelAllOfChainPK(ctx context.Context, chainPK cipher.PubKey) error
}

// SigStore represents a signature database implementation.
type SigStore interface {
	Sig(ctx context.Context, ot ObjectType, pk cipher.PubKey) (cipher.Sig, error)
	SigList(ctx context.Context, ot ObjectType, pks []cipher.PubKey) ([]cipher.Sig, error)
	AddSig(ctx context.Context, ot ObjectType, pk cipher.PubKey, sig cipher.Sig) error
	DelSig(ctx context.Context, ot ObjectType, pk cipher.PubKey) error
}
