package processor

import (
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ process.InterceptorProcessor = (*MiniblockInterceptorProcessor)(nil)

var log = logger.GetOrCreate("process/interceptors/processor")

// MiniblockInterceptorProcessor is the processor used when intercepting miniblocks
type MiniblockInterceptorProcessor struct {
	miniblockCache     storage.Cacher
	marshalizer        marshal.Marshalizer
	hasher             hashing.Hasher
	shardCoordinator   sharding.Coordinator
	whiteListHandler   process.WhiteListHandler
	registeredHandlers []func(topic string, hash []byte, data interface{})
	mutHandlers        sync.RWMutex
}

// NewMiniblockInterceptorProcessor creates a new MiniblockInterceptorProcessor instance
func NewMiniblockInterceptorProcessor(argument *ArgMiniblockInterceptorProcessor) (*MiniblockInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.MiniblockCache) {
		return nil, process.ErrNilMiniBlockPool
	}
	if check.IfNil(argument.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(argument.WhiteListHandler) {
		return nil, process.ErrNilWhiteListHandler
	}

	return &MiniblockInterceptorProcessor{
		miniblockCache:     argument.MiniblockCache,
		marshalizer:        argument.Marshalizer,
		hasher:             argument.Hasher,
		shardCoordinator:   argument.ShardCoordinator,
		whiteListHandler:   argument.WhiteListHandler,
		registeredHandlers: make([]func(topic string, hash []byte, data interface{}), 0),
	}, nil
}

// Validate checks if the intercepted data can be processed
// It returns nil as a body might consist of multiple miniblocks
// Since some might be valid and others not, we rather do the checking when
// we iterate the slice for processing as it is optimal to do so
func (mip *MiniblockInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the received miniblocks inside the miniblock cacher after a new validation round
// that will be done on each miniblock
func (mip *MiniblockInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, topic string) error {
	interceptedMiniblock, ok := data.(*interceptedBlocks.InterceptedMiniblock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	miniblock := interceptedMiniblock.Miniblock()
	hash := interceptedMiniblock.Hash()

	go mip.notify(miniblock, hash, topic)

	if !mip.whiteListHandler.IsWhiteListed(data) {
		log.Trace(
			"MiniblockInterceptorProcessor.Save: not whitelisted miniblocks will not be added in pool",
			"type", miniblock.Type,
			"sender shard", miniblock.SenderShardID,
			"receiver shard", miniblock.ReceiverShardID,
			"hash", hash,
		)
		return nil
	}

	mip.miniblockCache.HasOrAdd(hash, miniblock, miniblock.Size())

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming miniBlocks
func (mip *MiniblockInterceptorProcessor) RegisterHandler(handler func(topic string, hash []byte, data interface{})) {
	if handler == nil {
		return
	}

	mip.mutHandlers.Lock()
	mip.registeredHandlers = append(mip.registeredHandlers, handler)
	mip.mutHandlers.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mip *MiniblockInterceptorProcessor) IsInterfaceNil() bool {
	return mip == nil
}

func (mip *MiniblockInterceptorProcessor) notify(miniBlock *block.MiniBlock, hash []byte, topic string) {
	mip.mutHandlers.RLock()
	for _, handler := range mip.registeredHandlers {
		handler(topic, hash, miniBlock)
	}
	mip.mutHandlers.RUnlock()
}
