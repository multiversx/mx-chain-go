package proofscache

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/pkg/errors"
)

var ErrNilProof = errors.New("nil proof provided")

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
) error {
	if headerProof == nil {
		return ErrNilProof
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

	return nil
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

func (pp *proofsPool) GetAllNotarizedProofs(
	shardID uint32,
) (map[string]data.HeaderProofHandler, error) {
	pp.mutCache.RLock()
	defer pp.mutCache.RUnlock()

	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		return nil, fmt.Errorf("%w: proofs cache per shard not found, shard ID: %d", ErrMissingProof, shardID)
	}

	return proofsPerShard.getAllProofs(), nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pp *proofsPool) IsInterfaceNil() bool {
	return pp == nil
}
