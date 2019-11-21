package processor

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

// TxBodyInterceptorProcessor is the processor used when intercepting miniblocks grouped in a block.TxBlockBody structure
type TxBodyInterceptorProcessor struct {
	miniblockCache   storage.Cacher
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
}

// NewTxBodyInterceptorProcessor creates a new TxBodyInterceptorProcessor instance
func NewTxBodyInterceptorProcessor(argument *ArgTxBodyInterceptorProcessor) (*TxBodyInterceptorProcessor, error) {
	if argument == nil {
		return nil, process.ErrNilArguments
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

	return &TxBodyInterceptorProcessor{
		miniblockCache:   argument.MiniblockCache,
		marshalizer:      argument.Marshalizer,
		hasher:           argument.Hasher,
		shardCoordinator: argument.ShardCoordinator,
	}, nil
}

// Validate checks if the intercepted data can be processed
// It returns nil as a body might consist of multiple miniblocks
// Since some might be valid and others not, we rather do the checking when
// we iterate the slice for processing as it is optimal to do so
func (tbip *TxBodyInterceptorProcessor) Validate(data process.InterceptedData) error {
	return nil
}

// Save will save the received miniblocks inside the miniblock cacher after a new validation round
// that will be done on each miniblock
func (tbip *TxBodyInterceptorProcessor) Save(data process.InterceptedData) error {
	interceptedTxBody, ok := data.(*interceptedBlocks.InterceptedTxBlockBody)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	for _, miniblock := range interceptedTxBody.TxBlockBody() {
		err := tbip.processMiniblock(miniblock)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tbip *TxBodyInterceptorProcessor) processMiniblock(miniblock *block.MiniBlock) error {
	err := tbip.checkMiniblock(miniblock)
	if err != nil {
		log.Debug(err.Error())
		return nil
	}

	hash, err := core.CalculateHash(tbip.marshalizer, tbip.hasher, miniblock)
	if err != nil {
		return err
	}

	tbip.miniblockCache.HasOrAdd(hash, miniblock)

	return nil
}

func (tbip *TxBodyInterceptorProcessor) checkMiniblock(miniblock *block.MiniBlock) error {
	//TODO check for whitelisting

	isForCurrentShardRecv := miniblock.ReceiverShardID == tbip.shardCoordinator.SelfId()
	isForCurrentShardSender := miniblock.SenderShardID == tbip.shardCoordinator.SelfId()
	isForCurrentShard := isForCurrentShardRecv || isForCurrentShardSender
	if !isForCurrentShard {
		return process.ErrMiniblockNotForCurrentShard
	}

	return nil
}

// SignalEndOfProcessing signals the end of processing
func (tbip *TxBodyInterceptorProcessor) SignalEndOfProcessing(data []process.InterceptedData) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (tbip *TxBodyInterceptorProcessor) IsInterfaceNil() bool {
	if tbip == nil {
		return true
	}
	return false
}
