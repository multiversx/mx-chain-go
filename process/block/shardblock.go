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
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/dataPool"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/throttle"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

const maxCleanTime = time.Second

// shardProcessor implements shardProcessor interface and actually it tries to execute block
type shardProcessor struct {
	*baseProcessor
	dataPool          dataRetriever.PoolsHolder
	blocksTracker     process.BlocksTracker
	metaBlockFinality int

	chRcvAllMetaHdrs      chan bool
	mutUsedMetaHdrsHashes sync.Mutex
	usedMetaHdrsHashes    map[uint64][][]byte

	mutRequestedMetaHdrsHashes sync.RWMutex
	requestedMetaHdrsHashes    map[string]bool
	currHighestMetaHdrNonce    uint64
	allNeededMetaHdrsFound     bool

	processedMiniBlocks    map[string]map[string]struct{}
	mutProcessedMiniBlocks sync.RWMutex

	core          serviceContainer.Core
	txCoordinator process.TransactionCoordinator
	txCounter     *transactionCounter

	txsPoolsCleaner process.PoolsCleaner
}

// NewShardProcessor creates a new shardProcessor object
func NewShardProcessor(arguments ArgShardProcessor) (*shardProcessor, error) {

	err := checkProcessorNilParameters(
		arguments.Accounts,
		arguments.ForkDetector,
		arguments.Hasher,
		arguments.Marshalizer,
		arguments.Store,
		arguments.ShardCoordinator,
		arguments.NodesCoordinator,
		arguments.SpecialAddressHandler,
		arguments.Uint64Converter)
	if err != nil {
		return nil, err
	}

	if arguments.DataPool == nil || arguments.DataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	if arguments.BlocksTracker == nil || arguments.BlocksTracker.IsInterfaceNil() {
		return nil, process.ErrNilBlocksTracker
	}
	if arguments.RequestHandler == nil || arguments.RequestHandler.IsInterfaceNil() {
		return nil, process.ErrNilRequestHandler
	}
	if arguments.TxCoordinator == nil || arguments.TxCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilTransactionCoordinator
	}

	blockSizeThrottler, err := throttle.NewBlockSizeThrottle()
	if err != nil {
		return nil, err
	}

	base := &baseProcessor{
		accounts:                      arguments.Accounts,
		blockSizeThrottler:            blockSizeThrottler,
		forkDetector:                  arguments.ForkDetector,
		hasher:                        arguments.Hasher,
		marshalizer:                   arguments.Marshalizer,
		store:                         arguments.Store,
		shardCoordinator:              arguments.ShardCoordinator,
		nodesCoordinator:              arguments.NodesCoordinator,
		specialAddressHandler:         arguments.SpecialAddressHandler,
		uint64Converter:               arguments.Uint64Converter,
		onRequestHeaderHandlerByNonce: arguments.RequestHandler.RequestHeaderByNonce,
		appStatusHandler:              statusHandler.NewNilStatusHandler(),
	}
	err = base.setLastNotarizedHeadersSlice(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	if arguments.TxsPoolsCleaner == nil || arguments.TxsPoolsCleaner.IsInterfaceNil() {
		return nil, process.ErrNilTxsPoolsCleaner
	}

	sp := shardProcessor{
		core:            arguments.Core,
		baseProcessor:   base,
		dataPool:        arguments.DataPool,
		blocksTracker:   arguments.BlocksTracker,
		txCoordinator:   arguments.TxCoordinator,
		txCounter:       NewTransactionCounter(),
		txsPoolsCleaner: arguments.TxsPoolsCleaner,
	}
	sp.chRcvAllMetaHdrs = make(chan bool)

	transactionPool := sp.dataPool.Transactions()
	if transactionPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	sp.requestedMetaHdrsHashes = make(map[string]bool)
	sp.usedMetaHdrsHashes = make(map[uint64][][]byte)
	sp.processedMiniBlocks = make(map[string]map[string]struct{})

	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return nil, process.ErrNilMetaBlockPool
	}
	metaBlockPool.RegisterHandler(sp.receivedMetaBlock)
	sp.onRequestHeaderHandler = arguments.RequestHandler.RequestHeader

	sp.metaBlockFinality = process.MetaBlockFinality
	sp.allNeededMetaHdrsFound = true

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
		if err == process.ErrBlockHashDoesNotMatch {
			log.Info(fmt.Sprintf("requested missing shard header with hash %s for shard %d\n",
				core.ToB64(headerHandler.GetPrevHash()),
				headerHandler.GetShardID()))

			go sp.onRequestHeaderHandler(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return err
	}

	log.Debug(fmt.Sprintf("started processing block with round %d and nonce %d\n",
		headerHandler.GetRound(),
		headerHandler.GetNonce()))

	header, ok := headerHandler.(*block.Header)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	go getMetricsFromBlockBody(body, sp.marshalizer, sp.appStatusHandler)

	err = sp.checkHeaderBodyCorrelation(header, body)
	if err != nil {
		return err
	}

	numTxWithDst := sp.txCounter.getNumTxsFromPool(header.ShardId, sp.dataPool, sp.shardCoordinator.NumberOfShards())
	totalTxs := sp.txCounter.totalTxs
	go getMetricsFromHeader(header, uint64(numTxWithDst), totalTxs, sp.marshalizer, sp.appStatusHandler)

	log.Info(fmt.Sprintf("Total txs in pool: %d\n", numTxWithDst))

	err = sp.specialAddressHandler.SetShardConsensusData(
		headerHandler.GetPrevRandSeed(),
		headerHandler.GetRound(),
		headerHandler.GetEpoch(),
		headerHandler.GetShardID(),
	)
	if err != nil {
		return err
	}

	sp.txCoordinator.CreateBlockStarted()
	sp.txCoordinator.RequestBlockTransactions(body)
	requestedMetaHdrs, requestedFinalMetaHdrs := sp.requestMetaHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	err = sp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

	if requestedMetaHdrs > 0 || requestedFinalMetaHdrs > 0 {
		log.Info(fmt.Sprintf("requested %d missing meta headers and %d final meta headers\n", requestedMetaHdrs, requestedFinalMetaHdrs))
		err = sp.waitForMetaHdrHashes(haveTime())
		sp.mutRequestedMetaHdrsHashes.Lock()
		sp.allNeededMetaHdrsFound = true
		unreceivedMetaHdrs := len(sp.requestedMetaHdrsHashes)
		sp.mutRequestedMetaHdrsHashes.Unlock()
		if requestedMetaHdrs > 0 {
			log.Info(fmt.Sprintf("received %d missing meta headers\n", int(requestedMetaHdrs)-unreceivedMetaHdrs))
		}
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

	processedMetaHdrs, err := sp.getProcessedMetaBlocksFromMiniBlocks(body, header.MetaBlockHashes)
	if err != nil {
		return err
	}

	err = sp.setMetaConsensusData(processedMetaHdrs)
	if err != nil {
		return err
	}

	err = sp.txCoordinator.ProcessBlockTransaction(body, header.Round, haveTime)
	if err != nil {
		return err
	}

	if !sp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = sp.txCoordinator.VerifyCreatedBlockTransactions(body)
	if err != nil {
		return err
	}

	return nil
}

func (sp *shardProcessor) setMetaConsensusData(finalizedMetaBlocks []data.HeaderHandler) error {
	sp.specialAddressHandler.ClearMetaConsensusData()

	// for every finalized metablock header, reward the metachain consensus group members with accounts in shard
	for _, metaBlock := range finalizedMetaBlocks {
		round := metaBlock.GetRound()
		epoch := metaBlock.GetEpoch()
		err := sp.specialAddressHandler.SetMetaConsensusData(metaBlock.GetPrevRandSeed(), round, epoch)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetConsensusData - sets the reward data for the current consensus group
func (sp *shardProcessor) SetConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) {
	err := sp.specialAddressHandler.SetShardConsensusData(randomness, round, epoch, shardId)
	if err != nil {
		log.Error(err.Error())
	}
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
		err = sp.isHdrConstructionValid(metaHdr, tmpNotedHdr)
		if err != nil {
			return err
		}
		tmpNotedHdr = metaHdr
	}

	err = sp.checkMetaHdrFinality(tmpNotedHdr)
	if err != nil {
		return err
	}

	return nil
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (sp *shardProcessor) checkMetaHdrFinality(header data.HeaderHandler) error {
	if header == nil || header.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}

	sortedMetaHdrs, err := sp.getFinalityAttestingHeaders(header, process.MetaBlockFinality)
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
		if tmpHdr.hdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
			err = sp.isHdrConstructionValid(tmpHdr.hdr, lastVerifiedHdr)
			if err != nil {
				log.Debug(err.Error())
				continue
			}

			lastVerifiedHdr = tmpHdr.hdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified < sp.metaBlockFinality {
		go sp.onRequestHeaderHandlerByNonce(lastVerifiedHdr.GetShardID(), lastVerifiedHdr.GetNonce()+1)
		return process.ErrHeaderNotFinal
	}

	return nil
}

func (sp *shardProcessor) getFinalityAttestingHeaders(
	highestNonceHdr data.HeaderHandler,
	finality uint64,
) ([]*hashAndHdr, error) {

	if highestNonceHdr == nil || highestNonceHdr.IsInterfaceNil() {
		return nil, process.ErrNilBlockHeader
	}

	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return nil, process.ErrNilMetaBlockPool
	}

	orderedMetaBlocks := make([]*hashAndHdr, 0)
	// get keys and arrange them into shards
	for _, key := range metaBlockPool.Keys() {
		val, _ := metaBlockPool.Peek(key)
		if val == nil {
			continue
		}

		hdr, ok := val.(*block.MetaBlock)
		if !ok {
			continue
		}

		isHdrNonceLowerOrEqualThanHighestNonce := hdr.GetNonce() <= highestNonceHdr.GetNonce()
		isHdrNonceHigherThanFinalNonce := hdr.GetNonce() > highestNonceHdr.GetNonce()+finality

		if isHdrNonceLowerOrEqualThanHighestNonce ||
			isHdrNonceHigherThanFinalNonce {
			continue
		}

		orderedMetaBlocks = append(orderedMetaBlocks, &hashAndHdr{hdr: hdr, hash: key})
	}

	if len(orderedMetaBlocks) > 1 {
		sort.Slice(orderedMetaBlocks, func(i, j int) bool {
			return orderedMetaBlocks[i].hdr.GetNonce() < orderedMetaBlocks[j].hdr.GetNonce()
		})
	}

	return orderedMetaBlocks, nil
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

func (sp *shardProcessor) checkAndRequestIfMetaHeadersMissing(round uint64) {
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

	txPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)
	scPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)
	rewardPool := sp.txCoordinator.GetAllCurrentUsedTxs(block.RewardsBlock)

	for hash, tx := range scPool {
		txPool[hash] = tx
	}
	for hash, tx := range rewardPool {
		txPool[hash] = tx
	}

	go sp.core.Indexer().SaveBlock(body, header, txPool)
}

// RestoreBlockIntoPools restores the TxBlock and MetaBlock into associated pools
func (sp *shardProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	sp.removeLastNotarized()

	if headerHandler == nil || headerHandler.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}
	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
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

	restoredTxNr, err := sp.txCoordinator.RestoreBlockDataFromStorage(body)
	go sp.txCounter.subtractRestoredTxs(restoredTxNr)
	if err != nil {
		return err
	}

	miniBlockHashes := header.MapMiniBlockHashesToShards()
	err = sp.restoreMetaBlockIntoPool(miniBlockHashes, header.MetaBlockHashes)
	if err != nil {
		return err
	}

	return nil
}

func (sp *shardProcessor) restoreMetaBlockIntoPool(miniBlockHashes map[string]uint32, metaBlockHashes [][]byte) error {
	metaBlockPool := sp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		return process.ErrNilMetaBlockPool
	}

	metaHeaderNoncesPool := sp.dataPool.HeadersNonces()
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

		processedMiniBlocks := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for mbHash := range processedMiniBlocks {
			sp.addProcessedMiniBlock(metaBlockHash, []byte(mbHash))
		}

		metaBlockPool.Put(metaBlockHash, &metaBlock)
		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(metaBlock.GetShardID(), metaBlockHash)
		metaHeaderNoncesPool.Merge(metaBlock.Nonce, syncMap)

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
		metaBlock, ok := metaBlockPool.Peek(metaBlockKey)
		if !ok {
			log.Error(process.ErrNilMetaBlockHeader.Error())
			continue
		}

		hdr, ok := metaBlock.(data.HeaderHandler)
		if !ok {
			metaBlockPool.Remove(metaBlockKey)
			log.Error(process.ErrWrongTypeAssertion.Error())
			continue
		}

		crossMiniBlockHashes := hdr.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for key := range miniBlockHashes {
			_, ok = crossMiniBlockHashes[key]
			if !ok {
				continue
			}

			sp.removeProcessedMiniBlock(metaBlockKey, []byte(key))
		}
	}

	return nil
}

// CreateBlockBody creates a a list of miniblocks by filling them with transactions out of the transactions pools
// as long as the transactions limit for the block has not been reached and there is still time to add transactions
func (sp *shardProcessor) CreateBlockBody(round uint64, haveTime func() bool) (data.BodyHandler, error) {
	log.Debug(fmt.Sprintf("started creating block body in round %d\n", round))
	sp.txCoordinator.CreateBlockStarted()
	sp.blockSizeThrottler.ComputeMaxItems()

	miniBlocks, err := sp.createMiniBlocks(sp.shardCoordinator.NumberOfShards(), sp.blockSizeThrottler.MaxItemsToAdd(), round, haveTime)
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

	log.Debug(fmt.Sprintf("started committing block with round %d and nonce %d\n",
		headerHandler.GetRound(),
		headerHandler.GetNonce()))

	err = sp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
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
	nonceToByteSlice := sp.uint64Converter.ToByteSlice(header.Nonce)
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(header.ShardId)

	errNotCritical := sp.store.Put(hdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	log.LogIfError(errNotCritical)

	errNotCritical = sp.store.Put(dataRetriever.BlockHeaderUnit, headerHash, buff)
	log.LogIfError(errNotCritical)

	headerNoncePool := sp.dataPool.HeadersNonces()
	if headerNoncePool == nil {
		err = process.ErrNilDataPoolHolder
		return err
	}

	//TODO: Should be analyzed if put in pool is really necessary or not (right now there is no action of removing them)
	syncMap := &dataPool.ShardIdHashSyncMap{}
	syncMap.Store(headerHandler.GetShardID(), headerHash)
	headerNoncePool.Merge(headerHandler.GetNonce(), syncMap)

	body, ok := bodyHandler.(block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	err = sp.txCoordinator.SaveBlockDataToStorage(body)
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

	processedMetaHdrs, err := sp.getProcessedMetaBlocksFromHeader(header)
	if err != nil {
		return err
	}

	finalHeaders, finalHeadersHashes, err := sp.getHighestHdrForOwnShardFromMetachain(processedMetaHdrs)
	if err != nil {
		return err
	}

	err = sp.saveLastNotarizedHeader(sharding.MetachainShardId, processedMetaHdrs)
	if err != nil {
		return err
	}

	headerMeta, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return err
	}

	sp.appStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, fmt.Sprintf("meta %d", headerMeta.GetNonce()))

	_, err = sp.accounts.Commit()
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("shard block with nonce %d and hash %s has been committed successfully\n",
		header.Nonce,
		core.ToB64(headerHash)))

	sp.blocksTracker.AddBlock(header)

	errNotCritical = sp.txCoordinator.RemoveBlockDataFromPool(body)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	errNotCritical = sp.removeProcessedMetaBlocksFromPool(processedMetaHdrs)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	errNotCritical = sp.forkDetector.AddHeader(header, headerHash, process.BHProcessed, finalHeaders, finalHeadersHashes)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	highestFinalBlockNonce := sp.forkDetector.GetHighestFinalBlockNonce()
	log.Info(fmt.Sprintf("shard block with nonce %d is the highest final block in shard %d\n",
		highestFinalBlockNonce,
		sp.shardCoordinator.SelfId()))

	sp.appStatusHandler.SetStringValue(core.MetricCurrentBlockHash, core.ToB64(headerHash))
	sp.appStatusHandler.SetUInt64Value(core.MetricHighestFinalBlockInShard, highestFinalBlockNonce)

	hdrsToAttestPreviousFinal := uint32(header.Nonce-highestFinalBlockNonce) + 1
	sp.removeNotarizedHdrsBehindPreviousFinal(hdrsToAttestPreviousFinal)

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

	go sp.cleanTxsPools()

	// write data to log
	go sp.txCounter.displayLogInfo(
		header,
		body,
		headerHash,
		sp.shardCoordinator.NumberOfShards(),
		sp.shardCoordinator.SelfId(),
		sp.dataPool,
	)

	sp.blockSizeThrottler.Succeed(header.Round)

	return nil
}

func (sp *shardProcessor) cleanTxsPools() {
	_, err := sp.txsPoolsCleaner.Clean(maxCleanTime)
	log.LogIfError(err)
	log.Info(fmt.Sprintf("Total txs removed from pools cleaner %d", sp.txsPoolsCleaner.NumRemovedTxs()))
}

// getHighestHdrForOwnShardFromMetachain calculates the highest shard header notarized by metachain
func (sp *shardProcessor) getHighestHdrForOwnShardFromMetachain(
	processedHdrs []data.HeaderHandler,
) ([]data.HeaderHandler, [][]byte, error) {

	ownShIdHdrs := make([]data.HeaderHandler, 0)

	sort.Slice(processedHdrs, func(i, j int) bool {
		return processedHdrs[i].GetNonce() < processedHdrs[j].GetNonce()
	})

	for i := 0; i < len(processedHdrs); i++ {
		hdr, ok := processedHdrs[i].(*block.MetaBlock)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		hdrs, err := sp.getHighestHdrForShardFromMetachain(sp.shardCoordinator.SelfId(), hdr)
		if err != nil {
			return nil, nil, err
		}

		ownShIdHdrs = append(ownShIdHdrs, hdrs...)
	}

	if len(ownShIdHdrs) == 0 {
		ownShIdHdrs = append(ownShIdHdrs, &block.Header{})
	}

	sort.Slice(ownShIdHdrs, func(i, j int) bool {
		return ownShIdHdrs[i].GetNonce() < ownShIdHdrs[j].GetNonce()
	})

	ownShIdHdrsHashes := make([][]byte, 0)
	for i := 0; i < len(ownShIdHdrs); i++ {
		hash, _ := core.CalculateHash(sp.marshalizer, sp.hasher, ownShIdHdrs[i])
		ownShIdHdrsHashes = append(ownShIdHdrsHashes, hash)
	}

	return ownShIdHdrs, ownShIdHdrsHashes, nil
}

func (sp *shardProcessor) getHighestHdrForShardFromMetachain(shardId uint32, hdr *block.MetaBlock) ([]data.HeaderHandler, error) {
	ownShIdHdr := make([]data.HeaderHandler, 0)

	var errFound error
	// search for own shard id in shardInfo from metaHeaders
	for _, shardInfo := range hdr.ShardInfo {
		if shardInfo.ShardId != shardId {
			continue
		}

		ownHdr, err := process.GetShardHeader(shardInfo.HeaderHash, sp.dataPool.Headers(), sp.marshalizer, sp.store)
		if err != nil {
			go sp.onRequestHeaderHandler(shardInfo.ShardId, shardInfo.HeaderHash)

			log.Info(fmt.Sprintf("requested missing shard header with hash %s for shard %d\n",
				core.ToB64(shardInfo.HeaderHash),
				shardInfo.ShardId))

			errFound = err
			continue
		}

		ownShIdHdr = append(ownShIdHdr, ownHdr)
	}

	if errFound != nil {
		return nil, errFound
	}

	return ownShIdHdr, nil
}

// getProcessedMetaBlocksFromHeader returns all the meta blocks fully processed
func (sp *shardProcessor) getProcessedMetaBlocksFromHeader(header *block.Header) ([]data.HeaderHandler, error) {
	if header == nil {
		return nil, process.ErrNilBlockHeader
	}

	miniBlockHashes := make(map[int][]byte, 0)
	for i := 0; i < len(header.MiniBlockHeaders); i++ {
		miniBlockHashes[i] = header.MiniBlockHeaders[i].Hash
	}

	log.Debug(fmt.Sprintf("cross mini blocks in body: %d\n", len(miniBlockHashes)))

	processedMetaHeaders, usedMbs, err := sp.getProcessedMetaBlocksFromMiniBlockHashes(miniBlockHashes, header.MetaBlockHashes)
	if err != nil {
		return nil, err
	}

	for _, metaBlockKey := range header.MetaBlockHashes {
		obj, ok := sp.dataPool.MetaBlocks().Peek(metaBlockKey)
		if !ok {
			return nil, process.ErrNilMetaBlockHeader
		}

		metaBlock, ok := obj.(*block.MetaBlock)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for key := range crossMiniBlockHashes {
			if usedMbs[key] {
				sp.addProcessedMiniBlock(metaBlockKey, []byte(key))
			}
		}
	}

	return processedMetaHeaders, nil
}

// getProcessedMetaBlocks returns all the meta blocks fully processed
func (sp *shardProcessor) getProcessedMetaBlocksFromMiniBlocks(
	usedMiniBlocks []*block.MiniBlock,
	usedMetaBlockHashes [][]byte,
) ([]data.HeaderHandler, error) {
	if usedMiniBlocks == nil || usedMetaBlockHashes == nil {
		// not an error, it can happen that no metablock header or no miniblock is used.
		return make([]data.HeaderHandler, 0), nil
	}

	miniBlockHashes := make(map[int][]byte, 0)
	for i := 0; i < len(usedMiniBlocks); i++ {
		miniBlock := usedMiniBlocks[i]
		if miniBlock.SenderShardID == sp.shardCoordinator.SelfId() {
			continue
		}

		mbHash, err := core.CalculateHash(sp.marshalizer, sp.hasher, miniBlock)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
		miniBlockHashes[i] = mbHash
	}

	log.Debug(fmt.Sprintf("cross mini blocks in body: %d\n", len(miniBlockHashes)))
	processedMetaBlocks, _, err := sp.getProcessedMetaBlocksFromMiniBlockHashes(miniBlockHashes, usedMetaBlockHashes)

	return processedMetaBlocks, err
}

func (sp *shardProcessor) getProcessedMetaBlocksFromMiniBlockHashes(
	miniBlockHashes map[int][]byte,
	usedMetaBlockHashes [][]byte,
) ([]data.HeaderHandler, map[string]bool, error) {

	processedMetaHdrs := make([]data.HeaderHandler, 0)
	processedMBs := make(map[string]bool)

	for _, metaBlockKey := range usedMetaBlockHashes {
		obj, _ := sp.dataPool.MetaBlocks().Peek(metaBlockKey)
		if obj == nil {
			return nil, nil, process.ErrNilMetaBlockHeader
		}

		metaBlock, ok := obj.(*block.MetaBlock)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		log.Debug(fmt.Sprintf("meta header nonce: %d\n", metaBlock.Nonce))

		crossMiniBlockHashes := metaBlock.GetMiniBlockHeadersWithDst(sp.shardCoordinator.SelfId())
		for hash := range crossMiniBlockHashes {
			processedMBs[hash] = sp.isMiniBlockProcessed(metaBlockKey, []byte(hash))
		}

		for key := range miniBlockHashes {
			_, ok = crossMiniBlockHashes[string(miniBlockHashes[key])]
			if !ok {
				continue
			}

			processedMBs[string(miniBlockHashes[key])] = true

			delete(miniBlockHashes, key)
		}

		log.Debug(fmt.Sprintf("cross mini blocks in meta header: %d\n", len(crossMiniBlockHashes)))

		processedAll := true
		for key := range crossMiniBlockHashes {
			if !processedMBs[key] {
				processedAll = false
				break
			}
		}

		if processedAll {
			processedMetaHdrs = append(processedMetaHdrs, metaBlock)
		}
	}

	return processedMetaHdrs, processedMBs, nil
}

func (sp *shardProcessor) removeProcessedMetaBlocksFromPool(processedMetaHdrs []data.HeaderHandler) error {
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
		nonceToByteSlice := sp.uint64Converter.ToByteSlice(hdr.GetNonce())
		err = sp.store.Put(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice, headerHash)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		err = sp.store.Put(dataRetriever.MetaBlockUnit, headerHash, buff)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		sp.dataPool.MetaBlocks().Remove(headerHash)
		sp.dataPool.HeadersNonces().Remove(hdr.GetNonce(), sharding.MetachainShardId)
		sp.removeAllProcessedMiniBlocks(headerHash)

		log.Debug(fmt.Sprintf("metaBlock with round %d nonce %d and hash %s has been processed completely and removed from pool\n",
			hdr.GetRound(),
			hdr.GetNonce(),
			core.ToB64(headerHash)))

		processed++
	}

	if processed > 0 {
		log.Debug(fmt.Sprintf("%d meta blocks have been processed completely and removed from pool\n", processed))
	}

	notarized := unnotarized - len(sp.blocksTracker.UnnotarisedBlocks())
	if notarized > 0 {
		log.Debug(fmt.Sprintf("%d shard blocks have been notarised by metachain\n", notarized))
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

	metaHdrsNoncesCache := sp.dataPool.HeadersNonces()
	if metaHdrsNoncesCache == nil && sp.metaBlockFinality > 0 {
		return
	}

	miniBlksCache := sp.dataPool.MiniBlocks()
	if miniBlksCache == nil || miniBlksCache.IsInterfaceNil() {
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
				log.Info(fmt.Sprintf("received all final meta headers\n"))
				areFinalAttestingHdrsInCache = true
			} else {
				log.Info(fmt.Sprintf("requested %d missing final meta headers\n", requestedBlockHeaders))
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

	sp.txCoordinator.RequestMiniBlocks(metaBlock)
}

// requestFinalMissingHeaders requests the headers needed to accept the current selected headers for processing the
// current block. It requests the metaBlockFinality headers greater than the highest meta header related to the block
// which should be processed
func (sp *shardProcessor) requestFinalMissingHeaders() uint32 {
	requestedBlockHeaders := uint32(0)
	for i := sp.currHighestMetaHdrNonce + 1; i <= sp.currHighestMetaHdrNonce+uint64(sp.metaBlockFinality); i++ {
		if sp.currHighestMetaHdrNonce == uint64(0) {
			continue
		}

		_, _, err := process.GetMetaHeaderFromPoolWithNonce(
			i,
			sp.dataPool.MetaBlocks(),
			sp.dataPool.HeadersNonces())
		if err != nil {
			requestedBlockHeaders++
			go sp.onRequestHeaderHandlerByNonce(sharding.MetachainShardId, i)
		}
	}

	return requestedBlockHeaders
}

func (sp *shardProcessor) requestMetaHeaders(header *block.Header) (uint32, uint32) {
	_ = process.EmptyChannel(sp.chRcvAllMetaHdrs)

	sp.mutRequestedMetaHdrsHashes.Lock()

	sp.requestedMetaHdrsHashes = make(map[string]bool)
	sp.allNeededMetaHdrsFound = true

	if len(header.MetaBlockHashes) == 0 {
		sp.mutRequestedMetaHdrsHashes.Unlock()
		return 0, 0
	}

	missingHeaderHashes := sp.computeMissingHeaders(header)

	requestedBlockHeaders := uint32(0)
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

	sp.mutRequestedMetaHdrsHashes.Unlock()

	return requestedBlockHeaders, requestedFinalBlockHeaders
}

func (sp *shardProcessor) computeMissingHeaders(header *block.Header) [][]byte {
	missingHeaders := make([][]byte, 0)
	sp.currHighestMetaHdrNonce = uint64(0)

	for i := 0; i < len(header.MetaBlockHashes); i++ {
		hdr, err := process.GetMetaHeaderFromPool(
			header.MetaBlockHashes[i],
			sp.dataPool.MetaBlocks())
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

func (sp *shardProcessor) getAllMiniBlockDstMeFromMeta(round uint64, metaHashes [][]byte) (map[string][]byte, error) {
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

func (sp *shardProcessor) getOrderedMetaBlocks(round uint64) ([]*hashAndHdr, error) {
	metaBlocksPool := sp.dataPool.MetaBlocks()
	if metaBlocksPool == nil {
		return nil, process.ErrNilMetaBlockPool
	}

	lastHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, err
	}

	orderedMetaBlocks := make([]*hashAndHdr, 0)
	for _, key := range metaBlocksPool.Keys() {
		val, _ := metaBlocksPool.Peek(key)
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

	if len(orderedMetaBlocks) > 1 {
		sort.Slice(orderedMetaBlocks, func(i, j int) bool {
			return orderedMetaBlocks[i].hdr.GetNonce() < orderedMetaBlocks[j].hdr.GetNonce()
		})
	}

	return orderedMetaBlocks, nil
}

// isMetaHeaderFinal verifies if meta is trully final, in order to not do rollbacks
func (sp *shardProcessor) isMetaHeaderFinal(currHdr data.HeaderHandler, sortedHdrs []*hashAndHdr, startPos int) bool {
	if currHdr == nil || currHdr.IsInterfaceNil() {
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
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, [][]byte, uint32, error) {

	metaBlockCache := sp.dataPool.MetaBlocks()
	if metaBlockCache == nil || metaBlockCache.IsInterfaceNil() {
		return nil, nil, 0, process.ErrNilMetaBlockPool
	}

	miniBlockCache := sp.dataPool.MiniBlocks()
	if miniBlockCache == nil || miniBlockCache.IsInterfaceNil() {
		return nil, nil, 0, process.ErrNilMiniBlockPool
	}

	txPool := sp.dataPool.Transactions()
	if txPool == nil || txPool.IsInterfaceNil() {
		return nil, nil, 0, process.ErrNilTransactionPool
	}

	miniBlocks := make(block.MiniBlockSlice, 0)
	nrTxAdded := uint32(0)

	orderedMetaBlocks, err := sp.getOrderedMetaBlocks(round)
	if err != nil {
		return nil, nil, 0, err
	}

	log.Info(fmt.Sprintf("meta blocks ordered: %d\n", len(orderedMetaBlocks)))

	lastMetaHdr, err := sp.getLastNotarizedHdr(sharding.MetachainShardId)
	if err != nil {
		return nil, nil, 0, err
	}

	// do processing in order
	usedMetaHdrsHashes := make([][]byte, 0)
	for i := 0; i < len(orderedMetaBlocks); i++ {
		if !haveTime() {
			log.Info(fmt.Sprintf("time is up after putting %d cross txs with destination to current shard\n", nrTxAdded))
			break
		}

		itemsAddedInHeader := uint32(len(usedMetaHdrsHashes) + len(miniBlocks))
		if itemsAddedInHeader >= maxItemsInBlock {
			log.Info(fmt.Sprintf("%d max records allowed to be added in shard header has been reached\n", maxItemsInBlock))
			break
		}

		hdr, ok := orderedMetaBlocks[i].hdr.(*block.MetaBlock)
		if !ok {
			continue
		}

		err = sp.isHdrConstructionValid(hdr, lastMetaHdr)
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

		itemsAddedInBody := nrTxAdded
		if itemsAddedInBody >= maxItemsInBlock {
			continue
		}

		maxTxSpaceRemained := int32(maxItemsInBlock) - int32(itemsAddedInBody)
		maxMbSpaceRemained := int32(maxItemsInBlock) - int32(itemsAddedInHeader) - 1

		if maxTxSpaceRemained > 0 && maxMbSpaceRemained > 0 {
			processedMiniBlocksHashes := sp.getProcessedMiniBlocksHashes(orderedMetaBlocks[i].hash)
			currMBProcessed, currTxsAdded, hdrProcessFinished := sp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
				hdr,
				processedMiniBlocksHashes,
				uint32(maxTxSpaceRemained),
				uint32(maxMbSpaceRemained),
				round,
				haveTime)

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
	}

	sp.mutUsedMetaHdrsHashes.Lock()
	sp.usedMetaHdrsHashes[round] = usedMetaHdrsHashes
	sp.mutUsedMetaHdrsHashes.Unlock()

	return miniBlocks, usedMetaHdrsHashes, nrTxAdded, nil
}

func (sp *shardProcessor) createMiniBlocks(
	noShards uint32,
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.Body, error) {

	miniBlocks := make(block.Body, 0)

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

	destMeMiniBlocks, usedMetaHdrsHashes, txs, err := sp.createAndProcessCrossMiniBlocksDstMe(noShards, maxItemsInBlock, round, haveTime)
	if err != nil {
		log.Info(err.Error())
	}

	processedMetaHdrs, errNotCritical := sp.getProcessedMetaBlocksFromMiniBlocks(destMeMiniBlocks, usedMetaHdrsHashes)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	err = sp.setMetaConsensusData(processedMetaHdrs)
	if err != nil {
		return nil, err
	}

	log.Debug(fmt.Sprintf("processed %d miniblocks and %d txs with destination in self shard\n", len(destMeMiniBlocks), txs))

	if len(destMeMiniBlocks) > 0 {
		miniBlocks = append(miniBlocks, destMeMiniBlocks...)
	}

	maxTxSpaceRemained := int32(maxItemsInBlock) - int32(txs)
	maxMbSpaceRemained := int32(maxItemsInBlock) - int32(len(destMeMiniBlocks)) - int32(len(usedMetaHdrsHashes))

	if maxTxSpaceRemained > 0 && maxMbSpaceRemained > 0 {
		mbFromMe := sp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(
			uint32(maxTxSpaceRemained),
			uint32(maxMbSpaceRemained),
			round,
			haveTime)

		if len(mbFromMe) > 0 {
			miniBlocks = append(miniBlocks, mbFromMe...)
		}
	}

	log.Info(fmt.Sprintf("creating mini blocks has been finished: created %d mini blocks\n", len(miniBlocks)))
	return miniBlocks, nil
}

// CreateBlockHeader creates a miniblock header list given a block body
func (sp *shardProcessor) CreateBlockHeader(bodyHandler data.BodyHandler, round uint64, haveTime func() bool) (data.HeaderHandler, error) {
	log.Debug(fmt.Sprintf("started creating block header in round %d\n", round))
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

	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
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

	sp.appStatusHandler.SetUInt64Value(core.MetricNumTxInBlock, uint64(totalTxCount))
	sp.appStatusHandler.SetUInt64Value(core.MetricNumMiniBlocks, uint64(mbLen))

	sp.mutUsedMetaHdrsHashes.Lock()

	if usedMetaHdrsHashes, ok := sp.usedMetaHdrsHashes[round]; ok {
		header.MetaBlockHashes = usedMetaHdrsHashes
		delete(sp.usedMetaHdrsHashes, round)
	}

	sp.mutUsedMetaHdrsHashes.Unlock()

	sp.blockSizeThrottler.Add(
		round,
		core.Max(header.ItemsInBody(), header.ItemsInHeader()))

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
) (map[uint32][]byte, map[string][][]byte, error) {

	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
		return nil, nil, process.ErrNilMiniBlocks
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return nil, nil, process.ErrWrongTypeAssertion
	}

	mrsData := make(map[uint32][]byte)
	bodies, mrsTxs := sp.txCoordinator.CreateMarshalizedData(body)

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

// IsInterfaceNil returns true if there is no value under the interface
func (sp *shardProcessor) IsInterfaceNil() bool {
	if sp == nil {
		return true
	}
	return false
}

func (sp *shardProcessor) addProcessedMiniBlock(metaBlockHash []byte, miniBlockHash []byte) {
	sp.mutProcessedMiniBlocks.Lock()
	miniBlocksProcessed, ok := sp.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		miniBlocksProcessed := make(map[string]struct{})
		miniBlocksProcessed[string(miniBlockHash)] = struct{}{}
		sp.processedMiniBlocks[string(metaBlockHash)] = miniBlocksProcessed
		sp.mutProcessedMiniBlocks.Unlock()
		return
	}

	miniBlocksProcessed[string(miniBlockHash)] = struct{}{}
	sp.mutProcessedMiniBlocks.Unlock()
}

func (sp *shardProcessor) removeProcessedMiniBlock(metaBlockHash []byte, miniBlockHash []byte) {
	sp.mutProcessedMiniBlocks.Lock()
	miniBlocksProcessed, ok := sp.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		sp.mutProcessedMiniBlocks.Unlock()
		return
	}

	delete(miniBlocksProcessed, string(miniBlockHash))
	sp.mutProcessedMiniBlocks.Unlock()
}

func (sp *shardProcessor) removeAllProcessedMiniBlocks(metaBlockHash []byte) {
	sp.mutProcessedMiniBlocks.Lock()
	delete(sp.processedMiniBlocks, string(metaBlockHash))
	sp.mutProcessedMiniBlocks.Unlock()
}

func (sp *shardProcessor) getProcessedMiniBlocksHashes(metaBlockHash []byte) map[string]struct{} {
	sp.mutProcessedMiniBlocks.RLock()
	processedMiniBlocksHashes := sp.processedMiniBlocks[string(metaBlockHash)]
	sp.mutProcessedMiniBlocks.RUnlock()

	return processedMiniBlocksHashes
}

func (sp *shardProcessor) isMiniBlockProcessed(metaBlockHash []byte, miniBlockHash []byte) bool {
	sp.mutProcessedMiniBlocks.RLock()
	miniBlocksProcessed, ok := sp.processedMiniBlocks[string(metaBlockHash)]
	if !ok {
		sp.mutProcessedMiniBlocks.RUnlock()
		return false
	}

	_, isProcessed := miniBlocksProcessed[string(miniBlockHash)]
	sp.mutProcessedMiniBlocks.RUnlock()

	return isProcessed
}
