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

// metaProcessor implements metaProcessor interface and actually it tries to execute block
type metaProcessor struct {
	*baseProcessor
	core     serviceContainer.Core
	dataPool dataRetriever.MetaPoolsHolder

	shardsHeadersNonce *sync.Map
	shardBlockFinality uint32
	chRcvAllHdrs       chan bool
	headersCounter     *headersCounter
}

// NewMetaProcessor creates a new metaProcessor object
func NewMetaProcessor(arguments ArgMetaProcessor) (*metaProcessor, error) {
	err := checkProcessorNilParameters(arguments.ArgBaseProcessor)
	if err != nil {
		return nil, err
	}

	if arguments.DataPool == nil || arguments.DataPool.IsInterfaceNil() {
		return nil, process.ErrNilDataPoolHolder
	}
	if arguments.DataPool.ShardHeaders() == nil || arguments.DataPool.ShardHeaders().IsInterfaceNil() {
		return nil, process.ErrNilHeadersDataPool
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
		onRequestHeaderHandler:        arguments.RequestHandler.RequestHeader,
		onRequestHeaderHandlerByNonce: arguments.RequestHandler.RequestHeaderByNonce,
		appStatusHandler:              statusHandler.NewNilStatusHandler(),
		txCoordinator:                 arguments.TxCoordinator,
	}

	err = base.setLastNotarizedHeadersSlice(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	mp := metaProcessor{
		core:           arguments.Core,
		baseProcessor:  base,
		dataPool:       arguments.DataPool,
		headersCounter: NewHeaderCounter(),
	}

	mp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	mp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64)

	headerPool := mp.dataPool.ShardHeaders()
	headerPool.RegisterHandler(mp.receivedShardHeader)

	mp.chRcvAllHdrs = make(chan bool)

	mp.shardBlockFinality = process.ShardBlockFinality

	mp.shardsHeadersNonce = &sync.Map{}

	return &mp, nil
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	err := mp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		if err == process.ErrBlockHashDoesNotMatch {
			log.Info(fmt.Sprintf("requested missing meta header with hash %s for shard %d\n",
				core.ToB64(headerHandler.GetPrevHash()),
				headerHandler.GetShardID()))

			go mp.onRequestHeaderHandler(headerHandler.GetShardID(), headerHandler.GetPrevHash())
		}

		return err
	}

	log.Debug(fmt.Sprintf("started processing block with round %d and nonce %d\n",
		headerHandler.GetRound(),
		headerHandler.GetNonce()))

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = mp.checkHeaderBodyCorrelation(header, body)
	if err != nil {
		return err
	}

	go getMetricsFromMetaHeader(
		header,
		mp.marshalizer,
		mp.appStatusHandler,
		mp.dataPool.ShardHeaders().Len(),
		mp.headersCounter.getNumShardMBHeadersTotalProcessed(),
	)

	mp.createBlockStarted()
	mp.txCoordinator.RequestBlockTransactions(body)

	requestedShardHdrs, requestedFinalityAttestingShardHdrs := mp.requestShardHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	err = mp.txCoordinator.IsDataPreparedForProcessing(haveTime)
	if err != nil {
		return err
	}

	haveMissingShardHeaders := requestedShardHdrs > 0 || requestedFinalityAttestingShardHdrs > 0
	if haveMissingShardHeaders {
		log.Info(fmt.Sprintf("requested %d missing shard headers and %d finality attesting shard headers\n",
			requestedShardHdrs,
			requestedFinalityAttestingShardHdrs))

		err = mp.waitForBlockHeaders(haveTime())

		mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
		missingShardHdrs := mp.hdrsForCurrBlock.missingHdrs
		mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

		mp.resetMissingHdrs()

		if requestedShardHdrs > 0 {
			log.Info(fmt.Sprintf("received %d missing shard headers\n", requestedShardHdrs-missingShardHdrs))
		}

		if err != nil {
			return err
		}
	}

	if mp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing(header.Round)
	}()

	highestNonceHdrs, err := mp.checkShardHeadersValidity()
	if err != nil {
		return err
	}

	err = mp.checkShardHeadersFinality(highestNonceHdrs)
	if err != nil {
		return err
	}

	err = mp.verifyCrossShardMiniBlockDstMe(header)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			mp.RevertAccountState()
		}
	}()

	err = mp.processBlockHeaders(header, header.Round, haveTime)
	if err != nil {
		return err
	}

	err = mp.txCoordinator.ProcessBlockTransaction(body, header.Round, haveTime)
	if err != nil {
		return err
	}

	if !mp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	err = mp.txCoordinator.VerifyCreatedBlockTransactions(body)
	if err != nil {
		return err
	}

	return nil
}

// check if header has the same miniblocks as presented in body
func (mp *metaProcessor) checkHeaderBodyCorrelation(metaBlock *block.MetaBlock, body block.Body) error {
	mbHashesFromHdr := make(map[string]*block.ShardMiniBlockHeader)
	mbHeaders := 0
	for _, shardInfo := range metaBlock.ShardInfo {
		if shardInfo.ShardID != mp.shardCoordinator.SelfId() {
			continue
		}

		for _, mbHeader := range shardInfo.ShardMiniBlockHeaders {
			mbHashesFromHdr[string(mbHeader.Hash)] = &mbHeader
		}

		mbHeaders += len(shardInfo.ShardMiniBlockHeaders)
	}

	if mbHeaders != len(body) {
		return process.ErrHeaderBodyMismatch
	}

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]

		mbBytes, err := mp.marshalizer.Marshal(miniBlock)
		if err != nil {
			return err
		}
		mbHash := mp.hasher.Compute(string(mbBytes))

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

func (mp *metaProcessor) verifyCrossShardMiniBlockDstMe(header *block.MetaBlock) error {
	miniBlockMetaHashes, err := mp.getAllMiniBlockDstMeta(header)
	if err != nil {
		return err
	}

	crossMiniBlockHashes := header.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())
	for hash := range crossMiniBlockHashes {
		if _, ok := miniBlockMetaHashes[hash]; !ok {
			return process.ErrCrossShardMBWithoutConfirmationFromMeta
		}
	}

	return nil
}

func (mp *metaProcessor) getAllMiniBlockDstMeta(metaHdr *block.MetaBlock) (map[string][]byte, error) {
	miniBlockHashes := make(map[string][]byte)

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	for _, shardInfo := range metaHdr.ShardInfo {
		hdrInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardInfo.HeaderHash)]
		if !ok {
			continue
		}
		shardHeader, ok := hdrInfo.hdr.(*block.Header)
		if !ok {
			continue
		}

		lastHdr, err := mp.getLastNotarizedHdr(shardInfo.ShardID)
		if err != nil {
			return nil, err
		}

		if shardHeader.GetRound() > metaHdr.Round {
			continue
		}
		if shardHeader.GetRound() <= lastHdr.GetRound() {
			continue
		}
		if shardHeader.GetNonce() <= lastHdr.GetNonce() {
			continue
		}

		crossMiniBlockHashes := shardHeader.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())
		for hash := range crossMiniBlockHashes {
			miniBlockHashes[hash] = shardInfo.HeaderHash
		}
	}

	return miniBlockHashes, nil
}

// SetConsensusData - sets the reward addresses for the current consensus group
func (mp *metaProcessor) SetConsensusData(randomness []byte, round uint64, epoch uint32, shardId uint32) {
	// nothing to do
}

func (mp *metaProcessor) checkAndRequestIfShardHeadersMissing(round uint64) {
	_, _, sortedHdrPerShard, err := mp.getOrderedHdrs(round)
	if err != nil {
		log.Debug(err.Error())
		return
	}

	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		// map from *block.Header to dataHandler
		sortedHdrs := make([]data.HeaderHandler, len(sortedHdrPerShard[i]))
		for j := 0; j < len(sortedHdrPerShard[i]); j++ {
			sortedHdrs[j] = sortedHdrPerShard[i][j]
		}

		err := mp.requestHeadersIfMissing(sortedHdrs, i, round)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
	}

	return
}

func (mp *metaProcessor) indexBlock(
	metaBlock data.HeaderHandler,
	lastMetaBlock data.HeaderHandler,
) {
	if mp.core == nil || mp.core.Indexer() == nil {
		return
	}
	// Update tps benchmarks in the DB
	tpsBenchmark := mp.core.TPSBenchmark()
	if tpsBenchmark != nil {
		go mp.core.Indexer().UpdateTPS(tpsBenchmark)
	}

	publicKeys, err := mp.nodesCoordinator.GetValidatorsPublicKeys(metaBlock.GetPrevRandSeed(), metaBlock.GetRound(), sharding.MetachainShardId)
	if err != nil {
		return
	}

	signersIndexes := mp.nodesCoordinator.GetValidatorsIndexes(publicKeys)
	go mp.core.Indexer().SaveMetaBlock(metaBlock, signersIndexes)

	saveRoundInfoInElastic(mp.core.Indexer(), mp.nodesCoordinator, sharding.MetachainShardId, metaBlock, lastMetaBlock, signersIndexes)
}

// removeBlockInfoFromPool removes the block info from associated pools
func (mp *metaProcessor) removeBlockInfoFromPool(header *block.MetaBlock) error {
	if header == nil || header.IsInterfaceNil() {
		return process.ErrNilMetaBlockHeader
	}

	headerPool := mp.dataPool.ShardHeaders()
	if headerPool == nil || headerPool.IsInterfaceNil() {
		return process.ErrNilHeadersDataPool
	}

	headerNoncesPool := mp.dataPool.HeadersNonces()
	if headerNoncesPool == nil || headerNoncesPool.IsInterfaceNil() {
		return process.ErrNilHeadersNoncesDataPool
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for i := 0; i < len(header.ShardInfo); i++ {
		shardHeaderHash := header.ShardInfo[i].HeaderHash
		hdrInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardBlock, ok := hdrInfo.hdr.(*block.Header)
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		headerPool.Remove([]byte(shardHeaderHash))
		headerNoncesPool.Remove(shardBlock.Nonce, shardBlock.ShardId)
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	return nil
}

// RestoreBlockIntoPools restores the block into associated pools
func (mp *metaProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	mp.removeLastNotarized()

	if headerHandler == nil || headerHandler.IsInterfaceNil() {
		return process.ErrNilMetaBlockHeader
	}
	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
		return process.ErrNilTxBlockBody
	}

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	headerPool := mp.dataPool.ShardHeaders()
	if headerPool == nil || headerPool.IsInterfaceNil() {
		return process.ErrNilHeadersDataPool
	}

	headerNoncesPool := mp.dataPool.HeadersNonces()
	if headerNoncesPool == nil || headerNoncesPool.IsInterfaceNil() {
		return process.ErrNilHeadersNoncesDataPool
	}

	hdrHashes := make([][]byte, len(header.ShardInfo))
	for i := 0; i < len(header.ShardInfo); i++ {
		hdrHashes[i] = header.ShardInfo[i].HeaderHash
	}

	_, err := mp.txCoordinator.RestoreBlockDataFromStorage(body)
	if err != nil {
		return err
	}

	for _, hdrHash := range hdrHashes {
		buff, err := mp.store.Get(dataRetriever.BlockHeaderUnit, hdrHash)
		if err != nil {
			return err
		}

		hdr := block.Header{}
		err = mp.marshalizer.Unmarshal(&hdr, buff)
		if err != nil {
			return err
		}

		headerPool.Put(hdrHash, &hdr)
		syncMap := &dataPool.ShardIdHashSyncMap{}
		syncMap.Store(hdr.ShardId, hdrHash)
		headerNoncesPool.Merge(hdr.Nonce, syncMap)

		err = mp.store.GetStorer(dataRetriever.BlockHeaderUnit).Remove(hdrHash)
		if err != nil {
			return err
		}

		nonceToByteSlice := mp.uint64Converter.ToByteSlice(hdr.Nonce)
		err = mp.store.GetStorer(dataRetriever.ShardHdrNonceHashDataUnit).Remove(nonceToByteSlice)
		if err != nil {
			return err
		}

		mp.headersCounter.subtractRestoredMBHeaders(len(hdr.MiniBlockHeaders))
	}

	return nil
}

// CreateBlockBody creates block body of metachain
func (mp *metaProcessor) CreateBlockBody(round uint64, haveTime func() bool) (data.BodyHandler, error) {
	log.Debug(fmt.Sprintf("started creating block body in round %d\n", round))
	mp.createBlockStarted()
	mp.blockSizeThrottler.ComputeMaxItems()

	miniBlocks, err := mp.createMiniBlocks(mp.blockSizeThrottler.MaxItemsToAdd(), round, haveTime)
	if err != nil {
		return nil, err
	}

	return miniBlocks, nil
}

func (mp *metaProcessor) createMiniBlocks(
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.Body, error) {

	miniBlocks := make(block.Body, 0)

	if mp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after entered in createMiniBlocks method\n"))
		return nil, process.ErrTimeIsOut
	}

	txPool := mp.dataPool.Transactions()
	if txPool == nil {
		return nil, process.ErrNilTransactionPool
	}

	destMeMiniBlocks, nbTxs, nbHdrs, err := mp.createAndProcessCrossMiniBlocksDstMe(maxItemsInBlock, round, haveTime)
	if err != nil {
		log.Info(err.Error())
	}

	log.Info(fmt.Sprintf("processed %d miniblocks and %d txs with destination in self shard\n", len(destMeMiniBlocks), nbTxs))

	if len(destMeMiniBlocks) > 0 {
		miniBlocks = append(miniBlocks, destMeMiniBlocks...)
	}

	maxTxSpaceRemained := int32(maxItemsInBlock) - int32(nbTxs)
	maxMbSpaceRemained := mp.getMaxMiniBlocksSpaceRemained(
		maxItemsInBlock,
		uint32(len(destMeMiniBlocks))+nbHdrs,
		uint32(len(miniBlocks)))

	mbFromMe := mp.txCoordinator.CreateMbsAndProcessTransactionsFromMe(
		uint32(maxTxSpaceRemained),
		uint32(maxMbSpaceRemained),
		round,
		haveTime)

	if len(mbFromMe) > 0 {
		miniBlocks = append(miniBlocks, mbFromMe...)
	}

	return miniBlocks, nil
}

// full verification through metachain header
func (mp *metaProcessor) createAndProcessCrossMiniBlocksDstMe(
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) (block.MiniBlockSlice, uint32, uint32, error) {

	miniBlocks := make(block.MiniBlockSlice, 0)
	txsAdded := uint32(0)
	hdrsAdded := uint32(0)
	lastPushedHdr := make(map[uint32]data.HeaderHandler, mp.shardCoordinator.NumberOfShards())

	orderedHdrs, orderedHdrHashes, sortedHdrPerShard, err := mp.getOrderedHdrs(round)
	if err != nil {
		return nil, 0, 0, err
	}

	// save last committed header for verification
	mp.mutNotarizedHdrs.RLock()
	if mp.notarizedHdrs == nil {
		mp.mutNotarizedHdrs.RUnlock()
		return nil, 0, 0, process.ErrNotarizedHdrsSliceIsNil
	}
	for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
		lastPushedHdr[shardId] = mp.lastNotarizedHdrForShard(shardId)
	}
	mp.mutNotarizedHdrs.RUnlock()

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(orderedHdrs); i++ {
		if !haveTime() {
			log.Info(fmt.Sprintf("time is up after putting %d cross txs with destination to current shard\n", txsAdded))
			break
		}

		if len(miniBlocks) >= core.MaxMiniBlocksInBlock {
			log.Info(fmt.Sprintf("%d max number of mini blocks allowed to be added in one shard block has been reached\n", len(miniBlocks)))
			break
		}

		itemsAddedInHeader := uint32(len(mp.hdrsForCurrBlock.hdrHashAndInfo) + len(miniBlocks))
		if itemsAddedInHeader >= maxItemsInBlock {
			log.Info(fmt.Sprintf("%d max records allowed to be added in shard header has been reached\n", maxItemsInBlock))
			break
		}

		hdr := orderedHdrs[i]
		lastHdr, ok := lastPushedHdr[hdr.ShardId].(*block.Header)
		if !ok {
			continue
		}

		isFinal, _ := mp.isShardHeaderValidFinal(orderedHdrs[i], lastHdr, sortedHdrPerShard[hdr.ShardId])
		if !isFinal {
			continue
		}

		if len(hdr.GetMiniBlockHeadersWithDst(mp.shardCoordinator.SelfId())) == 0 {
			mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrHashes[i])] = &hdrInfo{hdr: hdr, usedInBlock: true}
			hdrsAdded++
			lastPushedHdr[hdr.ShardId] = hdr
			continue
		}

		itemsAddedInBody := txsAdded
		if itemsAddedInBody >= maxItemsInBlock {
			continue
		}

		maxTxSpaceRemained := int32(maxItemsInBlock) - int32(itemsAddedInBody)
		maxMbSpaceRemained := mp.getMaxMiniBlocksSpaceRemained(
			maxItemsInBlock,
			itemsAddedInHeader+1,
			uint32(len(miniBlocks)))

		if maxTxSpaceRemained > 0 && maxMbSpaceRemained > 0 {
			snapshot := mp.accounts.JournalLen()
			currMBProcessed, currTxsAdded, hdrProcessFinished := mp.txCoordinator.CreateMbsAndProcessCrossShardTransactionsDstMe(
				hdr,
				nil,
				uint32(maxTxSpaceRemained),
				uint32(maxMbSpaceRemained),
				round,
				haveTime)

			if !hdrProcessFinished {
				// shard header must be processed completely
				errAccountState := mp.accounts.RevertToSnapshot(snapshot)
				if errAccountState != nil {
					// TODO: evaluate if reloading the trie from disk will might solve the problem
					log.Error(errAccountState.Error())
				}
				break
			}

			// all txs processed, add to processed miniblocks
			miniBlocks = append(miniBlocks, currMBProcessed...)
			txsAdded = txsAdded + currTxsAdded

			mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrHashes[i])] = &hdrInfo{hdr: hdr, usedInBlock: true}
			hdrsAdded++

			lastPushedHdr[hdr.ShardId] = hdr
		}
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return miniBlocks, txsAdded, hdrsAdded, nil
}

func (mp *metaProcessor) processBlockHeaders(header *block.MetaBlock, round uint64, haveTime func() time.Duration) error {
	msg := ""
	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		for j := 0; j < len(shardData.ShardMiniBlockHeaders); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			headerHash := shardData.HeaderHash
			shardMiniBlockHeader := &shardData.ShardMiniBlockHeaders[j]
			err := mp.checkAndProcessShardMiniBlockHeader(
				headerHash,
				shardMiniBlockHeader,
				round,
				shardData.ShardID,
			)

			if err != nil {
				return err
			}

			msg = fmt.Sprintf("%s\n%s", msg, core.ToB64(shardMiniBlockHeader.Hash))
		}
	}

	if len(msg) > 0 {
		log.Debug(fmt.Sprintf("the following miniblocks hashes were successfully processed:%s\n", msg))
	}

	return nil
}

// CommitBlock commits the block in the blockchain if everything was checked successfully
func (mp *metaProcessor) CommitBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	var err error
	defer func() {
		if err != nil {
			mp.RevertAccountState()
		}
	}()

	err = checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	log.Debug(fmt.Sprintf("started committing block with round %d and nonce %d\n",
		headerHandler.GetRound(),
		headerHandler.GetNonce()))

	err = mp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	buff, err := mp.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	headerHash := mp.hasher.Compute(string(buff))
	nonceToByteSlice := mp.uint64Converter.ToByteSlice(header.Nonce)
	errNotCritical := mp.store.Put(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	log.LogIfError(errNotCritical)

	errNotCritical = mp.store.Put(dataRetriever.MetaBlockUnit, headerHash, buff)
	log.LogIfError(errNotCritical)

	headerNoncePool := mp.dataPool.HeadersNonces()
	if headerNoncePool == nil {
		err = process.ErrNilHeadersNoncesDataPool
		return err
	}

	metaBlockPool := mp.dataPool.MetaBlocks()
	if metaBlockPool == nil {
		err = process.ErrNilMetaBlockPool
		return err
	}

	headerNoncePool.Remove(header.GetNonce(), header.GetShardID())
	metaBlockPool.Remove(headerHash)

	body, ok := bodyHandler.(block.Body)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	err = mp.txCoordinator.SaveBlockDataToStorage(body)
	if err != nil {
		return err
	}

	for i := 0; i < len(body); i++ {
		buff, err = mp.marshalizer.Marshal(body[i])
		if err != nil {
			return err
		}

		miniBlockHash := mp.hasher.Compute(string(buff))
		errNotCritical = mp.store.Put(dataRetriever.MiniBlockUnit, miniBlockHash, buff)
		log.LogIfError(errNotCritical)
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for i := 0; i < len(header.ShardInfo); i++ {
		shardHeaderHash := header.ShardInfo[i].HeaderHash
		hdrInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardBlock, ok := hdrInfo.hdr.(*block.Header)
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		mp.updateShardHeadersNonce(shardBlock.ShardId, shardBlock.Nonce)

		buff, err = mp.marshalizer.Marshal(shardBlock)
		if err != nil {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return err
		}

		nonceToByteSlice := mp.uint64Converter.ToByteSlice(shardBlock.Nonce)
		hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(shardBlock.ShardId)
		errNotCritical = mp.store.Put(hdrNonceHashDataUnit, nonceToByteSlice, shardHeaderHash)
		log.LogIfError(errNotCritical)

		errNotCritical = mp.store.Put(dataRetriever.BlockHeaderUnit, shardHeaderHash, buff)
		log.LogIfError(errNotCritical)
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	mp.saveMetricCrossCheckBlockHeight()

	err = mp.saveLastNotarizedHeader(header)
	if err != nil {
		return err
	}

	_, err = mp.accounts.Commit()
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("meta block with nonce %d and hash %s has been committed successfully\n",
		header.Nonce,
		core.ToB64(headerHash)))

	errNotCritical = mp.removeBlockInfoFromPool(header)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	errNotCritical = mp.txCoordinator.RemoveBlockDataFromPool(body)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	errNotCritical = mp.forkDetector.AddHeader(header, headerHash, process.BHProcessed, nil, nil)
	if errNotCritical != nil {
		log.Debug(errNotCritical.Error())
	}

	log.Info(fmt.Sprintf("meta block with nonce %d is the highest final block in shard %d\n",
		mp.forkDetector.GetHighestFinalBlockNonce(),
		mp.shardCoordinator.SelfId()))

	hdrsToAttestPreviousFinal := mp.shardBlockFinality + 1
	mp.removeNotarizedHdrsBehindPreviousFinal(hdrsToAttestPreviousFinal)

	lastMetaBlock := chainHandler.GetCurrentBlockHeader()

	err = chainHandler.SetCurrentBlockBody(body)
	if err != nil {
		return err
	}

	err = chainHandler.SetCurrentBlockHeader(header)
	if err != nil {
		return err
	}

	chainHandler.SetCurrentBlockHeaderHash(headerHash)

	if mp.core != nil && mp.core.TPSBenchmark() != nil {
		mp.core.TPSBenchmark().Update(header)
	}

	mp.indexBlock(header, lastMetaBlock)

	saveMetachainCommitBlockMetrics(mp.appStatusHandler, header, headerHash, mp.nodesCoordinator)

	go mp.headersCounter.displayLogInfo(
		header,
		headerHash,
		mp.dataPool.ShardHeaders().Len(),
	)

	mp.blockSizeThrottler.Succeed(header.Round)

	return nil
}

func (mp *metaProcessor) updateShardHeadersNonce(key uint32, value uint64) {
	valueStoredI, ok := mp.shardsHeadersNonce.Load(key)
	if !ok {
		mp.shardsHeadersNonce.Store(key, value)
		return
	}

	valueStored, ok := valueStoredI.(uint64)
	if !ok {
		mp.shardsHeadersNonce.Store(key, value)
		return
	}

	if valueStored < value {
		mp.shardsHeadersNonce.Store(key, value)
	}
}

func (mp *metaProcessor) saveMetricCrossCheckBlockHeight() {
	crossCheckBlockHeight := ""
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		valueStoredI, ok := mp.shardsHeadersNonce.Load(i)
		if !ok {
			continue
		}

		valueStored, ok := valueStoredI.(uint64)
		if !ok {
			continue
		}

		crossCheckBlockHeight += fmt.Sprintf("%d: %d, ", i, valueStored)
	}

	mp.appStatusHandler.SetStringValue(core.MetricCrossCheckBlockHeight, crossCheckBlockHeight)
}

func (mp *metaProcessor) saveLastNotarizedHeader(header *block.MetaBlock) error {
	mp.mutNotarizedHdrs.Lock()
	defer mp.mutNotarizedHdrs.Unlock()

	if mp.notarizedHdrs == nil {
		return process.ErrNotarizedHdrsSliceIsNil
	}

	tmpLastNotarizedHdrForShard := make(map[uint32]data.HeaderHandler, mp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		tmpLastNotarizedHdrForShard[i] = mp.lastNotarizedHdrForShard(i)
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for i := 0; i < len(header.ShardInfo); i++ {
		shardHeaderHash := header.ShardInfo[i].HeaderHash
		hdrInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardHdr, ok := hdrInfo.hdr.(*block.Header)
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		if tmpLastNotarizedHdrForShard[shardHdr.ShardId].GetNonce() < shardHdr.Nonce {
			tmpLastNotarizedHdrForShard[shardHdr.ShardId] = shardHdr
		}
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		mp.notarizedHdrs[i] = append(mp.notarizedHdrs[i], tmpLastNotarizedHdrForShard[i])
		DisplayLastNotarized(mp.marshalizer, mp.hasher, tmpLastNotarizedHdrForShard[i], i)
	}

	return nil
}

// check if shard headers were signed and constructed correctly and returns headers which has to be
// checked for finality
func (mp *metaProcessor) checkShardHeadersValidity() (map[uint32]data.HeaderHandler, error) {
	mp.mutNotarizedHdrs.RLock()
	if mp.notarizedHdrs == nil {
		mp.mutNotarizedHdrs.RUnlock()
		return nil, process.ErrNotarizedHdrsSliceIsNil
	}

	tmpLastNotarized := make(map[uint32]data.HeaderHandler, mp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		tmpLastNotarized[i] = mp.lastNotarizedHdrForShard(i)
	}
	mp.mutNotarizedHdrs.RUnlock()

	highestNonceHdrs := make(map[uint32]data.HeaderHandler)

	usedShardHdrs := mp.sortHeadersForCurrentBlockByNonce(true)
	if len(usedShardHdrs) == 0 {
		return highestNonceHdrs, nil
	}

	for shardId, hdrsForShard := range usedShardHdrs {
		for _, shardHdr := range hdrsForShard {
			err := mp.isHdrConstructionValid(shardHdr, tmpLastNotarized[shardId])
			if err != nil {
				return nil, err
			}

			tmpLastNotarized[shardId] = shardHdr
			highestNonceHdrs[shardId] = shardHdr
		}
	}

	return highestNonceHdrs, nil
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (mp *metaProcessor) checkShardHeadersFinality(highestNonceHdrs map[uint32]data.HeaderHandler) error {
	finalityAttestingShardHdrs := mp.sortHeadersForCurrentBlockByNonce(false)

	for shardId, lastVerifiedHdr := range highestNonceHdrs {
		if lastVerifiedHdr == nil || lastVerifiedHdr.IsInterfaceNil() {
			return process.ErrNilBlockHeader
		}
		if lastVerifiedHdr.GetShardID() != shardId {
			return process.ErrShardIdMissmatch
		}

		// verify if there are "K" block after current to make this one final
		nextBlocksVerified := uint32(0)
		for _, shardHdr := range finalityAttestingShardHdrs[shardId] {
			if nextBlocksVerified >= mp.shardBlockFinality {
				break
			}

			// found a header with the next nonce
			if shardHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
				err := mp.isHdrConstructionValid(shardHdr, lastVerifiedHdr)
				if err != nil {
					log.Debug(err.Error())
					continue
				}

				lastVerifiedHdr = shardHdr
				nextBlocksVerified += 1
			}
		}

		if nextBlocksVerified < mp.shardBlockFinality {
			go mp.onRequestHeaderHandlerByNonce(lastVerifiedHdr.GetShardID(), lastVerifiedHdr.GetNonce()+1)
			return process.ErrHeaderNotFinal
		}
	}

	return nil
}

func (mp *metaProcessor) isShardHeaderValidFinal(currHdr *block.Header, lastHdr *block.Header, sortedShardHdrs []*block.Header) (bool, []uint32) {
	if currHdr == nil {
		return false, nil
	}
	if sortedShardHdrs == nil {
		return false, nil
	}
	if lastHdr == nil {
		return false, nil
	}

	err := mp.isHdrConstructionValid(currHdr, lastHdr)
	if err != nil {
		return false, nil
	}

	// verify if there are "K" block after current to make this one final
	lastVerifiedHdr := currHdr
	nextBlocksVerified := uint32(0)
	hdrIds := make([]uint32, 0)
	for i := 0; i < len(sortedShardHdrs); i++ {
		if nextBlocksVerified >= mp.shardBlockFinality {
			return true, hdrIds
		}

		// found a header with the next nonce
		tmpHdr := sortedShardHdrs[i]
		if tmpHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
			err := mp.isHdrConstructionValid(tmpHdr, lastVerifiedHdr)
			if err != nil {
				continue
			}

			lastVerifiedHdr = tmpHdr
			nextBlocksVerified += 1
			hdrIds = append(hdrIds, uint32(i))
		}
	}

	if nextBlocksVerified >= mp.shardBlockFinality {
		return true, hdrIds
	}

	return false, nil
}

// receivedShardHeader is a call back function which is called when a new header
// is added in the headers pool
func (mp *metaProcessor) receivedShardHeader(shardHeaderHash []byte) {
	shardHeaderPool := mp.dataPool.ShardHeaders()
	if shardHeaderPool == nil {
		return
	}

	obj, ok := shardHeaderPool.Peek(shardHeaderHash)
	if !ok {
		return
	}

	shardHeader, ok := obj.(*block.Header)
	if !ok {
		return
	}

	log.Debug(fmt.Sprintf("received shard block with hash %s and nonce %d from network\n",
		core.ToB64(shardHeaderHash),
		shardHeader.Nonce))

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()

	haveMissingShardHeaders := mp.hdrsForCurrBlock.missingHdrs > 0 || mp.hdrsForCurrBlock.missingFinalityAttestingHdrs > 0
	if haveMissingShardHeaders {
		hdrInfoForHash := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		receivedMissingShardHeader := hdrInfoForHash != nil && (hdrInfoForHash.hdr == nil || hdrInfoForHash.hdr.IsInterfaceNil())
		if receivedMissingShardHeader {
			hdrInfoForHash.hdr = shardHeader
			mp.hdrsForCurrBlock.missingHdrs--

			if shardHeader.Nonce > mp.hdrsForCurrBlock.highestHdrNonce[shardHeader.ShardId] {
				mp.hdrsForCurrBlock.highestHdrNonce[shardHeader.ShardId] = shardHeader.Nonce
			}
		}

		if mp.hdrsForCurrBlock.missingHdrs == 0 {
			missingFinalityAttestingShardHdrs := mp.hdrsForCurrBlock.missingFinalityAttestingHdrs
			mp.hdrsForCurrBlock.missingFinalityAttestingHdrs = mp.requestMissingFinalityAttestingHeaders()
			if mp.hdrsForCurrBlock.missingFinalityAttestingHdrs == 0 {
				log.Info(fmt.Sprintf("received %d missing finality attesting shard headers\n", missingFinalityAttestingShardHdrs))
			} else {
				log.Info(fmt.Sprintf("requested %d missing finality attesting shard headers\n", mp.hdrsForCurrBlock.missingFinalityAttestingHdrs))
			}
		}

		missingShardHdrs := mp.hdrsForCurrBlock.missingHdrs
		missingFinalityAttestingShardHdrs := mp.hdrsForCurrBlock.missingFinalityAttestingHdrs
		mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

		allMissingShardHeadersReceived := missingShardHdrs == 0 && missingFinalityAttestingShardHdrs == 0
		if allMissingShardHeadersReceived {
			mp.chRcvAllHdrs <- true
		}
	} else {
		mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
	}

	go mp.txCoordinator.RequestMiniBlocks(shardHeader)
}

// requestMissingFinalityAttestingHeaders requests the headers needed to accept the current selected headers for processing the
// current block. It requests the shardBlockFinality headers greater than the highest shard header, for each shard, related
// to the block which should be processed
func (mp *metaProcessor) requestMissingFinalityAttestingHeaders() uint32 {
	requestedBlockHeaders := uint32(0)
	for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
		highestHdrNonce := mp.hdrsForCurrBlock.highestHdrNonce[shardId]
		if highestHdrNonce == uint64(0) {
			continue
		}

		lastFinalityAttestingHeader := mp.hdrsForCurrBlock.highestHdrNonce[shardId] + uint64(mp.shardBlockFinality)
		for i := highestHdrNonce + 1; i <= lastFinalityAttestingHeader; i++ {
			shardHeader, shardHeaderHash, err := process.GetShardHeaderFromPoolWithNonce(
				i,
				shardId,
				mp.dataPool.ShardHeaders(),
				mp.dataPool.HeadersNonces())

			if err != nil {
				requestedBlockHeaders++
				go mp.onRequestHeaderHandlerByNonce(shardId, i)
				continue
			}

			mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)] = &hdrInfo{hdr: shardHeader, usedInBlock: false}
		}
	}

	return requestedBlockHeaders
}

func (mp *metaProcessor) requestShardHeaders(metaBlock *block.MetaBlock) (uint32, uint32) {
	_ = process.EmptyChannel(mp.chRcvAllHdrs)

	if len(metaBlock.ShardInfo) == 0 {
		return 0, 0
	}

	missingHeaderHashes := mp.computeMissingAndExistingShardHeaders(metaBlock)

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for shardId, shardHeaderHashes := range missingHeaderHashes {
		for _, hash := range shardHeaderHashes {
			mp.hdrsForCurrBlock.hdrHashAndInfo[string(hash)] = &hdrInfo{hdr: nil, usedInBlock: true}
			go mp.onRequestHeaderHandler(shardId, hash)
		}
	}

	if mp.hdrsForCurrBlock.missingHdrs == 0 {
		mp.hdrsForCurrBlock.missingFinalityAttestingHdrs = mp.requestMissingFinalityAttestingHeaders()
	}

	requestedHdrs := mp.hdrsForCurrBlock.missingHdrs
	requestedFinalityAttestingHdrs := mp.hdrsForCurrBlock.missingFinalityAttestingHdrs
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return requestedHdrs, requestedFinalityAttestingHdrs
}

func (mp *metaProcessor) computeMissingAndExistingShardHeaders(metaBlock *block.MetaBlock) map[uint32][][]byte {
	missingHeadersHashes := make(map[uint32][][]byte)

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for i := 0; i < len(metaBlock.ShardInfo); i++ {
		shardData := metaBlock.ShardInfo[i]
		hdr, err := process.GetShardHeaderFromPool(
			shardData.HeaderHash,
			mp.dataPool.ShardHeaders())

		if err != nil {
			missingHeadersHashes[shardData.ShardID] = append(missingHeadersHashes[shardData.ShardID], shardData.HeaderHash)
			mp.hdrsForCurrBlock.missingHdrs++
			continue
		}

		mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)] = &hdrInfo{hdr: hdr, usedInBlock: true}

		if hdr.Nonce > mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardID] {
			mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardID] = hdr.Nonce
		}
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	return missingHeadersHashes
}

func (mp *metaProcessor) checkAndProcessShardMiniBlockHeader(
	headerHash []byte,
	shardMiniBlockHeader *block.ShardMiniBlockHeader,
	round uint64,
	shardId uint32,
) error {
	// TODO: real processing has to be done here, using metachain state
	return nil
}

func (mp *metaProcessor) createShardInfo(
	round uint64,
	haveTime func() bool,
) ([]block.ShardData, error) {

	shardInfo := make([]block.ShardData, 0)
	lastPushedHdr := make(map[uint32]data.HeaderHandler, mp.shardCoordinator.NumberOfShards())

	if mp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	log.Info(fmt.Sprintf("creating shard info has been started \n"))

	// save last committed header for verification
	mp.mutNotarizedHdrs.RLock()
	if mp.notarizedHdrs == nil {
		mp.mutNotarizedHdrs.RUnlock()
		return nil, process.ErrNotarizedHdrsSliceIsNil
	}
	for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
		lastPushedHdr[shardId] = mp.lastNotarizedHdrForShard(shardId)
	}
	mp.mutNotarizedHdrs.RUnlock()

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	for hdrHash, hdrInfo := range mp.hdrsForCurrBlock.hdrHashAndInfo {
		shardHdr, ok := hdrInfo.hdr.(*block.Header)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		shardData := block.ShardData{}
		shardData.ShardMiniBlockHeaders = make([]block.ShardMiniBlockHeader, 0)
		shardData.TxCount = shardHdr.TxCount
		shardData.ShardID = shardHdr.ShardId
		shardData.HeaderHash = []byte(hdrHash)

		for i := 0; i < len(shardHdr.MiniBlockHeaders); i++ {
			shardMiniBlockHeader := block.ShardMiniBlockHeader{}
			shardMiniBlockHeader.SenderShardID = shardHdr.MiniBlockHeaders[i].SenderShardID
			shardMiniBlockHeader.ReceiverShardID = shardHdr.MiniBlockHeaders[i].ReceiverShardID
			shardMiniBlockHeader.Hash = shardHdr.MiniBlockHeaders[i].Hash
			shardMiniBlockHeader.TxCount = shardHdr.MiniBlockHeaders[i].TxCount

			// execute shard miniblock to change the trie root hash
			err := mp.checkAndProcessShardMiniBlockHeader(
				[]byte(hdrHash),
				&shardMiniBlockHeader,
				round,
				shardData.ShardID,
			)

			if err != nil {
				return nil, err
			}

			shardData.ShardMiniBlockHeaders = append(shardData.ShardMiniBlockHeaders, shardMiniBlockHeader)
		}

		shardInfo = append(shardInfo, shardData)
		mp.hdrsForCurrBlock.hdrHashAndInfo[string(hdrHash)].usedInBlock = true
	}
	mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	log.Info(fmt.Sprintf("creating shard info has been finished: created %d shard data\n", len(shardInfo)))
	return shardInfo, nil
}

func (mp *metaProcessor) createPeerInfo() ([]block.PeerData, error) {
	// TODO: to be implemented
	peerInfo := make([]block.PeerData, 0)
	return peerInfo, nil
}

// CreateBlockHeader creates a miniblock header list given a block body
func (mp *metaProcessor) CreateBlockHeader(bodyHandler data.BodyHandler, round uint64, haveTime func() bool) (data.HeaderHandler, error) {
	log.Debug(fmt.Sprintf("started creating block header in round %d\n", round))

	header := &block.MetaBlock{
		ShardInfo:    make([]block.ShardData, 0),
		PeerInfo:     make([]block.PeerData, 0),
		PrevRandSeed: make([]byte, 0),
		RandSeed:     make([]byte, 0),
	}

	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing(round)
	}()

	shardInfo, err := mp.createShardInfo(round, haveTime)
	if err != nil {
		return nil, err
	}

	peerInfo, err := mp.createPeerInfo()
	if err != nil {
		return nil, err
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	totalTxCount, miniBlockHeaders, err := mp.createMiniBlockHeaders(body)
	if err != nil {
		return nil, err
	}

	header.MiniBlockHeaders = miniBlockHeaders
	header.ShardInfo = shardInfo
	header.PeerInfo = peerInfo
	header.RootHash = mp.getRootHash()
	header.TxCount = getTxCount(shardInfo) + uint32(totalTxCount)

	mp.blockSizeThrottler.Add(
		round,
		core.MaxUint32(header.ItemsInBody(), header.ItemsInHeader()))

	return header, nil
}

func (mp *metaProcessor) waitForBlockHeaders(waitTime time.Duration) error {
	select {
	case <-mp.chRcvAllHdrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

// MarshalizedDataToBroadcast prepares underlying data into a marshalized object according to destination
func (mp *metaProcessor) MarshalizedDataToBroadcast(
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
	bodies, mrsTxs := mp.txCoordinator.CreateMarshalizedData(body)

	for shardId, subsetBlockBody := range bodies {
		buff, err := mp.marshalizer.Marshal(subsetBlockBody)
		if err != nil {
			log.Debug(process.ErrMarshalWithoutSuccess.Error())
			continue
		}
		mrsData[shardId] = buff
	}

	return mrsData, mrsTxs, nil
}

func (mp *metaProcessor) getOrderedHdrs(round uint64) ([]*block.Header, [][]byte, map[uint32][]*block.Header, error) {
	shardBlocksPool := mp.dataPool.ShardHeaders()
	if shardBlocksPool == nil {
		return nil, nil, nil, process.ErrNilShardBlockPool
	}

	hashAndBlockMap := make(map[uint32][]*hashAndHdr)
	headersMap := make(map[uint32][]*block.Header)
	headers := make([]*block.Header, 0)
	hdrHashes := make([][]byte, 0)

	mp.mutNotarizedHdrs.RLock()
	if mp.notarizedHdrs == nil {
		mp.mutNotarizedHdrs.RUnlock()
		return nil, nil, nil, process.ErrNotarizedHdrsSliceIsNil
	}

	// get keys and arrange them into shards
	for _, key := range shardBlocksPool.Keys() {
		val, _ := shardBlocksPool.Peek(key)
		if val == nil {
			continue
		}

		hdr, ok := val.(*block.Header)
		if !ok {
			continue
		}

		if hdr.GetRound() > round {
			continue
		}

		currShardId := hdr.ShardId
		if mp.lastNotarizedHdrForShard(currShardId) == nil {
			continue
		}

		if hdr.GetRound() <= mp.lastNotarizedHdrForShard(currShardId).GetRound() {
			continue
		}

		if hdr.GetNonce() <= mp.lastNotarizedHdrForShard(currShardId).GetNonce() {
			continue
		}

		hashAndBlockMap[currShardId] = append(hashAndBlockMap[currShardId],
			&hashAndHdr{hdr: hdr, hash: key})
	}
	mp.mutNotarizedHdrs.RUnlock()

	// sort headers for each shard
	maxHdrLen := 0
	for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
		hdrsForShard := hashAndBlockMap[shardId]
		if len(hdrsForShard) == 0 {
			continue
		}

		sort.Slice(hdrsForShard, func(i, j int) bool {
			return hdrsForShard[i].hdr.GetNonce() < hdrsForShard[j].hdr.GetNonce()
		})

		tmpHdrLen := len(hdrsForShard)
		if maxHdrLen < tmpHdrLen {
			maxHdrLen = tmpHdrLen
		}
	}

	// copy from map to lists - equality between number of headers per shard
	for i := 0; i < maxHdrLen; i++ {
		for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
			hdrsForShard := hashAndBlockMap[shardId]
			if i >= len(hdrsForShard) {
				continue
			}

			hdr, ok := hdrsForShard[i].hdr.(*block.Header)
			if !ok {
				continue
			}

			headers = append(headers, hdr)
			hdrHashes = append(hdrHashes, hdrsForShard[i].hash)
			headersMap[shardId] = append(headersMap[shardId], hdr)
		}
	}

	return headers, hdrHashes, headersMap, nil
}

func getTxCount(shardInfo []block.ShardData) uint32 {
	txs := uint32(0)
	for i := 0; i < len(shardInfo); i++ {
		for j := 0; j < len(shardInfo[i].ShardMiniBlockHeaders); j++ {
			txs += shardInfo[i].ShardMiniBlockHeaders[j].TxCount
		}
	}

	return txs
}

// DecodeBlockBody method decodes block body from a given byte array
func (mp *metaProcessor) DecodeBlockBody(dta []byte) data.BodyHandler {
	if dta == nil {
		return nil
	}

	var body block.Body

	err := mp.marshalizer.Unmarshal(&body, dta)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return body
}

// DecodeBlockHeader method decodes block header from a given byte array
func (mp *metaProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	var header block.MetaBlock

	err := mp.marshalizer.Unmarshal(&header, dta)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &header
}

// IsInterfaceNil returns true if there is no value under the interface
func (mp *metaProcessor) IsInterfaceNil() bool {
	if mp == nil {
		return true
	}
	return false
}
