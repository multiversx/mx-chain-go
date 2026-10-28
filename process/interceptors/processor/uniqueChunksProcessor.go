package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

type uniqueChunksProcessor struct {
	km         sync.KeyRWMutexHandler
	cache      storage.Cacher
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
}

// NewUniqueChunksProcessor creates a new instance of unique chunks processor
func NewUniqueChunksProcessor(
	cache storage.Cacher,
	marshaller marshal.Marshalizer,
	hasher hashing.Hasher,
) (*uniqueChunksProcessor, error) {
	if check.IfNil(cache) {
		return nil, process.ErrNilInterceptedDataCache
	}
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(hasher) {
		return nil, process.ErrNilHasher
	}

	return &uniqueChunksProcessor{
		km:         sync.NewKeyRWMutex(),
		cache:      cache,
		marshaller: marshaller,
		hasher:     hasher,
	}, nil
}

// CheckBatch verifies if the provided batch is already received
func (proc *uniqueChunksProcessor) CheckBatch(b *batch.Batch, _ process.WhiteListHandler, broadcastMethod p2p.BroadcastMethod) (process.CheckedChunkResult, error) {
	if b == nil || len(b.Data) == 0 {
		return process.CheckedChunkResult{}, nil
	}
	if isDirectSend(broadcastMethod) {
		return process.CheckedChunkResult{}, nil
	}

	batchHash, err := core.CalculateHash(proc.marshaller, proc.hasher, b)
	if err != nil {
		return process.CheckedChunkResult{}, err
	}
	batchHashStr := string(batchHash)

	proc.km.RLock(batchHashStr)
	defer proc.km.RUnlock(batchHashStr)

	if _, ok := proc.cache.Get(batchHash); ok {
		return process.CheckedChunkResult{}, process.ErrDuplicatedInterceptedDataNotAllowed
	}

	return process.CheckedChunkResult{}, nil
}

// MarkVerified marks the batch as verified
func (proc *uniqueChunksProcessor) MarkVerified(b *batch.Batch, broadcastMethod p2p.BroadcastMethod) {
	if b == nil || len(b.Data) == 0 {
		return
	}
	if isDirectSend(broadcastMethod) {
		return
	}
	batchHash, err := core.CalculateHash(proc.marshaller, proc.hasher, b)
	if err != nil {
		return
	}
	batchHashStr := string(batchHash)
	proc.km.Lock(batchHashStr)
	defer proc.km.Unlock(batchHashStr)

	proc.cache.Put(batchHash, struct{}{}, 0)
}

func isDirectSend(broadcastMethod p2p.BroadcastMethod) bool {
	return broadcastMethod == p2p.Direct
}

// Close closes the internal cache
func (proc *uniqueChunksProcessor) Close() error {
	return proc.cache.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (proc *uniqueChunksProcessor) IsInterfaceNil() bool {
	return proc == nil
}
