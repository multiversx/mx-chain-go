package proofscache

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("dataRetriever/proofscache")

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

// AddProof will add the provided proof to the pool
func (pp *proofsPool) AddProof(
	headerProof data.HeaderProofHandler,
) error {
	if headerProof == nil {
		return ErrNilProof
	}

	shardID := headerProof.GetHeaderShardId()
	headerHash := headerProof.GetHeaderHash()

	hasProof := pp.HasProof(shardID, headerHash)
	if hasProof {
		log.Trace("there was already a valid proof for header, headerHash: %s", headerHash)
		return nil
	}

	pp.mutCache.Lock()
	defer pp.mutCache.Unlock()

	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		proofsPerShard = newProofsCache()
		pp.cache[shardID] = proofsPerShard
	}

	log.Trace("added proof to pool",
		"header hash", headerProof.GetHeaderHash(),
		"epoch", headerProof.GetHeaderEpoch(),
		"nonce", headerProof.GetHeaderNonce(),
		"shardID", headerProof.GetHeaderShardId(),
		"pubKeys bitmap", headerProof.GetPubKeysBitmap(),
	)

	proofsPerShard.addProof(headerProof)

	return nil
}

// CleanupProofsBehindNonce will cleanup proofs from pool based on nonce
func (pp *proofsPool) CleanupProofsBehindNonce(shardID uint32, nonce uint64) error {
	if nonce == 0 {
		return nil
	}

	pp.mutCache.RLock()
	defer pp.mutCache.RUnlock()

	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		return fmt.Errorf("%w: proofs cache per shard not found, shard ID: %d", ErrMissingProof, shardID)
	}

	log.Trace("cleanup proofs behind nonce",
		"nonce", nonce,
		"shardID", shardID,
	)

	proofsPerShard.cleanupProofsBehindNonce(nonce)

	return nil
}

// GetProof will get the proof from pool
func (pp *proofsPool) GetProof(
	shardID uint32,
	headerHash []byte,
) (data.HeaderProofHandler, error) {
	if headerHash == nil {
		return nil, fmt.Errorf("nil header hash")
	}

	pp.mutCache.RLock()
	defer pp.mutCache.RUnlock()

	log.Trace("trying to get proof",
		"headerHash", headerHash,
		"shardID", shardID,
	)

	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		return nil, fmt.Errorf("%w: proofs cache per shard not found, shard ID: %d", ErrMissingProof, shardID)
	}

	return proofsPerShard.getProofByHash(headerHash)
}

// HasProof will check if there is a proof for the provided hash
func (pp *proofsPool) HasProof(
	shardID uint32,
	headerHash []byte,
) bool {
	_, err := pp.GetProof(shardID, headerHash)
	return err == nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pp *proofsPool) IsInterfaceNil() bool {
	return pp == nil
}
