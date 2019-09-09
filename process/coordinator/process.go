package coordinator

import (
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type transactionCoordinator struct {
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
	miniBlockPool    storage.Cacher

	mutPreProcessor sync.RWMutex
	txPreProcessors map[block.Type]process.PreProcessor
	keysTxPreProcs  []block.Type

	mutInterimProcessors sync.RWMutex
	interimProcessors    map[block.Type]process.IntermediateTransactionHandler
	keysInterimProcs     []block.Type

	mutRequestedTxs sync.RWMutex
	requestedTxs    map[block.Type]int

	onRequestMiniBlock func(shardId uint32, mbHash []byte)
}

var log = logger.DefaultLogger()

// NewTransactionCoordinator creates a transaction coordinator to run and coordinate preprocessors and processors
func NewTransactionCoordinator(
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	dataPool dataRetriever.PoolsHolder,
	requestHandler process.RequestHandler,
	preProcessors process.PreProcessorsContainer,
	interProcessors process.IntermediateProcessorContainer,
) (*transactionCoordinator, error) {
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if requestHandler == nil {
		return nil, process.ErrNilRequestHandler
	}
	if interProcessors == nil {
		return nil, process.ErrNilIntermediateProcessorContainer
	}
	if preProcessors == nil {
		return nil, process.ErrNilPreProcessorsContainer
	}

	tc := &transactionCoordinator{
		shardCoordinator: shardCoordinator,
		accounts:         accounts,
	}

	tc.miniBlockPool = dataPool.MiniBlocks()
	if tc.miniBlockPool == nil {
		return nil, process.ErrNilMiniBlockPool
	}
	tc.miniBlockPool.RegisterHandler(tc.receivedMiniBlock)

	tc.onRequestMiniBlock = requestHandler.RequestMiniBlock
	tc.requestedTxs = make(map[block.Type]int)
	tc.txPreProcessors = make(map[block.Type]process.PreProcessor)
	tc.interimProcessors = make(map[block.Type]process.IntermediateTransactionHandler)

	tc.keysTxPreProcs = preProcessors.Keys()
	sort.Slice(tc.keysTxPreProcs, func(i, j int) bool {
		return tc.keysTxPreProcs[i] < tc.keysTxPreProcs[j]
	})
	for _, value := range tc.keysTxPreProcs {
		preProc, err := preProcessors.Get(value)
		if err != nil {
			return nil, err
		}
		tc.txPreProcessors[value] = preProc
	}

	tc.keysInterimProcs = interProcessors.Keys()
	sort.Slice(tc.keysInterimProcs, func(i, j int) bool {
		return tc.keysInterimProcs[i] < tc.keysInterimProcs[j]
	})
	for _, value := range tc.keysInterimProcs {
		interProc, err := interProcessors.Get(value)
		if err != nil {
			return nil, err
		}
		tc.interimProcessors[value] = interProc
	}

	return tc, nil
}

// separateBodyByType creates a map of bodies according to type
func (tc *transactionCoordinator) separateBodyByType(body block.Body) map[block.Type]block.Body {
	separatedBodies := make(map[block.Type]block.Body)

	for i := 0; i < len(body); i++ {
		mb := body[i]

		if separatedBodies[mb.Type] == nil {
			separatedBodies[mb.Type] = block.Body{}
		}

		separatedBodies[mb.Type] = append(separatedBodies[mb.Type], mb)
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
func (tc *transactionCoordinator) RequestBlockTransactions(body block.Body) {
	separatedBodies := tc.separateBodyByType(body)

	tc.initRequestedTxs()

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if preproc == nil {
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
			if preproc == nil {
				wg.Done()

				return
			}

			err := preproc.IsDataPrepared(requestedTxs, haveTime)
			if err != nil {
				log.Debug(err.Error())

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

// SaveBlockDataToStorage saves the data from block body into storage units
func (tc *transactionCoordinator) SaveBlockDataToStorage(body block.Body) error {
	separatedBodies := tc.separateBodyByType(body)

	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	// Length of body types + another go routine for the intermediate transactions
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if preproc == nil {
				wg.Done()
				return
			}

			err := preproc.SaveTxBlockToStorage(blockBody)
			if err != nil {
				log.Debug(err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}

			wg.Done()
		}(key, value)
	}

	wg.Wait()

	intermediatePreproc := tc.getInterimProcessor(block.SmartContractResultBlock)
	if intermediatePreproc == nil {
		return errFound
	}

	err := intermediatePreproc.SaveCurrentIntermediateTxToStorage()
	if err != nil {
		log.Debug(err.Error())

		errMutex.Lock()
		errFound = err
		errMutex.Unlock()
	}

	return errFound
}

// RestoreBlockDataFromStorage restores block data from storage to pool
func (tc *transactionCoordinator) RestoreBlockDataFromStorage(body block.Body) (int, map[int][][]byte, error) {
	separatedBodies := tc.separateBodyByType(body)

	var errFound error
	localMutex := sync.Mutex{}
	totalRestoredTx := 0
	restoredMbHashes := make(map[int][][]byte)

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody block.Body) {
			restoredMbs := make(map[int][]byte)

			preproc := tc.getPreProcessor(blockType)
			if preproc == nil {
				wg.Done()
				return
			}

			restoredTxs, restoredMbs, err := preproc.RestoreTxBlockIntoPools(blockBody, tc.miniBlockPool)
			if err != nil {
				log.Debug(err.Error())

				localMutex.Lock()
				errFound = err
				localMutex.Unlock()
			}

			localMutex.Lock()
			totalRestoredTx += restoredTxs

			for shId, mbHash := range restoredMbs {
				restoredMbHashes[shId] = append(restoredMbHashes[shId], mbHash)
			}

			localMutex.Unlock()

			wg.Done()
		}(key, value)
	}

	wg.Wait()

	return totalRestoredTx, restoredMbHashes, errFound
}

// RemoveBlockDataFromPool deletes block data from pools
func (tc *transactionCoordinator) RemoveBlockDataFromPool(body block.Body) error {
	separatedBodies := tc.separateBodyByType(body)

	var errFound error
	errMutex := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func(blockType block.Type, blockBody block.Body) {
			preproc := tc.getPreProcessor(blockType)
			if preproc == nil {
				wg.Done()
				return
			}

			err := preproc.RemoveTxBlockFromPools(blockBody, tc.miniBlockPool)
			if err != nil {
				log.Debug(err.Error())

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
	body block.Body,
	round uint64,
	haveTime func() time.Duration,
) error {
	separatedBodies := tc.separateBodyByType(body)

	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	for _, blockType := range tc.keysTxPreProcs {
		if separatedBodies[blockType] == nil {
			continue
		}

		preproc := tc.getPreProcessor(blockType)
		if preproc == nil {
			return process.ErrMissingPreProcessor
		}

		err := preproc.ProcessBlockTransactions(separatedBodies[blockType], round, haveTime)
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
	maxTxRemaining uint32,
	maxMbRemaining uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, bool) {
	miniBlocks := make(block.MiniBlockSlice, 0)
	nrTxAdded := uint32(0)
	nrMBprocessed := 0

	if hdr == nil || hdr.IsInterfaceNil() {
		return miniBlocks, nrTxAdded, true
	}

	crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(tc.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		if !haveTime() {
			break
		}

		if hdr.GetMiniBlockProcessed([]byte(key)) {
			nrMBprocessed++
			continue
		}

		miniVal, _ := tc.miniBlockPool.Peek([]byte(key))
		if miniVal == nil {
			go tc.onRequestMiniBlock(senderShardId, []byte(key))
			continue
		}

		miniBlock, ok := miniVal.(*block.MiniBlock)
		if !ok {
			continue
		}

		preproc := tc.getPreProcessor(miniBlock.Type)
		if preproc == nil {
			continue
		}

		// overflow would happen if processing would continue
		txOverFlow := nrTxAdded+uint32(len(miniBlock.TxHashes)) > maxTxRemaining
		if txOverFlow {
			return miniBlocks, nrTxAdded, false
		}

		requestedTxs := preproc.RequestTransactionsForMiniBlock(*miniBlock)
		if requestedTxs > 0 {
			continue
		}

		err := tc.processCompleteMiniBlock(preproc, miniBlock, round, haveTime)
		if err != nil {
			continue
		}

		// all txs processed, add to processed miniblocks
		miniBlocks = append(miniBlocks, miniBlock)
		nrTxAdded = nrTxAdded + uint32(len(miniBlock.TxHashes))
		nrMBprocessed++

		mbOverFlow := uint32(len(miniBlocks)) >= maxMbRemaining
		if mbOverFlow {
			return miniBlocks, nrTxAdded, false
		}
	}

	allMBsProcessed := nrMBprocessed == len(crossMiniBlockHashes)
	return miniBlocks, nrTxAdded, allMBsProcessed
}

// CreateMbsAndProcessTransactionsFromMe creates miniblocks and processes transactions from pool
func (tc *transactionCoordinator) CreateMbsAndProcessTransactionsFromMe(
	maxTxSpaceRemained uint32,
	maxMbSpaceRemained uint32,
	round uint64,
	haveTime func() bool,
) block.MiniBlockSlice {

	txPreProc := tc.getPreProcessor(block.TxBlock)
	if txPreProc == nil {
		return nil
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	txSpaceRemained := int(maxTxSpaceRemained)

	newMBAdded := true
	for newMBAdded {
		newMBAdded = false

		for shardId := uint32(0); shardId < tc.shardCoordinator.NumberOfShards(); shardId++ {
			if txSpaceRemained <= 0 {
				break
			}

			mbSpaceRemained := int(maxMbSpaceRemained) - len(miniBlocks)
			if mbSpaceRemained <= 0 {
				break
			}

			miniBlock, err := txPreProc.CreateAndProcessMiniBlock(
				tc.shardCoordinator.SelfId(),
				shardId,
				txSpaceRemained,
				haveTime,
				round)
			if err != nil {
				continue
			}

			if len(miniBlock.TxHashes) > 0 {
				txSpaceRemained -= len(miniBlock.TxHashes)
				miniBlocks = append(miniBlocks, miniBlock)
				newMBAdded = true
			}
		}
	}

	interMBs := tc.processAddedInterimTransactions()
	if len(interMBs) > 0 {
		miniBlocks = append(miniBlocks, interMBs...)
	}

	rewardsMBs := tc.createRewardsMiniBlocks()
	if len(interMBs) > 0 {
		miniBlocks = append(miniBlocks, rewardsMBs...)
	}
	
	return miniBlocks
}

func (tc *transactionCoordinator) createRewardsMiniBlocks() block.MiniBlockSlice {
	// add rewards transactions to separate miniBlocks
	interimProc := tc.getInterimProcessor(block.RewardsBlock)
	if interimProc == nil {
		return nil
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	rewardsMbs := interimProc.CreateAllInterMiniBlocks()
	for key, mb := range rewardsMbs {
		mb.ReceiverShardID = key
		mb.SenderShardID = tc.shardCoordinator.SelfId()
		mb.Type = block.RewardsBlock

		miniBlocks = append(miniBlocks, mb)
	}

	return miniBlocks
}

func (tc *transactionCoordinator) processAddedInterimTransactions() block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0)

	// processing has to be done in order, as the order of different type of transactions over the same account is strict
	for _, blockType := range tc.keysInterimProcs {
		if blockType == block.RewardsBlock {
			// this has to be processed last
			continue
		}

		interimProc := tc.getInterimProcessor(blockType)
		if interimProc == nil {
			// this will never be reached as keysInterimProcs are the actual keys from the interimMap
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
func (tc *transactionCoordinator) CreateMarshalizedData(body block.Body) (map[uint32]block.MiniBlockSlice, map[string][][]byte) {
	mrsTxs := make(map[string][][]byte)
	bodies := make(map[uint32]block.MiniBlockSlice)

	for i := 0; i < len(body); i++ {
		miniblock := body[i]
		receiverShardId := miniblock.ReceiverShardID
		if receiverShardId == tc.shardCoordinator.SelfId() { // not taking into account miniblocks for current shard
			continue
		}

		broadcastTopic, err := createBroadcastTopic(tc.shardCoordinator, receiverShardId, miniblock.Type)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		preproc := tc.getPreProcessor(miniblock.Type)
		if preproc == nil {
			continue
		}

		bodies[receiverShardId] = append(bodies[receiverShardId], miniblock)

		currMrsTxs, err := preproc.CreateMarshalizedData(miniblock.TxHashes)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		if len(currMrsTxs) > 0 {
			mrsTxs[broadcastTopic] = append(mrsTxs[broadcastTopic], currMrsTxs...)
		}

		interimProc := tc.getInterimProcessor(miniblock.Type)
		if interimProc == nil {
			continue
		}

		currMrsInterTxs, err := interimProc.CreateMarshalizedData(miniblock.TxHashes)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		if len(currMrsInterTxs) > 0 {
			mrsTxs[broadcastTopic] = append(mrsTxs[broadcastTopic], currMrsInterTxs...)
		}
	}

	return bodies, mrsTxs
}

// GetAllCurrentUsedTxs returns the cached transaction data for current round
func (tc *transactionCoordinator) GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler {
	tc.mutPreProcessor.RLock()
	defer tc.mutPreProcessor.RUnlock()

	if _, ok := tc.txPreProcessors[blockType]; !ok {
		return nil
	}

	return tc.txPreProcessors[blockType].GetAllCurrentUsedTxs()
}

// RequestMiniBlocks request miniblocks if missing
func (tc *transactionCoordinator) RequestMiniBlocks(header data.HeaderHandler) {
	if header == nil || header.IsInterfaceNil() {
		return
	}

	crossMiniBlockHashes := header.GetMiniBlockHeadersWithDst(tc.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		obj, _ := tc.miniBlockPool.Peek([]byte(key))
		if obj == nil {
			go tc.onRequestMiniBlock(senderShardId, []byte(key))
		}
	}
}

// receivedMiniBlock is a callback function when a new miniblock was received
// it will further ask for missing transactions
func (tc *transactionCoordinator) receivedMiniBlock(miniBlockHash []byte) {
	val, ok := tc.miniBlockPool.Peek(miniBlockHash)
	if !ok {
		return
	}

	miniBlock, ok := val.(block.MiniBlock)
	if !ok {
		return
	}

	preproc := tc.getPreProcessor(miniBlock.Type)
	if preproc == nil {
		return
	}

	_ = preproc.RequestTransactionsForMiniBlock(miniBlock)
}

// processMiniBlockComplete - all transactions must be processed together, otherwise error
func (tc *transactionCoordinator) processCompleteMiniBlock(
	preproc process.PreProcessor,
	miniBlock *block.MiniBlock,
	round uint64,
	haveTime func() bool,
) error {

	snapshot := tc.accounts.JournalLen()
	err := preproc.ProcessMiniBlock(miniBlock, haveTime, round)
	if err != nil {
		log.Debug(err.Error())
		errAccountState := tc.accounts.RevertToSnapshot(snapshot)
		if errAccountState != nil {
			// TODO: evaluate if reloading the trie from disk will might solve the problem
			log.Error(errAccountState.Error())
		}

		return err
	}

	return nil
}

// VerifyCreatedBlockTransactions checks whether the created transactions are the same as the one proposed
func (tc *transactionCoordinator) VerifyCreatedBlockTransactions(body block.Body) error {
	tc.mutInterimProcessors.RLock()
	defer tc.mutInterimProcessors.RUnlock()
	errMutex := sync.Mutex{}
	var errFound error
	// TODO: think if it is good in parallel or it is needed in sequences
	wg := sync.WaitGroup{}
	wg.Add(len(tc.interimProcessors))

	for key, interimProc := range tc.interimProcessors {
		if key == block.RewardsBlock {
			// this has to be processed last
			wg.Done()
			continue
		}

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

	interimProc := tc.getInterimProcessor(block.RewardsBlock)
	if interimProc == nil {
		return nil
	}

	return interimProc.VerifyInterMiniBlocks(body)
}
