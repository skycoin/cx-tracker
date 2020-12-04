package store

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/SkycoinProject/cx/cxgo/cxspec"
	"github.com/skycoin/dmsg/cipher"
)

var (
	randInst = rand.New(rand.NewSource(int64(binaryEnc.Uint64(cipher.RandByte(8)))))
)

type chainAggregate struct {
	m map[cxspec.CXChainAddresses]int64 // value: last_seen timestamp
	mx sync.Mutex
}

func newChainAggregate() *chainAggregate {
	return &chainAggregate{ m: make(map[cxspec.CXChainAddresses]int64, 1) }
}

func (ca *chainAggregate) Update(addrs cxspec.CXChainAddresses) {
	ca.mx.Lock()
	ca.m[addrs] = time.Now().Unix()
	ca.mx.Unlock()
}

func (ca *chainAggregate) Rand(max int) []cxspec.CXChainAddresses {
	out := make([]cxspec.CXChainAddresses, 0, max)
	ca.mx.Lock()

	// determine skip range
	var skipN = 0
	if skipMax := len(ca.m) - max; skipMax > 0 {
		skipN = randInst.Intn(skipMax)
	}

	// populate results
	i := 0
	for addrs := range ca.m {
		if i++; i < skipN {
			continue
		}
		if out = append(out, addrs); len(out) >= max {
			break
		}
	}

	ca.mx.Unlock()
	return out
}

func (ca *chainAggregate) GarbageCollect(timeout time.Duration) int {
	now := time.Now().Unix()
	timeoutS := int64(timeout.Seconds())

	ca.mx.Lock()
	for addrs, lastSeen := range ca.m {
		if lastSeen+timeoutS < now {
			delete(ca.m, addrs)
		}
	}
	size := len(ca.m)
	ca.mx.Unlock()

	return size
}

type MemoryPeersStore struct {
	timeout    time.Duration
	entries    map[cipher.PubKey]cxspec.SignedPeerEntry
	aggregates map[cipher.SHA256]*chainAggregate
	mx         sync.Mutex
}

func NewMemoryPeersStore(timeout time.Duration, size int) *MemoryPeersStore {
	return &MemoryPeersStore{
		timeout:    timeout,
		entries:    make(map[cipher.PubKey]cxspec.SignedPeerEntry, size),
		aggregates: make(map[cipher.SHA256]*chainAggregate, size),
	}
}

func (ps *MemoryPeersStore) UpdateEntry(_ context.Context, entry cxspec.SignedPeerEntry) error {
	pk := entry.Entry.PublicKey

	ps.mx.Lock()

	// check 'last_seen' value
	oldEntry, ok := ps.entries[pk]
	if ok && entry.Entry.LastSeen <= oldEntry.Entry.LastSeen {
		return fmt.Errorf("updated entry's 'last_seen' field should be higher than that of last entry '%d'", oldEntry.Entry.LastSeen)
	}

	ps.entries[pk] = entry

	for hashStr, addrs := range entry.Entry.CXChains {
		var hash cipher.SHA256
		if _, err := hex.Decode(hash[:], []byte(hashStr)); err != nil {
			return fmt.Errorf("internal database error: %w", err)
		}

		aggregate, ok := ps.aggregates[hash]
		if !ok {
			// create if not exist
			aggregate = newChainAggregate()
			ps.aggregates[hash] = aggregate
		}

		aggregate.Update(addrs)
	}

	ps.mx.Unlock()
	return nil
}

func (ps *MemoryPeersStore) Entry(_ context.Context, pk cipher.PubKey) (cxspec.SignedPeerEntry, error) {
	ps.mx.Lock()
	defer ps.mx.Unlock()

	entry, ok := ps.entries[pk]
	if !ok {
		return cxspec.SignedPeerEntry{}, fmt.Errorf("entry of pk '%s' has timed out or does not exist", pk.Hex())
	}

	return entry, nil
}

func (ps *MemoryPeersStore) RandPeersOfChain(_ context.Context, hash cipher.SHA256, max int) ([]cxspec.CXChainAddresses, error) {
	ps.mx.Lock()
	aggregate, ok := ps.aggregates[hash]
	ps.mx.Unlock()

	if !ok {
		return []cxspec.CXChainAddresses{}, nil
	}

	out := aggregate.Rand(max)
	return out, nil
}

func (ps *MemoryPeersStore) GarbageCollect(_ context.Context) {
	ps.mx.Lock()
	for _, aggregate := range ps.aggregates {
		aggregate.GarbageCollect(ps.timeout)
	}
	ps.mx.Unlock()
}
