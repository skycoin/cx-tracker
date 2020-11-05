package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SkycoinProject/cx-chains/src/cipher"
	"github.com/SkycoinProject/cx/cxgo/cxspec"
)

// SignedSpec contains a chain spec alongside a valid signature.
type SignedSpec struct {
	Spec cxspec.ChainSpec `json:"spec"`
	Sig  string           `json:"sig"` // hex representation of signature
}

// Verify checks the following:
// - Spec is of right era, has valid chain pk, and generates valid genesis block.
// - Signature is valid
func (ss *SignedSpec) Verify() error {
	if era := ss.Spec.SpecEra; era != cxspec.Era {
		return fmt.Errorf("unexpected chain spec era '%s' (expected '%s')",
			era, cxspec.Era)
	}

	if _, err := ss.Spec.GenerateGenesisBlock(); err != nil {
		return fmt.Errorf("chain spec failed to generate genesis block: %w", err)
	}

	sig, err := cipher.SigFromHex(ss.Sig)
	if err != nil {
		return fmt.Errorf("failed to decode spec signature: %w", err)
	}

	pk := ss.Spec.ProcessedChainPubKey()
	hash := ss.Spec.SpecHash()

	if err := cipher.VerifyPubKeySignedHash(pk, sig, hash); err != nil {
		return fmt.Errorf("failed to verify spec signature: %w", err)
	}

	return nil
}

// SignedTrustedNodes contains a trusted nodes object alongside a valid signature.
type SignedTrustedNodes struct {
	TrustedNodes cxspec.TrustedNodes `json:"trusted_nodes"`
	Sig          string              `json:"sig"`
}

// Verify checks the following:
// - TrustedNodes has valid public key and iteration higher than previous.
// - Signature checks out with pk and is valid.
func (sn *SignedTrustedNodes) Verify(lastI uint64) error {
	pk, err := cipher.PubKeyFromHex(sn.TrustedNodes.ChainPubKey)
	if err != nil {
		return fmt.Errorf("object's pk is invalid: %w", err)
	}

	if lastI <= sn.TrustedNodes.Iteration {
		return fmt.Errorf("expected iteration >= %d", lastI)
	}

	sig, err := cipher.SigFromHex(sn.Sig)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}

	j, err := json.Marshal(sn.TrustedNodes)
	if err != nil {
		panic(err) // should not happen
	}

	hash := cipher.SumSHA256(j)
	if err := cipher.VerifyPubKeySignedHash(pk, sig, hash); err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	return nil
}

// Aggregate is the aggregate of all store implementations.
type Aggregate struct {
	Specs        SpecStore
	TrustedNodes TrustedNodesStore
	ClientNodes  ClientNodesStore
	Sigs         SigStore
}

func (a Aggregate) InjectSpec(ctx context.Context, spec *SignedSpec) {

}