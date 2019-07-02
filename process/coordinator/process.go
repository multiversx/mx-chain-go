package coordinator

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type transactionCoordinator struct {
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
	miniBlockPool    storage.Cacher

	mutPreprocessor sync.RWMutex
	txPreprocessors map[block.Type]process.PreProcessor

	mutInterimProcessors sync.RWMutex
	interimProcessors    map[block.Type]process.IntermediateTransactionHandler

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
	tc.txPreprocessors = make(map[block.Type]process.PreProcessor)
	tc.interimProcessors = make(map[block.Type]process.IntermediateTransactionHandler)

	keys := preProcessors.Keys()
	for _, value := range keys {
		preproc, err := preProcessors.Get(value)
		if err != nil {
			return nil, err
		}
		tc.txPreprocessors[value] = preproc
	}

	keys = interProcessors.Keys()
	for _, value := range keys {
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
		go func() {
			preproc := tc.getPreprocessor(key)
			if preproc == nil {
				wg.Done()
				return
			}
			requestedTxs := preproc.RequestBlockTransactions(value)

			tc.mutRequestedTxs.Lock()
			tc.requestedTxs[key] = requestedTxs
			tc.mutRequestedTxs.Unlock()

			wg.Done()
		}()
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
		go func() {
			preproc := tc.getPreprocessor(key)
			if preproc == nil {
				wg.Done()

				return
			}

			err := preproc.IsDataPrepared(value, haveTime)
			if err != nil {
				log.Debug(err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}()
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
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func() {
			preproc := tc.getPreprocessor(key)
			if preproc == nil {
				wg.Done()
				return
			}

			err := preproc.SaveTxBlockToStorage(value)
			if err != nil {
				log.Debug(err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()

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
		go func() {
			restoredMbs := make(map[int][]byte)

			preproc := tc.getPreprocessor(key)
			if preproc == nil {
				wg.Done()
				return
			}

			restoredTxs, restoredMbs, err := preproc.RestoreTxBlockIntoPools(value, tc.miniBlockPool)
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
		}()
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
		go func() {
			preproc := tc.getPreprocessor(key)
			if preproc == nil {
				wg.Done()
				return
			}

			err := preproc.RemoveTxBlockFromPools(value, tc.miniBlockPool)
			if err != nil {
				log.Debug(err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	return errFound
}

// ProcessBlockTransaction processes transactions and updates state tries
func (tc *transactionCoordinator) ProcessBlockTransaction(body block.Body, round uint32, haveTime func() time.Duration) error {
	separatedBodies := tc.separateBodyByType(body)

	var errFound error
	errMutex := sync.Mutex{}

	// TODO: think if it is good in parallel or it is needed in sequences
	wg := sync.WaitGroup{}
	wg.Add(len(separatedBodies))

	for key, value := range separatedBodies {
		go func() {
			preproc := tc.getPreprocessor(key)
			if preproc == nil {
				wg.Done()
				return
			}

			err := preproc.ProcessBlockTransactions(value, round, haveTime)
			if err != nil {
				log.Debug(err.Error())

				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	return errFound
}

// CreateMbsAndProcessCrossShardTransactionsDstMe creates miniblocks and processes cross shard transaction
// with destination of current shard
func (tc *transactionCoordinator) CreateMbsAndProcessCrossShardTransactionsDstMe(
	hdr data.HeaderHandler,
	maxTxRemaining uint32,
	round uint32,
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

		preproc := tc.getPreprocessor(miniBlock.Type)
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
	}

	allMBsProcessed := nrMBprocessed == len(crossMiniBlockHashes)
	return miniBlocks, nrTxAdded, allMBsProcessed
}

// CreateMbsAndProcessTransactionsFromMe creates miniblocks and processes transactions from pool
func (tc *transactionCoordinator) CreateMbsAndProcessTransactionsFromMe(maxTxRemaining uint32, round uint32, haveTime func() bool) block.MiniBlockSlice {
	txPreProc := tc.getPreprocessor(block.TxBlock)
	if txPreProc == nil {
		return nil
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	addedTxs := 0
	for i := 0; i < int(tc.shardCoordinator.NumberOfShards()); i++ {
		remainingSpace := int(maxTxRemaining) - addedTxs
		if remainingSpace <= 0 {
			break
		}

		miniBlock, err := txPreProc.CreateAndProcessMiniBlock(tc.shardCoordinator.SelfId(), uint32(i), remainingSpace, haveTime, round)
		if err != nil {
			break
		}

		if len(miniBlock.TxHashes) > 0 {
			addedTxs += len(miniBlock.TxHashes)
			miniBlocks = append(miniBlocks, miniBlock)
		}
	}

	interMBs := tc.processAddedInterimTransactions()
	if len(interMBs) > 0 {
		miniBlocks = append(miniBlocks, interMBs...)
	}

	miniBlocks = append(miniBlocks, tc.processAddedInterimTransactions()...)
	return miniBlocks
}

func (tc *transactionCoordinator) processAddedInterimTransactions() block.MiniBlockSlice {
	miniBlocks := make(block.MiniBlockSlice, 0)

	tc.mutInterimProcessors.RLock()

	resMutex := sync.Mutex{}
	// TODO: think if it is good in parallel or it is needed in sequences
	wg := sync.WaitGroup{}
	wg.Add(len(tc.interimProcessors))

	for _, interimProc := range tc.interimProcessors {
		go func() {
			currMbs := interimProc.CreateAllInterMiniBlocks()
			resMutex.Lock()
			for _, value := range currMbs {
				miniBlocks = append(miniBlocks, value)
			}
			resMutex.Unlock()
			wg.Done()
		}()
	}

	wg.Wait()
	tc.mutInterimProcessors.RUnlock()

	return miniBlocks
}

// CreateBlockStarted initializes necessary data for preprocessors at block create or block process
func (tc *transactionCoordinator) CreateBlockStarted() {
	tc.mutPreprocessor.RLock()
	for _, value := range tc.txPreprocessors {
		value.CreateBlockStarted()
	}
	tc.mutPreprocessor.RUnlock()
}

func (tc *transactionCoordinator) getPreprocessor(blockType block.Type) process.PreProcessor {
	tc.mutPreprocessor.RLock()
	preprocessor, exists := tc.txPreprocessors[blockType]
	tc.mutPreprocessor.RUnlock()

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

// CreateMarshalizedData creates marshalized data for broadcasting
func (tc *transactionCoordinator) CreateMarshalizedData(body block.Body) (map[uint32]block.MiniBlockSlice, map[uint32][][]byte) {
	mrsTxs := make(map[uint32][][]byte)
	bodies := make(map[uint32]block.MiniBlockSlice)

	for i := 0; i < len(body); i++ {
		miniblock := body[i]
		receiverShardId := miniblock.ReceiverShardID
		if receiverShardId == tc.shardCoordinator.SelfId() { // not taking into account miniblocks for current shard
			continue
		}
		preproc := tc.getPreprocessor(miniblock.Type)
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
			mrsTxs[receiverShardId] = append(mrsTxs[receiverShardId], currMrsTxs...)
		}
	}

	return bodies, mrsTxs
}

// GetAllCurrentUsedTxs returns the cached transaction data for current round
func (tc *transactionCoordinator) GetAllCurrentUsedTxs(blockType block.Type) map[string]data.TransactionHandler {
	tc.mutPreprocessor.RLock()
	defer tc.mutPreprocessor.RUnlock()

	if _, ok := tc.txPreprocessors[blockType]; !ok {
		return nil
	}

	return tc.txPreprocessors[blockType].GetAllCurrentUsedTxs()
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

	preproc := tc.getPreprocessor(miniBlock.Type)
	if preproc == nil {
		return
	}

	_ = preproc.RequestTransactionsForMiniBlock(miniBlock)
}

// processMiniBlockComplete - all transactions must be processed together, otherwise error
func (tc *transactionCoordinator) processCompleteMiniBlock(
	preproc process.PreProcessor,
	miniBlock *block.MiniBlock,
	round uint32,
	haveTime func() bool,
) error {

	snapshot := tc.accounts.JournalLen()
	err := preproc.ProcessMiniBlock(miniBlock, haveTime, round)
	if err != nil {
		log.Error(err.Error())
		errAccountState := tc.accounts.RevertToSnapshot(snapshot)
		if errAccountState != nil {
			// TODO: evaluate if reloading the trie from disk will might solve the problem
			log.Error(errAccountState.Error())
		}

		return err
	}

	return nil
}

func (tc *transactionCoordinator) VerifyCreatedBlockTransactions(body block.Body) error {
	tc.mutInterimProcessors.RLock()

	errMutex := sync.Mutex{}
	var errFound error
	// TODO: think if it is good in parallel or it is needed in sequences
	wg := sync.WaitGroup{}
	wg.Add(len(tc.interimProcessors))

	for _, interimProc := range tc.interimProcessors {
		go func() {
			err := interimProc.VerifyInterMiniBlocks(body)
			if err != nil {
				errMutex.Lock()
				errFound = err
				errMutex.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Wait()
	tc.mutInterimProcessors.RUnlock()

	return errFound
}
