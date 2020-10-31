package store

import (
	"context"
	"errors"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
)

var (
	// ErrNotFound occurs when the requested object cannot be found in database.
	ErrNotFound = errors.New("not found")
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
	ChainSpecExists(ctx context.Context, chainPK cipher.PubKey) (bool, error)
	AddSpec(ctx context.Context, spec cxspec.ChainSpec) error
	DelSpec(ctx context.Context, chainPK cipher.PubKey) error
}

type TrustedNodesStore interface {
	TrustedNodesByChainPK(ctx context.Context, chainPK cipher.PubKey) (cxspec.TrustedNodes, error)
	AddTrustedNodes(ctx context.Context, nodes cxspec.TrustedNodes) error
	DelTrustedNodes(ctx context.Context, chainPK cipher.PubKey) error
}

type ClientNodesStore interface {
	ClientNodesRand(ctx context.Context, max int) ([]string, error)
	AddClientNodes(ctx context.Context, chainPK cipher.PubKey, addresses []string) error
	DelClientNodes(ctx context.Context, chainPK cipher.PubKey, addresses []string) error
}

// SigStore represents a signature database implementation.
type SigStore interface {
	Sig(ot ObjectType, pk cipher.PubKey) (cipher.Sig, uint64, error)
	SigList(ot ObjectType, pks []cipher.PubKey) ([]cipher.Sig, error)
	AddSig(ot ObjectType, pk cipher.PubKey, sig cipher.Sig, oi uint64) error
	DelSig(ot ObjectType, pk cipher.PubKey) error
}
