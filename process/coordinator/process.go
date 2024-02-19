package coordinator

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
	"github.com/multiversx/mx-chain-go/process/block/processedMb"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/cache"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.TransactionCoordinator = (*transactionCoordinator)(nil)

var log = logger.GetOrCreate("process/coordinator")

type createMiniBlockDestMeExecutionInfo struct {
	processedTxHashes             [][]byte
	miniBlocks                    block.MiniBlockSlice
	numTxAdded                    uint32
	numNewMiniBlocksProcessed     int
	numAlreadyMiniBlocksProcessed int
}

type processedIndexes struct {
	indexOfLastTxProcessed           int32
	indexOfLastTxProcessedByProposer int32
}

// ArgTransactionCoordinator holds all dependencies required by the transaction coordinator factory in order to create new instances
type ArgTransactionCoordinator struct {
	Hasher                       hashing.Hasher
	Marshalizer                  marshal.Marshalizer
	ShardCoordinator             sharding.Coordinator
	Accounts                     state.AccountsAdapter
	MiniBlockPool                storage.Cacher
	RequestHandler               process.RequestHandler
	PreProcessors                process.PreProcessorsContainer
	InterProcessors              process.IntermediateProcessorContainer
	GasHandler                   process.GasHandler
	FeeHandler                   process.TransactionFeeHandler
	BlockSizeComputation         preprocess.BlockSizeComputationHandler
	BalanceComputation           preprocess.BalanceComputationHandler
	EconomicsFee                 process.FeeHandler
	TxTypeHandler                process.TxTypeHandler
	TransactionsLogProcessor     process.TransactionLogProcessor
	EnableEpochsHandler          common.EnableEpochsHandler
	ScheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	DoubleTransactionsDetector   process.DoubleTransactionDetector
	ProcessedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	TxExecutionOrderHandler      common.TxExecutionOrderHandler
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

	onRequestMiniBlocks          func(shardId uint32, mbHashes [][]byte)
	gasHandler                   process.GasHandler
	feeHandler                   process.TransactionFeeHandler
	blockSizeComputation         preprocess.BlockSizeComputationHandler
	balanceComputation           preprocess.BalanceComputationHandler
	requestedItemsHandler        process.TimeCacher
	economicsFee                 process.FeeHandler
	txTypeHandler                process.TxTypeHandler
	transactionsLogProcessor     process.TransactionLogProcessor
	scheduledTxsExecutionHandler process.ScheduledTxsExecutionHandler
	doubleTransactionsDetector   process.DoubleTransactionDetector
	processedMiniBlocksTracker   process.ProcessedMiniBlocksTracker
	enableEpochsHandler          common.EnableEpochsHandler
	txExecutionOrderHandler      common.TxExecutionOrderHandler
}

// NewTransactionCoordinator creates a transaction coordinator to run and coordinate preprocessors and processors
func NewTransactionCoordinator(args ArgTransactionCoordinator) (*transactionCoordinator, error) {
	err := checkTransactionCoordinatorNilParameters(args)
	if err != nil {
		return nil, err
	}

	tc := &transactionCoordinator{
		shardCoordinator:             args.ShardCoordinator,
		accounts:                     args.Accounts,
		gasHandler:                   args.GasHandler,
		hasher:                       args.Hasher,
		marshalizer:                  args.Marshalizer,
		feeHandler:                   args.FeeHandler,
		blockSizeComputation:         args.BlockSizeComputation,
		balanceComputation:           args.BalanceComputation,
		economicsFee:                 args.EconomicsFee,
		txTypeHandler:                args.TxTypeHandler,
		transactionsLogProcessor:     args.TransactionsLogProcessor,
		scheduledTxsExecutionHandler: args.ScheduledTxsExecutionHandler,
		doubleTransactionsDetector:   args.DoubleTransactionsDetector,
		processedMiniBlocksTracker:   args.ProcessedMiniBlocksTracker,
		enableEpochsHandler:          args.EnableEpochsHandler,
		txExecutionOrderHandler:      args.TxExecutionOrderHandler,
	}

	tc.miniBlockPool = args.MiniBlockPool
	tc.onRequestMiniBlocks = args.RequestHandler.RequestMiniBlocks
	tc.requestedTxs = make(map[block.Type]int)
	tc.txPreProcessors = make(map[block.Type]process.PreProcessor)
	tc.interimProcessors = make(map[block.Type]process.IntermediateTransactionHandler)

	tc.keysTxPreProcs = args.PreProcessors.Keys()
	sort.Slice(tc.keysTxPreProcs, func(i, j int) bool {
		return tc.keysTxPreProcs[i] < tc.keysTxPreProcs[j]
	})
	for _, value := range tc.keysTxPreProcs {
		preProc, errGet := args.PreProcessors.Get(value)
		if errGet != nil {
			return nil, errGet
		}
		tc.txPreProcessors[value] = preProc
	}

	tc.keysInterimProcs = args.InterProcessors.Keys()
	sort.Slice(tc.keysInterimProcs, func(i, j int) bool {
		return tc.keysInterimProcs[i] < tc.keysInterimProcs[j]
	})
	for _, value := range tc.keysInterimProcs {
		interProc, errGet := args.InterProcessors.Get(value)
		if errGet != nil {
			return nil, errGet
		}
		tc.interimProcessors[value] = interProc
	}

	tc.requestedItemsHandler = cache.NewTimeCache(common.MaxWaitingTimeToReceiveRequestedItem)
	tc.miniBlockPool.RegisterHandler(tc.receivedMiniBlock, core.UniqueIdentifier())

	return tc, nil
}

// separateBodyByType creates a map of bodies according to type
func (tc *transactionCoordinator) separateBodyByType(body *block.Body) map[block.Type]*block.Body {
	separatedBodies := make(map[block.Type]*block.Body)
	for i := 0; i < len(body.MiniBlocks); i++ {
		mb := body.MiniBlocks[i]

		separatedMbType := mb.Type
		if mb.Type == block.InvalidBlock {
			separatedMbType = block.TxBlock
		}

		if _, ok := separatedBodies[separatedMbType]; !ok {
			separatedBodies[separatedMbType] = &block.Body{}
		}

		separatedBodies[separatedMbType].MiniBlocks = append(separatedBodies[separatedMbType].MiniBlocks, mb)
	}

	return separatedBodies
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

	separatedBodies := tc.separateBodyByType(body)

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

	wg.Wait()
	tc.mutRequestedTxs.RUnlock()

	return errFound
}

// SaveTxsToStorage saves transactions from block body into storage units
func (tc *transactionCoordinator) SaveTxsToStorage(body *block.Body) {
	if check.IfNil(body) {
		return
	}

	separatedBodies := tc.separateBodyByType(body)
	for key, value := range separatedBodies {
		tc.saveTxsToStorage(key, value)
	}

	for _, blockType := range tc.keysInterimProcs {
		tc.saveCurrentIntermediateTxToStorage(blockType)
	}
}

func (tc *transactionCoordinator) saveTxsToStorage(blockType block.Type, blockBody *block.Body) {
	preproc := tc.getPreProcessor(blockType)
	if check.IfNil(preproc) {
		return
	}

	err := preproc.SaveTxsToStorage(blockBody)
	if err != nil {
		log.Trace("SaveTxsToStorage", "error", err.Error())
	}
}

func (tc *transactionCoordinator) saveCurrentIntermediateTxToStorage(blockType block.Type) {
	intermediateProc := tc.getInterimProcessor(blockType)
	if check.IfNil(intermediateProc) {
		return
	}

	intermediateProc.SaveCurrentIntermediateTxToStorage()
}

// RestoreBlockDataFromStorage restores block data from storage to pool
func (tc *transactionCoordinator) RestoreBlockDataFromStorage(body *block.Body) (int, error) {
	if check.IfNil(body) {
		return 0, nil
	}

	separatedBodies := tc.separateBodyByType(body)

	var errFound error
	localMutex := sync.Mutex{}
	totalRestoredTx := 0

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody *block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if check.IfNil(preproc) {
				wg.Done()
				return
			}

			restoredTxs, err := preproc.RestoreBlockDataIntoPools(blockBody, tc.miniBlockPool)
			if err != nil {
				log.Trace("RestoreBlockDataIntoPools", "error", err.Error())

				localMutex.Lock()
				errFound = err
				localMutex.Unlock()
			}

			localMutex.Lock()
			totalRestoredTx += restoredTxs

			localMutex.Unlock()

			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return totalRestoredTx, errFound
}

// RemoveBlockDataFromPool deletes block data from pools
func (tc *transactionCoordinator) RemoveBlockDataFromPool(body *block.Body) error {
	if check.IfNil(body) {
		return nil
	}

	separatedBodies := tc.separateBodyByType(body)

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

			err := preproc.RemoveBlockDataFromPools(blockBody, tc.miniBlockPool)
			if err != nil {
				log.Trace("RemoveBlockDataFromPools", "error", err.Error())

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

	separatedBodies := tc.separateBodyByType(body)

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
) (block.MiniBlockSlice, error) {
	if check.IfNil(body) {
		return nil, process.ErrNilBlockBody
	}

	if tc.isMaxBlockSizeReached(body) {
		return nil, process.ErrMaxBlockSizeReached
	}

	haveTime := func() bool {
		return timeRemaining() >= 0
	}

	tc.doubleTransactionsDetector.ProcessBlockBody(body)

	startTime := time.Now()
	createdMbsToMe, mbIndex, err := tc.processMiniBlocksToMe(header, body, haveTime)
	elapsedTime := time.Since(startTime)
	log.Debug("elapsed time to processMiniBlocksToMe",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, err
	}

	if mbIndex == len(body.MiniBlocks) {
		return createdMbsToMe, nil
	}

	miniBlocksFromMe := body.MiniBlocks[mbIndex:]
	startTime = time.Now()
	createdMbsFromMe, err := tc.processMiniBlocksFromMe(header, &block.Body{MiniBlocks: miniBlocksFromMe}, haveTime)
	elapsedTime = time.Since(startTime)
	log.Debug("elapsed time to processMiniBlocksFromMe",
		"time [s]", elapsedTime,
	)
	if err != nil {
		return nil, err
	}

	finalMiniBlocks := append(createdMbsToMe, createdMbsFromMe...)

	return finalMiniBlocks, nil
}

func (tc *transactionCoordinator) processMiniBlocksFromMe(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) (block.MiniBlockSlice, error) {
	for _, mb := range body.MiniBlocks {
		if mb.SenderShardID != tc.shardCoordinator.SelfId() {
			return nil, process.ErrMiniBlocksInWrongOrder
		}
	}

	numMiniBlocksProcessed := 0
	separatedBodies := tc.separateBodyByType(body)

	defer func() {
		log.Debug("transactionCoordinator.processMiniBlocksFromMe: gas consumed, refunded and penalized info",
			"num mini blocks processed", numMiniBlocksProcessed,
			"total gas provided", tc.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", tc.gasHandler.TotalGasRefunded(),
			"total gas penalized", tc.gasHandler.TotalGasPenalized())
	}()

	createdMBs := make(block.MiniBlockSlice, 0)
	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	for _, blockType := range tc.keysTxPreProcs {
		if separatedBodies[blockType] == nil {
			continue
		}

		preProc := tc.getPreProcessor(blockType)
		if check.IfNil(preProc) {
			return nil, process.ErrMissingPreProcessor
		}

		miniBlocks, err := preProc.ProcessBlockTransactions(header, separatedBodies[blockType], haveTime)
		if err != nil {
			return nil, err
		}

		createdMBs = append(createdMBs, miniBlocks...)
		numMiniBlocksProcessed += len(separatedBodies[blockType].MiniBlocks)
	}

	return createdMBs, nil
}

func (tc *transactionCoordinator) processMiniBlocksToMe(
	header data.HeaderHandler,
	body *block.Body,
	haveTime func() bool,
) (block.MiniBlockSlice, int, error) {
	numMiniBlocksProcessed := 0

	defer func() {
		log.Debug("transactionCoordinator.processMiniBlocksToMe: gas provided, refunded and penalized info",
			"num mini blocks processed", numMiniBlocksProcessed,
			"total gas provided", tc.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", tc.gasHandler.TotalGasRefunded(),
			"total gas penalized", tc.gasHandler.TotalGasPenalized())
	}()

	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	// processing destination ME miniblocks first
	mbIndex := 0
	createdMBs := make(block.MiniBlockSlice, 0)
	for mbIndex = 0; mbIndex < len(body.MiniBlocks); mbIndex++ {
		miniBlock := body.MiniBlocks[mbIndex]
		if miniBlock.SenderShardID == tc.shardCoordinator.SelfId() {
			return createdMBs, mbIndex, nil
		}

		preProc := tc.getPreProcessor(miniBlock.Type)
		if check.IfNil(preProc) {
			return nil, mbIndex, process.ErrMissingPreProcessor
		}

		log.Debug("processMiniBlocksToMe: miniblock", "type", miniBlock.Type)
		miniBlocks, err := preProc.ProcessBlockTransactions(header, &block.Body{MiniBlocks: []*block.MiniBlock{miniBlock}}, haveTime)
		if err != nil {
			return nil, mbIndex, err
		}

		createdMBs = append(createdMBs, miniBlocks...)
		numMiniBlocksProcessed++
	}

	return createdMBs, mbIndex, nil
}

// CreateMbsAndProcessCrossShardTransactionsDstMe creates miniblocks and processes cross shard transaction
// with destination of current shard
func (tc *transactionCoordinator) CreateMbsAndProcessCrossShardTransactionsDstMe(
	hdr data.HeaderHandler,
	processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
) (block.MiniBlockSlice, uint32, bool, error) {

	createMBDestMeExecutionInfo := initMiniBlockDestMeExecutionInfo()

	if check.IfNil(hdr) {
		log.Warn("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe header is nil")

		// we return the nil error here as to allow the proposer execute as much as it can, even if it ends up in a
		// totally unlikely situation in which it needs to process a nil block.

		return createMBDestMeExecutionInfo.miniBlocks, createMBDestMeExecutionInfo.numTxAdded, false, nil
	}

	shouldSkipShard := make(map[uint32]bool)

	headerHash, err := core.CalculateHash(tc.marshalizer, tc.hasher, hdr)
	if err != nil {
		log.Warn("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe CalculateHash error",
			"error", err)

		// we return the nil error here as to allow the proposer execute as much as it can, even if it ends up in a
		// totally unlikely situation in which it can not marshall a block.
		return createMBDestMeExecutionInfo.miniBlocks, createMBDestMeExecutionInfo.numTxAdded, false, nil
	}

	tc.handleCreateMiniBlocksDestMeInit(headerHash)

	finalCrossMiniBlockInfos := tc.getFinalCrossMiniBlockInfos(hdr.GetOrderedCrossMiniblocksWithDst(tc.shardCoordinator.SelfId()), hdr)

	defer func() {
		log.Debug("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: gas provided, refunded and penalized info",
			"header round", hdr.GetRound(),
			"header nonce", hdr.GetNonce(),
			"num mini blocks to be processed", len(finalCrossMiniBlockInfos),
			"num already mini blocks processed", createMBDestMeExecutionInfo.numAlreadyMiniBlocksProcessed,
			"num new mini blocks processed", createMBDestMeExecutionInfo.numNewMiniBlocksProcessed,
			"total gas provided", tc.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", tc.gasHandler.TotalGasRefunded(),
			"total gas penalized", tc.gasHandler.TotalGasPenalized())
	}()

	tc.requestMissingMiniBlocksAndTransactions(finalCrossMiniBlockInfos)

	for _, miniBlockInfo := range finalCrossMiniBlockInfos {
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

		processedMbInfo := getProcessedMiniBlockInfo(processedMiniBlocksInfo, miniBlockInfo.Hash)
		if processedMbInfo.FullyProcessed {
			createMBDestMeExecutionInfo.numAlreadyMiniBlocksProcessed++
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
			log.Error("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: mini block assertion type failed",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
			)
			continue
		}

		if scheduledMode && (miniBlock.Type != block.TxBlock || processedMbInfo.IndexOfLastTxProcessed > -1) {
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Debug("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: mini block can not be processed in scheduled mode",
				"scheduled mode", scheduledMode,
				"type", miniBlock.Type,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
				"index of last tx processed", processedMbInfo.IndexOfLastTxProcessed,
			)
			continue
		}

		preproc := tc.getPreProcessor(miniBlock.Type)
		if check.IfNil(preproc) {
			return nil, 0, false, fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type)
		}

		requestedTxs := preproc.RequestTransactionsForMiniBlock(miniBlock)
		if requestedTxs > 0 {
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Trace("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: transactions not found and were requested",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"round", miniBlockInfo.Round,
				"requested txs", requestedTxs,
			)
			continue
		}

		oldIndexOfLastTxProcessed := processedMbInfo.IndexOfLastTxProcessed

		errProc := tc.processCompleteMiniBlock(preproc, miniBlock, miniBlockInfo.Hash, haveTime, haveAdditionalTime, scheduledMode, processedMbInfo)
		tc.handleProcessMiniBlockExecution(oldIndexOfLastTxProcessed, miniBlock, processedMbInfo, createMBDestMeExecutionInfo)
		if errProc != nil {
			shouldSkipShard[miniBlockInfo.SenderShardID] = true
			log.Debug("transactionCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: processed complete mini block failed",
				"scheduled mode", scheduledMode,
				"sender shard", miniBlockInfo.SenderShardID,
				"hash", miniBlockInfo.Hash,
				"type", miniBlock.Type,
				"round", miniBlockInfo.Round,
				"num txs", len(miniBlock.TxHashes),
				"num all txs processed", processedMbInfo.IndexOfLastTxProcessed+1,
				"num current txs processed", processedMbInfo.IndexOfLastTxProcessed-oldIndexOfLastTxProcessed,
				"fully processed", processedMbInfo.FullyProcessed,
				"total gas provided", tc.gasHandler.TotalGasProvided(),
				"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
				"total gas refunded", tc.gasHandler.TotalGasRefunded(),
				"total gas penalized", tc.gasHandler.TotalGasPenalized(),
				"error", errProc,
			)

			continue
		}

		log.Debug("transactionsCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe: processed complete mini block succeeded",
			"scheduled mode", scheduledMode,
			"sender shard", miniBlockInfo.SenderShardID,
			"hash", miniBlockInfo.Hash,
			"type", miniBlock.Type,
			"round", miniBlockInfo.Round,
			"num txs", len(miniBlock.TxHashes),
			"num all txs processed", processedMbInfo.IndexOfLastTxProcessed+1,
			"num current txs processed", processedMbInfo.IndexOfLastTxProcessed-oldIndexOfLastTxProcessed,
			"fully processed", processedMbInfo.FullyProcessed,
			"total gas provided", tc.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", tc.gasHandler.TotalGasRefunded(),
			"total gas penalized", tc.gasHandler.TotalGasPenalized(),
		)
	}

	numTotalMiniBlocksProcessed := createMBDestMeExecutionInfo.numAlreadyMiniBlocksProcessed + createMBDestMeExecutionInfo.numNewMiniBlocksProcessed
	allMBsProcessed := numTotalMiniBlocksProcessed == len(finalCrossMiniBlockInfos)
	if !allMBsProcessed {
		tc.revertIfNeeded(createMBDestMeExecutionInfo, headerHash)
	}

	return createMBDestMeExecutionInfo.miniBlocks, createMBDestMeExecutionInfo.numTxAdded, allMBsProcessed, nil
}

func (tc *transactionCoordinator) requestMissingMiniBlocksAndTransactions(mbsInfo []*data.MiniBlockInfo) {
	mapMissingMiniBlocksPerShard := make(map[uint32][][]byte)

	tc.requestedItemsHandler.Sweep()

	for _, mbInfo := range mbsInfo {
		object, isMiniBlockFound := tc.miniBlockPool.Peek(mbInfo.Hash)
		if !isMiniBlockFound {
			log.Debug("transactionCoordinator.requestMissingMiniBlocksAndTransactions: mini block not found and was requested",
				"sender shard", mbInfo.SenderShardID,
				"hash", mbInfo.Hash,
				"round", mbInfo.Round,
			)
			mapMissingMiniBlocksPerShard[mbInfo.SenderShardID] = append(mapMissingMiniBlocksPerShard[mbInfo.SenderShardID], mbInfo.Hash)
			_ = tc.requestedItemsHandler.Add(string(mbInfo.Hash))
			continue
		}

		miniBlock, isMiniBlock := object.(*block.MiniBlock)
		if !isMiniBlock {
			log.Warn("transactionCoordinator.requestMissingMiniBlocksAndTransactions", "mb hash", mbInfo.Hash, "error", process.ErrWrongTypeAssertion)
			continue
		}

		preproc := tc.getPreProcessor(miniBlock.Type)
		if check.IfNil(preproc) {
			log.Warn("transactionCoordinator.requestMissingMiniBlocksAndTransactions: getPreProcessor", "mb type", miniBlock.Type, "error", process.ErrNilPreProcessor)
			continue
		}

		numTxsRequested := preproc.RequestTransactionsForMiniBlock(miniBlock)
		if numTxsRequested > 0 {
			log.Debug("transactionCoordinator.requestMissingMiniBlocksAndTransactions: RequestTransactionsForMiniBlock", "mb hash", mbInfo.Hash,
				"num txs requested", numTxsRequested)
		}
	}

	for senderShardID, mbsHashes := range mapMissingMiniBlocksPerShard {
		go tc.onRequestMiniBlocks(senderShardID, mbsHashes)
	}
}

func initMiniBlockDestMeExecutionInfo() *createMiniBlockDestMeExecutionInfo {
	return &createMiniBlockDestMeExecutionInfo{
		processedTxHashes:             make([][]byte, 0),
		miniBlocks:                    make(block.MiniBlockSlice, 0),
		numTxAdded:                    0,
		numNewMiniBlocksProcessed:     0,
		numAlreadyMiniBlocksProcessed: 0,
	}
}

func (tc *transactionCoordinator) handleCreateMiniBlocksDestMeInit(headerHash []byte) {
	if tc.shardCoordinator.SelfId() != core.MetachainShardId {
		return
	}

	tc.InitProcessedTxsResults(headerHash)
	tc.gasHandler.Reset(headerHash)
}

func (tc *transactionCoordinator) handleProcessMiniBlockExecution(
	oldIndexOfLastTxProcessed int32,
	miniBlock *block.MiniBlock,
	processedMbInfo *processedMb.ProcessedMiniBlockInfo,
	createMBDestMeExecutionInfo *createMiniBlockDestMeExecutionInfo,
) {
	if oldIndexOfLastTxProcessed >= processedMbInfo.IndexOfLastTxProcessed {
		return
	}

	newProcessedTxHashes := miniBlock.TxHashes[oldIndexOfLastTxProcessed+1 : processedMbInfo.IndexOfLastTxProcessed+1]
	createMBDestMeExecutionInfo.processedTxHashes = append(createMBDestMeExecutionInfo.processedTxHashes, newProcessedTxHashes...)
	createMBDestMeExecutionInfo.miniBlocks = append(createMBDestMeExecutionInfo.miniBlocks, miniBlock)
	createMBDestMeExecutionInfo.numTxAdded = createMBDestMeExecutionInfo.numTxAdded + uint32(len(newProcessedTxHashes))

	if processedMbInfo.FullyProcessed {
		createMBDestMeExecutionInfo.numNewMiniBlocksProcessed++
	}
}

func getProcessedMiniBlockInfo(
	processedMiniBlocksInfo map[string]*processedMb.ProcessedMiniBlockInfo,
	miniBlockHash []byte,
) *processedMb.ProcessedMiniBlockInfo {

	if processedMiniBlocksInfo == nil {
		return &processedMb.ProcessedMiniBlockInfo{
			IndexOfLastTxProcessed: -1,
			FullyProcessed:         false,
		}
	}

	processedMbInfo, ok := processedMiniBlocksInfo[string(miniBlockHash)]
	if !ok {
		processedMbInfo = &processedMb.ProcessedMiniBlockInfo{
			IndexOfLastTxProcessed: -1,
			FullyProcessed:         false,
		}
		processedMiniBlocksInfo[string(miniBlockHash)] = processedMbInfo
	}

	return processedMbInfo
}

func (tc *transactionCoordinator) getFinalCrossMiniBlockInfos(
	crossMiniBlockInfos []*data.MiniBlockInfo,
	header data.HeaderHandler,
) []*data.MiniBlockInfo {

	if !tc.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return crossMiniBlockInfos
	}

	miniBlockInfos := make([]*data.MiniBlockInfo, 0)
	for _, crossMiniBlockInfo := range crossMiniBlockInfos {
		miniBlockHeader := process.GetMiniBlockHeaderWithHash(header, crossMiniBlockInfo.Hash)
		if miniBlockHeader != nil && !miniBlockHeader.IsFinal() {
			log.Debug("transactionCoordinator.getFinalCrossMiniBlockInfos: do not execute mini block which is not final", "mb hash", miniBlockHeader.GetHash())
			continue
		}

		miniBlockInfos = append(miniBlockInfos, crossMiniBlockInfo)
	}

	return miniBlockInfos
}

func (tc *transactionCoordinator) revertIfNeeded(createMBDestMeExecutionInfo *createMiniBlockDestMeExecutionInfo, key []byte) {
	shouldRevert := tc.shardCoordinator.SelfId() == core.MetachainShardId && len(createMBDestMeExecutionInfo.processedTxHashes) > 0
	if !shouldRevert {
		return
	}

	tc.gasHandler.RestoreGasSinceLastReset(key)
	tc.RevertProcessedTxsResults(createMBDestMeExecutionInfo.processedTxHashes, key)

	createMBDestMeExecutionInfo.miniBlocks = make(block.MiniBlockSlice, 0)
	createMBDestMeExecutionInfo.numTxAdded = 0
}

// CreateMbsAndProcessTransactionsFromMe creates miniblocks and processes transactions from pool
func (tc *transactionCoordinator) CreateMbsAndProcessTransactionsFromMe(
	haveTime func() bool,
	randomness []byte,
) block.MiniBlockSlice {

	numMiniBlocksProcessed := 0
	miniBlocks := make(block.MiniBlockSlice, 0)

	defer func() {
		log.Debug("transactionCoordinator.CreateMbsAndProcessTransactionsFromMe: gas provided, refunded and penalized info",
			"num mini blocks processed", numMiniBlocksProcessed,
			"total gas provided", tc.gasHandler.TotalGasProvided(),
			"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
			"total gas refunded", tc.gasHandler.TotalGasRefunded(),
			"total gas penalized", tc.gasHandler.TotalGasPenalized())
	}()

	for _, blockType := range tc.keysTxPreProcs {
		txPreProc := tc.getPreProcessor(blockType)
		if check.IfNil(txPreProc) {
			return nil
		}

		mbs, err := txPreProc.CreateAndProcessMiniBlocks(haveTime, randomness)
		if err != nil {
			log.Debug("CreateAndProcessMiniBlocks", "error", err.Error())
		}

		if len(mbs) > 0 {
			miniBlocks = append(miniBlocks, mbs...)
		}

		numMiniBlocksProcessed += len(mbs)
	}

	interMBs := tc.CreatePostProcessMiniBlocks()
	if len(interMBs) > 0 {
		miniBlocks = append(miniBlocks, interMBs...)
	}

	return miniBlocks
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
	tc.txExecutionOrderHandler.Clear()

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
				tc.appendMarshalledItems(
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
				tc.appendMarshalledItems(
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

func (tc *transactionCoordinator) appendMarshalledItems(
	dataMarshalizer process.DataMarshalizer,
	txHashes [][]byte,
	mrsTxs map[string][][]byte,
	broadcastTopic string,
) {
	currMrsTxs, err := dataMarshalizer.CreateMarshalledData(txHashes)
	if err != nil {
		log.Debug("appendMarshalledItems.CreateMarshalledData", "error", err.Error())
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
func (tc *transactionCoordinator) GetAllCurrentLogs() []*data.LogData {
	return tc.transactionsLogProcessor.GetAllCurrentLogs()
}

// RequestMiniBlocksAndTransactions requests mini blocks and transactions if missing
func (tc *transactionCoordinator) RequestMiniBlocksAndTransactions(header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	finalCrossMiniBlockHashes := tc.getFinalCrossMiniBlockHashes(header)
	mbsInfo := make([]*data.MiniBlockInfo, 0, len(finalCrossMiniBlockHashes))
	for mbHash, senderShardID := range finalCrossMiniBlockHashes {
		mbsInfo = append(mbsInfo, &data.MiniBlockInfo{
			Hash:          []byte(mbHash),
			SenderShardID: senderShardID,
			Round:         header.GetRound(),
		})
	}

	tc.requestMissingMiniBlocksAndTransactions(mbsInfo)
}

func (tc *transactionCoordinator) getFinalCrossMiniBlockHashes(headerHandler data.HeaderHandler) map[string]uint32 {
	if !tc.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		return headerHandler.GetMiniBlockHeadersWithDst(tc.shardCoordinator.SelfId())
	}
	return process.GetFinalCrossMiniBlockHashes(headerHandler, tc.shardCoordinator.SelfId())
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

	preproc := tc.getPreProcessor(miniBlock.Type)
	if check.IfNil(preproc) {
		log.Warn("transactionCoordinator.receivedMiniBlock",
			"error", fmt.Errorf("%w unknown block type %d", process.ErrNilPreProcessor, miniBlock.Type))
		return
	}

	numTxsRequested := preproc.RequestTransactionsForMiniBlock(miniBlock)
	if numTxsRequested > 0 {
		log.Debug("transactionCoordinator.receivedMiniBlock", "hash", key,
			"num txs requested", numTxsRequested)
	}
}

// processMiniBlockComplete - all transactions must be processed together, otherwise error
func (tc *transactionCoordinator) processCompleteMiniBlock(
	preproc process.PreProcessor,
	miniBlock *block.MiniBlock,
	miniBlockHash []byte,
	haveTime func() bool,
	haveAdditionalTime func() bool,
	scheduledMode bool,
	processedMbInfo *processedMb.ProcessedMiniBlockInfo,
) error {

	snapshot := tc.handleProcessMiniBlockInit(miniBlockHash)

	log.Debug("transactionsCoordinator.processCompleteMiniBlock: before processing",
		"scheduled mode", scheduledMode,
		"sender shard", miniBlock.SenderShardID,
		"hash", miniBlockHash,
		"type", miniBlock.Type,
		"num txs to be processed", len(miniBlock.TxHashes),
		"total gas provided", tc.gasHandler.TotalGasProvided(),
		"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
		"total gas refunded", tc.gasHandler.TotalGasRefunded(),
		"total gas penalized", tc.gasHandler.TotalGasPenalized(),
	)

	txsToBeReverted, indexOfLastTxProcessed, shouldRevert, err := preproc.ProcessMiniBlock(
		miniBlock,
		haveTime,
		haveAdditionalTime,
		scheduledMode,
		tc.enableEpochsHandler.IsFlagEnabled(common.MiniBlockPartialExecutionFlag),
		int(processedMbInfo.IndexOfLastTxProcessed),
		tc,
	)

	log.Debug("transactionsCoordinator.processCompleteMiniBlock: after processing",
		"num all txs processed", indexOfLastTxProcessed+1,
		"num current txs processed", indexOfLastTxProcessed-int(processedMbInfo.IndexOfLastTxProcessed),
		"txs to be reverted", len(txsToBeReverted),
		"total gas provided", tc.gasHandler.TotalGasProvided(),
		"total gas provided as scheduled", tc.gasHandler.TotalGasProvidedAsScheduled(),
		"total gas refunded", tc.gasHandler.TotalGasRefunded(),
		"total gas penalized", tc.gasHandler.TotalGasPenalized(),
	)

	if err != nil {
		log.Debug("processCompleteMiniBlock.ProcessMiniBlock",
			"scheduled mode", scheduledMode,
			"hash", miniBlockHash,
			"type", miniBlock.Type,
			"snd shard", miniBlock.SenderShardID,
			"rcv shard", miniBlock.ReceiverShardID,
			"num txs", len(miniBlock.TxHashes),
			"txs to be reverted", len(txsToBeReverted),
			"num all txs processed", indexOfLastTxProcessed+1,
			"num current txs processed", indexOfLastTxProcessed-int(processedMbInfo.IndexOfLastTxProcessed),
			"should revert", shouldRevert,
			"error", err.Error(),
		)

		if shouldRevert {
			tc.handleProcessTransactionError(snapshot, miniBlockHash, txsToBeReverted)
		} else {
			if tc.enableEpochsHandler.IsFlagEnabled(common.MiniBlockPartialExecutionFlag) {
				processedMbInfo.IndexOfLastTxProcessed = int32(indexOfLastTxProcessed)
				processedMbInfo.FullyProcessed = false
			}
		}

		return err
	}

	processedMbInfo.IndexOfLastTxProcessed = int32(indexOfLastTxProcessed)
	processedMbInfo.FullyProcessed = true

	return nil
}

func (tc *transactionCoordinator) handleProcessMiniBlockInit(miniBlockHash []byte) int {
	snapshot := tc.accounts.JournalLen()
	tc.InitProcessedTxsResults(miniBlockHash)
	tc.gasHandler.Reset(miniBlockHash)

	return snapshot
}

func (tc *transactionCoordinator) handleProcessTransactionError(snapshot int, miniBlockHash []byte, txsToBeReverted [][]byte) {
	tc.gasHandler.RestoreGasSinceLastReset(miniBlockHash)

	err := tc.accounts.RevertToSnapshot(snapshot)
	if err != nil {
		log.Debug("transactionCoordinator.handleProcessTransactionError: RevertToSnapshot", "error", err.Error())
	}

	if len(txsToBeReverted) > 0 {
		tc.RevertProcessedTxsResults(txsToBeReverted, miniBlockHash)
	}
	tc.txExecutionOrderHandler.RemoveMultiple(txsToBeReverted)
}

// InitProcessedTxsResults inits processed txs results for the given key
func (tc *transactionCoordinator) InitProcessedTxsResults(key []byte) {
	tc.mutInterimProcessors.RLock()
	defer tc.mutInterimProcessors.RUnlock()

	for _, value := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[value]
		if !ok {
			continue
		}
		interProc.InitProcessedResults(key)
	}
}

// RevertProcessedTxsResults reverts processed txs results for the given hashes and key
func (tc *transactionCoordinator) RevertProcessedTxsResults(txHashes [][]byte, key []byte) {
	tc.mutInterimProcessors.RLock()
	defer tc.mutInterimProcessors.RUnlock()

	for _, value := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[value]
		if !ok {
			continue
		}
		resultHashes := interProc.RemoveProcessedResults(key)
		accFeesBeforeRevert := tc.feeHandler.GetAccumulatedFees()
		tc.feeHandler.RevertFees(resultHashes)
		tc.txExecutionOrderHandler.RemoveMultiple(resultHashes)
		accFeesAfterRevert := tc.feeHandler.GetAccumulatedFees()

		if accFeesBeforeRevert.Cmp(accFeesAfterRevert) != 0 {
			log.Debug("revertProcessedTxsResults.RevertFees with result hashes",
				"num resultHashes", len(resultHashes),
				"value", big.NewInt(0).Sub(accFeesBeforeRevert, accFeesAfterRevert))
		}
	}

	accFeesBeforeRevert := tc.feeHandler.GetAccumulatedFees()
	tc.feeHandler.RevertFees(txHashes)
	tc.txExecutionOrderHandler.RemoveMultiple(txHashes)
	accFeesAfterRevert := tc.feeHandler.GetAccumulatedFees()
	if accFeesBeforeRevert.Cmp(accFeesAfterRevert) != 0 {
		log.Debug("revertProcessedTxsResults.RevertFees with tx hashes",
			"num txHashes", len(txHashes),
			"value", big.NewInt(0).Sub(accFeesBeforeRevert, accFeesAfterRevert))
	}
}

// VerifyCreatedBlockTransactions checks whether the created transactions are the same as the one proposed
func (tc *transactionCoordinator) VerifyCreatedBlockTransactions(hdr data.HeaderHandler, body *block.Body) error {
	errMutex := sync.Mutex{}
	var errFound error

	wg := sync.WaitGroup{}

	tc.mutInterimProcessors.RLock()
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
	tc.mutInterimProcessors.RUnlock()

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
	tc.mutInterimProcessors.RLock()
	defer tc.mutInterimProcessors.RUnlock()

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

// GetCreatedInShardMiniBlocks will return the intra-shard created miniblocks
func (tc *transactionCoordinator) GetCreatedInShardMiniBlocks() []*block.MiniBlock {
	tc.mutInterimProcessors.RLock()
	defer tc.mutInterimProcessors.RUnlock()

	miniBlocks := make([]*block.MiniBlock, 0)
	for _, blockType := range tc.keysInterimProcs {
		interProc, ok := tc.interimProcessors[blockType]
		if !ok {
			continue
		}

		miniBlock := interProc.GetCreatedInShardMiniBlock()
		if miniBlock == nil {
			continue
		}

		miniBlocks = append(miniBlocks, miniBlock)
	}

	return miniBlocks
}

// GetNumOfCrossInterMbsAndTxs gets the number of cross intermediate transactions and mini blocks
func (tc *transactionCoordinator) GetNumOfCrossInterMbsAndTxs() (int, int) {
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
func (tc *transactionCoordinator) VerifyCreatedMiniBlocks(
	header data.HeaderHandler,
	body *block.Body,
) error {
	if header.GetEpoch() < tc.enableEpochsHandler.GetActivationEpoch(common.BlockGasAndFeesReCheckFlag) {
		return nil
	}

	mapMiniBlockTypeAllTxs := tc.getAllTransactions(body)

	err := tc.verifyGasLimit(header, body, mapMiniBlockTypeAllTxs)
	if err != nil {
		return err
	}

	err = tc.verifyFees(header, body, mapMiniBlockTypeAllTxs)
	if err != nil {
		return err
	}

	return nil
}

func (tc *transactionCoordinator) getAllTransactions(body *block.Body) map[block.Type]map[string]data.TransactionHandler {
	mapMiniBlockTypeAllTxs := make(map[block.Type]map[string]data.TransactionHandler)
	for _, miniBlock := range body.MiniBlocks {
		_, ok := mapMiniBlockTypeAllTxs[miniBlock.Type]
		if !ok {
			mapMiniBlockTypeAllTxs[miniBlock.Type] = tc.GetAllCurrentUsedTxs(miniBlock.Type)
		}
	}

	return mapMiniBlockTypeAllTxs
}

func (tc *transactionCoordinator) verifyGasLimit(
	header data.HeaderHandler,
	body *block.Body,
	mapMiniBlockTypeAllTxs map[block.Type]map[string]data.TransactionHandler,
) error {

	if len(body.MiniBlocks) != len(header.GetMiniBlockHeaderHandlers()) {
		log.Warn("transactionCoordinator.verifyGasLimit: num of mini blocks and mini blocks headers does not match", "num of mb", len(body.MiniBlocks), "num of mbh", len(header.GetMiniBlockHeaderHandlers()))
		return process.ErrNumOfMiniBlocksAndMiniBlocksHeadersMismatch
	}

	for index, miniBlock := range body.MiniBlocks {
		isCrossShardMiniBlockFromMe := miniBlock.SenderShardID == tc.shardCoordinator.SelfId() &&
			miniBlock.ReceiverShardID != tc.shardCoordinator.SelfId()
		if !isCrossShardMiniBlockFromMe {
			continue
		}
		if miniBlock.Type == block.SmartContractResultBlock {
			continue
		}
		if tc.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
			miniBlockHeader := header.GetMiniBlockHeaderHandlers()[index]
			if miniBlockHeader.GetProcessingType() == int32(block.Processed) {
				log.Debug("transactionCoordinator.verifyGasLimit: do not verify gas limit for mini block executed as scheduled in previous block", "mb hash", miniBlockHeader.GetHash())
				continue
			}
		}

		err := tc.checkGasProvidedByMiniBlockInReceiverShard(miniBlock, mapMiniBlockTypeAllTxs[miniBlock.Type])
		if err != nil {
			return err
		}
	}

	return nil
}

func (tc *transactionCoordinator) checkGasProvidedByMiniBlockInReceiverShard(
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
) error {
	var err error
	var gasProvidedByTxInReceiverShard uint64
	gasProvidedByMiniBlockInReceiverShard := uint64(0)

	for _, txHash := range miniBlock.TxHashes {
		txHandler, ok := mapHashTx[string(txHash)]
		if !ok {
			log.Debug("missing transaction in checkGasProvidedByMiniBlockInReceiverShard ", "type", miniBlock.Type, "txHash", txHash)
			return process.ErrMissingTransaction
		}

		_, txTypeDstShard := tc.txTypeHandler.ComputeTransactionType(txHandler)
		moveBalanceGasLimit := tc.economicsFee.ComputeGasLimit(txHandler)
		if txTypeDstShard == process.MoveBalance {
			gasProvidedByTxInReceiverShard = moveBalanceGasLimit
		} else {
			gasProvidedByTxInReceiverShard, err = core.SafeSubUint64(txHandler.GetGasLimit(), moveBalanceGasLimit)
			if err != nil {
				return err
			}
		}

		gasProvidedByMiniBlockInReceiverShard, err = core.SafeAddUint64(gasProvidedByMiniBlockInReceiverShard, gasProvidedByTxInReceiverShard)
		if err != nil {
			return err
		}
	}

	// the max gas limit to be compared with, should be the maximum value between max gas limit per mini block and max gas limit per tx,
	// as the mini blocks with only one tx inside could have gas limit higher than gas limit per mini block but lower or equal than max gas limit per tx.
	// This is done to accept at least one tx in each mini block, if the tx gas limit respects the max gas limit per tx, even if its gas limit is higher than gas limit per mini block.
	if gasProvidedByMiniBlockInReceiverShard > core.MaxUint64(tc.economicsFee.MaxGasLimitPerMiniBlockForSafeCrossShard(), tc.economicsFee.MaxGasLimitPerTx()) {
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

	if tc.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
		scheduledGasAndFees := tc.scheduledTxsExecutionHandler.GetScheduledGasAndFees()
		totalMaxAccumulatedFees.Add(totalMaxAccumulatedFees, scheduledGasAndFees.AccumulatedFees)
		totalMaxDeveloperFees.Add(totalMaxDeveloperFees, scheduledGasAndFees.DeveloperFees)
	}

	if len(body.MiniBlocks) != len(header.GetMiniBlockHeaderHandlers()) {
		log.Warn("transactionCoordinator.verifyFees: num of mini blocks and mini blocks headers does not match", "num of mb", len(body.MiniBlocks), "num of mbh", len(header.GetMiniBlockHeaderHandlers()))
		return process.ErrNumOfMiniBlocksAndMiniBlocksHeadersMismatch
	}

	for index, miniBlock := range body.MiniBlocks {
		if miniBlock.Type == block.PeerBlock {
			continue
		}
		if tc.enableEpochsHandler.IsFlagEnabled(common.ScheduledMiniBlocksFlag) {
			miniBlockHeader := header.GetMiniBlockHeaderHandlers()[index]
			if miniBlockHeader.GetProcessingType() == int32(block.Processed) {
				log.Debug("transactionCoordinator.verifyFees: do not verify fees for mini block executed as scheduled in previous block", "mb hash", miniBlockHeader.GetHash())
				continue
			}
		}

		maxAccumulatedFeesFromMiniBlock, maxDeveloperFeesFromMiniBlock, err := tc.getMaxAccumulatedAndDeveloperFees(
			header.GetMiniBlockHeaderHandlers()[index],
			miniBlock,
			mapMiniBlockTypeAllTxs[miniBlock.Type],
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
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
	miniBlock *block.MiniBlock,
	mapHashTx map[string]data.TransactionHandler,
) (*big.Int, *big.Int, error) {
	maxAccumulatedFeesFromMiniBlock := big.NewInt(0)
	maxDeveloperFeesFromMiniBlock := big.NewInt(0)

	pi, err := tc.getIndexesOfLastTxProcessed(miniBlock, miniBlockHeaderHandler)
	if err != nil {
		return nil, nil, err
	}

	indexOfFirstTxToBeProcessed := pi.indexOfLastTxProcessed + 1
	err = process.CheckIfIndexesAreOutOfBound(indexOfFirstTxToBeProcessed, pi.indexOfLastTxProcessedByProposer, miniBlock)
	if err != nil {
		return nil, nil, err
	}

	for index := indexOfFirstTxToBeProcessed; index <= pi.indexOfLastTxProcessedByProposer; index++ {
		txHash := miniBlock.TxHashes[index]
		txHandler, ok := mapHashTx[string(txHash)]
		if !ok {
			log.Debug("missing transaction in getMaxAccumulatedFeesAndDeveloperFees ", "type", miniBlock.Type, "txHash", txHash)
			return nil, nil, process.ErrMissingTransaction
		}

		maxAccumulatedFeesFromTx := core.SafeMul(txHandler.GetGasLimit(), txHandler.GetGasPrice())
		maxAccumulatedFeesFromMiniBlock.Add(maxAccumulatedFeesFromMiniBlock, maxAccumulatedFeesFromTx)

		maxDeveloperFeesFromTx := core.GetIntTrimmedPercentageOfValue(maxAccumulatedFeesFromTx, tc.economicsFee.DeveloperPercentage())
		maxDeveloperFeesFromMiniBlock.Add(maxDeveloperFeesFromMiniBlock, maxDeveloperFeesFromTx)
	}

	return maxAccumulatedFeesFromMiniBlock, maxDeveloperFeesFromMiniBlock, nil
}

func (tc *transactionCoordinator) getIndexesOfLastTxProcessed(
	miniBlock *block.MiniBlock,
	miniBlockHeaderHandler data.MiniBlockHeaderHandler,
) (*processedIndexes, error) {

	miniBlockHash, err := core.CalculateHash(tc.marshalizer, tc.hasher, miniBlock)
	if err != nil {
		return nil, err
	}

	pi := &processedIndexes{}

	processedMiniBlockInfo, _ := tc.processedMiniBlocksTracker.GetProcessedMiniBlockInfo(miniBlockHash)
	pi.indexOfLastTxProcessed = processedMiniBlockInfo.IndexOfLastTxProcessed
	pi.indexOfLastTxProcessedByProposer = miniBlockHeaderHandler.GetIndexOfLastTxProcessed()

	return pi, nil
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
	if check.IfNil(arguments.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	err := core.CheckHandlerCompatibility(arguments.EnableEpochsHandler, []core.EnableEpochFlag{
		common.ScheduledMiniBlocksFlag,
		common.MiniBlockPartialExecutionFlag,
		common.BlockGasAndFeesReCheckFlag,
	})
	if err != nil {
		return err
	}
	if check.IfNil(arguments.ScheduledTxsExecutionHandler) {
		return process.ErrNilScheduledTxsExecutionHandler
	}
	if check.IfNil(arguments.DoubleTransactionsDetector) {
		return process.ErrNilDoubleTransactionsDetector
	}
	if check.IfNil(arguments.ProcessedMiniBlocksTracker) {
		return process.ErrNilProcessedMiniBlocksTracker
	}
	if check.IfNil(arguments.TxExecutionOrderHandler) {
		return process.ErrNilTxExecutionOrderHandler
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

// AddTxsFromMiniBlocks adds transactions from given mini blocks needed by the current block
func (tc *transactionCoordinator) AddTxsFromMiniBlocks(miniBlocks block.MiniBlockSlice) {
	for _, mb := range miniBlocks {
		preProc := tc.getPreProcessor(mb.Type)
		if check.IfNil(preProc) {
			log.Warn("transactionCoordinator.AddTxsFromMiniBlocks: preProc is nil", "blockType", mb.Type)
			continue
		}

		preProc.AddTxsFromMiniBlocks(block.MiniBlockSlice{mb})
	}
}

// AddTransactions adds the given transactions to the preprocessor
func (tc *transactionCoordinator) AddTransactions(txs []data.TransactionHandler, blockType block.Type) {
	preProc := tc.getPreProcessor(blockType)
	if check.IfNil(preProc) {
		log.Warn("transactionCoordinator.AddTransactions preProc is nil", "blockType", blockType)
		return
	}

	preProc.AddTransactions(txs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (tc *transactionCoordinator) IsInterfaceNil() bool {
	return tc == nil
}
