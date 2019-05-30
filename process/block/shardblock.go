package block

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var txCounterMutex = sync.RWMutex{}
var txsCurrentBlockProcessed = 0
var txsTotalProcessed = 0

const maxTransactionsInBlock = 15000

// shardProcessor implements shardProcessor interface and actually it tries to execute block
type shardProcessor struct {
	*baseProcessor
	dataPool             dataRetriever.PoolsHolder
	txProcessor          process.TransactionProcessor
	blocksTracker        process.BlocksTracker
	chRcvAllTxs          chan bool
	onRequestTransaction func(shardID uint32, txHashes [][]byte)
	mutRequestedTxHashes sync.RWMutex
	requestedTxHashes    map[string]bool
	mutCrossTxsForBlock  sync.RWMutex
	crossTxsForBlock     map[string]*transaction.Transaction
	onRequestMiniBlock   func(shardId uint32, mbHash []byte)
}

// NewShardProcessor creates a new shardProcessor object
func NewShardProcessor(
	dataPool dataRetriever.PoolsHolder,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txProcessor process.TransactionProcessor,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	forkDetector process.ForkDetector,
	blocksTracker process.BlocksTracker,
	requestTransactionHandler func(shardId uint32, txHashes [][]byte),
	requestMiniBlockHandler func(shardId uint32, miniblockHash []byte),
) (*shardProcessor, error) {

	err := checkProcessorNilParameters(
		accounts,
		forkDetector,
		hasher,
		marshalizer,
		store,
		shardCoordinator)
	if err != nil {
		return nil, err
	}

	if dataPool == nil {
		return nil, process.ErrNilDataPoolHolder
	}
	if txProcessor == nil {
		return nil, process.ErrNilTxProcessor
	}
	if blocksTracker == nil {
		return nil, process.ErrNilBlocksTracker
	}
	if requestTransactionHandler == nil {
		return nil, process.ErrNilTransactionHandler
	}
	if requestMiniBlockHandler == nil {
		return nil, process.ErrNilMiniBlocksRequestHandler
	}

	base := &baseProcessor{
		accounts:         accounts,
		forkDetector:     forkDetector,
		hasher:           hasher,
		marshalizer:      marshalizer,
		store:            store,
		shardCoordinator: shardCoordinator,
	}

	sp := shardProcessor{
		baseProcessor: base,
		dataPool:      dataPool,
		txProcessor:   txProcessor,
		blocksTracker: blocksTracker,
	}

	sp.chRcvAllTxs = make(chan bool)
	sp.onRequestTransaction = requestTransactionHandler

	transactionPool := sp.dataPool.Transactions()
	if transactionPool == nil {
		return nil, process.ErrNilTransactionPool
	}
	transactionPool.RegisterHandler(sp.receivedTransaction)

	sp.onRequestMiniBlock = requestMiniBlockHandler
	sp.requestedTxHashes = make(map[string]bool)
	sp.crossTxsForBlock = make(map[string]*transaction.Transaction)

	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return nil, process.ErrNilMetaBlockPool
	}
	metaBlockPool.RegisterHandler(sp.receivedMetaBlock)

	miniBlockPool := sp.dataPool.MiniBlocks()
	if miniBlockPool == nil {
		return nil, process.ErrNilMiniBlockPool
	}
	miniBlockPool.RegisterHandler(sp.receivedMiniBlock)

	return &sp, nil
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (sp *shardProcessor) ProcessBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	err := sp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	log.Info(fmt.Sprintf("Total txs in pool: %d\n", sp.getNrTxsWithDst(header.ShardId)))

	requestedTxs := sp.requestBlockTransactions(body)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	if requestedTxs > 0 {
		log.Info(fmt.Sprintf("requested %d missing txs\n", requestedTxs))
		err = sp.waitForTxHashes(haveTime())
		log.Info(fmt.Sprintf("received %d missing txs\n", requestedTxs-len(sp.requestedTxHashes)))
		if err != nil {
			return err
		}
	}

	if sp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		if err != nil {
			sp.RevertAccountState()
		}
	}()

	err = sp.processBlockTransactions(body, header.Round, haveTime)
	if err != nil {
		return err
	}

	if !sp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateMissmatch
		return err
	}

	go sp.checkAndRequestIfMetaHeadersMissing(header.GetRound(), header.GetNonce())

	return nil
}

func (sp *shardProcessor) checkAndRequestIfMetaHeadersMissing(round uint32, nonce uint64) error {
	orderedMetaBlocks, err := sp.getOrderedMetaBlocks(round)
	if err != nil {
		return err
	}

	sortedHdrs := make([]data.HeaderHandler, 0)
	for i := 0; i < len(orderedMetaBlocks); i++ {
		hdr, ok := orderedMetaBlocks[i].hdr.(*block.MetaBlock)
		if !ok {
			continue
		}
		sortedHdrs = append(sortedHdrs, hdr)
	}

	err = sp.requestHeadersIfMissing(sortedHdrs, sharding.MetachainShardId, nonce)
	if err != nil {
		log.Info(err.Error())
	}

	return nil
}

// removeTxBlockFromPools removes transactions and miniblocks from associated pools
func (sp *shardProcessor) removeTxBlockFromPools(body block.Body) error {
	if body == nil {
		return process.ErrNilTxBlockBody
	}

	transactionPool := sp.dataPool.Transactions()
	if transactionPool == nil {
		return process.ErrNilTransactionPool
	}

	miniBlockPool := sp.dataPool.MiniBlocks()
	if miniBlockPool == nil {
		return process.ErrNilMiniBlockPool
	}

	for i := 0; i < len(body); i++ {
		currentMiniBlock := body[i]
		strCache := process.ShardCacherIdentifier(currentMiniBlock.SenderShardID, currentMiniBlock.ReceiverShardID)
		transactionPool.RemoveSetOfDataFromPool(currentMiniBlock.TxHashes, strCache)

		buff, err := sp.marshalizer.Marshal(currentMiniBlock)
		if err != nil {
			return err
		}

		miniBlockHash := sp.hasher.Compute(string(buff))
		miniBlockPool.Remove(miniBlockHash)
	}

	return nil
}

// RestoreBlockIntoPools restores the TxBlock and MetaBlock into associated pools
func (sp *shardProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if bodyHandler == nil {
		return process.ErrNilTxBlockBody
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	miniBlockHashes := make(map[int][]byte, 0)
	err := sp.restoreTxBlockIntoPools(body, miniBlockHashes)
	if err != nil {
		return err
	}

	err = sp.restoreMetaBlockIntoPool(miniBlockHashes)
	if err != nil {
		return err
	}

	return nil
}

func (sp *shardProcessor) restoreTxBlockIntoPools(
	body block.Body,
	miniBlockHashes map[int][]byte,
) error {

	transactionPool := sp.dataPool.Transactions()
	if transactionPool == nil {
		return process.ErrNilTransactionPool
	}

	miniBlockPool := sp.dataPool.MiniBlocks()
	if miniBlockPool == nil {
		return process.ErrNilMiniBlockPool
	}

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		strCache := process.ShardCacherIdentifier(miniBlock.SenderShardID, miniBlock.ReceiverShardID)
		txsBuff, err := sp.store.GetAll(dataRetriever.TransactionUnit, miniBlock.TxHashes)
		if err != nil {
			return err
		}

		for txHash, txBuff := range txsBuff {
			tx := transaction.Transaction{}
			err = sp.marshalizer.Unmarshal(&tx, txBuff)
			if err != nil {
				return err
			}

			transactionPool.AddData([]byte(txHash), &tx, strCache)
		}

		buff, err := sp.marshalizer.Marshal(miniBlock)
		if err != nil {
			return err
		}

		miniBlockHash := sp.hasher.Compute(string(buff))
		miniBlockPool.Put(miniBlockHash, miniBlock)
		if miniBlock.SenderShardID != sp.shardCoordinator.SelfId() {
			miniBlockHashes[i] = miniBlockHash
		}
		txCounterMutex.Lock()
		txsTotalProcessed -= len(miniBlock.TxHashes)
		txCounterMutex.Unlock()
	}

	return nil
}

func (sp *shardProcessor) restoreMetaBlockIntoPool(miniBlockHashes map[int][]byte) error {
	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return process.ErrNilMetaBlockPool
	}

	for _, metaBlockKey := range metaBlockPool.Keys() {
		if len(miniBlockHashes) == 0 {
			break
		}
		metaBlock, _ := metaBlockPool.Peek(metaBlockKey)
		if metaBlock == nil {
			return process.ErrNilMetaBlockHeader
		}

		hdr, _ := metaBlock.(data.HeaderHandler)
		if hdr == nil {
			return process.ErrWrongTypeAssertion
		}

		crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for key := range miniBlockHashes {
			_, ok := crossMiniBlockHashes[string(miniBlockHashes[key])]
			if !ok {
				continue
			}

			hdr.SetMiniBlockProcessed(miniBlockHashes[key], false)
			delete(miniBlockHashes, key)
		}
	}

	return nil
}

// CreateBlockBody creates a a list of miniblocks by filling them with transactions out of the transactions pools
// as long as the transactions limit for the block has not been reached and there is still time to add transactions
func (sp *shardProcessor) CreateBlockBody(round uint32, haveTime func() bool) (data.BodyHandler, error) {
	miniBlocks, err := sp.createMiniBlocks(sp.shardCoordinator.NumberOfShards(), maxTransactionsInBlock, round, haveTime)

	if err != nil {
		return nil, err
	}

	return miniBlocks, nil
}

func (sp *shardProcessor) processBlockTransactions(body block.Body, round uint32, haveTime func() time.Duration) error {
	// basic validation already done in interceptors
	txPool := sp.dataPool.Transactions()

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			txHash := miniBlock.TxHashes[j]
			tx := sp.getTransactionFromPool(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHash)
			err := sp.processAndRemoveBadTransaction(
				txHash,
				tx,
				txPool,
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

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (sp *shardProcessor) CommitBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	var err error
	defer func() {
		if err != nil {
			sp.RevertAccountState()
		}
	}()

	err = checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	buff, err := sp.marshalizer.Marshal(headerHandler)
	if err != nil {
		return err
	}

	headerHash := sp.hasher.Compute(string(buff))
	err = sp.store.Put(dataRetriever.BlockHeaderUnit, headerHash, buff)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	for i := 0; i < len(body); i++ {
		buff, err = sp.marshalizer.Marshal(body[i])
		if err != nil {
			return err
		}

		miniBlockHash := sp.hasher.Compute(string(buff))
		err = sp.store.Put(dataRetriever.MiniBlockUnit, miniBlockHash, buff)
		if err != nil {
			return err
		}
	}

	headerNoncePool := sp.dataPool.HeadersNonces()
	if headerNoncePool == nil {
		err = process.ErrNilDataPoolHolder
		return err
	}

	_ = headerNoncePool.Put(headerHandler.GetNonce(), headerHash)

	for i := 0; i < len(body); i++ {
		miniBlock := (body)[i]
		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := sp.getTransactionFromPool(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHash)
			if tx == nil {
				err = process.ErrMissingTransaction
				return err
			}

			buff, err = sp.marshalizer.Marshal(tx)
			if err != nil {
				return err
			}

			err = sp.store.Put(dataRetriever.TransactionUnit, txHash, buff)
			if err != nil {
				return err
			}
		}
	}

	_, err = sp.accounts.Commit()
	if err != nil {
		return err
	}

	sp.blocksTracker.AddBlock(header)

	log.Info(fmt.Sprintf("shardBlock with nonce %d and hash %s has been committed successfully\n",
		header.Nonce,
		core.ToB64(headerHash)))

	errNotCritical := sp.removeTxBlockFromPools(body)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	processedMetaHdrs, errNotCritical := sp.getProcessedMetaBlocksFromPool(body)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	err = sp.saveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	if err != nil {
		return err
	}

	errNotCritical = sp.removeProcessedMetablocksFromPool(processedMetaHdrs)
	if errNotCritical != nil {
		return err
	}

	errNotCritical = sp.forkDetector.AddHeader(header, headerHash, process.BHProcessed)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	err = chainHandler.SetCurrentBlockBody(body)
	if err != nil {
		return err
	}

	err = chainHandler.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	chainHandler.SetCurrentBlockHeaderHash(headerHash)

	// write data to log
	go sp.displayShardBlock(header, body)

	return nil
}

// removeMetaBlockFromPool removes meta blocks from associated pool
func (sp *shardProcessor) getProcessedMetaBlocksFromPool(body block.Body) ([]data.HeaderHandler, error) {
	if body == nil {
		return nil, process.ErrNilTxBlockBody
	}

	miniBlockHashes := make(map[int][]byte, 0)
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.SenderShardID == sp.shardCoordinator.SelfId() {
			continue
		}

		buff, err := sp.marshalizer.Marshal(miniBlock)
		if err != nil {
			return nil, err
		}
		mbHash := sp.hasher.Compute(string(buff))
		miniBlockHashes[i] = mbHash
	}

	processedMetaHdrs := make([]data.HeaderHandler, 0)
	for _, metaBlockKey := range sp.dataPool.MetaBlocks().Keys() {
		metaBlock, _ := sp.dataPool.MetaBlocks().Peek(metaBlockKey)
		if metaBlock == nil {
			return processedMetaHdrs, process.ErrNilMetaBlockHeader
		}

		hdr, ok := metaBlock.(*block.MetaBlock)
		if !ok {
			return processedMetaHdrs, process.ErrWrongTypeAssertion
		}

		crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for key := range miniBlockHashes {
			_, ok := crossMiniBlockHashes[string(miniBlockHashes[key])]
			if !ok {
				continue
			}

			hdr.SetMiniBlockProcessed(miniBlockHashes[key], true)
			delete(miniBlockHashes, key)
		}

		processedAll := true
		for key := range crossMiniBlockHashes {
			if !hdr.GetMiniBlockProcessed([]byte(key)) {
				processedAll = false
				break
			}
		}

		if processedAll {
			processedMetaHdrs = append(processedMetaHdrs, hdr)
		}
	}

	return processedMetaHdrs, nil
}

func (sp *shardProcessor) removeProcessedMetablocksFromPool(processedMetaHdrs []data.HeaderHandler) error {
	lastNoterizedMetaHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	// processedMetaHdrs is also sorted
	for i := 0; i < len(processedMetaHdrs); i++ {
		hdr := processedMetaHdrs[i]

		// remove process finished
		if hdr.GetNonce() > lastNoterizedMetaHdr.GetNonce() {
			continue
		}

		errNotCritical := sp.blocksTracker.RemoveNotarisedBlocks(hdr)
		if errNotCritical != nil {
			log.Debug(errNotCritical.Error())
		}

		// metablock was processed and finalized
		buff, err := sp.marshalizer.Marshal(hdr)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		key := sp.hasher.Compute(string(buff))
		err = sp.store.Put(dataRetriever.MetaBlockUnit, key, buff)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
		sp.dataPool.MetaBlocks().Remove(key)
		log.Info(fmt.Sprintf("metablock with nonce %d has been processed completely and removed from pool\n",
			hdr.GetNonce()))
	}

	return nil
}

// getTransactionFromPool gets the transaction from a given shard id and a given transaction hash
func (sp *shardProcessor) getTransactionFromPool(
	senderShardID uint32,
	destShardID uint32,
	txHash []byte,
) *transaction.Transaction {

	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		log.Error(process.ErrNilTransactionPool.Error())
		return nil
	}

	strCache := process.ShardCacherIdentifier(senderShardID, destShardID)
	txStore := txPool.ShardDataStore(strCache)
	if txStore == nil {
		log.Error(process.ErrNilStorage.Error())
		return nil
	}

	val, ok := txStore.Peek(txHash)
	if !ok {
		return nil
	}

	tx, ok := val.(*transaction.Transaction)
	if !ok {
		log.Error(process.ErrInvalidTxInPool.Error())
		return nil
	}

	return tx
}

// receivedTransaction is a call back function which is called when a new transaction
// is added in the transaction pool
func (sp *shardProcessor) receivedTransaction(txHash []byte) {
	sp.mutRequestedTxHashes.Lock()

	if len(sp.requestedTxHashes) > 0 {
		if sp.requestedTxHashes[string(txHash)] {
			delete(sp.requestedTxHashes, string(txHash))
		}

		lenReqTxHashes := len(sp.requestedTxHashes)
		sp.mutRequestedTxHashes.Unlock()

		if lenReqTxHashes == 0 {
			sp.chRcvAllTxs <- true
		}

		return
	}

	sp.mutRequestedTxHashes.Unlock()
}

// receivedMetaBlock is a callback function when a new metablock was received
// upon receiving, it parses the new metablock and requests miniblocks and transactions
// which destination is the current shard
func (sp *shardProcessor) receivedMetaBlock(metaBlockHash []byte) {
	metaBlockCache := sp.dataPool.MetaBlocks()
	if metaBlockCache == nil {
		return
	}

	miniBlockCache := sp.dataPool.MiniBlocks()
	if miniBlockCache == nil {
		return
	}

	metaBlock, ok := metaBlockCache.Peek(metaBlockHash)
	if !ok {
		return
	}

	hdr, ok := metaBlock.(data.HeaderHandler)
	if !ok {
		return
	}

	log.Info(fmt.Sprintf("received metablock with hash %s and nonce %d from network\n",
		core.ToB64(metaBlockHash),
		hdr.GetNonce()))

	lastHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return
	}
	if hdr.GetNonce() < lastHdr.GetNonce() {
		return
	}
	if hdr.GetRound() < lastHdr.GetRound() {
		return
	}

	crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		miniVal, _ := miniBlockCache.Peek([]byte(key))
		if miniVal == nil {
			//TODO: It should be analyzed if launching the next line(request) on go routine is better or not
			go sp.onRequestMiniBlock(senderShardId, []byte(key))
		}
	}
}

// receivedMiniBlock is a callback function when a new miniblock was received
// it will further ask for missing transactions
func (sp *shardProcessor) receivedMiniBlock(miniBlockHash []byte) {
	metaBlockCache := sp.dataPool.MetaBlocks()
	if metaBlockCache == nil {
		return
	}

	miniBlockCache := sp.dataPool.MiniBlocks()
	if miniBlockCache == nil {
		return
	}

	val, ok := miniBlockCache.Peek(miniBlockHash)
	if !ok {
		return
	}

	miniBlock, ok := val.(block.MiniBlock)
	if !ok {
		return
	}

	requestedTxs := make([][]byte, 0)
	for _, txHash := range miniBlock.TxHashes {
		tx := sp.getTransactionFromPool(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHash)
		if tx == nil {
			requestedTxs = append(requestedTxs, txHash)
		}
	}

	sp.onRequestTransaction(miniBlock.SenderShardID, requestedTxs)
}

func (sp *shardProcessor) requestBlockTransactions(body block.Body) int {
	sp.mutRequestedTxHashes.Lock()

	requestedTxs := 0
	missingTxsForShards := sp.computeMissingTxsForShards(body)
	sp.requestedTxHashes = make(map[string]bool)

	for shardId, txHashes := range missingTxsForShards {
		requestedTxs += len(txHashes)
		for _, txHash := range txHashes {
			sp.requestedTxHashes[string(txHash)] = true
		}

		sp.onRequestTransaction(shardId, txHashes)
	}

	sp.mutRequestedTxHashes.Unlock()

	return requestedTxs
}

func (sp *shardProcessor) computeMissingTxsForShards(body block.Body) map[uint32][][]byte {
	missingTxsForShard := make(map[uint32][][]byte)

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		currentShardMissingTransactions := make([][]byte, 0)

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			txHash := miniBlock.TxHashes[j]
			tx := sp.getTransactionFromPool(miniBlock.SenderShardID, miniBlock.ReceiverShardID, txHash)

			if tx == nil {
				currentShardMissingTransactions = append(currentShardMissingTransactions, txHash)
			}
		}

		if len(currentShardMissingTransactions) > 0 {
			missingTxsForShard[miniBlock.SenderShardID] = currentShardMissingTransactions
		}
	}

	return missingTxsForShard
}

func (sp *shardProcessor) processAndRemoveBadTransaction(
	transactionHash []byte,
	transaction *transaction.Transaction,
	txPool dataRetriever.ShardedDataCacherNotifier,
	round uint32,
	sndShardId uint32,
	dstShardId uint32,
) error {

	if txPool == nil {
		return process.ErrNilTransactionPool
	}

	err := sp.txProcessor.ProcessTransaction(transaction, round)
	if err == process.ErrLowerNonceInTransaction ||
		err == process.ErrInsufficientFunds {
		strCache := process.ShardCacherIdentifier(sndShardId, dstShardId)
		txPool.RemoveData(transactionHash, strCache)
	}

	return err
}

func (sp *shardProcessor) requestBlockTransactionsForMiniBlock(mb *block.MiniBlock) int {
	missingTxsForMiniBlock := sp.computeMissingTxsForMiniBlock(mb)
	sp.onRequestTransaction(mb.SenderShardID, missingTxsForMiniBlock)

	return len(missingTxsForMiniBlock)
}

func (sp *shardProcessor) computeMissingTxsForMiniBlock(mb *block.MiniBlock) [][]byte {
	missingTransactions := make([][]byte, 0)
	for _, txHash := range mb.TxHashes {
		tx := sp.getTransactionFromPool(mb.SenderShardID, mb.ReceiverShardID, txHash)
		if tx == nil {
			missingTransactions = append(missingTransactions, txHash)
		}
	}

	return missingTransactions
}

func (sp *shardProcessor) getAllTxsFromMiniBlock(
	mb *block.MiniBlock,
	haveTime func() bool,
) ([]*transaction.Transaction, [][]byte, error) {

	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		return nil, nil, process.ErrNilTransactionPool
	}

	strCache := process.ShardCacherIdentifier(mb.SenderShardID, mb.ReceiverShardID)
	txCache := txPool.ShardDataStore(strCache)
	if txCache == nil {
		return nil, nil, process.ErrNilTransactionPool
	}

	// verify if all transaction exists
	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)
	for _, txHash := range mb.TxHashes {
		if !haveTime() {
			return nil, nil, process.ErrTimeIsOut
		}

		tmp, _ := txCache.Peek(txHash)
		if tmp == nil {
			return nil, nil, process.ErrNilTransaction
		}

		tx, ok := tmp.(*transaction.Transaction)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}
		txHashes = append(txHashes, txHash)
		transactions = append(transactions, tx)
	}

	return transactions, txHashes, nil
}

// processMiniBlockComplete - all transactions must be processed together, otherwise error
func (sp *shardProcessor) processMiniBlockComplete(
	miniBlock *block.MiniBlock,
	round uint32,
	haveTime func() bool,
) error {

	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		return process.ErrNilTransactionPool
	}

	miniBlockTxs, _, err := sp.getAllTxsFromMiniBlock(miniBlock, haveTime)
	if err != nil {
		return err
	}

	snapshot := sp.accounts.JournalLen()
	for index := range miniBlockTxs {
		if !haveTime() {
			err = process.ErrTimeIsOut
			break
		}

		err = sp.txProcessor.ProcessTransaction(miniBlockTxs[index], round)
		if err != nil {
			break
		}
	}
	// all txs from miniblock has to be processed together
	if err != nil {
		log.Error(err.Error())
		errAccountState := sp.accounts.RevertToSnapshot(snapshot)
		if errAccountState != nil {
			// TODO: evaluate if reloading the trie from disk will might solve the problem
			log.Error(errAccountState.Error())
		}
	}
	return err
}

func (sp *shardProcessor) getOrderedMetaBlocks(round uint32) ([]*hashAndHdr, error) {
	metaBlockCache := sp.dataPool.MetaBlocks()
	if metaBlockCache == nil {
		return nil, process.ErrNilMetaBlockPool
	}

	lastHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, err
	}

	orderedMetaBlocks := make([]*hashAndHdr, 0)
	for _, key := range metaBlockCache.Keys() {
		val, _ := metaBlockCache.Peek(key)
		if val == nil {
			continue
		}

		hdr, ok := val.(*block.MetaBlock)
		if !ok {
			continue
		}

		if hdr.GetRound() > round {
			continue
		}
		if hdr.GetRound() < lastHdr.GetRound() {
			continue
		}
		if hdr.GetNonce() < lastHdr.GetNonce() {
			continue
		}

		orderedMetaBlocks = append(orderedMetaBlocks, &hashAndHdr{hdr: hdr, hash: key})
	}

	// sort headers for each shard
	if len(orderedMetaBlocks) == 0 {
		return nil, process.ErrNoNewMetablocks
	}

	sort.Slice(orderedMetaBlocks, func(i, j int) bool {
		return orderedMetaBlocks[i].hdr.GetNonce() < orderedMetaBlocks[j].hdr.GetNonce()
	})

	return orderedMetaBlocks, nil
}

func (sp *shardProcessor) processMiniBlocksFromHeader(
	hdr data.HeaderHandler,
	maxTxRemaining uint32,
	round uint32,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, bool) {
	// verification of hdr and miniblock validity is done before the function is called
	miniBlockCache := sp.dataPool.MiniBlocks()

	miniBlocks := make(block.MiniBlockSlice, 0)
	nrTxAdded := uint32(0)
	nrMBprocessed := 0
	// get mini block hashes which contain cross shard txs with destination in self shard
	crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		if !haveTime() {
			break
		}

		if hdr.GetMiniBlockProcessed([]byte(key)) {
			nrMBprocessed++
			continue
		}

		miniVal, _ := miniBlockCache.Peek([]byte(key))
		if miniVal == nil {
			go sp.onRequestMiniBlock(senderShardId, []byte(key))
			continue
		}

		miniBlock, ok := miniVal.(*block.MiniBlock)
		if !ok {
			continue
		}

		// overflow would happen if processing would continue
		txOverFlow := nrTxAdded+uint32(len(miniBlock.TxHashes)) > maxTxRemaining
		if txOverFlow {
			return miniBlocks, nrTxAdded, false
		}

		requestedTxs := sp.requestBlockTransactionsForMiniBlock(miniBlock)
		if requestedTxs > 0 {
			continue
		}

		err := sp.processMiniBlockComplete(miniBlock, round, haveTime)
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

// full verification through metachain header
func (sp *shardProcessor) createAndProcessCrossMiniBlocksDstMe(
	noShards uint32,
	maxTxInBlock int,
	round uint32,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, error) {

	metaBlockCache := sp.dataPool.MetaBlocks()
	if metaBlockCache == nil {
		return nil, 0, process.ErrNilMetaBlockPool
	}

	miniBlockCache := sp.dataPool.MiniBlocks()
	if miniBlockCache == nil {
		return nil, 0, process.ErrNilMiniBlockPool
	}

	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		return nil, 0, process.ErrNilTransactionPool
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	nrTxAdded := uint32(0)

	orderedMetaBlocks, err := sp.getOrderedMetaBlocks(uint32(round))
	if err != nil {
		return nil, 0, err
	}

	lastMetaHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, 0, err
	}

	// do processing in order
	for i := 0; i < len(orderedMetaBlocks); i++ {
		if !haveTime() {
			log.Info(fmt.Sprintf("time is up after putting %d cross txs with destination to current shard \n", nrTxAdded))
			return miniBlocks, nrTxAdded, nil
		}

		hdr, ok := orderedMetaBlocks[i].hdr.(*block.MetaBlock)
		if !ok {
			continue
		}

		err := sp.isHdrConstructionValid(hdr, lastMetaHdr)
		if err != nil {
			continue
		}

		maxTxRemaining := uint32(maxTxInBlock) - nrTxAdded
		currMBProcessed, currTxsAdded, hdrProcessFinished := sp.processMiniBlocksFromHeader(hdr, maxTxRemaining, round, haveTime)

		// all txs processed, add to processed miniblocks
		miniBlocks = append(miniBlocks, currMBProcessed...)
		nrTxAdded = nrTxAdded + currTxsAdded

		if !hdrProcessFinished {
			return miniBlocks, nrTxAdded, nil
		}

		lastMetaHdr = hdr
	}

	return miniBlocks, nrTxAdded, nil
}

func (sp *shardProcessor) createMiniBlocks(
	noShards uint32,
	maxTxInBlock int,
	round uint32,
	haveTime func() bool,
) (block.Body, error) {

	miniBlocks := make(block.Body, 0)
	sp.mutCrossTxsForBlock.Lock()
	sp.crossTxsForBlock = make(map[string]*transaction.Transaction)
	sp.mutCrossTxsForBlock.Unlock()

	if sp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after entered in createMiniBlocks method\n"))
		return miniBlocks, nil
	}

	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	destMeMiniBlocks, txs, err := sp.createAndProcessCrossMiniBlocksDstMe(noShards, maxTxInBlock, round, haveTime)
	if err != nil {
		log.Info(err.Error())
	}

	if len(destMeMiniBlocks) > 0 {
		miniBlocks = append(miniBlocks, destMeMiniBlocks...)
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up added %d transactions\n", txs))
		return miniBlocks, nil
	}

	if txs > uint32(maxTxInBlock) {
		log.Info(fmt.Sprintf("block is full: added %d transactions\n", txs))
		return miniBlocks, nil
	}

	for i := 0; i < int(noShards); i++ {
		strCache := process.ShardCacherIdentifier(sp.shardCoordinator.SelfId(), uint32(i))
		txStore := txPool.ShardDataStore(strCache)

		timeBefore := time.Now()
		orderedTxes, orderedTxHashes, err := getTxs(txStore)
		timeAfter := time.Now()

		if !haveTime() {
			log.Info(fmt.Sprintf("time is up after ordered %d txs in %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))
			return miniBlocks, nil
		}

		log.Info(fmt.Sprintf("time elapsed to ordered %d txs: %v sec\n", len(orderedTxes), timeAfter.Sub(timeBefore).Seconds()))

		if err != nil {
			log.Debug(fmt.Sprintf("when trying to order txs: %s", err.Error()))
			continue
		}

		miniBlock := block.MiniBlock{}
		miniBlock.SenderShardID = sp.shardCoordinator.SelfId()
		miniBlock.ReceiverShardID = uint32(i)
		miniBlock.TxHashes = make([][]byte, 0)
		log.Info(fmt.Sprintf("creating mini blocks has been started: have %d txs in pool for shard id %d\n", len(orderedTxes), miniBlock.ReceiverShardID))

		for index := range orderedTxes {
			if !haveTime() {
				break
			}

			snapshot := sp.accounts.JournalLen()

			// execute transaction to change the trie root hash
			err := sp.processAndRemoveBadTransaction(
				orderedTxHashes[index],
				orderedTxes[index],
				txPool,
				round,
				miniBlock.SenderShardID,
				miniBlock.ReceiverShardID,
			)

			if err != nil {
				log.Error(err.Error())
				err = sp.accounts.RevertToSnapshot(snapshot)
				if err != nil {
					log.Error(err.Error())
				}
				continue
			}

			sp.mutCrossTxsForBlock.Lock()
			sp.crossTxsForBlock[string(orderedTxHashes[index])] = orderedTxes[index]
			sp.mutCrossTxsForBlock.Unlock()
			miniBlock.TxHashes = append(miniBlock.TxHashes, orderedTxHashes[index])
			txs++

			if txs >= uint32(maxTxInBlock) { // max transactions count in one block was reached
				log.Info(fmt.Sprintf("max txs accepted in one block is reached: added %d txs from %d txs\n", len(miniBlock.TxHashes), len(orderedTxes)))

				if len(miniBlock.TxHashes) > 0 {
					miniBlocks = append(miniBlocks, &miniBlock)
				}

				log.Info(fmt.Sprintf("creating mini blocks has been finished: created %d mini blocks\n", len(miniBlocks)))
				return miniBlocks, nil
			}
		}

		if !haveTime() {
			log.Info(fmt.Sprintf("time is up: added %d txs from %d txs\n", len(miniBlock.TxHashes), len(orderedTxes)))

			if len(miniBlock.TxHashes) > 0 {
				miniBlocks = append(miniBlocks, &miniBlock)
			}

			log.Info(fmt.Sprintf("creating mini blocks has been finished: created %d mini blocks\n", len(miniBlocks)))
			return miniBlocks, nil
		}

		if len(miniBlock.TxHashes) > 0 {
			miniBlocks = append(miniBlocks, &miniBlock)
		}
	}

	log.Info(fmt.Sprintf("creating mini blocks has been finished: created %d mini blocks\n", len(miniBlocks)))
	return miniBlocks, nil
}

// CreateBlockHeader creates a miniblock header list given a block body
func (sp *shardProcessor) CreateBlockHeader(bodyHandler data.BodyHandler, round uint32, haveTime func() bool) (data.HeaderHandler, error) {
	// TODO: add PrevRandSeed and RandSeed when BLS signing is completed
	header := &block.Header{
		MiniBlockHeaders: make([]block.MiniBlockHeader, 0),
		RootHash:         sp.getRootHash(),
		ShardId:          sp.shardCoordinator.SelfId(),
		PrevRandSeed:     make([]byte, 0),
		RandSeed:         make([]byte, 0),
	}

	go sp.checkAndRequestIfMetaHeadersMissing(header.GetRound(), header.GetNonce())

	if bodyHandler == nil {
		return header, nil
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	mbLen := len(body)
	totalTxCount := 0
	miniBlockHeaders := make([]block.MiniBlockHeader, mbLen)
	for i := 0; i < mbLen; i++ {
		txCount := len(body[i].TxHashes)
		totalTxCount += txCount
		mbBytes, err := sp.marshalizer.Marshal(body[i])
		if err != nil {
			return nil, err
		}
		mbHash := sp.hasher.Compute(string(mbBytes))

		miniBlockHeaders[i] = block.MiniBlockHeader{
			Hash:            mbHash,
			SenderShardID:   sp.shardCoordinator.SelfId(),
			ReceiverShardID: body[i].ReceiverShardID,
			TxCount:         uint32(txCount),
		}
	}

	header.MiniBlockHeaders = miniBlockHeaders
	header.TxCount = uint32(totalTxCount)
	return header, nil
}

func (sp *shardProcessor) waitForTxHashes(waitTime time.Duration) error {
	select {
	case <-sp.chRcvAllTxs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

func (sp *shardProcessor) displayShardBlock(header *block.Header, body block.Body) {
	if header == nil || body == nil {
		return
	}

	headerHash, err := sp.computeHeaderHash(header)
	if err != nil {
		log.Error(err.Error())
		return
	}

	sp.displayLogInfo(header, body, headerHash)
}

func (sp *shardProcessor) displayLogInfo(
	header *block.Header,
	body block.Body,
	headerHash []byte,
) {
	dispHeader, dispLines := createDisplayableShardHeaderAndBlockBody(header, body)

	tblString, err := display.CreateTableString(dispHeader, dispLines)
	if err != nil {
		log.Error(err.Error())
		return
	}

	txCounterMutex.RLock()
	tblString = tblString + fmt.Sprintf("\nHeader hash: %s\n\n"+
		"Total txs processed until now: %d. Total txs processed for this block: %d. Total txs remained in pool: %d\n\n"+
		"Total shards: %d. Current shard id: %d\n",
		core.ToB64(headerHash),
		txsTotalProcessed,
		txsCurrentBlockProcessed,
		sp.getNrTxsWithDst(header.ShardId),
		sp.shardCoordinator.NumberOfShards(),
		sp.shardCoordinator.SelfId())
	txCounterMutex.RUnlock()

	log.Info(tblString)
}

func createDisplayableShardHeaderAndBlockBody(
	header *block.Header,
	body block.Body,
) ([]string, []*display.LineData) {

	tableHeader := []string{"Part", "Parameter", "Value"}

	lines := displayHeader(header)

	shardLines := make([]*display.LineData, 0)
	shardLines = append(shardLines, display.NewLineData(false, []string{
		"Header",
		"Block type",
		"TxBlock"}))
	shardLines = append(shardLines, display.NewLineData(false, []string{
		"",
		"Shard",
		fmt.Sprintf("%d", header.ShardId)}))
	shardLines = append(shardLines, lines...)

	if header.BlockBodyType == block.TxBlock {
		shardLines = displayTxBlockBody(shardLines, body)

		return tableHeader, shardLines
	}

	// TODO: implement the other block bodies

	shardLines = append(shardLines, display.NewLineData(false, []string{"Unknown", "", ""}))
	return tableHeader, shardLines
}

func displayTxBlockBody(lines []*display.LineData, body block.Body) []*display.LineData {
	txCounterMutex.Lock()
	txsCurrentBlockProcessed = 0
	txCounterMutex.Unlock()

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]

		part := fmt.Sprintf("MiniBlock_%d", miniBlock.ReceiverShardID)

		if miniBlock.TxHashes == nil || len(miniBlock.TxHashes) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<EMPTY>"}))
		}

		txCounterMutex.Lock()
		txsCurrentBlockProcessed += len(miniBlock.TxHashes)
		txsTotalProcessed += len(miniBlock.TxHashes)
		txCounterMutex.Unlock()

		for j := 0; j < len(miniBlock.TxHashes); j++ {
			if j == 0 || j >= len(miniBlock.TxHashes)-1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("TxHash_%d", j+1),
					core.ToB64(miniBlock.TxHashes[j])}))

				part = ""
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("..."),
					fmt.Sprintf("...")}))

				part = ""
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}

func sortTxByNonce(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)

	mTxHashes := make(map[uint64][][]byte)
	mTransactions := make(map[uint64][]*transaction.Transaction)

	nonces := make([]uint64, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Peek(key)
		if val == nil {
			continue
		}

		tx, ok := val.(*transaction.Transaction)
		if !ok {
			continue
		}

		if mTxHashes[tx.Nonce] == nil {
			nonces = append(nonces, tx.Nonce)
			mTxHashes[tx.Nonce] = make([][]byte, 0)
			mTransactions[tx.Nonce] = make([]*transaction.Transaction, 0)
		}

		mTxHashes[tx.Nonce] = append(mTxHashes[tx.Nonce], key)
		mTransactions[tx.Nonce] = append(mTransactions[tx.Nonce], tx)
	}

	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i] < nonces[j]
	})

	for _, nonce := range nonces {
		keys := mTxHashes[nonce]

		for idx, key := range keys {
			txHashes = append(txHashes, key)
			transactions = append(transactions, mTransactions[nonce][idx])
		}
	}

	return transactions, txHashes, nil
}

func (sp *shardProcessor) getNrTxsWithDst(dstShardId uint32) int {
	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		return 0
	}

	sumTxs := 0

	for i := uint32(0); i < sp.shardCoordinator.NumberOfShards(); i++ {
		strCache := process.ShardCacherIdentifier(i, dstShardId)
		txStore := txPool.ShardDataStore(strCache)
		if txStore == nil {
			continue
		}
		sumTxs += txStore.Len()
	}

	return sumTxs
}

// MarshalizedDataToBroadcast prepares underlying data into a marshalized object according to destination
func (sp *shardProcessor) MarshalizedDataToBroadcast(
	header data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (map[uint32][]byte, map[uint32][][]byte, error) {

	if bodyHandler == nil {
		return nil, nil, process.ErrNilMiniBlocks
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	mrsData := make(map[uint32][]byte)
	mrsTxs := make(map[uint32][][]byte)
	bodies := make(map[uint32]block.Body)

	for i := 0; i < len(body); i++ {
		miniblock := body[i]
		receiverShardId := miniblock.ReceiverShardID
		if receiverShardId == sp.shardCoordinator.SelfId() {
			//not taking into account miniblocks for current shard
			continue
		}
		bodies[receiverShardId] = append(bodies[receiverShardId], miniblock)

		for _, txHash := range miniblock.TxHashes {
			sp.mutCrossTxsForBlock.RLock()
			tx := sp.crossTxsForBlock[string(txHash)]
			sp.mutCrossTxsForBlock.RUnlock()
			if tx != nil {
				txMrs, err := sp.marshalizer.Marshal(tx)
				if err != nil {
					return nil, nil, process.ErrMarshalWithoutSuccess
				}
				mrsTxs[receiverShardId] = append(mrsTxs[receiverShardId], txMrs)
			}
		}
	}

	for shardId, subsetBlockBody := range bodies {
		buff, err := sp.marshalizer.Marshal(subsetBlockBody)
		if err != nil {
			return nil, nil, process.ErrMarshalWithoutSuccess
		}
		mrsData[shardId] = buff
	}

	return mrsData, mrsTxs, nil
}

func getTxs(txShardStore storage.Cacher) ([]*transaction.Transaction, [][]byte, error) {
	if txShardStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	transactions := make([]*transaction.Transaction, 0)
	txHashes := make([][]byte, 0)

	for _, key := range txShardStore.Keys() {
		val, _ := txShardStore.Peek(key)
		if val == nil {
			continue
		}

		tx, ok := val.(*transaction.Transaction)
		if !ok {
			continue
		}

		txHashes = append(txHashes, key)
		transactions = append(transactions, tx)
	}

	return transactions, txHashes, nil
}

// DecodeBlockBody method decodes block body from a given byte array
func (sp *shardProcessor) DecodeBlockBody(dta []byte) data.BodyHandler {
	if dta == nil {
		return nil
	}

	var body block.Body

	err := sp.marshalizer.Unmarshal(&body, dta)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return body
}

// DecodeBlockHeader method decodes block header from a given byte array
func (sp *shardProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	var header block.Header

	err := sp.marshalizer.Unmarshal(&header, dta)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &header
}
