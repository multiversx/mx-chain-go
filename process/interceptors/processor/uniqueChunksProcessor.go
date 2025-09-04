package processor

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

type uniqueChunksProcessor struct {
	km         sync.KeyRWMutexHandler
	cache      storage.Cacher
	marshaller marshal.Marshalizer
}

// NewUniqueChunksProcessor creates a new instance of unique chunks processor
func NewUniqueChunksProcessor(cache storage.Cacher, marshaller marshal.Marshalizer) (*uniqueChunksProcessor, error) {
	if check.IfNil(cache) {
		return nil, process.ErrNilInterceptedDataCache
	}
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}

	return &uniqueChunksProcessor{
		km:         sync.NewKeyRWMutex(),
		cache:      cache,
		marshaller: marshaller,
	}, nil
}

// CheckBatch verifies if the provided batch is already received
func (proc *uniqueChunksProcessor) CheckBatch(b *batch.Batch, _ process.WhiteListHandler) (process.CheckedChunkResult, error) {
	if b == nil || len(b.Data) == 0 {
		return process.CheckedChunkResult{}, nil
	}

	batchHash, err := proc.marshaller.Marshal(b)
	if err != nil {
		return process.CheckedChunkResult{}, err
	}
	batchHashStr := string(batchHash)

	proc.km.Lock(batchHashStr)
	defer proc.km.Unlock(batchHashStr)

	if _, ok := proc.cache.Get(batchHash); ok {
		return process.CheckedChunkResult{}, process.DuplicatedInterceptedDataNotAllowed
	}

	proc.cache.Put(batchHash, struct{}{}, 0)

	return process.CheckedChunkResult{}, nil
}

// Close closes the internal cache
func (proc *uniqueChunksProcessor) Close() error {
	return proc.cache.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (proc *uniqueChunksProcessor) IsInterfaceNil() bool {
	return proc == nil
}
