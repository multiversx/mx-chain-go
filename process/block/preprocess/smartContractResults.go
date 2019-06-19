package preprocess

import (
	"fmt"
	"sort"
	"sync"
	"time"

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
	chRcvAllScrs                 chan bool
	onRequestSmartContractResult func(shardID uint32, txHashes [][]byte)
	missingScrs                  int
	mutScrsForBlock              sync.RWMutex
	scrForBlock                  map[string]*txInfo
	txPool                       dataRetriever.ShardedDataCacherNotifier
	storage                      dataRetriever.StorageService
	hasher                       hashing.Hasher
	marshalizer                  marshal.Marshalizer
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
		return nil, process.ErrNilSmartContractResultPool
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

	scr := &smartContractResults{
		hasher:                       hasher,
		marshalizer:                  marshalizer,
		shardCoordinator:             shardCoordinator,
		storage:                      store,
		txPool:                       scrDataPool,
		onRequestSmartContractResult: onRequestSmartContractResult,
		scrProcessor:                 scrProcessor,
		accounts:                     accounts,
	}

	scr.chRcvAllScrs = make(chan bool)
	scr.txPool.RegisterHandler(scr.receivedSmartContractResult)
	scr.scrForBlock = make(map[string]*txInfo)

	return scr, nil
}

// waitForTxHashes waits for a call whether all the requested smartContractResults appeared
func (scr *smartContractResults) waitForTxHashes(waitTime time.Duration) error {
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
		err := scr.waitForTxHashes(haveTime())
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
	if miniBlockPool == nil {
		return process.ErrNilMiniBlockPool
	}

	for i := 0; i < len(body); i++ {
		currentMiniBlock := body[i]
		strCache := process.ShardCacherIdentifier(currentMiniBlock.SenderShardID, currentMiniBlock.ReceiverShardID)
		scr.txPool.RemoveSetOfDataFromPool(currentMiniBlock.TxHashes, strCache)

		buff, err := scr.marshalizer.Marshal(currentMiniBlock)
		if err != nil {
			return err
		}

		miniBlockHash := scr.hasher.Compute(string(buff))
		miniBlockPool.Remove(miniBlockHash)
	}

	return nil
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

			scr.txPool.AddData([]byte(txHash), &tx, strCache)
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
func (scr *smartContractResults) ProcessBlockSmartContractResults(body block.Body, round uint32, haveTime func() time.Duration) error {
	// basic validation already done in interceptors
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			scr.mutScrsForBlock.RLock()
			txInfo := scr.scrForBlock[string(txHash)]
			scr.mutScrsForBlock.RUnlock()
			if txInfo == nil || txInfo.tx == nil {
				return process.ErrMissingSmartContractResult
			}

			err := scr.processAndRemoveBadSmartContractResult(
				txHash,
				txInfo.tx,
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
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]

			scr.mutScrsForBlock.RLock()
			txInfo := scr.scrForBlock[string(txHash)]
			scr.mutScrsForBlock.RUnlock()

			if txInfo == nil || txInfo.tx == nil {
				return process.ErrMissingSmartContractResult
			}

			buff, err := scr.marshalizer.Marshal(txInfo.tx)
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
	txStore := scr.txPool.ShardDataStore(strCache)
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
			txInfoForHash.txShardInfo != nil &&
			!txInfoForHash.has {
			tx := scr.getSmartContractResultFromPool(txInfoForHash.senderShardID, txInfoForHash.receiverShardID, txHash)
			if tx != nil {
				scr.scrForBlock[string(txHash)].tx = tx
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
	scr.scrForBlock = make(map[string]*txInfo)
	scr.mutScrsForBlock.Unlock()
}

// RequestBlockSmartContractResults request for smartContractResults if missing from a block.Body
func (scr *smartContractResults) RequestBlockSmartContractResults(body block.Body) int {
	scr.CreateBlockStarted()

	requestedScrs := 0
	missingScrsForShards := scr.computeMissingAndExistingScrsForShards(body)

	scr.mutScrsForBlock.Lock()
	for senderShardID, scrHashesInfo := range missingScrsForShards {
		txShardInfo := &txShardInfo{senderShardID: senderShardID, receiverShardID: scrHashesInfo.receiverShardID}
		for _, txHash := range scrHashesInfo.txHashes {
			scr.scrForBlock[string(txHash)] = &txInfo{tx: nil, txShardInfo: txShardInfo, has: false}
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
		txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}
		txHashes := make([][]byte, 0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := scr.getSmartContractResultFromPool(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHash)

			if tx == nil {
				txHashes = append(txHashes, txHash)
				scr.missingScrs++
			} else {
				scr.scrForBlock[string(txHash)] = &txInfo{tx: tx, txShardInfo: txShardInfo, has: true}
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
func (scr *smartContractResults) processAndRemoveBadSmartContractResult(
	smartContractResultHash []byte,
	smartContractResult *smartContractResult.SmartContractResult,
	round uint32,
	sndShardId uint32,
	dstShardId uint32,
) error {

	_, err := scr.scrProcessor.ProcessSmartContractResult(smartContractResult, round)
	if err == process.ErrLowerNonceInSmartContractResult ||
		err == process.ErrInsufficientFunds {
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		scr.txPool.RemoveData(smartContractResultHash, strCache)
	}

	if err != nil {
		return err
	}

	txShardInfo := &txShardInfo{senderShardID: sndShardId, receiverShardID: dstShardId}
	scr.mutScrsForBlock.Lock()
	scr.scrForBlock[string(smartContractResultHash)] = &txInfo{tx: smartContractResult, txShardInfo: txShardInfo, has: true}
	scr.mutScrsForBlock.Unlock()

	return nil
}

// RequestSmartContractResultsForMiniBlock requests missing smartContractResults for a certain miniblock
func (scr *smartContractResults) RequestSmartContractResultsForMiniBlock(mb block.MiniBlock) int {
	missingScrsForMiniBlock := scr.computeMissingScrsForMiniBlock(mb)
	scr.onRequestSmartContractResult(mb.SenderShardID, missingScrsForMiniBlock)

	return len(missingScrsForMiniBlock)
}

// computeMissingScrsForMiniBlock computes missing smartContractResults for a certain miniblock
func (scr *smartContractResults) computeMissingScrsForMiniBlock(mb block.MiniBlock) [][]byte {
	missingSmartContractResults := make([][]byte, 0)
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
	txCache := scr.txPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilSmartContractResultPool
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
	strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
	txStore := scr.txPool.ShardDataStore(strCache)

	timeBefore := time.Now()
	orderedTxes, orderedTxHashes, err := scr.getScrs(txStore)
	timeAfter := time.Now()

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after ordered %d scr in %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))
		return nil, nil
	}

	log.Info(fmt.Sprintf("time elapsed to ordered %d scr: %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))

	if err != nil {
		log.Debug(fmt.Sprintf("when trying to order scr: %s", err.Error()))
	}

	miniBlock := &block.MiniBlock{}
	miniBlock.SenderShardID = sndShardId
	miniBlock.ReceiverShardID = dstShardId
	miniBlock.TxHashes = make([][]byte, 0)
	log.Info(fmt.Sprintf("creating mini blocks has been started: have %d scr in pool for shard id %d\n", len(orderedTxes), miniBlock.ReceiverShardID))

	addedScrs := 0
	for index := range orderedTxes {
		if !haveTime() {
			break
		}

		snapshot := scr.accounts.JournalLen()

		// execute smartContractResult to change the trie root hash
		err := scr.processAndRemoveBadSmartContractResult(
			orderedTxHashes[index],
			orderedTxes[index],
			round,
			miniBlock.SenderShardID,
			miniBlock.ReceiverShardID,
		)

		if err != nil {
			log.Error(err.Error())
			err = scr.accounts.RevertToSnapshot(snapshot)
			if err != nil {
				log.Error(err.Error())
			}
			continue
		}

		miniBlock.TxHashes = append(miniBlock.TxHashes, orderedTxHashes[index])
		addedScrs++

		if addedScrs >= spaceRemained { // max smartContractResults count in one block was reached
			log.Info(fmt.Sprintf("max scr accepted in one block is reached: added %d scr from %d scr\n", len(miniBlock.TxHashes), len(orderedTxes)))
			return miniBlock, nil
		}
	}

	return miniBlock, nil
}

// ProcessMiniBlock processes all the smartContractResults from a and saves the processed smartContractResults in local cache complete miniblock
func (scr *smartContractResults) ProcessMiniBlock(miniBlock *block.MiniBlock, haveTime func() bool, round uint32) error {
	miniBlockScrs, miniBlockTxHashes, err := scr.getAllScrsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	for index := range miniBlockScrs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			return err
		}

		_, err = scr.scrProcessor.ProcessSmartContractResult(miniBlockScrs[index], round)
		if err != nil {
			return err
		}
	}

	txShardInfo := &txShardInfo{senderShardID: miniBlock.SenderShardID, receiverShardID: miniBlock.ReceiverShardID}

	scr.mutScrsForBlock.Lock()
	for index, txHash := range miniBlockTxHashes {
		scr.scrForBlock[string(txHash)] = &txInfo{tx: miniBlockScrs[index], txShardInfo: txShardInfo, has: true}
	}
	scr.mutScrsForBlock.Unlock()

	return nil
}

// SortTxByNonce sort smartContractResults according to nonces
func SortTxByNonce(txShardStore storage.Cacher) ([]*smartContractResult.SmartContractResult, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	smartContractResults := make([]*smartContractResult.SmartContractResult, 0)
	txHashes := make([][]byte, 0)

	mTxHashes := make(map[uint64][][]byte)
	mSmartContractResults := make(map[uint64][]*smartContractResult.SmartContractResult)

	nonces := make([]uint64, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Peek(key)
		if val == nil {
			continue
		}

		tx, ok := val.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		if mTxHashes[tx.Nonce] == nil {
			nonces = append(nonces, tx.Nonce)
			mTxHashes[tx.Nonce] = make([][]byte, 0)
			mSmartContractResults[tx.Nonce] = make([]*smartContractResult.SmartContractResult, 0)
		}

		mTxHashes[tx.Nonce] = append(mTxHashes[tx.Nonce], key)
		mSmartContractResults[tx.Nonce] = append(mSmartContractResults[tx.Nonce], tx)
	}

	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i] < nonces[j]
	})

	for _, nonce := range nonces {
		keys := mTxHashes[nonce]

		for idx, key := range keys {
			txHashes = append(txHashes, key)
			smartContractResults = append(smartContractResults, mSmartContractResults[nonce][idx])
		}
	}

	return smartContractResults, txHashes, nil
}

// CreateMarshalizedData marshalizes smartContractResults and creates and saves them into a new structure
func (scr *smartContractResults) CreateMarshalizedData(txHashes [][]byte) ([][]byte, error) {
	mrsScrs := make([][]byte, 0)
	for _, txHash := range txHashes {
		scr.mutScrsForBlock.RLock()
		txInfo := scr.scrForBlock[string(txHash)]
		scr.mutScrsForBlock.RUnlock()

		if txInfo == nil || txInfo.tx == nil {
			continue
		}

		txMrs, err := scr.marshalizer.Marshal(txInfo.tx)
		if err != nil {
			return nil, process.ErrMarshalWithoutSuccess
		}
		mrsScrs = append(mrsScrs, txMrs)
	}

	return mrsScrs, nil
}

// getScrs gets all the available smartContractResults from the pool
func (scr *smartContractResults) getScrs(txShardStore storage.Cacher) ([]*smartContractResult.SmartContractResult, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	smartContractResults := make([]*smartContractResult.SmartContractResult, 0)
	txHashes := make([][]byte, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Peek(key)
		if val == nil {
			continue
		}

		tx, ok := val.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}

		txHashes = append(txHashes, key)
		smartContractResults = append(smartContractResults, tx)
	}

	return smartContractResults, txHashes, nil
}

// GetAllCurrentUsedScrs returns all the smartContractResults used at current creation / processing
func (scr *smartContractResults) GetAllCurrentUsedScrs() map[string]*smartContractResult.SmartContractResult {
	txPool := make(map[string]*smartContractResult.SmartContractResult)

	scr.mutScrsForBlock.RLock()
	for txHash, txInfo := range scr.scrForBlock {
		txPool[txHash] = txInfo.tx
	}
	scr.mutScrsForBlock.RUnlock()

	return txPool
}
