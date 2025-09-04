package coordinator

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
)

// BlockDataRequestArgs holds the arguments needed to create a CoordinatorRequest
type BlockDataRequestArgs struct {
	RequestHandler      process.RequestHandler
	MiniBlockPool       storage.Cacher
	PreProcessors       process.PreProcessorsContainer
	ShardCoordinator    process.ShardCoordinator
	EnableEpochsHandler common.EnableEpochsHandler
}

// BlockDataRequest implements the BlockDataRequester interface
type BlockDataRequest struct {
	mutRequestedTxs       sync.RWMutex
	requestedTxs          map[block.Type]int
	requestedItemsHandler process.TimeCacher
	miniBlockPool         storage.Cacher
	preProcessors         process.PreProcessorsContainer
	shardCoordinator      process.ShardCoordinator
	enableEpochsHandler   common.EnableEpochsHandler
	requestHandler        process.RequestHandler
}

// NewBlockDataRequester creates a new instance of BlockDataRequest
func NewBlockDataRequester(args BlockDataRequestArgs) (*BlockDataRequest, error) {
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}
	if check.IfNil(args.MiniBlockPool) {
		return nil, process.ErrNilMiniBlockPool
	}
	if check.IfNil(args.PreProcessors) {
		return nil, process.ErrNilPreProcessorsContainer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return nil, process.ErrNilEnableEpochsHandler
	}

	bdr := &BlockDataRequest{
		requestedTxs:        make(map[block.Type]int),
		preProcessors:       args.PreProcessors,
		miniBlockPool:       args.MiniBlockPool,
		shardCoordinator:    args.ShardCoordinator,
		enableEpochsHandler: args.EnableEpochsHandler,
		requestHandler:      args.RequestHandler,
	}

	bdr.requestedItemsHandler = cache.NewTimeCache(common.MaxWaitingTimeToReceiveRequestedItem)
	bdr.miniBlockPool.RegisterHandler(bdr.receivedMiniBlock, core.UniqueIdentifier())

	return bdr, nil
}

// Reset resets the internal state of BlockDataRequest
func (bdr *BlockDataRequest) Reset() {
	bdr.initRequestedTxs()
	bdr.requestedItemsHandler.Sweep()

	// clean up the preprocessors used only for missing data requests
	preprocessors := bdr.preProcessors.Keys()
	for _, preprocessorType := range preprocessors {
		preprocessor, err := bdr.preProcessors.Get(preprocessorType)
		if err != nil {
			log.Warn("BlockDataRequest.Reset: GetPreProcessor", "type", preprocessorType, "error", err)
			continue
		}
		preprocessor.CreateBlockStarted()
	}
}

// RequestBlockTransactions verifies missing transaction and requests them
func (bdr *BlockDataRequest) RequestBlockTransactions(body *block.Body) {
	if check.IfNil(body) {
		return
	}

	separatedBodies := process.SeparateBodyByType(body)

	bdr.initRequestedTxs()

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc, err := bdr.preProcessors.Get(blockType)
			if err != nil {
				wg.Done()
				return
			}
			requestedTxs := preproc.RequestBlockTransactions(blockBody)

			bdr.mutRequestedTxs.Lock()
			bdr.requestedTxs[blockType] = requestedTxs
			bdr.mutRequestedTxs.Unlock()

			wg.Done()
		}(key, value)
	}

	wg.Wait()
}

// RequestMiniBlocksAndTransactions requests mini blocks and transactions if missing
func (bdr *BlockDataRequest) RequestMiniBlocksAndTransactions(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	finalCrossMiniBlockHashes := bdr.getFinalCrossMiniBlockHashes(header)
	mbsInfo := make([]*data.MiniBlockInfo, 0, len(finalCrossMiniBlockHashes))
	for mbHash, senderShardID := range finalCrossMiniBlockHashes {
		mbsInfo = append(mbsInfo, &data.MiniBlockInfo{
			Hash:          []byte(mbHash),
			SenderShardID: senderShardID,
			Round:         header.GetRound(),
		})
	}

	bdr.requestMissingMiniBlocksAndTransactions(mbsInfo)
}

// GetFinalCrossMiniBlockInfoAndRequestMissing returns the final cross mini block infos and requests missing mini blocks and transactions
func (bdr *BlockDataRequest) GetFinalCrossMiniBlockInfoAndRequestMissing(header data.HeaderHandler) []*data.MiniBlockInfo {
	finalCrossMiniBlockInfos := bdr.getFinalCrossMiniBlockInfos(header.GetOrderedCrossMiniblocksWithDst(bdr.shardCoordinator.SelfId()), header)
	bdr.requestMissingMiniBlocksAndTransactions(finalCrossMiniBlockInfos)

	return finalCrossMiniBlockInfos
}

// IsDataPreparedForProcessing verifies if all the needed data is prepared
func (bdr *BlockDataRequest) IsDataPreparedForProcessing(haveTime func() time.Duration) error {
	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}

	bdr.mutRequestedTxs.RLock()
	wg.Add(len(bdr.requestedTxs))

	for key, value := range bdr.requestedTxs {
		go func(blockType block.Type, requestedTxs int) {
			preproc, err := bdr.preProcessors.Get(blockType)
			if err != nil {
				wg.Done()
				return
			}

			err = preproc.IsDataPrepared(requestedTxs, haveTime)
			if err != nil {
				log.Trace("IsDataPrepared", "error", err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(key, value)
	}

	wg.Wait()
	bdr.mutRequestedTxs.RUnlock()

	return errFound
}

// IsInterfaceNil checks if there is no value under the interface
func (bdr *BlockDataRequest) IsInterfaceNil() bool {
	return bdr == nil
}

// initRequestedTxs init the requested txs number
func (bdr *BlockDataRequest) initRequestedTxs() {
	bdr.mutRequestedTxs.Lock()
	bdr.requestedTxs = make(map[block.Type]int)
	bdr.mutRequestedTxs.Unlock()
}

func (bdr *BlockDataRequest) requestMissingMiniBlocksAndTransactions(mbsInfo []*data.MiniBlockInfo) {
	mapMissingMiniBlocksPerShard := make(map[uint32][][]byte)

	bdr.requestedItemsHandler.Sweep()

	for _, mbInfo := range mbsInfo {
		object, isMiniBlockFound := bdr.miniBlockPool.Peek(mbInfo.Hash)
		if !isMiniBlockFound {
			log.Debug("BlockDataRequest.requestMissingMiniBlocksAndTransactions: mini block not found and was requested",
				"sender shard", mbInfo.SenderShardID,
				"hash", mbInfo.Hash,
				"round", mbInfo.Round,
			)
			mapMissingMiniBlocksPerShard[mbInfo.SenderShardID] = append(mapMissingMiniBlocksPerShard[mbInfo.SenderShardID], mbInfo.Hash)
			_ = bdr.requestedItemsHandler.Add(string(mbInfo.Hash))
			continue
		}

		miniBlock, isMiniBlock := object.(*block.MiniBlock)
		if !isMiniBlock {
			log.Warn("BlockDataRequest.requestMissingMiniBlocksAndTransactions", "mb hash", mbInfo.Hash, "error", process.ErrWrongTypeAssertion)
			continue
		}

		preproc, err := bdr.preProcessors.Get(miniBlock.Type)
		if err != nil {
			log.Warn("BlockDataRequest.requestMissingMiniBlocksAndTransactions: GetPreProcessor", "mb type", miniBlock.Type, "error", err)
			continue
		}

		numTxsRequested := preproc.RequestTransactionsForMiniBlock(miniBlock)
		if numTxsRequested > 0 {
			log.Debug("BlockDataRequest.requestMissingMiniBlocksAndTransactions: RequestTransactionsForMiniBlock", "mb hash", mbInfo.Hash,
				"num txs requested", numTxsRequested)
		}
	}

	for senderShardID, mbsHashes := range mapMissingMiniBlocksPerShard {
		go bdr.requestHandler.RequestMiniBlocks(senderShardID, mbsHashes)
	}
}

func (bdr *BlockDataRequest) receivedMiniBlock(key []byte, value interface{}) {
	if key == nil {
		return
	}

	if !bdr.requestedItemsHandler.Has(string(key)) {
		return
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		log.Warn("BlockDataRequest.receivedMiniBlock", "error", process.ErrWrongTypeAssertion)
		return
	}

	log.Trace("BlockDataRequest.receivedMiniBlock", "hash", key)

	preproc, err := bdr.preProcessors.Get(miniBlock.Type)
	if err != nil {
		log.Warn("BlockDataRequest.receivedMiniBlock",
			"error", fmt.Errorf("%w unknown block type %d", err, miniBlock.Type))
		return
	}

	numTxsRequested := preproc.RequestTransactionsForMiniBlock(miniBlock)
	if numTxsRequested > 0 {
		log.Debug("BlockDataRequest.receivedMiniBlock", "hash", key,
			"num txs requested", numTxsRequested)
	}
}

func (bdr *BlockDataRequest) getFinalCrossMiniBlockHashes(headerHandler data.HeaderHandler) map[string]uint32 {
	if !bdr.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return headerHandler.GetMiniBlockHeadersWithDst(bdr.shardCoordinator.SelfId())
	}
	return process.GetFinalCrossMiniBlockHashes(headerHandler, bdr.shardCoordinator.SelfId())
}

func (bdr *BlockDataRequest) getFinalCrossMiniBlockInfos(
	crossMiniBlockInfos []*data.MiniBlockInfo,
	header data.HeaderHandler,
) []*data.MiniBlockInfo {

	if !bdr.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return crossMiniBlockInfos
	}

	miniBlockInfos := make([]*data.MiniBlockInfo, 0)
	for _, crossMiniBlockInfo := range crossMiniBlockInfos {
		miniBlockHeader := process.GetMiniBlockHeaderWithHash(header, crossMiniBlockInfo.Hash)
		if miniBlockHeader != nil && !miniBlockHeader.IsFinal() {
			log.Debug("BlockDataRequest.getFinalCrossMiniBlockInfos: do not execute mini block which is not final", "mb hash", miniBlockHeader.GetHash())
			continue
		}

		miniBlockInfos = append(miniBlockInfos, crossMiniBlockInfo)
	}

	return miniBlockInfos
}
