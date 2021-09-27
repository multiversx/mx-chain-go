package coordinator

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
)

var _ process.TransactionCoordinator = (*transactionCoordinator)(nil)

var log = logger.GetOrCreate("process/coordinator")

// ArgTransactionCoordinator holds all dependencies required by the transaction coordinator factory in order to create new instances
type ArgTransactionCoordinator struct {
	Hasher                            hashing.Hasher
	Marshalizer                       marshal.Marshalizer
	ShardCoordinator                  sharding.Coordinator
	Accounts                          state.AccountsAdapter
	MiniBlockPool                     storage.Cacher
	RequestHandler                    process.RequestHandler
	PreProcessors                     process.PreProcessorsContainer
	InterProcessors                   process.IntermediateProcessorContainer
	GasHandler                        process.GasHandler
	FeeHandler                        process.TransactionFeeHandler
	BlockSizeComputation              preprocess.BlockSizeComputationHandler
	BalanceComputation                preprocess.BalanceComputationHandler
	EconomicsFee                      process.FeeHandler
	TxTypeHandler                     process.TxTypeHandler
	TransactionsLogProcessor          process.TransactionLogProcessor
	BlockGasAndFeesReCheckEnableEpoch uint32
	EpochNotifier                     process.EpochNotifier
	PostProcessorTxsHandler           process.PostProcessorTxsHandler
	MixedTxsInMiniBlocksEnableEpoch   uint32
	ScheduledMiniBlocksEnableEpoch    uint32
}

type transactionCoordinator struct {
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
	miniBlockPool    storage.Cacher
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer

	mutPreProcessor sync.RWMutex
	txPreProcessors map[block.Type]process.PreProcessor
	keysTxPreProcs  []block.Type

	mutInterimProcessors sync.RWMutex
	interimProcessors    map[block.Type]process.IntermediateTransactionHandler
	keysInterimProcs     []block.Type

	mutRequestedTxs sync.RWMutex
	requestedTxs    map[block.Type]int

	onRequestMiniBlock                func(shardId uint32, mbHash []byte)
	gasHandler                        process.GasHandler
	feeHandler                        process.TransactionFeeHandler
	blockSizeComputation              preprocess.BlockSizeComputationHandler
	balanceComputation                preprocess.BalanceComputationHandler
	requestedItemsHandler             process.TimeCacher
	economicsFee                      process.FeeHandler
	txTypeHandler                     process.TxTypeHandler
	transactionsLogProcessor          process.TransactionLogProcessor
	blockGasAndFeesReCheckEnableEpoch uint32
	mixedTxsInMiniBlocksEnableEpoch   uint32
	scheduledMiniBlocksEnableEpoch    uint32
	epochNotifier                     process.EpochNotifier
	postProcessorTxsHandler           process.PostProcessorTxsHandler
	flagMixedTxsInMiniBlocks          atomic.Flag
	flagScheduledMiniBlocks           atomic.Flag
}

// NewTransactionCoordinator creates a transaction coordinator to run and coordinate preprocessors and processors
func NewTransactionCoordinator(arguments ArgTransactionCoordinator) (*transactionCoordinator, error) {
	err := checkTransactionCoordinatorNilParameters(arguments)
	if err != nil {
		return nil, err
	}

	tc := &transactionCoordinator{
		shardCoordinator:                  arguments.ShardCoordinator,
		accounts:                          arguments.Accounts,
		gasHandler:                        arguments.GasHandler,
		hasher:                            arguments.Hasher,
		marshalizer:                       arguments.Marshalizer,
		feeHandler:                        arguments.FeeHandler,
		blockSizeComputation:              arguments.BlockSizeComputation,
		balanceComputation:                arguments.BalanceComputation,
		economicsFee:                      arguments.EconomicsFee,
		txTypeHandler:                     arguments.TxTypeHandler,
		blockGasAndFeesReCheckEnableEpoch: arguments.BlockGasAndFeesReCheckEnableEpoch,
		transactionsLogProcessor:          arguments.TransactionsLogProcessor,
		epochNotifier:                     arguments.EpochNotifier,
		postProcessorTxsHandler:           arguments.PostProcessorTxsHandler,
		mixedTxsInMiniBlocksEnableEpoch:   arguments.MixedTxsInMiniBlocksEnableEpoch,
		scheduledMiniBlocksEnableEpoch:    arguments.ScheduledMiniBlocksEnableEpoch,
	}
	log.Debug("coordinator/process: enable epoch for block gas and fees re-check", "epoch", tc.blockGasAndFeesReCheckEnableEpoch)
	log.Debug("coordinator/process: enable epoch for mixed txs in mini blocks", "epoch", tc.mixedTxsInMiniBlocksEnableEpoch)
	log.Debug("coordinator/process: enable epoch for scheduled mini blocks", "epoch", tc.scheduledMiniBlocksEnableEpoch)

	tc.miniBlockPool = arguments.MiniBlockPool
	tc.onRequestMiniBlock = arguments.RequestHandler.RequestMiniBlock
	tc.requestedTxs = make(map[block.Type]int)
	tc.txPreProcessors = make(map[block.Type]process.PreProcessor)
	tc.interimProcessors = make(map[block.Type]process.IntermediateTransactionHandler)

	tc.keysTxPreProcs = arguments.PreProcessors.Keys()
	sort.Slice(tc.keysTxPreProcs, func(i, j int) bool {
		return tc.keysTxPreProcs[i] < tc.keysTxPreProcs[j]
	})
	for _, value := range tc.keysTxPreProcs {
		preProc, err := arguments.PreProcessors.Get(value)
		if err != nil {
			return nil, err
		}
		tc.txPreProcessors[value] = preProc
	}

	tc.keysInterimProcs = arguments.InterProcessors.Keys()
	sort.Slice(tc.keysInterimProcs, func(i, j int) bool {
		return tc.keysInterimProcs[i] < tc.keysInterimProcs[j]
	})
	for _, value := range tc.keysInterimProcs {
		interProc, err := arguments.InterProcessors.Get(value)
		if err != nil {
			return nil, err
		}
		tc.interimProcessors[value] = interProc
	}

	tc.requestedItemsHandler = timecache.NewTimeCache(common.MaxWaitingTimeToReceiveRequestedItem)
	tc.miniBlockPool.RegisterHandler(tc.receivedMiniBlock, core.UniqueIdentifier())

	tc.epochNotifier.RegisterNotifyHandler(tc)

	return tc, nil
}

// separateBodyByType creates a map of bodies according to type
func (tc *transactionCoordinator) separateBodyByType(body *block.Body, withMixedMiniBlocksSplit bool) (map[block.Type]*block.Body, error) {
	var splitAndCompactedMiniBlocks block.MiniBlockSlice
	var err error

	mapSeparatedBodies := make(map[block.Type]*block.Body)
	for _, miniBlock := range body.MiniBlocks {
		if withMixedMiniBlocksSplit {
			splitAndCompactedMiniBlocks, err = tc.splitAndCompactMiniBlockBasedOnTxsType(miniBlock)
			if err != nil {
				return nil, err
			}
		} else {
			splitAndCompactedMiniBlocks = block.MiniBlockSlice{miniBlock}
		}

		createSeparateBodiesFromMiniBlocks(mapSeparatedBodies, splitAndCompactedMiniBlocks)
	}

	return mapSeparatedBodies, nil
}

func createSeparateBodiesFromMiniBlocks(mapSeparatedBodies map[block.Type]*block.Body, miniBlocks block.MiniBlockSlice) {
	for _, mb := range miniBlocks {
		blockType := mb.Type
		if blockType == block.InvalidBlock {
			blockType = block.TxBlock
		}

		if _, ok := mapSeparatedBodies[blockType]; !ok {
			mapSeparatedBodies[blockType] = &block.Body{}
		}

		mapSeparatedBodies[blockType].MiniBlocks = append(mapSeparatedBodies[blockType].MiniBlocks, mb)
	}
}

// initRequestedTxs init the requested txs number
func (tc *transactionCoordinator) initRequestedTxs() {
	tc.mutRequestedTxs.Lock()
	tc.requestedTxs = make(map[block.Type]int)
	tc.mutRequestedTxs.Unlock()
}

// RequestBlockTransactions verifies missing transaction and requests them
func (tc *transactionCoordinator) RequestBlockTransactions(body *block.Body) {
	if check.IfNil(body) {
		return
	}

	separatedBodies, err := tc.separateBodyByType(body, true)
	if err != nil {
		log.Error("RequestBlockTransactions", "error", err.Error())
		return
	}

	tc.initRequestedTxs()

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			requestedTxs := preproc.RequestBlockTransactions(blockBody)

			tc.mutRequestedTxs.Lock()
			tc.requestedTxs[blockType] = requestedTxs
			tc.mutRequestedTxs.Unlock()

			wg.Done()
		}(key, value)
	}

	wg.Wait()
}

// IsDataPreparedForProcessing verifies if all the needed data is prepared
func (tc *transactionCoordinator) IsDataPreparedForProcessing(haveTime func() time.Duration) error {
	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}

	tc.mutRequestedTxs.RLock()
	wg.Add(len(tc.requestedTxs))

	for key, value := range tc.requestedTxs {
		go func(blockType block.Type, requestedTxs int) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			err := preproc.IsDataPrepared(requestedTxs, haveTime)
			if err != nil {
				log.Trace("IsDataPrepared", "error", err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(key, value)
	}

	tc.mutRequestedTxs.RUnlock()
	wg.Wait()

	return errFound
}

// SaveTxsToStorage saves transactions from block body into storage units
func (tc *transactionCoordinator) SaveTxsToStorage(body *block.Body) error {
	if check.IfNil(body) {
		return nil
	}

	separatedBodies, err := tc.separateBodyByType(body, true)
	if err != nil {
		return err
	}

	for key, value := range separatedBodies {
		err := tc.saveTxsToStorage(key, value)
		if err != nil {
			return err
		}
	}

	for _, blockType := range tc.keysInterimProcs {
		err := tc.saveCurrentIntermediateTxToStorage(blockType)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *transactionCoordinator) saveTxsToStorage(blockType block.Type, blockBody *block.Body) error {
	preproc := tc.getPreProcessor(blockType)
	if check.IfNil(preproc) {
		return nil
	}

	err := preproc.SaveTxsToStorage(blockBody)
	if err != nil {
		log.Trace("SaveTxsToStorage", "error", err.Error())

		return err
	}

	return nil
}

func (tc *transactionCoordinator) saveCurrentIntermediateTxToStorage(blockType block.Type) error {
	intermediateProc := tc.getInterimProcessor(blockType)
	if check.IfNil(intermediateProc) {
		return nil
	}

	err := intermediateProc.SaveCurrentIntermediateTxToStorage()
	if err != nil {
		log.Trace("SaveCurrentIntermediateTxToStorage", "error", err.Error())
		return err
	}

	return nil
}

// RestoreBlockDataFromStorage restores block data from storage to pool
func (tc *transactionCoordinator) RestoreBlockDataFromStorage(body *block.Body) (int, error) {
	if check.IfNil(body) {
		return 0, nil
	}

	totalRestoredTxs, err := tc.restoreTxsIntoPool(body)
	if err != nil {
		return totalRestoredTxs, err
	}

	separatedBodies, err := tc.separateBodyByType(body, false)
	if err != nil {
		return totalRestoredTxs, err
	}

	var errFound error
	localMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			errRestore := preproc.RestoreMiniBlocksIntoPools(blockBody, tc.miniBlockPool)
			if errRestore != nil {
				log.Trace("RestoreMiniBlocksIntoPools", "error", errRestore.Error())

				localMutex.Lock()
				errFound = errRestore
				localMutex.Unlock()
			}

			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return totalRestoredTxs, errFound
}

func (tc *transactionCoordinator) restoreTxsIntoPool(body *block.Body) (int, error) {
	if check.IfNil(body) {
		return 0, nil
	}

	separatedBodies, err := tc.separateBodyByType(body, true)
	if err != nil {
		return 0, err
	}

	var errFound error
	localMutex := sync.Mutex{}
	totalRestoredTxs := 0

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			restoredTxs, errRestore := preproc.RestoreTxsIntoPools(blockBody)
			if errRestore != nil {
				log.Trace("RestoreTxsIntoPools", "error", errRestore.Error())

				localMutex.Lock()
				errFound = errRestore
				localMutex.Unlock()
			}

			localMutex.Lock()
			totalRestoredTxs += restoredTxs
			localMutex.Unlock()

			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return totalRestoredTxs, errFound
}

// RemoveBlockDataFromPool deletes block data from pools
func (tc *transactionCoordinator) RemoveBlockDataFromPool(body *block.Body) error {
	if check.IfNil(body) {
		return nil
	}

	err := tc.RemoveTxsFromPool(body)
	if err != nil {
		return err
	}

	separatedBodies, err := tc.separateBodyByType(body, false)
	if err != nil {
		return err
	}

	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			err := preproc.RemoveMiniBlocksFromPools(blockBody, tc.miniBlockPool)
			if err != nil {
				log.Trace("RemoveMiniBlocksFromPools", "error", err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return errFound
}

// RemoveTxsFromPool deletes txs from pools
func (tc *transactionCoordinator) RemoveTxsFromPool(body *block.Body) error {
	if check.IfNil(body) {
		return nil
	}

	separatedBodies, err := tc.separateBodyByType(body, true)
	if err != nil {
		return err
	}

	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			err := preproc.RemoveTxsFromPools(blockBody)
			if err != nil {
				log.Trace("RemoveTxsFromPools", "error", err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return errFound
}

// ProcessBlockTransaction processes transactions and updates state tries
func (tc *transactionCoordinator) ProcessBlockTransaction(
	header data.HeaderHandler,
	body *block.Body,
	timeRemaining func() time.Duration,
) error {
	if check.IfNil(body) {
		return process.ErrNilBlockBody
	}

	if tc.isMaxBlockSizeReached(body) {
		return process.ErrMaxBlockSizeReached
	}

	haveTime := func() bool {
		return timeRemaining() >= 0
	}

	for _, miniBlock := range body.MiniBlocks {
		log.Trace("ProcessBlockTransaction: miniblock",
			"sender shard", miniBlock.SenderShardID,
			"receiver shard", miniBlock.ReceiverShardID,
			"type", miniBlock.Type,
			"num txs", len(miniBlock.TxHashes))
	}

	startTime := time.Now()
	mbIndex, err := tc.processMiniBlocksToMe(header, body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to processMiniBlocksToMe",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	if mbIndex == len(body.MiniBlocks) {
		return nil
	}

	miniBlocksFromMe := body.MiniBlocks[mbIndex:]
	startTime = time.Now()
	err = tc.processMiniBlocksFromMe(header, &block.Body{MiniBlocks: miniBlocksFromMe}, haveTime)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to processMiniBlocksFromMe",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return err
	}

	return nil
}

func (tc *transactionCoordinator) processMiniBlocksFromMe(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) error {
	for _, mb := range body.MiniBlocks {
		if mb.SenderShardID != tc.shardCoordinator.SelfId() {
			return process.ErrMiniBlocksInWrongOrder
		}
	}

	separatedBodies, err := tc.separateBodyByType(body, true)
	if err != nil {
		return err
	}

	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	for _, blockType := range tc.keysTxPreProcs {
		if separatedBodies[blockType] == nil {
			continue
		}

		preProc := tc.getPreProcessor(blockType)
		if check.IfNil(preProc) {
			return process.ErrMissingPreProcessor
		}

		err := preProc.ProcessBlockTransactions(header, separatedBodies[blockType], haveTime, false, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *transactionCoordinator) processMiniBlocksToMe(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) (int, error) {
	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	// processing destination ME miniblocks first
	mbIndex := 0
	for mbIndex = 0; mbIndex < len(body.MiniBlocks); mbIndex++ {
		miniBlock := body.MiniBlocks[mbIndex]
		if miniBlock.SenderShardID == tc.shardCoordinator.SelfId() {
			return mbIndex, nil
		}

		err := tc.processMiniBlock(header, haveTime, miniBlock)
		if err != nil {
			return mbIndex, err
		}
	}

	return mbIndex, nil
}

func (tc *transactionCoordinator) processMiniBlock(
	header data.HeaderHandler,
	haveTime func() bool,
	miniBlock *block.MiniBlock,
) error {

	var err error
	var miniBlocks block.MiniBlockSlice
	scheduledMode := false
	if tc.flagScheduledMiniBlocks.IsSet() {
		scheduledMode, err = process.IsScheduledMode(header, &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock}}, tc.hasher, tc.marshalizer)
		if err != nil {
			return err
		}
	}

	var totalGasConsumed uint64
	if scheduledMode {
		totalGasConsumed = tc.gasHandler.TotalGasConsumedAsScheduled()
	} else {
		totalGasConsumed = tc.gasHandler.TotalGasConsumed()
	}

	gasConsumedInfo := &process.GasConsumedInfo{
		GasConsumedByMiniBlockInReceiverShard: uint64(0),
		GasConsumedByMiniBlocksInSenderShard:  uint64(0),
		TotalGasConsumedInSelfShard:           totalGasConsumed,
	}

	miniBlocks, err = tc.splitMiniBlockBasedOnTxsType(miniBlock)
	if err != nil {
		return err
	}

	for _, mb := range miniBlocks {
		preProc := tc.getPreProcessor(mb.Type)
		if check.IfNil(preProc) {
			return process.ErrMissingPreProcessor
		}

		err = preProc.ProcessBlockTransactions(header, &block.Body{MiniBlocks: []*block.MiniBlock{mb}}, haveTime, scheduledMode, gasConsumedInfo)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateMbsAndProcessCrossShardTransactionsDstMe creates miniblocks and processes cross shard transaction
// with destination of current shard
func (tc *transactionCoordinator) CreateMbsAndProcessCrossShardTransactionsDstMe(
	hdr data.HeaderHandler,
	processedMiniBlocksHashes map[string]struct{},
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
) (block.MiniBlockSlice, uint32, bool, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	nrTxAdded := uint32(0)
	nrMiniBlocksProcessed := 0

	if check.IfNil(hdr) {
		return miniBlocks, nrTxAdded, false, nil
	}

	shouldSkipShard := make(map[uint32]bool)

	crossMiniBlockInfos := hdr.GetOrderedCrossMiniblocksWithDst(tc.shardCoordinator.SelfId())
	for _, miniBlockInfo := range crossMiniBlockInfos {
		if !haveTime() && !haveAdditionalTime() {
			log.Debug("CreateMbsAndProcessCrossShardTransactionsDstMe",
				"scheduled mode", scheduledMode,
				"stop creating", "time is out")
			break
		}

		if tc.blockSizeComputation.IsMaxBlockSizeReached(0, 0) {
			log.Debug("CreateMbsAndProcessCrossShardTransactionsDstMe",
				"scheduled mode", scheduledMode,
				"stop creating", "max block size has been reached")
			break
		}

		if shouldSkipShard[miniBlockInfo.SenderShardID] {
			log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: should skip shard",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		_, ok := processedMiniBlocksHashes[string(miniBlockInfo.Hash)]
		if ok {
			nrMiniBlocksProcessed++
			log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: mini block already processed",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		miniVal, _ := tc.miniBlockPool.Peek(miniBlockInfo.Hash)
		if miniVal == nil {
			go tc.onRequestMiniBlock(miniBlockInfo.SenderShardID, miniBlockInfo.Hash)
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: mini block not found and was requested",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		miniBlock, ok := miniVal.(*block.MiniBlock)
		if !ok {
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: mini block assertion type failed",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		//if scheduledMode && !miniBlock.IsScheduledMiniBlock() {
		//	shouldSkipShard[miniBlockInfo.SenderShardID] = true
		//	log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: mini block is not scheduled",
		//		"scheduled mode", scheduledMode,
		//		"sender shard", miniBlockInfo.SenderShardID,
		//		"hash", miniBlockInfo.Hash,
		//		"round", miniBlockInfo.Round,
		//	)
		//	continue
		//}

		shouldSkip, err := tc.requestMissingTxsAndProcessMiniBlock(
			miniBlock,
			miniBlockInfo,
			haveTime,
			haveAdditionalTime,
			scheduledMode,
			shouldSkipShard,
		)
		if err != nil {
			return nil, 0, false, err
		}
		if shouldSkip {
			continue
		}

		// all txs processed, add to processed miniblocks
		miniBlocks = append(miniBlocks, miniBlock)
		nrTxAdded = nrTxAdded + uint32(len(miniBlock.TxHashes))
		nrMiniBlocksProcessed++
		if processedMiniBlocksHashes != nil {
			processedMiniBlocksHashes[string(miniBlockInfo.Hash)] = struct{}{}
		}
	}

	allMBsProcessed := nrMiniBlocksProcessed == len(crossMiniBlockInfos)

	return miniBlocks, nrTxAdded, allMBsProcessed, nil
}

func (tc *transactionCoordinator) requestMissingTxsAndProcessMiniBlock(
	miniBlock *block.MiniBlock,
	miniBlockInfo *data.MiniBlockInfo,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
	shouldSkipShard map[uint32]bool,
) (bool, error) {

	splitAndCompactedMiniBlocks, err := tc.splitAndCompactMiniBlockBasedOnTxsType(miniBlock)
	if err != nil {
		return true, err
	}

	numTxsRequested := 0
	for _, mb := range splitAndCompactedMiniBlocks {
		preproc := tc.getPreProcessor(mb.Type)
		if check.IfNil(preproc) {
			return true, fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, mb.Type)
		}

		numTxsRequested += preproc.RequestTransactionsForMiniBlock(mb)
	}

	if numTxsRequested > 0 {
		shouldSkipShard[miniBlockInfo.SenderShardID] = true
		log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: transactions not found and were requested",
			"scheduled mode", scheduledMode,
			"sender shard", miniBlockInfo.SenderShardID,
			"hash", miniBlockInfo.Hash,
			"round", miniBlockInfo.Round,
			"num txs requested", numTxsRequested,
		)
		return true, nil
	}

	err = tc.processCompleteMiniBlock(miniBlock, miniBlockInfo.Hash, haveTime, haveAdditionalTime, scheduledMode)
	if err != nil {
		shouldSkipShard[miniBlockInfo.SenderShardID] = true
		log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: processed complete mini block failed",
			"scheduled mode", scheduledMode,
			"sender shard", miniBlockInfo.SenderShardID,
			"hash", miniBlockInfo.Hash,
			"round", miniBlockInfo.Round,
		)
		return true, nil
	}

	return false, nil
}

func (tc *transactionCoordinator) splitAndCompactMiniBlockBasedOnTxsType(miniBlock *block.MiniBlock) (block.MiniBlockSlice, error) {
	if !tc.flagMixedTxsInMiniBlocks.IsSet() {
		return block.MiniBlockSlice{miniBlock}, nil
	}

	shouldSkipSplit := miniBlock.Type != block.TxBlock
	if shouldSkipSplit {
		return block.MiniBlockSlice{miniBlock}, nil
	}

	mapTxTypeTxHashes, err := getMapTxTypeTxHashes(miniBlock)
	if err != nil {
		return nil, err
	}

	return tc.splitAndCompactMiniBlock(miniBlock, mapTxTypeTxHashes)
}

func getMapTxTypeTxHashes(miniBlock *block.MiniBlock) (map[block.Type][][]byte, error) {
	txsType, err := miniBlock.GetTxsTypeFromMiniBlock()
	if err != nil {
		return nil, err
	}

	mapTxTypeTxHashes := make(map[block.Type][][]byte)
	for index, txHash := range miniBlock.TxHashes {
		txType := txsType[index]
		_, ok := mapTxTypeTxHashes[txType]
		if !ok {
			mapTxTypeTxHashes[txType] = make([][]byte, 0)
		}

		mapTxTypeTxHashes[txType] = append(mapTxTypeTxHashes[txType], txHash)
	}

	return mapTxTypeTxHashes, nil
}

func (tc *transactionCoordinator) splitAndCompactMiniBlock(
	miniBlock *block.MiniBlock,
	mapTxTypeTxHashes map[block.Type][][]byte,
) (block.MiniBlockSlice, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)

	mbr, err := miniBlock.GetMiniBlockReserved()
	if err != nil {
		return nil, err
	}
	if mbr != nil {
		mbr.TransactionsType = nil
	}

	for _, mbType := range tc.keysTxPreProcs {
		txHashes, ok := mapTxTypeTxHashes[mbType]
		if !ok {
			continue
		}

		mb := &block.MiniBlock{
			TxHashes:        txHashes,
			ReceiverShardID: miniBlock.ReceiverShardID,
			SenderShardID:   miniBlock.SenderShardID,
			Type:            mbType,
		}

		err = mb.SetMiniBlockReserved(mbr)
		if err != nil {
			return nil, err
		}

		miniBlocks = append(miniBlocks, mb)
	}

	return miniBlocks, nil
}

// splitMiniBlockBasedOnTxsType splits the given mini block in different small mini blocks based on tx type,
// preserving the transactions order in the original mini block
func (tc *transactionCoordinator) splitMiniBlockBasedOnTxsType(miniBlock *block.MiniBlock) (block.MiniBlockSlice, error) {
	if !tc.flagMixedTxsInMiniBlocks.IsSet() {
		return block.MiniBlockSlice{miniBlock}, nil
	}

	shouldSkipSplit := miniBlock.Type != block.TxBlock
	if shouldSkipSplit {
		return block.MiniBlockSlice{miniBlock}, nil
	}

	return splitMiniBlock(miniBlock)
}

func splitMiniBlock(miniBlock *block.MiniBlock) (block.MiniBlockSlice, error) {
	var lastTxType block.Type
	var splitMiniBlock *block.MiniBlock

	txsType, err := miniBlock.GetTxsTypeFromMiniBlock()
	if err != nil {
		return nil, err
	}

	mbr, err := miniBlock.GetMiniBlockReserved()
	if err != nil {
		return nil, err
	}
	if mbr != nil {
		mbr.TransactionsType = nil
	}

	splitMiniBlocks := make(block.MiniBlockSlice, 0)
	for i := 0; i < len(miniBlock.TxHashes); i++ {
		shouldCreateNewMiniBlock := i == 0 || txsType[i] != lastTxType
		if shouldCreateNewMiniBlock {
			lastTxType = txsType[i]
			if splitMiniBlock != nil {
				splitMiniBlocks = append(splitMiniBlocks, splitMiniBlock)
			}

			splitMiniBlock, err = createNewMiniBlock(miniBlock, txsType[i], mbr)
			if err != nil {
				return nil, err
			}
		}

		splitMiniBlock.TxHashes = append(splitMiniBlock.TxHashes, miniBlock.TxHashes[i])
	}

	splitMiniBlocks = append(splitMiniBlocks, splitMiniBlock)
	return splitMiniBlocks, nil
}

func createNewMiniBlock(mb *block.MiniBlock, mbType block.Type, mbr *block.MiniBlockReserved) (*block.MiniBlock, error) {
	splitMiniBlock := &block.MiniBlock{
		TxHashes:        make([][]byte, 0),
		ReceiverShardID: mb.ReceiverShardID,
		SenderShardID:   mb.SenderShardID,
		Type:            mbType,
	}

	err := splitMiniBlock.SetMiniBlockReserved(mbr)
	if err != nil {
		return nil, err
	}

	return splitMiniBlock, nil
}

// CreateMbsAndProcessTransactionsFromMe creates miniblocks and processes transactions from pool
func (tc *transactionCoordinator) CreateMbsAndProcessTransactionsFromMe(
	haveTime func() bool,
) block.MiniBlockSlice {

	miniBlocks := make(block.MiniBlockSlice, 0)
	for _, blockType := range tc.keysTxPreProcs {
		txPreProc := tc.getPreProcessor(blockType)
		if check.IfNil(txPreProc) {
			return nil
		}

		mbs, err := txPreProc.CreateAndProcessMiniBlocks(haveTime)
		if err != nil {
			log.Debug("CreateAndProcessMiniBlocks", "error", err.Error())
		}

		if len(mbs) > 0 {
			miniBlocks = append(miniBlocks, mbs...)
		}
	}

	interMBs := tc.createPostProcessMiniBlocks()
	if len(interMBs) > 0 {
		miniBlocks = append(miniBlocks, interMBs...)
	}

	return miniBlocks
}

func (tc *transactionCoordinator) createPostProcessMiniBlocks() block.MiniBlockSlice {
	miniBlocks := tc.CreatePostProcessMiniBlocks()
	if !tc.flagMixedTxsInMiniBlocks.IsSet() {
		return miniBlocks
	}

	finalMiniBlocks := make(block.MiniBlockSlice, 0)
	for i := 0; i < len(miniBlocks); i++ {
		txHashes := make([][]byte, 0)
		for _, txHash := range miniBlocks[i].TxHashes {
			if tc.postProcessorTxsHandler.IsPostProcessorTxAdded(txHash) {
				continue
			}
			txHashes = append(txHashes, txHash)
		}

		if len(txHashes) == 0 {
			continue
		}

		miniBlocks[i].TxHashes = txHashes
		finalMiniBlocks = append(finalMiniBlocks, miniBlocks[i])
	}

	return finalMiniBlocks
}

// CreatePostProcessMiniBlocks returns all the post processed miniblocks
func (tc *transactionCoordinator) CreatePostProcessMiniBlocks() block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0)

	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	for _, blockType := range tc.keysInterimProcs {
		interimProc := tc.getInterimProcessor(blockType)
		if check.IfNil(interimProc) {
			continue
		}

		currMbs := interimProc.CreateAllInterMiniBlocks()
		for _, value := range currMbs {
			miniBlocks = append(miniBlocks, value)
		}
	}

	return miniBlocks
}

// CreateBlockStarted initializes necessary data for preprocessors at block create or block process
func (tc *transactionCoordinator) CreateBlockStarted() {
	tc.gasHandler.Init()
	tc.blockSizeComputation.Init()
	tc.balanceComputation.Init()

	tc.mutPreProcessor.RLock()
	for _, value := range tc.txPreProcessors {
		value.CreateBlockStarted()
	}
	tc.mutPreProcessor.RUnlock()

	tc.mutInterimProcessors.RLock()
	for _, value := range tc.interimProcessors {
		value.CreateBlockStarted()
	}
	tc.mutInterimProcessors.RUnlock()

	tc.transactionsLogProcessor.Clean()
}

func (tc *transactionCoordinator) getPreProcessor(blockType block.Type) process.PreProcessor {
	tc.mutPreProcessor.RLock()
	preprocessor, exists := tc.txPreProcessors[blockType]
	tc.mutPreProcessor.RUnlock()

	if !exists {
		return nil
	}

	return preprocessor
}

func (tc *transactionCoordinator) getInterimProcessor(blockType block.Type) process.IntermediateTransactionHandler {
	tc.mutInterimProcessors.RLock()
	interProcessor, exists := tc.interimProcessors[blockType]
	tc.mutInterimProcessors.RUnlock()

	if !exists {
		return nil
	}

	return interProcessor
}

func createBroadcastTopic(shardC sharding.Coordinator, destShId uint32, mbType block.Type) (string, error) {
	var baseTopic string

	switch mbType {
	case block.TxBlock:
		baseTopic = factory.TransactionTopic
	case block.PeerBlock:
		baseTopic = factory.PeerChBodyTopic
	case block.SmartContractResultBlock:
		baseTopic = factory.UnsignedTransactionTopic
	case block.RewardsBlock:
		baseTopic = factory.RewardsTransactionTopic
	default:
		return "", process.ErrUnknownBlockType
	}

	transactionTopic := baseTopic +
		shardC.CommunicationIdentifier(destShId)

	return transactionTopic, nil
}

// CreateMarshalizedData creates marshalized data for broadcasting
func (tc *transactionCoordinator) CreateMarshalizedData(body *block.Body) map[string][][]byte {
	mrsTxs := make(map[string][][]byte)

	if check.IfNil(body) {
		return mrsTxs
	}

	for i := 0; i < len(body.MiniBlocks); i++ {
		miniBlock := body.MiniBlocks[i]
		if miniBlock.SenderShardID != tc.shardCoordinator.SelfId() ||
			miniBlock.ReceiverShardID == tc.shardCoordinator.SelfId() {
			continue
		}

		broadcastTopic, err := createBroadcastTopic(tc.shardCoordinator, miniBlock.ReceiverShardID, miniBlock.Type)
		if err != nil {
			log.Warn("CreateMarshalizedData.createBroadcastTopic", "error", err.Error())
			continue
		}

		isPreProcessMiniBlock := miniBlock.Type == block.TxBlock
		preproc := tc.getPreProcessor(miniBlock.Type)
		if !check.IfNil(preproc) && isPreProcessMiniBlock {
			dataMarshalizer, ok := preproc.(process.DataMarshalizer)
			if ok {
				// preproc supports marshalizing items
				tc.appendMarshalizedItems(
					dataMarshalizer,
					miniBlock.TxHashes,
					mrsTxs,
					broadcastTopic,
				)
			}
		}

		interimProc := tc.getInterimProcessor(miniBlock.Type)
		if !check.IfNil(interimProc) && !isPreProcessMiniBlock {
			dataMarshalizer, ok := interimProc.(process.DataMarshalizer)
			if ok {
				// interimProc supports marshalizing items
				tc.appendMarshalizedItems(
					dataMarshalizer,
					miniBlock.TxHashes,
					mrsTxs,
					broadcastTopic,
				)
			}
		}
	}

	return mrsTxs
}

func (tc *transactionCoordinator) appendMarshalizedItems(
	dataMarshalizer process.DataMarshalizer,
	txHashes [][]byte,
	mrsTxs map[string][][]byte,
	broadcastTopic string,
) {
	currMrsTxs, err := dataMarshalizer.CreateMarshalizedData(txHashes)
	if err != nil {
		log.Debug("appendMarshalizedItems.CreateMarshalizedData", "error", err.Error())
		return
	}

	if len(currMrsTxs) > 0 {
		mrsTxs[broadcastTopic] = append(mrsTxs[broadcastTopic], currMrsTxs...)
	}
}

// GetAllCurrentUsedTxs returns the cached transaction data for current round
func (tc *transactionCoordinator) GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler {
	txPool := make(map[string]data.TransactionHandler)
	interTxPool := make(map[string]data.TransactionHandler)

	preProc := tc.getPreProcessor(blockType)
	if preProc != nil {
		txPool = preProc.GetAllCurrentUsedTxs()
	}

	interProc := tc.getInterimProcessor(blockType)
	if interProc != nil {
		interTxPool = interProc.GetAllCurrentFinishedTxs()
	}

	for hash, tx := range interTxPool {
		txPool[hash] = tx
	}

	return txPool
}

// GetAllCurrentLogs return the cached logs data from current round
func (tc *transactionCoordinator) GetAllCurrentLogs() map[string]data.LogHandler {
	return tc.transactionsLogProcessor.GetAllCurrentLogs()
}

// RequestMiniBlocks request miniblocks if missing
func (tc *transactionCoordinator) RequestMiniBlocks(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	tc.requestedItemsHandler.Sweep()

	crossMiniBlockHashes := header.GetMiniBlockHeadersWithDst(tc.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		obj, _ := tc.miniBlockPool.Peek([]byte(key))
		if obj == nil {
			go tc.onRequestMiniBlock(senderShardId, []byte(key))
			_ = tc.requestedItemsHandler.Add(key)
		}
	}
}

func (tc *transactionCoordinator) receivedMiniBlock(key []byte, value interface{}) {
	if key == nil {
		return
	}

	if !tc.requestedItemsHandler.Has(string(key)) {
		return
	}

	miniBlock, ok := value.(*block.MiniBlock)
	if !ok {
		log.Warn("transactionCoordinator.receivedMiniBlock", "error", process.ErrWrongTypeAssertion)
		return
	}

	log.Trace("transactionCoordinator.receivedMiniBlock", "hash", key)

	splitAndCompactedMiniBlocks, err := tc.splitAndCompactMiniBlockBasedOnTxsType(miniBlock)
	if err != nil {
		log.Warn("transactionCoordinator.receivedMiniBlock", "error", err)
		return
	}

	numTxsRequested := 0
	for _, mb := range splitAndCompactedMiniBlocks {
		preproc := tc.getPreProcessor(mb.Type)
		if check.IfNil(preproc) {
			log.Warn("transactionCoordinator.receivedMiniBlock",
				"error", fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, mb.Type))
			return
		}

		numTxsRequested += preproc.RequestTransactionsForMiniBlock(mb)
	}

	if numTxsRequested > 0 {
		log.Trace("transactionCoordinator.receivedMiniBlock", "hash", key, "num txs requested", numTxsRequested)
	}
}

// processMiniBlockComplete - all transactions must be processed together, otherwise error
func (tc *transactionCoordinator) processCompleteMiniBlock(
	miniBlock *block.MiniBlock,
	miniBlockHash []byte,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
) error {

	snapshot := tc.accounts.JournalLen()

	var totalGasConsumed uint64
	if scheduledMode {
		totalGasConsumed = tc.gasHandler.TotalGasConsumedAsScheduled()
	} else {
		totalGasConsumed = tc.gasHandler.TotalGasConsumed()
	}

	gasConsumedInfo := &process.GasConsumedInfo{
		GasConsumedByMiniBlockInReceiverShard: uint64(0),
		GasConsumedByMiniBlocksInSenderShard:  uint64(0),
		TotalGasConsumedInSelfShard:           totalGasConsumed,
	}

	splitMiniBlocks, err := tc.splitMiniBlockBasedOnTxsType(miniBlock)
	if err != nil {
		return err
	}

	allTxsToBeReverted := make([][]byte, 0)
	allNumTxsProcessed := 0
	for index, mb := range splitMiniBlocks {
		preproc := tc.getPreProcessor(mb.Type)
		if check.IfNil(preproc) {
			return fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, mb.Type)
		}

		txsToBeReverted, numTxsProcessed, errProcess := preproc.ProcessMiniBlock(miniBlock, haveTime, haveAdditionalTime, tc.getNumOfCrossInterMbsAndTxs, scheduledMode, index == 0, gasConsumedInfo)
		allTxsToBeReverted = append(allTxsToBeReverted, txsToBeReverted...)
		allNumTxsProcessed += numTxsProcessed
		if errProcess != nil {
			log.Debug("processCompleteMiniBlock.ProcessMiniBlock",
				"scheduled mode", scheduledMode,
				"hash", miniBlockHash,
				"type", miniBlock.Type,
				"snd shard", miniBlock.SenderShardID,
				"rcv shard", miniBlock.ReceiverShardID,
				"num txs", len(miniBlock.TxHashes),
				"txs to be reverted", len(allTxsToBeReverted),
				"num txs processed", allNumTxsProcessed,
				"error", errProcess.Error(),
			)

			errAccountState := tc.accounts.RevertToSnapshot(snapshot)
			if errAccountState != nil {
				// TODO: evaluate if reloading the trie from disk will might solve the problem
				log.Debug("RevertToSnapshot", "error", errAccountState.Error())
			}

			if len(allTxsToBeReverted) > 0 {
				tc.revertProcessedTxsResults(allTxsToBeReverted)
			}

			return errProcess
		}
	}

	return nil
}

func (tc *transactionCoordinator) revertProcessedTxsResults(txHashes [][]byte) {
	for _, value := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[value]
		if !ok {
			continue
		}
		interProc.RemoveProcessedResultsFor(txHashes)
	}
	tc.feeHandler.RevertFees(txHashes)
}

// VerifyCreatedBlockTransactions checks whether the created transactions are the same as the one proposed
func (tc *transactionCoordinator) VerifyCreatedBlockTransactions(hdr data.HeaderHandler, body *block.Body) error {
	tc.mutInterimProcessors.RLock()
	defer tc.mutInterimProcessors.RUnlock()
	errMutex := sync.Mutex{}
	var errFound error

	wg := sync.WaitGroup{}
	wg.Add(len(tc.interimProcessors))

	for _, interimProc := range tc.interimProcessors {
		go func(intermediateProcessor process.IntermediateTransactionHandler) {
			err := intermediateProcessor.VerifyInterMiniBlocks(body)
			if err != nil {
				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}(interimProc)
	}

	wg.Wait()

	if errFound != nil {
		return errFound
	}

	if check.IfNil(hdr) {
		return process.ErrNilBlockHeader
	}

	createdReceiptHash, err := tc.CreateReceiptsHash()
	if err != nil {
		return err
	}

	if !bytes.Equal(createdReceiptHash, hdr.GetReceiptsHash()) {
		log.Debug("VerifyCreatedBlockTransactions", "error", process.ErrReceiptsHashMissmatch,
			"createdReceiptHash", createdReceiptHash,
			"headerReceiptHash", hdr.GetReceiptsHash(),
		)
		return process.ErrReceiptsHashMissmatch
	}

	return nil
}

// CreateReceiptsHash will return the hash for the receipts
func (tc *transactionCoordinator) CreateReceiptsHash() ([]byte, error) {
	allReceiptsHashes := make([][]byte, 0)

	for _, value := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[value]
		if !ok {
			continue
		}

		mb := interProc.GetCreatedInShardMiniBlock()
		if mb == nil {
			log.Trace("CreateReceiptsHash nil inshard miniblock for type", "type", value)
			continue
		}

		log.Trace("CreateReceiptsHash.GetCreatedInShardMiniBlock",
			"type", mb.Type,
			"senderShardID", mb.SenderShardID,
			"receiverShardID", mb.ReceiverShardID,
			"numTxHashes", len(mb.TxHashes),
			"interimProcType", value,
		)

		for _, hash := range mb.TxHashes {
			log.Trace("tx", "hash", hash)
		}

		currHash, err := core.CalculateHash(tc.marshalizer, tc.hasher, mb)
		if err != nil {
			return nil, err
		}

		allReceiptsHashes = append(allReceiptsHashes, currHash)
	}

	finalReceiptHash, err := core.CalculateHash(tc.marshalizer, tc.hasher, &batch.Batch{Data: allReceiptsHashes})
	return finalReceiptHash, err
}

// CreateMarshalizedReceipts will return all the receipts list in one marshalized object
func (tc *transactionCoordinator) CreateMarshalizedReceipts() ([]byte, error) {
	receiptsBatch := &batch.Batch{}
	for _, blockType := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[blockType]
		if !ok {
			continue
		}

		miniBlock := interProc.GetCreatedInShardMiniBlock()
		if miniBlock == nil {
			continue
		}

		marshalizedMiniBlock, err := tc.marshalizer.Marshal(miniBlock)
		if err != nil {
			return nil, err
		}

		receiptsBatch.Data = append(receiptsBatch.Data, marshalizedMiniBlock)
	}

	if len(receiptsBatch.Data) == 0 {
		return make([]byte, 0), nil
	}

	return tc.marshalizer.Marshal(receiptsBatch)
}

func (tc *transactionCoordinator) getNumOfCrossInterMbsAndTxs() (int, int) {
	totalNumMbs := 0
	totalNumTxs := 0

	for _, blockType := range tc.keysInterimProcs {
		interimProc := tc.getInterimProcessor(blockType)
		if check.IfNil(interimProc) {
			continue
		}

		numMbs, numTxs := interimProc.GetNumOfCrossInterMbsAndTxs()
		totalNumMbs += numMbs
		totalNumTxs += numTxs
	}

	return totalNumMbs, totalNumTxs
}

func (tc *transactionCoordinator) isMaxBlockSizeReached(body *block.Body) bool {
	numMbs := len(body.MiniBlocks)
	numTxs := 0
	numCrossShardScCallsOrSpecialTxs := 0

	allTxs := make(map[string]data.TransactionHandler)

	preProc := tc.getPreProcessor(block.TxBlock)
	if check.IfNil(preProc) {
		log.Warn("transactionCoordinator.isMaxBlockSizeReached: preProc is nil", "blockType", block.TxBlock)
	} else {
		allTxs = preProc.GetAllCurrentUsedTxs()
	}

	for _, mb := range body.MiniBlocks {
		numTxs += len(mb.TxHashes)
		numCrossShardScCallsOrSpecialTxs += getNumOfCrossShardScCallsOrSpecialTxs(mb, allTxs, tc.shardCoordinator.SelfId()) * common.AdditionalScrForEachScCallOrSpecialTx
	}

	if numCrossShardScCallsOrSpecialTxs > 0 {
		numMbs++
	}

	isMaxBlockSizeReached := tc.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(numMbs, numTxs+numCrossShardScCallsOrSpecialTxs)

	log.Trace("transactionCoordinator.isMaxBlockSizeReached",
		"isMaxBlockSizeReached", isMaxBlockSizeReached,
		"numMbs", numMbs,
		"numTxs", numTxs,
		"numCrossShardScCallsOrSpecialTxs", numCrossShardScCallsOrSpecialTxs,
	)

	return isMaxBlockSizeReached
}

func getNumOfCrossShardScCallsOrSpecialTxs(
	mb *block.MiniBlock,
	allTxs map[string]data.TransactionHandler,
	selfShardID uint32,
) int {
	isCrossShardTxBlockFromSelf := mb.Type == block.TxBlock && mb.SenderShardID == selfShardID && mb.ReceiverShardID != selfShardID
	if !isCrossShardTxBlockFromSelf {
		return 0
	}

	numCrossShardScCallsOrSpecialTxs := 0
	for _, txHash := range mb.TxHashes {
		tx, ok := allTxs[string(txHash)]
		if !ok {
			log.Warn("transactionCoordinator.getNumOfCrossShardScCallsOrSpecialTxs: tx not found",
				"mb type", mb.Type,
				"senderShardID", mb.SenderShardID,
				"receiverShardID", mb.ReceiverShardID,
				"numTxHashes", len(mb.TxHashes),
				"tx hash", txHash)

			// If the tx is not found we assume that it is the smart contract call or a special tx to handle the worst case scenario
			numCrossShardScCallsOrSpecialTxs++
			continue
		}

		if core.IsSmartContractAddress(tx.GetRcvAddr()) || len(tx.GetRcvUserName()) > 0 {
			numCrossShardScCallsOrSpecialTxs++
		}
	}

	return numCrossShardScCallsOrSpecialTxs
}

// VerifyCreatedMiniBlocks re-checks gas used and generated fees in the given block
func (tc *transactionCoordinator) VerifyCreatedMiniBlocks(header data.HeaderHandler, body *block.Body) error {
	if header.GetEpoch() < tc.blockGasAndFeesReCheckEnableEpoch {
		return nil
	}

	mapMiniBlockTypeAllTxs := tc.getAllTransactions()

	err := tc.verifyGasLimit(body, mapMiniBlockTypeAllTxs)
	if err != nil {
		return err
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	if err != nil {
		return err
	}

	return nil
}

func (tc *transactionCoordinator) getAllTransactions() map[block.Type]map[string]data.TransactionHandler {
	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)

	for _, blockType := range tc.keysTxPreProcs {
		_, ok := mapMiniBlockTypeAllTxs[blockType]
		if !ok {
			mapMiniBlockTypeAllTxs[blockType] = tc.GetAllCurrentUsedTxs(blockType)
		}
	}

	for _, blockType := range tc.keysInterimProcs {
		_, ok := mapMiniBlockTypeAllTxs[blockType]
		if !ok {
			mapMiniBlockTypeAllTxs[blockType] = tc.GetAllCurrentUsedTxs(blockType)
		}
	}

	return mapMiniBlockTypeAllTxs
}

func (tc *transactionCoordinator) verifyGasLimit(
	body *block.Body,
	mapMiniBlockTypeAllTxs map[block.Type]map[string]data.TransactionHandler,
) error {
	for _, miniBlock := range body.MiniBlocks {
		isCrossShardMiniBlockFromMe := miniBlock.SenderShardID == tc.shardCoordinator.SelfId() &&
			miniBlock.ReceiverShardID != tc.shardCoordinator.SelfId()
		if !isCrossShardMiniBlockFromMe {
			continue
		}

		if miniBlock.Type == block.SmartContractResultBlock {
			continue
		}

		err := tc.checkGasConsumedByMiniBlockInReceiverShard(miniBlock, mapMiniBlockTypeAllTxs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *transactionCoordinator) checkGasConsumedByMiniBlockInReceiverShard(
	miniBlock *block.MiniBlock,
	mapMiniBlockTypeAllTxs map[block.Type]map[string]data.TransactionHandler,
) error {
	var err error
	var gasConsumedByTxInReceiverShard uint64
	gasConsumedByMiniBlockInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		txHandler, ok := isTxHashInMap(txHash, mapMiniBlockTypeAllTxs)
		if !ok {
			log.Debug("missing transaction in checkGasConsumedByMiniBlockInReceiverShard ", "type", miniBlock.Type, "txHash", txHash)
			return process.ErrMissingTransaction
		}

		_, txTypeDstShard := tc.txTypeHandler.ComputeTransactionType(txHandler)
		moveBalanceGasLimit := tc.economicsFee.ComputeGasLimit(txHandler)
		if txTypeDstShard == process.MoveBalance {
			gasConsumedByTxInReceiverShard = moveBalanceGasLimit
		} else {
			gasConsumedByTxInReceiverShard, err = core.SafeSubUint64(txHandler.GetGasLimit(), moveBalanceGasLimit)
			if err != nil {
				return err
			}
		}

		gasConsumedByMiniBlockInReceiverShard, err = core.SafeAddUint64(gasConsumedByMiniBlockInReceiverShard, gasConsumedByTxInReceiverShard)
		if err != nil {
			return err
		}
	}

	if gasConsumedByMiniBlockInReceiverShard > tc.economicsFee.MaxGasLimitPerBlock(miniBlock.ReceiverShardID) {
		return process.ErrMaxGasLimitPerMiniBlockInReceiverShardIsReached
	}

	return nil
}

func (tc *transactionCoordinator) verifyFees(
	header data.HeaderHandler,
	body *block.Body,
	mapMiniBlockTypeAllTxs map[block.Type]map[string]data.TransactionHandler,
) error {
	totalMaxAccumulatedFees := big.NewInt(0)
	totalMaxDeveloperFees := big.NewInt(0)

	for _, miniBlock := range body.MiniBlocks {
		if miniBlock.Type == block.PeerBlock {
			continue
		}

		maxAccumulatedFeesFromMiniBlock, maxDeveloperFeesFromMiniBlock, err := tc.getMaxAccumulatedAndDeveloperFees(
			miniBlock,
			mapMiniBlockTypeAllTxs,
		)
		if err != nil {
			return err
		}

		totalMaxAccumulatedFees.Add(totalMaxAccumulatedFees, maxAccumulatedFeesFromMiniBlock)
		totalMaxDeveloperFees.Add(totalMaxDeveloperFees, maxDeveloperFeesFromMiniBlock)
	}

	if header.GetAccumulatedFees().Cmp(totalMaxAccumulatedFees) > 0 {
		return process.ErrMaxAccumulatedFeesExceeded
	}
	if header.GetDeveloperFees().Cmp(totalMaxDeveloperFees) > 0 {
		return process.ErrMaxDeveloperFeesExceeded
	}

	return nil
}

func (tc *transactionCoordinator) getMaxAccumulatedAndDeveloperFees(
	miniBlock *block.MiniBlock,
	mapMiniBlockTypeAllTxs map[block.Type]map[string]data.TransactionHandler,
) (*big.Int, *big.Int, error) {
	maxAccumulatedFeesFromMiniBlock := big.NewInt(0)
	maxDeveloperFeesFromMiniBlock := big.NewInt(0)

	for _, txHash := range miniBlock.TxHashes {
		txHandler, ok := isTxHashInMap(txHash, mapMiniBlockTypeAllTxs)
		if !ok {
			log.Debug("missing transaction in getMaxAccumulatedFeesAndDeveloperFees ", "type", miniBlock.Type, "txHash", txHash)
			return big.NewInt(0), big.NewInt(0), process.ErrMissingTransaction
		}

		maxAccumulatedFeesFromTx := core.SafeMul(txHandler.GetGasLimit(), txHandler.GetGasPrice())
		maxAccumulatedFeesFromMiniBlock.Add(maxAccumulatedFeesFromMiniBlock, maxAccumulatedFeesFromTx)

		maxDeveloperFeesFromTx := core.GetIntTrimmedPercentageOfValue(maxAccumulatedFeesFromTx, tc.economicsFee.DeveloperPercentage())
		maxDeveloperFeesFromMiniBlock.Add(maxDeveloperFeesFromMiniBlock, maxDeveloperFeesFromTx)
	}

	return maxAccumulatedFeesFromMiniBlock, maxDeveloperFeesFromMiniBlock, nil
}

func isTxHashInMap(
	txHash []byte,
	mapMiniBlockTypeAllTxs map[block.Type]map[string]data.TransactionHandler,
) (data.TransactionHandler, bool) {

	for _, mapHashTx := range mapMiniBlockTypeAllTxs {
		txHandler, ok := mapHashTx[string(txHash)]
		if ok {
			return txHandler, true
		}
	}

	return nil, false
}

func checkTransactionCoordinatorNilParameters(arguments ArgTransactionCoordinator) error {
	if check.IfNil(arguments.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arguments.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(arguments.MiniBlockPool) {
		return process.ErrNilMiniBlockPool
	}
	if check.IfNil(arguments.RequestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(arguments.InterProcessors) {
		return process.ErrNilIntermediateProcessorContainer
	}
	if check.IfNil(arguments.PreProcessors) {
		return process.ErrNilPreProcessorsContainer
	}
	if check.IfNil(arguments.GasHandler) {
		return process.ErrNilGasHandler
	}
	if check.IfNil(arguments.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arguments.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arguments.FeeHandler) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(arguments.BlockSizeComputation) {
		return process.ErrNilBlockSizeComputationHandler
	}
	if check.IfNil(arguments.BalanceComputation) {
		return process.ErrNilBalanceComputationHandler
	}
	if check.IfNil(arguments.EconomicsFee) {
		return process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(arguments.TxTypeHandler) {
		return process.ErrNilTxTypeHandler
	}
	if check.IfNil(arguments.TransactionsLogProcessor) {
		return process.ErrNilTxLogsProcessor
	}
	if check.IfNil(arguments.EpochNotifier) {
		return process.ErrNilEpochNotifier
	}
	if check.IfNil(arguments.PostProcessorTxsHandler) {
		return process.ErrNilPostProcessorTxsHandler
	}

	return nil
}

// AddIntermediateTransactions adds the given intermediate transactions
func (tc *transactionCoordinator) AddIntermediateTransactions(mapSCRs map[block.Type][]data.TransactionHandler) error {
	for blockType, scrs := range mapSCRs {
		interimProc := tc.getInterimProcessor(blockType)
		if check.IfNil(interimProc) {
			return process.ErrNilIntermediateProcessor
		}

		err := interimProc.AddIntermediateTransactions(scrs)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAllIntermediateTxs gets all the intermediate transactions separated by block type
func (tc *transactionCoordinator) GetAllIntermediateTxs() map[block.Type]map[string]data.TransactionHandler {
	mapIntermediateTxs := make(map[block.Type]map[string]data.TransactionHandler)
	for _, blockType := range tc.keysInterimProcs {
		interimProc := tc.getInterimProcessor(blockType)
		if check.IfNil(interimProc) {
			continue
		}

		mapIntermediateTxs[blockType] = interimProc.GetAllCurrentFinishedTxs()
	}

	return mapIntermediateTxs
}

// GetAllIntermediateTxsForTxHash gets all the intermediate transactions, for a given transaction hash, separated by block type
func (tc *transactionCoordinator) GetAllIntermediateTxsForTxHash(txHash []byte) map[block.Type]map[uint32][]*process.TxInfo {
	mapIntermediateTxs := make(map[block.Type]map[uint32][]*process.TxInfo)
	for _, blockType := range tc.keysInterimProcs {
		interimProc := tc.getInterimProcessor(blockType)
		if check.IfNil(interimProc) {
			continue
		}

		mapIntermediateTxs[blockType] = interimProc.GetAllIntermediateTxsForTxHash(txHash)
	}

	return mapIntermediateTxs
}

// EpochConfirmed is called whenever a new epoch is confirmed
func (tc *transactionCoordinator) EpochConfirmed(epoch uint32, _ uint64) {
	tc.flagMixedTxsInMiniBlocks.Toggle(epoch >= tc.mixedTxsInMiniBlocksEnableEpoch)
	log.Debug("transactionCoordinator: mixed txs in mini blocks", "enabled", tc.flagMixedTxsInMiniBlocks.IsSet())
	tc.flagScheduledMiniBlocks.Toggle(epoch >= tc.scheduledMiniBlocksEnableEpoch)
	log.Debug("transactionCoordinator: scheduled mini blocks", "enabled", tc.flagScheduledMiniBlocks.IsSet())
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *transactionCoordinator) IsInterfaceNil() bool {
	return tc == nil
}
