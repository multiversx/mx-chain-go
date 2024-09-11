package proofscache

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
)

type proofsPool struct {
	mutCache sync.RWMutex
	cache    map[uint32]*proofsCache
}

// NewProofsPool creates a new proofs pool component
func NewProofsPool() *proofsPool {
	return &proofsPool{
		cache: make(map[uint32]*proofsCache),
	}
}

func (pp *proofsPool) AddNotarizedProof(
	headerProof data.HeaderProofHandler,
) {
	if headerProof == nil {
		return
	}

	pp.mutCache.Lock()
	defer pp.mutCache.Unlock()

	shardID := headerProof.GetHeaderShardId()

	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		proofsPerShard = newProofsCache()
		pp.cache[shardID] = proofsPerShard
	}

	proofsPerShard.addProof(headerProof)
}

func (pp *proofsPool) CleanupNotarizedProofsBehindNonce(shardID uint32, nonce uint64) error {
	if nonce == 0 {
		return nil
	}

	pp.mutCache.RLock()
	defer pp.mutCache.RUnlock()

	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		return fmt.Errorf("%w: proofs cache per shard not found, shard ID: %d", ErrMissingProof, shardID)
	}

	proofsPerShard.cleanupProofsBehindNonce(nonce)

	return nil
}

func (pp *proofsPool) GetNotarizedProof(
	shardID uint32,
	headerHash []byte,
) (data.HeaderProofHandler, error) {
	pp.mutCache.RLock()
	defer pp.mutCache.RUnlock()

	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		return nil, fmt.Errorf("%w: proofs cache per shard not found, shard ID: %d", ErrMissingProof, shardID)
	}

	return proofsPerShard.getProofByHash(headerHash)
}

// IsInterfaceNil returns true if there is no value under the interface
func (pp *proofsPool) IsInterfaceNil() bool {
	return pp == nil
}
