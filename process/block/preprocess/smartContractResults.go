package preprocess

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type scrShardInfo struct {
	senderShardID   uint32
	receiverShardID uint32
}

type scrInfo struct {
	scr *smartContractResult.SmartContractResult
	*scrShardInfo
	has bool
}

type scrHashesInfo struct {
	txHashes        [][]byte
	receiverShardID uint32
}

type smartContractResults struct {
	*basePreProcess
	chRcvAllScrs                 chan bool
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte)
	missingScrs                  int
	mutScrsForBlock              sync.RWMutex
	scrForBlock                  map[string]*scrInfo
	scrPool                      dataRetriever.ShardedDataCacherNotifier
	storage                      dataRetriever.StorageService
	scrProcessor                 process.SmartContractResultProcessor
	shardCoordinator             sharding.Coordinator
	accounts                     state.AccountsAdapter
}

// NewSmartContractResultPreprocessor creates a new smartContractResult preprocessor object
func NewSmartContractResultPreprocessor(
	scrDataPool dataRetriever.ShardedDataCacherNotifier,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	scrProcessor process.SmartContractResultProcessor,
	shardCoordinator sharding.Coordinator,
	accounts state.AccountsAdapter,
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte),
) (*smartContractResults, error) {

	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if scrDataPool == nil {
		return nil, process.ErrNilScrDataPool
	}
	if store == nil {
		return nil, process.ErrNilScrStorage
	}
	if scrProcessor == nil {
		return nil, process.ErrNilTxProcessor
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}
	if accounts == nil {
		return nil, process.ErrNilAccountsAdapter
	}
	if onRequestSmartContractResult == nil {
		return nil, process.ErrNilRequestHandler
	}

	bpp := &basePreProcess{hasher: hasher, marshalizer: marshalizer}

	scr := &smartContractResults{
		basePreProcess:               bpp,
		shardCoordinator:             shardCoordinator,
		storage:                      store,
		scrPool:                      scrDataPool,
		onRequestSmartContractResult: onRequestSmartContractResult,
		scrProcessor:                 scrProcessor,
		accounts:                     accounts,
	}

	scr.chRcvAllScrs = make(chan bool)
	scr.scrPool.RegisterHandler(scr.receivedSmartContractResult)
	scr.scrForBlock = make(map[string]*scrInfo)

	return scr, nil
}

// waitForScrHashes waits for a call whether all the requested smartContractResults appeared
func (scr *smartContractResults) waitForScrHashes(waitTime time.Duration) error {
	select {
	case <-scr.chRcvAllScrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// IsDataPrepared returns non error if all the requested smartContractResults arrived and were saved into the pool
func (scr *smartContractResults) IsDataPrepared(requestedScrs int, haveTime func() time.Duration) error {
	if requestedScrs > 0 {
		log.Info(fmt.Sprintf("requested %d missing scr\n", requestedScrs))
		err := scr.waitForScrHashes(haveTime())
		scr.mutScrsForBlock.RLock()
		missingScrs := scr.missingScrs
		scr.mutScrsForBlock.RUnlock()
		log.Info(fmt.Sprintf("received %d missing scr\n", requestedScrs-missingScrs))
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveTxBlockFromPools removes smartContractResults and miniblocks from associated pools
func (scr *smartContractResults) RemoveTxBlockFromPools(body block.Body, miniBlockPool storage.Cacher) error {
	if body == nil {
		return process.ErrNilTxBlockBody
	}

	err := scr.removeDataFromPools(body, miniBlockPool, scr.scrPool, block.SmartContractResultBlock)

	return err
}

// RestoreTxBlockIntoPools restores the smartContractResults and miniblocks to associated pools
func (scr *smartContractResults) RestoreTxBlockIntoPools(
	body block.Body,
	miniBlockHashes map[int][]byte,
	miniBlockPool storage.Cacher,
) (int, error) {
	if miniBlockPool == nil {
		return 0, process.ErrNilMiniBlockPool
	}

	scrRestored := 0
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}

		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		scrBuff, err := scr.storage.GetAll(dataRetriever.SmartContractResultUnit, miniBlock.TxHashes)
		if err != nil {
			return scrRestored, err
		}

		for txHash, txBuff := range scrBuff {
			tx := smartContractResult.SmartContractResult{}
			err = scr.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return scrRestored, err
			}

			scr.scrPool.AddData([]byte(txHash), &tx, strCache)
		}

		buff, err := scr.marshalizer.Marshal(miniBlock)
		if err != nil {
			return scrRestored, err
		}

		miniBlockHash := scr.hasher.Compute(string(buff))
		miniBlockPool.Put(miniBlockHash, miniBlock)
		if miniBlock.SenderShardID != scr.shardCoordinator.SelfId() {
			miniBlockHashes[i] = miniBlockHash
		}

		scrRestored += len(miniBlock.TxHashes)
	}

	return scrRestored, nil
}

// ProcessBlockSmartContractResults processes all the smartContractResult from the block.Body, updates the state
func (scr *smartContractResults) ProcessBlockTransactions(body block.Body, round uint32, haveTime func() time.Duration) error {
	// basic validation already done in interceptors
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}
		if miniBlock.ReceiverShardID != scr.shardCoordinator.SelfId() {
			continue
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			scr.mutScrsForBlock.RLock()
			txInfo := scr.scrForBlock[string(txHash)]
			scr.mutScrsForBlock.RUnlock()
			if txInfo == nil || txInfo.scr == nil {
				return process.ErrMissingTransaction
			}

			err := scr.processSmartContractResult(
				txHash,
				txInfo.scr,
				round,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
			)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

// SaveTxBlockToStorage saves smartContractResults from body into storage
func (scr *smartContractResults) SaveTxBlockToStorage(body block.Body) error {
	for i := 0; i < len(body); i++ {
		miniBlock := (body)[i]
		if miniBlock.Type != block.SmartContractResultBlock {
			continue
		}

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]

			scr.mutScrsForBlock.RLock()
			txInfo := scr.scrForBlock[string(txHash)]
			scr.mutScrsForBlock.RUnlock()

			if txInfo == nil || txInfo.scr == nil {
				return process.ErrMissingTransaction
			}

			buff, err := scr.marshalizer.Marshal(txInfo.scr)
			if err != nil {
				return err
			}

			errNotCritical := scr.storage.Put(dataRetriever.SmartContractResultUnit, txHash, buff)
			if errNotCritical != nil {
				log.Error(errNotCritical.Error())
			}
		}
	}

	return nil
}

// getSmartContractResultFromPool gets the smartContractResult from a given shard id and a given smartContractResult hash
func (scr *smartContractResults) getSmartContractResultFromPool(
	senderShardID uint32,
	destShardID uint32,
	txHash []byte,
) *smartContractResult.SmartContractResult {
	strCache := process.ShardCacherIdentifier(senderShardID, destShardID)
	txStore := scr.scrPool.ShardDataStore(strCache)
	if txStore == nil {
		log.Error(process.ErrNilStorage.Error())
		return nil
	}

	val, ok := txStore.Peek(txHash)
	if !ok {
		log.Debug(process.ErrTxNotFound.Error())
		return nil
	}

	tx, ok := val.(*smartContractResult.SmartContractResult)
	if !ok {
		log.Error(process.ErrInvalidTxInPool.Error())
		return nil
	}

	return tx
}

// receivedSmartContractResult is a call back function which is called when a new smartContractResult
// is added in the smartContractResult pool
func (scr *smartContractResults) receivedSmartContractResult(txHash []byte) {
	scr.mutScrsForBlock.Lock()
	if scr.missingScrs > 0 {
		txInfoForHash := scr.scrForBlock[string(txHash)]
		if txInfoForHash != nil &&
			txInfoForHash.scrShardInfo != nil &&
			!txInfoForHash.has {
			tx := scr.getSmartContractResultFromPool(txInfoForHash.senderShardID, txInfoForHash.receiverShardID, txHash)
			if tx != nil {
				scr.scrForBlock[string(txHash)].scr = tx
				scr.scrForBlock[string(txHash)].has = true
				scr.missingScrs--
			}
		}

		missingScrs := scr.missingScrs
		scr.mutScrsForBlock.Unlock()

		if missingScrs == 0 {
			scr.chRcvAllScrs <- true
		}
	} else {
		scr.mutScrsForBlock.Unlock()
	}
}

// CreateBlockStarted cleans the local cache map for processed/created smartContractResults at this round
func (scr *smartContractResults) CreateBlockStarted() {
	scr.mutScrsForBlock.Lock()
	scr.scrForBlock = make(map[string]*scrInfo)
	scr.mutScrsForBlock.Unlock()
}

// RequestBlockSmartContractResults request for smartContractResults if missing from a block.Body
func (scr *smartContractResults) RequestBlockTransactions(body block.Body) int {
	scr.CreateBlockStarted()

	requestedScrs := 0
	missingScrsForShards := scr.computeMissingAndExistingScrsForShards(body)

	scr.mutScrsForBlock.Lock()
	for senderShardID, scrHashesInfo := range missingScrsForShards {
		txShardInfo := &scrShardInfo{senderShardID: senderShardID, receiverShardID: scrHashesInfo.receiverShardID}
		for _, txHash := range scrHashesInfo.txHashes {
			scr.scrForBlock[string(txHash)] = &scrInfo{scr: nil, scrShardInfo: txShardInfo, has: false}
		}
	}
	scr.mutScrsForBlock.Unlock()

	for senderShardID, scrHashesInfo := range missingScrsForShards {
		requestedScrs += len(scrHashesInfo.txHashes)
		scr.onRequestSmartContractResult(senderShardID, scrHashesInfo.txHashes)
	}

	return requestedScrs
}

// computeMissingAndExistingScrsForShards calculates what smartContractResults are available and what are missing from block.Body
func (scr *smartContractResults) computeMissingAndExistingScrsForShards(body block.Body) map[uint32]*scrHashesInfo {
	missingScrsForShard := make(map[uint32]*scrHashesInfo)
	scr.missingScrs = 0

	scr.mutScrsForBlock.Lock()
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		txShardInfo := &scrShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}
		txHashes := make([][]byte, 0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := scr.getSmartContractResultFromPool(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHash)

			if tx == nil {
				txHashes = append(txHashes, txHash)
				scr.missingScrs++
			} else {
				scr.scrForBlock[string(txHash)] = &scrInfo{scr: tx, scrShardInfo: txShardInfo, has: true}
			}
		}

		if len(txHashes) > 0 {
			missingScrsForShard[miniBlock.SenderShardID] = &scrHashesInfo{
				txHashes:        txHashes,
				receiverShardID: miniBlock.ReceiverShardID,
			}
		}
	}
	scr.mutScrsForBlock.Unlock()

	return missingScrsForShard
}

// processAndRemoveBadSmartContractResults processed smartContractResults, if scr are with error it removes them from pool
func (scr *smartContractResults) processSmartContractResult(
	smartContractResultHash []byte,
	smartContractResult *smartContractResult.SmartContractResult,
	round uint32,
	sndShardId uint32,
	dstShardId uint32,
) error {

	err := scr.scrProcessor.ProcessSmartContractResult(smartContractResult)
	if err != nil {
		return err
	}

	txShardInfo := &scrShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	scr.mutScrsForBlock.Lock()
	scr.scrForBlock[string(smartContractResultHash)] = &scrInfo{scr: smartContractResult, scrShardInfo: txShardInfo, has: true}
	scr.mutScrsForBlock.Unlock()

	return nil
}

// RequestSmartContractResultsForMiniBlock requests missing smartContractResults for a certain miniblock
func (scr *smartContractResults) RequestTransactionsForMiniBlock(mb block.MiniBlock) int {
	missingScrsForMiniBlock := scr.computeMissingScrsForMiniBlock(mb)
	scr.onRequestSmartContractResult(mb.SenderShardID, missingScrsForMiniBlock)

	return len(missingScrsForMiniBlock)
}

// computeMissingScrsForMiniBlock computes missing smartContractResults for a certain miniblock
func (scr *smartContractResults) computeMissingScrsForMiniBlock(mb block.MiniBlock) [][]byte {
	missingSmartContractResults := make([][]byte, 0)
	if mb.Type != block.SmartContractResultBlock {
		return missingSmartContractResults
	}

	for _, txHash := range mb.TxHashes {
		tx := scr.getSmartContractResultFromPool(mb.SenderShardID, mb.ReceiverShardID, txHash)
		if tx == nil {
			missingSmartContractResults = append(missingSmartContractResults, txHash)
		}
	}

	return missingSmartContractResults
}

// getAllScrsFromMiniBlock gets all the smartContractResults from a miniblock into a new structure
func (scr *smartContractResults) getAllScrsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*smartContractResult.SmartContractResult, [][]byte, error) {

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := scr.scrPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilScrDataPool
	}

	// verify if all smartContractResult exists
	smartContractResults := make([]*smartContractResult.SmartContractResult, 0)
	txHashes := make([][]byte, 0)
	for _, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, _ := txCache.Peek(txHash)
		if tmp == nil {
			return nil, nil, process.ErrNilSmartContractResult
		}

		tx, ok := tmp.(*smartContractResult.SmartContractResult)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}
		txHashes = append(txHashes, txHash)
		smartContractResults = append(smartContractResults, tx)
	}

	return smartContractResults, txHashes, nil
}

// CreateAndProcessMiniBlock creates the miniblock from storage and processes the smartContractResults added into the miniblock
func (scr *smartContractResults) CreateAndProcessMiniBlock(sndShardId, dstShardId uint32, spaceRemained int, haveTime func() bool, round uint32) (*block.MiniBlock, error) {
	return nil, nil
}

// ProcessMiniBlock processes all the smartContractResults from a and saves the processed smartContractResults in local cache complete miniblock
func (scr *smartContractResults) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint32) error {
	if miniBlock.Type != block.SmartContractResultBlock {
		return process.ErrWrongTypeAssertion
	}

	miniBlockScrs, miniBlockTxHashes, err := scr.getAllScrsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	for index := range miniBlockScrs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return err
		}

		err = scr.scrProcessor.ProcessSmartContractResult(miniBlockScrs[index])
		if err != nil {
			return err
		}
	}

	txShardInfo := &scrShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	scr.mutScrsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		scr.scrForBlock[string(txHash)] = &scrInfo{scr: miniBlockScrs[index], scrShardInfo: txShardInfo, has: true}
	}
	scr.mutScrsForBlock.Unlock()

	return nil
}

// CreateMarshalizedData marshalizes smartContractResults and creates and saves them into a new structure
func (scr *smartContractResults) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsScrs := make([][]byte, 0)
	for _, txHash := range txHashes {
		scr.mutScrsForBlock.RLock()
		txInfo := scr.scrForBlock[string(txHash)]
		scr.mutScrsForBlock.RUnlock()

		if txInfo == nil || txInfo.scr == nil {
			continue
		}

		txMrs, err := scr.marshalizer.Marshal(txInfo.scr)
		if err != nil {
			return nil, process.ErrMarshalWithoutSuccess
		}
		mrsScrs = append(mrsScrs, txMrs)
	}

	return mrsScrs, nil
}

// GetAllCurrentUsedScrs returns all the smartContractResults used at current creation / processing
func (scr *smartContractResults) GetAllCurrentUsedTxs() map[string]data.TransactionHandler {
	scrPool := make(map[string]data.TransactionHandler)

	scr.mutScrsForBlock.RLock()
	for txHash, txInfo := range scr.scrForBlock {
		scrPool[txHash] = txInfo.scr
	}
	scr.mutScrsForBlock.RUnlock()

	return scrPool
}
