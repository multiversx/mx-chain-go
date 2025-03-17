package proofscache

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const defaultCleanupNonceDelta = 3
const defaultBucketSize = 100

var log = logger.GetOrCreate("dataRetriever/proofscache")

type proofsPool struct {
	mutCache sync.RWMutex
	cache    map[uint32]*proofsCache

	mutAddedProofSubscribers sync.RWMutex
	addedProofSubscribers    []func(headerProof data.HeaderProofHandler)
	cleanupNonceDelta        uint64
	bucketSize               int
}

// NewProofsPool creates a new proofs pool component
func NewProofsPool(cleanupNonceDelta uint64, bucketSize int) *proofsPool {
	if cleanupNonceDelta < defaultCleanupNonceDelta {
		log.Debug("proofs pool: using default cleanup nonce delta", "cleanupNonceDelta", defaultCleanupNonceDelta)
		cleanupNonceDelta = defaultCleanupNonceDelta
	}
	if bucketSize < defaultBucketSize {
		log.Debug("proofs pool: using default bucket size", "bucketSize", defaultBucketSize)
		bucketSize = defaultBucketSize
	}

	return &proofsPool{
		cache:                 make(map[uint32]*proofsCache),
		addedProofSubscribers: make([]func(headerProof data.HeaderProofHandler), 0),
		cleanupNonceDelta:     cleanupNonceDelta,
		bucketSize:            bucketSize,
	}
}

// AddProof will add the provided proof to the pool
func (pp *proofsPool) AddProof(
	headerProof data.HeaderProofHandler,
) bool {
	if check.IfNilReflect(headerProof) {
		return false
	}

	shardID := headerProof.GetHeaderShardId()
	headerHash := headerProof.GetHeaderHash()

	hasProof := pp.HasProof(shardID, headerHash)
	if hasProof {
		return false
	}

	pp.mutCache.Lock()
	proofsPerShard, ok := pp.cache[shardID]
	if !ok {
		proofsPerShard = newProofsCache(pp.bucketSize)
		pp.cache[shardID] = proofsPerShard
	}
	pp.mutCache.Unlock()

	log.Debug("added proof to pool",
		"header hash", headerProof.GetHeaderHash(),
		"epoch", headerProof.GetHeaderEpoch(),
		"nonce", headerProof.GetHeaderNonce(),
		"shardID", headerProof.GetHeaderShardId(),
		"pubKeys bitmap", headerProof.GetPubKeysBitmap(),
		"round", headerProof.GetHeaderRound(),
		"nonce", headerProof.GetHeaderNonce(),
		"isStartOfEpoch", headerProof.GetIsStartOfEpoch(),
	)

	proofsPerShard.addProof(headerProof)

	pp.callAddedProofSubscribers(headerProof)

	return true
}

func (pp *proofsPool) callAddedProofSubscribers(headerProof data.HeaderProofHandler) {
	pp.mutAddedProofSubscribers.RLock()
	defer pp.mutAddedProofSubscribers.RUnlock()

	for _, handler := range pp.addedProofSubscribers {
		go handler(headerProof)
	}
}

// CleanupProofsBehindNonce will cleanup proofs from pool based on nonce
func (pp *proofsPool) CleanupProofsBehindNonce(shardID uint32, nonce uint64) error {
	if nonce == 0 {
		return nil
	}

	if nonce <= pp.cleanupNonceDelta {
		return nil
	}

	nonce -= pp.cleanupNonceDelta

	pp.mutCache.RLock()
	proofsPerShard, ok := pp.cache[shardID]
	pp.mutCache.RUnlock()
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
	log.Trace("trying to get proof",
		"headerHash", headerHash,
		"shardID", shardID,
	)

	pp.mutCache.RLock()
	proofsPerShard, ok := pp.cache[shardID]
	pp.mutCache.RUnlock()
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

// RegisterHandler registers a new handler to be called when a new data is added
func (pp *proofsPool) RegisterHandler(handler func(headerProof data.HeaderProofHandler)) {
	if handler == nil {
		log.Error("attempt to register a nil handler to proofs pool")
		return
	}

	pp.mutAddedProofSubscribers.Lock()
	pp.addedProofSubscribers = append(pp.addedProofSubscribers, handler)
	pp.mutAddedProofSubscribers.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (pp *proofsPool) IsInterfaceNil() bool {
	return pp == nil
}
