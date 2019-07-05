package block

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/preprocess"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

const maxTransactionsInBlock = 15000

// shardProcessor implements shardProcessor interface and actually it tries to execute block
type shardProcessor struct {
	*baseProcessor
	dataPool          dataRetriever.PoolsHolder
	txProcessor       process.TransactionProcessor
	blocksTracker     process.BlocksTracker
	metaBlockFinality int

	onRequestMiniBlock    func(shardId uint32, mbHash []byte)
	chRcvAllMetaHdrs      chan bool
	mutUsedMetaHdrsHashes sync.Mutex
	usedMetaHdrsHashes    map[uint32][][]byte

	mutRequestedMetaHdrsHashes sync.RWMutex
	requestedMetaHdrsHashes    map[string]bool
	currHighestMetaHdrNonce    uint64
	allNeededMetaHdrsFound     bool

	core         serviceContainer.Core
	txPreProcess process.PreProcessor
}

// NewShardProcessor creates a new shardProcessor object
func NewShardProcessor(
	core serviceContainer.Core,
	dataPool dataRetriever.PoolsHolder,
	store dataRetriever.StorageService,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	txProcessor process.TransactionProcessor,
	accounts state.AccountsAdapter,
	shardCoordinator sharding.Coordinator,
	forkDetector process.ForkDetector,
	blocksTracker process.BlocksTracker,
	startHeaders map[uint32]data.HeaderHandler,
	metaChainActive bool,
	requestHandler process.RequestHandler,
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
	if requestHandler == nil {
		return nil, process.ErrNilRequestHandler
	}

	base := &baseProcessor{
		accounts:                      accounts,
		forkDetector:                  forkDetector,
		hasher:                        hasher,
		marshalizer:                   marshalizer,
		store:                         store,
		shardCoordinator:              shardCoordinator,
		onRequestHeaderHandlerByNonce: requestHandler.RequestHeaderByNonce,
	}
	err = base.setLastNotarizedHeadersSlice(startHeaders, metaChainActive)
	if err != nil {
		return nil, err
	}

	sp := shardProcessor{
		core:          core,
		baseProcessor: base,
		dataPool:      dataPool,
		txProcessor:   txProcessor,
		blocksTracker: blocksTracker,
	}

	sp.chRcvAllMetaHdrs = make(chan bool)

	transactionPool := sp.dataPool.Transactions()
	if transactionPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	sp.onRequestMiniBlock = requestHandler.RequestMiniBlock
	sp.requestedMetaHdrsHashes = make(map[string]bool)
	sp.usedMetaHdrsHashes = make(map[uint32][][]byte)

	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return nil, process.ErrNilMetaBlockPool
	}
	metaBlockPool.RegisterHandler(sp.receivedMetaBlock)
	sp.onRequestHeaderHandler = requestHandler.RequestHeader

	miniBlockPool := sp.dataPool.MiniBlocks()
	if miniBlockPool == nil {
		return nil, process.ErrNilMiniBlockPool
	}
	miniBlockPool.RegisterHandler(sp.receivedMiniBlock)

	sp.metaBlockFinality = process.MetaBlockFinality

	//TODO: This should be injected when BlockProcessor will be refactored
	sp.uint64Converter = uint64ByteSlice.NewBigEndianConverter()

	sp.allNeededMetaHdrsFound = true

	// TODO: inject all preprocessors together, find a good structure for them
	sp.txPreProcess, err = preprocess.NewTransactionPreprocessor(
		sp.dataPool.Transactions(),
		sp.store,
		sp.hasher,
		sp.marshalizer,
		sp.txProcessor,
		sp.shardCoordinator,
		accounts,
		requestHandler.RequestTransaction)
	if err != nil {
		return nil, err
	}

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

	err = sp.checkHeaderBodyCorrelation(header, body)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Total txs in pool: %d\n", GetNumTxsWithDst(header.ShardId, sp.dataPool, sp.shardCoordinator.NumberOfShards())))

	requestedTxs := sp.txPreProcess.RequestBlockTransactions(body)
	requestedMetaHdrs, requestedFinalMetaHdrs := sp.requestMetaHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	err = sp.txPreProcess.IsDataPrepared(requestedTxs, haveTime)
	if err != nil {
		return err
	}

	if requestedMetaHdrs > 0 || requestedFinalMetaHdrs > 0 {
		log.Info(fmt.Sprintf("requested %d missing meta headers and %d final meta headers to confirm cross shard txs\n", requestedMetaHdrs, requestedFinalMetaHdrs))
		err = sp.waitForMetaHdrHashes(haveTime())
		sp.mutRequestedMetaHdrsHashes.Lock()
		sp.allNeededMetaHdrsFound = true
		unreceivedMetaHdrs := len(sp.requestedMetaHdrsHashes)
		sp.mutRequestedMetaHdrsHashes.Unlock()
		log.Info(fmt.Sprintf("received %d missing meta headers\n", int(requestedMetaHdrs)-unreceivedMetaHdrs))
		if err != nil {
			return err
		}
	}

	if sp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing(header.Round)
	}()

	err = sp.checkMetaHeadersValidityAndFinality(header)
	if err != nil {
		return err
	}

	err = sp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			sp.RevertAccountState()
		}
	}()

	err = sp.txPreProcess.ProcessBlockTransactions(body, header.Round, haveTime)
	if err != nil {
		return err
	}

	if !sp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateMissmatch
		return err
	}

	return nil
}

// checkMetaHeadersValidity - checks if listed metaheaders are valid as construction
func (sp *shardProcessor) checkMetaHeadersValidityAndFinality(header *block.Header) error {
	metablockCache := sp.dataPool.MetaBlocks()
	if metablockCache == nil {
		return process.ErrNilMetaBlockPool
	}

	tmpNotedHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	currAddedMetaHdrs := make([]*block.MetaBlock, 0)
	for _, metaHash := range header.MetaBlockHashes {
		value, ok := metablockCache.Peek(metaHash)
		if !ok {
			return process.ErrNilMetaBlockHeader
		}

		metaHdr, ok := value.(*block.MetaBlock)
		if !ok {
			return process.ErrWrongTypeAssertion
		}

		currAddedMetaHdrs = append(currAddedMetaHdrs, metaHdr)
	}

	if len(currAddedMetaHdrs) == 0 {
		return nil
	}

	sort.Slice(currAddedMetaHdrs, func(i, j int) bool {
		return currAddedMetaHdrs[i].Nonce < currAddedMetaHdrs[j].Nonce
	})

	for _, metaHdr := range currAddedMetaHdrs {
		err := sp.isHdrConstructionValid(metaHdr, tmpNotedHdr)
		if err != nil {
			return err
		}
		tmpNotedHdr = metaHdr
	}

	err = sp.checkMetaHdrFinality(tmpNotedHdr, header.Round)
	if err != nil {
		return err
	}

	return nil
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (sp *shardProcessor) checkMetaHdrFinality(header data.HeaderHandler, round uint32) error {
	if header == nil {
		return process.ErrNilBlockHeader
	}

	sortedMetaHdrs, err := sp.getOrderedMetaBlocks(round)
	if err != nil {
		return err
	}

	lastVerifiedHdr := header
	// verify if there are "K" block after current to make this one final
	nextBlocksVerified := 0
	for _, tmpHdr := range sortedMetaHdrs {
		if nextBlocksVerified >= sp.metaBlockFinality {
			break
		}

		// found a header with the next nonce
		if tmpHdr.hdr.GetNonce() == header.GetNonce()+1 {
			err := sp.isHdrConstructionValid(tmpHdr.hdr, lastVerifiedHdr)
			if err != nil {
				continue
			}

			lastVerifiedHdr = tmpHdr.hdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified < sp.metaBlockFinality {
		return process.ErrHeaderNotFinal
	}

	return nil
}

// check if header has the same miniblocks as presented in body
func (sp *shardProcessor) checkHeaderBodyCorrelation(hdr *block.Header, body block.Body) error {
	mbHashesFromHdr := make(map[string]*block.MiniBlockHeader)
	for i := 0; i < len(hdr.MiniBlockHeaders); i++ {
		mbHashesFromHdr[string(hdr.MiniBlockHeaders[i].Hash)] = &hdr.MiniBlockHeaders[i]
	}

	if len(hdr.MiniBlockHeaders) != len(body) {
		return process.ErrHeaderBodyMismatch
	}

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]

		mbBytes, err := sp.marshalizer.Marshal(miniBlock)
		if err != nil {
			return err
		}
		mbHash := sp.hasher.Compute(string(mbBytes))

		mbHdr, ok := mbHashesFromHdr[string(mbHash)]
		if !ok {
			return process.ErrHeaderBodyMismatch
		}

		if mbHdr.TxCount != uint32(len(miniBlock.TxHashes)) {
			return process.ErrHeaderBodyMismatch
		}

		if mbHdr.ReceiverShardID != miniBlock.ReceiverShardID {
			return process.ErrHeaderBodyMismatch
		}

		if mbHdr.SenderShardID != miniBlock.SenderShardID {
			return process.ErrHeaderBodyMismatch
		}
	}

	return nil
}

func (sp *shardProcessor) checkAndRequestIfMetaHeadersMissing(round uint32) {
	orderedMetaBlocks, err := sp.getOrderedMetaBlocks(round)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	sortedHdrs := make([]data.HeaderHandler, 0)
	for i := 0; i < len(orderedMetaBlocks); i++ {
		hdr, ok := orderedMetaBlocks[i].hdr.(*block.MetaBlock)
		if !ok {
			continue
		}
		sortedHdrs = append(sortedHdrs, hdr)
	}

	err = sp.requestHeadersIfMissing(sortedHdrs, sharding.MetachainShardId, round)
	if err != nil {
		log.Info(err.Error())
	}

	return
}

func (sp *shardProcessor) indexBlockIfNeeded(
	body data.BodyHandler,
	header data.HeaderHandler) {
	if sp.core == nil || sp.core.Indexer() == nil {
		return
	}

	txPool := sp.txPreProcess.GetAllCurrentUsedTxs()

	go sp.core.Indexer().SaveBlock(body, header, txPool)
}

// RestoreBlockIntoPools restores the TxBlock and MetaBlock into associated pools
func (sp *shardProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if headerHandler == nil {
		return process.ErrNilBlockHeader
	}
	if bodyHandler == nil {
		return process.ErrNilTxBlockBody
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	header, ok := headerHandler.(*block.Header)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	restoredTxNr, miniBlockHashes, err := sp.txPreProcess.RestoreTxBlockIntoPools(body, sp.dataPool.MiniBlocks())
	go SubstractRestoredTxs(restoredTxNr)
	if err != nil {
		return err
	}

	err = sp.restoreMetaBlockIntoPool(miniBlockHashes, header.MetaBlockHashes)
	if err != nil {
		return err
	}

	sp.restoreLastNotarized()

	return nil
}

func (sp *shardProcessor) restoreMetaBlockIntoPool(miniBlockHashes map[int][]byte, metaBlockHashes [][]byte) error {
	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return process.ErrNilMetaBlockPool
	}

	metaHeaderNoncesPool := sp.dataPool.MetaHeadersNonces()
	if metaHeaderNoncesPool == nil {
		return process.ErrNilMetaHeadersNoncesDataPool
	}

	for _, metaBlockHash := range metaBlockHashes {
		buff, err := sp.store.Get(dataRetriever.MetaBlockUnit, metaBlockHash)
		if err != nil {
			continue
		}

		metaBlock := block.MetaBlock{}
		err = sp.marshalizer.Unmarshal(&metaBlock, buff)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		metaBlockPool.Put(metaBlockHash, &metaBlock)
		metaHeaderNoncesPool.Put(metaBlock.Nonce, &metaBlock)

		err = sp.store.GetStorer(dataRetriever.MetaBlockUnit).Remove(metaBlockHash)
		if err != nil {
			log.Error(err.Error())
		}

		nonceToByteSlice := sp.uint64Converter.ToByteSlice(metaBlock.Nonce)
		err = sp.store.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit).Remove(nonceToByteSlice)
		if err != nil {
			log.Error(err.Error())
		}
	}

	for _, metaBlockKey := range metaBlockPool.Keys() {
		if len(miniBlockHashes) == 0 {
			break
		}
		metaBlock, _ := metaBlockPool.Peek(metaBlockKey)
		if metaBlock == nil {
			log.Error(process.ErrNilMetaBlockHeader.Error())
			continue
		}

		hdr, _ := metaBlock.(data.HeaderHandler)
		if hdr == nil {
			log.Error(process.ErrWrongTypeAssertion.Error())
			continue
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

	header, ok := headerHandler.(*block.Header)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	buff, err := sp.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	headerHash := sp.hasher.Compute(string(buff))
	errNotCritical := sp.store.Put(dataRetriever.BlockHeaderUnit, headerHash, buff)
	log.LogIfError(errNotCritical)

	nonceToByteSlice := sp.uint64Converter.ToByteSlice(header.Nonce)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(header.ShardId)
	errNotCritical = sp.store.Put(hdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	log.LogIfError(errNotCritical)

	headerNoncePool := sp.dataPool.HeadersNonces()
	if headerNoncePool == nil {
		err = process.ErrNilDataPoolHolder
		return err
	}

	//TODO: Should be analyzed if put in pool is really necessary or not (right now there is no action of removing them)
	_ = headerNoncePool.Put(headerHandler.GetNonce(), headerHash)

	body, ok := bodyHandler.(block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	err = sp.txPreProcess.SaveTxBlockToStorage(body)
	if err != nil {
		return err
	}

	for i := 0; i < len(body); i++ {
		buff, err = sp.marshalizer.Marshal(body[i])
		if err != nil {
			return err
		}

		miniBlockHash := sp.hasher.Compute(string(buff))
		errNotCritical = sp.store.Put(dataRetriever.MiniBlockUnit, miniBlockHash, buff)
		log.LogIfError(errNotCritical)
	}

	processedMetaHdrs, errNotCritical := sp.getProcessedMetaBlocksFromPool(body, header)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	err = sp.saveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	if err != nil {
		return err
	}

	sp.mutNotarizedHdrs.RLock()
	lastNotarizedNonce := sp.lastNotarizedHdrs[sharding.MetachainShardId].GetNonce()
	sp.mutNotarizedHdrs.RUnlock()

	log.Info(fmt.Sprintf("last notarized block nonce from metachain is %d\n",
		lastNotarizedNonce))

	_, err = sp.accounts.Commit()
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("shardBlock with nonce %d and hash %s has been committed successfully\n",
		header.Nonce,
		core.ToB64(headerHash)))

	sp.blocksTracker.AddBlock(header)

	errNotCritical = sp.txPreProcess.RemoveTxBlockFromPools(body, sp.dataPool.MiniBlocks())
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	errNotCritical = sp.removeProcessedMetablocksFromPool(processedMetaHdrs)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	errNotCritical = sp.forkDetector.AddHeader(header, headerHash, process.BHProcessed)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
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

	sp.indexBlockIfNeeded(bodyHandler, headerHandler)

	// write data to log
	go DisplayLogInfo(header, body, headerHash, sp.shardCoordinator.NumberOfShards(), sp.shardCoordinator.SelfId(), sp.dataPool)

	return nil
}

// getProcessedMetaBlocksFromPool returns all the meta blocks fully processed
func (sp *shardProcessor) getProcessedMetaBlocksFromPool(body block.Body, header *block.Header) ([]data.HeaderHandler, error) {
	if body == nil {
		return nil, process.ErrNilTxBlockBody
	}
	if header == nil {
		return nil, process.ErrNilBlockHeader
	}

	miniBlockHashes := make(map[int][]byte, 0)
	for i := 0; i < len(body); i++ {
		miniBlock := body[i]
		if miniBlock.SenderShardID == sp.shardCoordinator.SelfId() {
			continue
		}

		buff, err := sp.marshalizer.Marshal(miniBlock)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		mbHash := sp.hasher.Compute(string(buff))
		miniBlockHashes[i] = mbHash
	}

	log.Debug(fmt.Sprintf("cross mini blocks in body: %d\n", len(miniBlockHashes)))

	processedMetaHdrs := make([]data.HeaderHandler, 0)
	for _, metaBlockKey := range header.MetaBlockHashes {
		metaBlock, _ := sp.dataPool.MetaBlocks().Peek(metaBlockKey)
		if metaBlock == nil {
			log.Debug(process.ErrNilMetaBlockHeader.Error())
			continue
		}

		hdr, ok := metaBlock.(*block.MetaBlock)
		if !ok {
			log.Debug(process.ErrWrongTypeAssertion.Error())
			continue
		}

		log.Debug(fmt.Sprintf("meta header nonce: %d\n", hdr.Nonce))

		crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for key := range miniBlockHashes {
			_, ok := crossMiniBlockHashes[string(miniBlockHashes[key])]
			if !ok {
				continue
			}

			hdr.SetMiniBlockProcessed(miniBlockHashes[key], true)
			delete(miniBlockHashes, key)
		}

		log.Debug(fmt.Sprintf("cross mini blocks in meta header: %d\n", len(crossMiniBlockHashes)))

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
	lastNotarizedMetaHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	processed := 0
	unnotarized := len(sp.blocksTracker.UnnotarisedBlocks())
	// processedMetaHdrs is also sorted
	for i := 0; i < len(processedMetaHdrs); i++ {
		hdr := processedMetaHdrs[i]

		// remove process finished
		if hdr.GetNonce() > lastNotarizedMetaHdr.GetNonce() {
			continue
		}

		errNotCritical := sp.blocksTracker.RemoveNotarisedBlocks(hdr)
		log.LogIfError(errNotCritical)

		// metablock was processed and finalized
		buff, err := sp.marshalizer.Marshal(hdr)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		headerHash := sp.hasher.Compute(string(buff))
		err = sp.store.Put(dataRetriever.MetaBlockUnit, headerHash, buff)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		nonceToByteSlice := sp.uint64Converter.ToByteSlice(hdr.GetNonce())
		err = sp.store.Put(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice, headerHash)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		sp.dataPool.MetaBlocks().Remove(headerHash)
		sp.dataPool.MetaHeadersNonces().Remove(hdr.GetNonce())

		log.Debug(fmt.Sprintf("metaBlock with nonce %d has been processed completely and removed from pool\n",
			hdr.GetNonce()))

		processed++
	}

	if processed > 0 {
		log.Info(fmt.Sprintf("%d meta blocks have been processed completely and removed from pool\n", processed))
	}

	notarized := unnotarized - len(sp.blocksTracker.UnnotarisedBlocks())
	if notarized > 0 {
		log.Info(fmt.Sprintf("%d shard blocks have been notarised by metachain\n", notarized))
	}

	return nil
}

// receivedMetaBlock is a callback function when a new metablock was received
// upon receiving, it parses the new metablock and requests miniblocks and transactions
// which destination is the current shard
func (sp *shardProcessor) receivedMetaBlock(metaBlockHash []byte) {
	metaBlksCache := sp.dataPool.MetaBlocks()
	if metaBlksCache == nil {
		return
	}

	metaHdrsNoncesCache := sp.dataPool.MetaHeadersNonces()
	if metaHdrsNoncesCache == nil && sp.metaBlockFinality > 0 {
		return
	}

	miniBlksCache := sp.dataPool.MiniBlocks()
	if miniBlksCache == nil {
		return
	}

	obj, ok := metaBlksCache.Peek(metaBlockHash)
	if !ok {
		return
	}

	metaBlock, ok := obj.(data.HeaderHandler)
	if !ok {
		return
	}

	log.Debug(fmt.Sprintf("received metablock with hash %s and nonce %d from network\n",
		core.ToB64(metaBlockHash),
		metaBlock.GetNonce()))

	sp.mutRequestedMetaHdrsHashes.Lock()

	if !sp.allNeededMetaHdrsFound {
		if sp.requestedMetaHdrsHashes[string(metaBlockHash)] {
			delete(sp.requestedMetaHdrsHashes, string(metaBlockHash))

			if metaBlock.GetNonce() > sp.currHighestMetaHdrNonce {
				sp.currHighestMetaHdrNonce = metaBlock.GetNonce()
			}
		}

		lenReqMetaHdrsHashes := len(sp.requestedMetaHdrsHashes)
		areFinalAttestingHdrsInCache := false
		if lenReqMetaHdrsHashes == 0 {
			requestedBlockHeaders := sp.requestFinalMissingHeaders()
			if requestedBlockHeaders == 0 {
				areFinalAttestingHdrsInCache = true
			}
		}

		sp.allNeededMetaHdrsFound = lenReqMetaHdrsHashes == 0 && areFinalAttestingHdrsInCache

		sp.mutRequestedMetaHdrsHashes.Unlock()

		if lenReqMetaHdrsHashes == 0 && areFinalAttestingHdrsInCache {
			sp.chRcvAllMetaHdrs <- true
		}
	} else {
		sp.mutRequestedMetaHdrsHashes.Unlock()
	}

	lastNotarizedHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return
	}
	if metaBlock.GetNonce() <= lastNotarizedHdr.GetNonce() {
		return
	}
	if metaBlock.GetRound() <= lastNotarizedHdr.GetRound() {
		return
	}

	crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
	for key, senderShardId := range crossMiniBlockHashes {
		obj, _ := miniBlksCache.Peek([]byte(key))
		if obj == nil {
			//TODO: It should be analyzed if launching the next line(request) on go routine is better or not
			go sp.onRequestMiniBlock(senderShardId, []byte(key))
		}
	}
}

// requestFinalMissingHeaders requests the headers needed to accept the current selected headers for processing the
// current block. It requests the metaBlockFinality headers greater than the highest meta header related to the block
// which should be processed
func (sp *shardProcessor) requestFinalMissingHeaders() uint32 {
	requestedBlockHeaders := uint32(0)
	for i := sp.currHighestMetaHdrNonce + 1; i <= sp.currHighestMetaHdrNonce+uint64(sp.metaBlockFinality); i++ {
		if !sp.dataPool.MetaHeadersNonces().Has(i) {
			requestedBlockHeaders++
			go sp.onRequestHeaderHandlerByNonce(sharding.MetachainShardId, i)
		}
	}

	return requestedBlockHeaders
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

	_ = sp.txPreProcess.RequestTransactionsForMiniBlock(miniBlock)
}

func (sp *shardProcessor) requestMetaHeaders(header *block.Header) (uint32, uint32) {
	sp.mutRequestedMetaHdrsHashes.Lock()

	sp.allNeededMetaHdrsFound = true

	if len(header.MetaBlockHashes) == 0 {
		sp.mutRequestedMetaHdrsHashes.Unlock()
		return 0, 0
	}

	missingHeaderHashes := sp.computeMissingHeaders(header)

	requestedBlockHeaders := uint32(0)
	sp.requestedMetaHdrsHashes = make(map[string]bool)
	for _, hash := range missingHeaderHashes {
		requestedBlockHeaders++
		sp.requestedMetaHdrsHashes[string(hash)] = true
		go sp.onRequestHeaderHandler(sharding.MetachainShardId, hash)
	}

	requestedFinalBlockHeaders := uint32(0)
	if requestedBlockHeaders > 0 {
		sp.allNeededMetaHdrsFound = false
	} else {
		requestedFinalBlockHeaders = sp.requestFinalMissingHeaders()
		if requestedFinalBlockHeaders > 0 {
			sp.allNeededMetaHdrsFound = false
		}
	}

	if !sp.allNeededMetaHdrsFound {
		process.EmptyChannel(sp.chRcvAllMetaHdrs)
	}

	sp.mutRequestedMetaHdrsHashes.Unlock()

	return requestedBlockHeaders, requestedFinalBlockHeaders
}

func (sp *shardProcessor) computeMissingHeaders(header *block.Header) [][]byte {
	missingHeaders := make([][]byte, 0)
	sp.currHighestMetaHdrNonce = uint64(0)

	for i := 0; i < len(header.MetaBlockHashes); i++ {
		hdr, err := process.GetMetaHeaderFromPool(header.MetaBlockHashes[i], sp.dataPool.MetaBlocks())
		if err != nil {
			missingHeaders = append(missingHeaders, header.MetaBlockHashes[i])
			continue
		}

		if hdr.Nonce > sp.currHighestMetaHdrNonce {
			sp.currHighestMetaHdrNonce = hdr.Nonce
		}
	}

	return missingHeaders
}

// processMiniBlockComplete - all transactions must be processed together, otherwise error
func (sp *shardProcessor) createAndProcessMiniBlockComplete(
	miniBlock *block.MiniBlock,
	round uint32,
	haveTime func() bool,
) error {

	snapshot := sp.accounts.JournalLen()
	err := sp.txPreProcess.ProcessMiniBlock(miniBlock, haveTime, round)
	if err != nil {
		log.Debug(err.Error())
		errAccountState := sp.accounts.RevertToSnapshot(snapshot)
		if errAccountState != nil {
			// TODO: evaluate if reloading the trie from disk will might solve the problem
			log.Debug(errAccountState.Error())
		}

		return err
	}

	return nil
}

func (sp *shardProcessor) verifyCrossShardMiniBlockDstMe(hdr *block.Header) error {
	mMiniBlockMeta, err := sp.getAllMiniBlockDstMeFromMeta(hdr.Round, hdr.MetaBlockHashes)
	if err != nil {
		return err
	}

	miniBlockDstMe := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
	for mbHash := range miniBlockDstMe {
		if _, ok := mMiniBlockMeta[mbHash]; !ok {
			return process.ErrCrossShardMBWithoutConfirmationFromMeta
		}
	}

	return nil
}

func (sp *shardProcessor) getAllMiniBlockDstMeFromMeta(round uint32, metaHashes [][]byte) (map[string][]byte, error) {
	metaBlockCache := sp.dataPool.MetaBlocks()
	if metaBlockCache == nil {
		return nil, process.ErrNilMetaBlockPool
	}

	lastHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, err
	}

	mMiniBlockMeta := make(map[string][]byte)
	for _, metaHash := range metaHashes {
		val, _ := metaBlockCache.Peek(metaHash)
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
		if hdr.GetRound() <= lastHdr.GetRound() {
			continue
		}
		if hdr.GetNonce() <= lastHdr.GetNonce() {
			continue
		}

		miniBlockDstMe := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for mbHash := range miniBlockDstMe {
			mMiniBlockMeta[mbHash] = metaHash
		}
	}

	return mMiniBlockMeta, nil
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
		if hdr.GetRound() <= lastHdr.GetRound() {
			continue
		}
		if hdr.GetNonce() <= lastHdr.GetNonce() {
			continue
		}

		orderedMetaBlocks = append(orderedMetaBlocks, &hashAndHdr{hdr: hdr, hash: key})
	}

	sort.Slice(orderedMetaBlocks, func(i, j int) bool {
		return orderedMetaBlocks[i].hdr.GetNonce() < orderedMetaBlocks[j].hdr.GetNonce()
	})

	return orderedMetaBlocks, nil
}

func (sp *shardProcessor) createAndprocessMiniBlocksFromHeader(
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

		requestedTxs := sp.txPreProcess.RequestTransactionsForMiniBlock(*miniBlock)
		if requestedTxs > 0 {
			continue
		}

		err := sp.createAndProcessMiniBlockComplete(miniBlock, round, haveTime)
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

// isMetaHeaderFinal verifies if meta is trully final, in order to not do rollbacks
func (sp *shardProcessor) isMetaHeaderFinal(currHdr data.HeaderHandler, sortedHdrs []*hashAndHdr, startPos int) bool {
	if currHdr == nil {
		return false
	}
	if sortedHdrs == nil {
		return false
	}

	// verify if there are "K" block after current to make this one final
	lastVerifiedHdr := currHdr
	nextBlocksVerified := 0

	for i := startPos; i < len(sortedHdrs); i++ {
		if nextBlocksVerified >= sp.metaBlockFinality {
			return true
		}

		// found a header with the next nonce
		tmpHdr := sortedHdrs[i].hdr
		if tmpHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
			err := sp.isHdrConstructionValid(tmpHdr, lastVerifiedHdr)
			if err != nil {
				continue
			}

			lastVerifiedHdr = tmpHdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified >= sp.metaBlockFinality {
		return true
	}

	return false
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

	orderedMetaBlocks, err := sp.getOrderedMetaBlocks(round)
	if err != nil {
		return nil, 0, err
	}

	log.Info(fmt.Sprintf("meta blocks ordered: %d \n", len(orderedMetaBlocks)))

	lastMetaHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, 0, err
	}

	// do processing in order
	usedMetaHdrsHashes := make([][]byte, 0)
	for i := 0; i < len(orderedMetaBlocks); i++ {
		if !haveTime() {
			log.Info(fmt.Sprintf("time is up after putting %d cross txs with destination to current shard \n", nrTxAdded))
			break
		}

		hdr, ok := orderedMetaBlocks[i].hdr.(*block.MetaBlock)
		if !ok {
			continue
		}

		err := sp.isHdrConstructionValid(hdr, lastMetaHdr)
		if err != nil {
			continue
		}

		isFinal := sp.isMetaHeaderFinal(hdr, orderedMetaBlocks, i+1)
		if !isFinal {
			continue
		}

		if len(hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())) == 0 {
			usedMetaHdrsHashes = append(usedMetaHdrsHashes, orderedMetaBlocks[i].hash)
			lastMetaHdr = hdr
			continue
		}

		maxTxRemaining := uint32(maxTxInBlock) - nrTxAdded
		currMBProcessed, currTxsAdded, hdrProcessFinished := sp.createAndprocessMiniBlocksFromHeader(hdr, maxTxRemaining, round, haveTime)

		// all txs processed, add to processed miniblocks
		miniBlocks = append(miniBlocks, currMBProcessed...)
		nrTxAdded = nrTxAdded + currTxsAdded

		if currTxsAdded > 0 {
			usedMetaHdrsHashes = append(usedMetaHdrsHashes, orderedMetaBlocks[i].hash)
		}

		if !hdrProcessFinished {
			break
		}

		lastMetaHdr = hdr
	}

	sp.mutUsedMetaHdrsHashes.Lock()
	sp.usedMetaHdrsHashes[round] = usedMetaHdrsHashes
	sp.mutUsedMetaHdrsHashes.Unlock()

	return miniBlocks, nrTxAdded, nil
}

func (sp *shardProcessor) createMiniBlocks(
	noShards uint32,
	maxTxInBlock int,
	round uint32,
	haveTime func() bool,
) (block.Body, error) {

	miniBlocks := make(block.Body, 0)

	sp.txPreProcess.CreateBlockStarted()

	if sp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after entered in createMiniBlocks method\n"))
		return nil, process.ErrTimeIsOut
	}

	txPool := sp.dataPool.Transactions()
	if txPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	destMeMiniBlocks, txs, err := sp.createAndProcessCrossMiniBlocksDstMe(noShards, maxTxInBlock, round, haveTime)
	if err != nil {
		log.Info(err.Error())
	}

	log.Debug(fmt.Sprintf("processed %d miniblocks and %d txs with destination in self shard\n", len(destMeMiniBlocks), txs))

	if len(destMeMiniBlocks) > 0 {
		miniBlocks = append(miniBlocks, destMeMiniBlocks...)
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up added %d transactions\n", txs))
		return miniBlocks, nil
	}

	addedTxs := int(txs)
	for i := 0; i < int(noShards); i++ {
		miniBlock, err := sp.txPreProcess.CreateAndProcessMiniBlock(sp.shardCoordinator.SelfId(), uint32(i), maxTxInBlock-addedTxs, haveTime, round)
		if err != nil {
			return miniBlocks, nil
		}

		if len(miniBlock.TxHashes) > 0 {
			addedTxs += len(miniBlock.TxHashes)
			miniBlocks = append(miniBlocks, miniBlock)
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

	defer func() {
		go sp.checkAndRequestIfMetaHeadersMissing(round)
	}()

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
			SenderShardID:   body[i].SenderShardID,
			ReceiverShardID: body[i].ReceiverShardID,
			TxCount:         uint32(txCount),
			Type:            body[i].Type,
		}
	}

	header.MiniBlockHeaders = miniBlockHeaders
	header.TxCount = uint32(totalTxCount)

	sp.mutUsedMetaHdrsHashes.Lock()

	if usedMetaHdrsHashes, ok := sp.usedMetaHdrsHashes[round]; ok {
		header.MetaBlockHashes = usedMetaHdrsHashes
		delete(sp.usedMetaHdrsHashes, round)
	}

	sp.mutUsedMetaHdrsHashes.Unlock()

	return header, nil
}

func (sp *shardProcessor) waitForMetaHdrHashes(waitTime time.Duration) error {
	select {
	case <-sp.chRcvAllMetaHdrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
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
		if receiverShardId == sp.shardCoordinator.SelfId() { // not taking into account miniblocks for current shard
			continue
		}

		bodies[receiverShardId] = append(bodies[receiverShardId], miniblock)

		currMrsTxs, err := sp.txPreProcess.CreateMarshalizedData(miniblock.TxHashes)
		if err != nil {
			log.Debug(err.Error())
			continue
		}

		if len(currMrsTxs) > 0 {
			mrsTxs[receiverShardId] = append(mrsTxs[receiverShardId], currMrsTxs...)
		}
	}

	for shardId, subsetBlockBody := range bodies {
		buff, err := sp.marshalizer.Marshal(subsetBlockBody)
		if err != nil {
			log.Debug(process.ErrMarshalWithoutSuccess.Error())
			continue
		}
		mrsData[shardId] = buff
	}

	return mrsData, mrsTxs, nil
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
