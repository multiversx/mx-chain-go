package block

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
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
	core              serviceContainer.Core
	dataPool          dataRetriever.MetaPoolsHolder
	pendingMiniBlocks process.PendingMiniBlocksHandler

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

	if check.IfNil(arguments.DataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(arguments.DataPool.ShardHeaders()) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(arguments.PendingMiniBlocks) {
		return nil, process.ErrNilPendingMiniBlocksHandler
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
		validatorStatisticsProcessor:  arguments.ValidatorStatisticsProcessor,
		endOfEpochTrigger:             arguments.EndOfEpochTrigger,
	}

	err = base.setLastNotarizedHeadersSlice(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	mp := metaProcessor{
		core:              arguments.Core,
		baseProcessor:     base,
		dataPool:          arguments.DataPool,
		headersCounter:    NewHeaderCounter(),
		pendingMiniBlocks: arguments.PendingMiniBlocks,
	}

	mp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	mp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64)
	mp.hdrsForCurrBlock.requestedFinalityAttestingHdrs = make(map[uint32][]uint64)

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

	go getMetricsFromMetaHeader(
		header,
		mp.marshalizer,
		mp.appStatusHandler,
		mp.dataPool.ShardHeaders().Len(),
		mp.headersCounter.getNumShardMBHeadersTotalProcessed(),
	)

	mp.createBlockStarted()

	requestedShardHdrs, requestedFinalityAttestingShardHdrs := mp.requestShardHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
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

	highestNonceHdrs, err := mp.checkShardHeadersValidity(header)
	if err != nil {
		return err
	}

	err = mp.checkShardHeadersFinality(highestNonceHdrs)
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

	if !mp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateDoesNotMatch
		return err
	}

	validatorStatsRH, err := mp.validatorStatisticsProcessor.UpdatePeerState(header)
	if err != nil {
		return err
	}

	if !bytes.Equal(validatorStatsRH, header.GetValidatorStatsRootHash()) {
		err = process.ErrValidatorStatsRootHashDoesNotMatch
		return err
	}

	err = mp.pendingMiniBlocks.AddProcessedHeader(header)
	if err != nil {
		return err
	}

	return nil
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

		err := mp.requestHeadersIfMissing(sortedHdrs, i, round, mp.dataPool.ShardHeaders())
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
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardBlock, ok := headerInfo.hdr.(*block.Header)
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrWrongTypeAssertion
		}

		headerPool.Remove(shardHeaderHash)
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

	header, ok := headerHandler.(*block.MetaBlock)
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

	err := mp.pendingMiniBlocks.RevertHeader(header)
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
	return &block.MetaBlockBody{}, nil
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
				shardData.ShardId,
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
	// metaBlockPool.Remove(headerHash) should not be called, just to have a restore point if it is needed later

	body, ok := bodyHandler.(*block.MetaBlockBody)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	mp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for i := 0; i < len(header.ShardInfo); i++ {
		shardHeaderHash := header.ShardInfo[i].HeaderHash
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardBlock, ok := headerInfo.hdr.(*block.Header)
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

	err = mp.commitAll()
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

	log.Info(fmt.Sprintf("pools info: MetaBlocks = %d from %d, ShardHeaders = %d from %d\n",
		mp.dataPool.MetaBlocks().Len(),
		mp.dataPool.MetaBlocks().MaxSize(),
		mp.dataPool.ShardHeaders().Len(),
		mp.dataPool.ShardHeaders().MaxSize(),
	))

	return nil
}

// RevertStateToBlock recreates thee state tries to the root hashes indicated by the provided header
func (mp *metaProcessor) RevertStateToBlock(header data.HeaderHandler) error {
	err := mp.accounts.RecreateTrie(header.GetRootHash())
	if err != nil {
		return err
	}

	err = mp.validatorStatisticsProcessor.RevertPeerState(header)
	if err != nil {
		return err
	}

	return nil
}

// RevertAccountState reverts the account state for cleanup failed process
func (mp *metaProcessor) RevertAccountState() {
	err := mp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Error(err.Error())
	}

	err = mp.validatorStatisticsProcessor.RevertPeerStateToSnapshot(0)
	if err != nil {
		log.Error(err.Error())
	}
}

func (mp *metaProcessor) getPrevHeader(header *block.MetaBlock) (*block.MetaBlock, error) {
	metaBlockStore := mp.store.GetStorer(dataRetriever.MetaBlockUnit)
	buff, err := metaBlockStore.Get(header.GetPrevHash())
	if err != nil {
		return nil, err
	}

	prevMetaHeader := &block.MetaBlock{}
	err = mp.marshalizer.Unmarshal(prevMetaHeader, buff)
	if err != nil {
		return nil, err
	}

	return prevMetaHeader, nil
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

func (mp *metaProcessor) commitAll() error {
	_, err := mp.accounts.Commit()
	if err != nil {
		return err
	}

	_, err = mp.validatorStatisticsProcessor.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (mp *metaProcessor) saveMetricCrossCheckBlockHeight() {
	crossCheckBlockHeight := ""
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		heightValue := uint64(0)

		valueStoredI, isValueInMap := mp.shardsHeadersNonce.Load(i)
		if isValueInMap {
			valueStored, ok := valueStoredI.(uint64)
			if ok {
				heightValue = valueStored
			}
		}

		crossCheckBlockHeight += fmt.Sprintf("%d: %d, ", i, heightValue)
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
		headerInfo, ok := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardHeaderHash)]
		if !ok {
			mp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()
			return process.ErrMissingHeader
		}

		shardHdr, ok := headerInfo.hdr.(*block.Header)
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
func (mp *metaProcessor) checkShardHeadersValidity(metaHdr *block.MetaBlock) (map[uint32]data.HeaderHandler, error) {
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

	mp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	defer mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()

	mapShardMiniBlockHeadersData := make(map[string]struct{})
	mapMiniBlockHeadersInMetaBlock := make(map[string]struct{})

	for _, shardData := range metaHdr.ShardInfo {
		actualHdr := mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)].hdr
		shardHdr, ok := actualHdr.(*block.Header)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		if len(shardData.ShardMiniBlockHeaders) != len(shardHdr.MiniBlockHeaders) {
			return nil, process.ErrHeaderShardDataMismatch
		}

		for _, shardMiniBlockHdr := range shardData.ShardMiniBlockHeaders {
			mapMiniBlockHeadersInMetaBlock[string(shardMiniBlockHdr.Hash)] = struct{}{}
		}

		for _, actualMiniBlockHdr := range shardHdr.MiniBlockHeaders {
			mapShardMiniBlockHeadersData[string(actualMiniBlockHdr.Hash)] = struct{}{}
		}
	}

	for mbHdrHash, _ := range mapMiniBlockHeadersInMetaBlock {
		if _, ok := mapShardMiniBlockHeadersData[mbHdrHash]; !ok {
			return nil, process.ErrHeaderShardDataMismatch
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
			mp.hdrsForCurrBlock.missingFinalityAttestingHdrs = mp.requestMissingFinalityAttestingShardHeaders()
			if mp.hdrsForCurrBlock.missingFinalityAttestingHdrs == 0 {
				log.Info(fmt.Sprintf("received all missing finality attesting shard headers\n"))
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

	if mp.isHeaderOutOfRange(shardHeader, shardHeaderPool) {
		shardHeaderPool.Remove(shardHeaderHash)
		return
	}

	// request miniblocks for which metachain is destination
	for _, mb := range shardHeader.MiniBlockHeaders {
		if mb.ReceiverShardID == mp.shardCoordinator.SelfId() {
			//TODO continue implementation: go mp.onRequestMiniBlock(mb.Hash)
		}
	}
}

// requestMissingFinalityAttestingShardHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the shardBlockFinality headers greater than the highest shard header,
// for each shard, related to the block which should be processed
func (mp *metaProcessor) requestMissingFinalityAttestingShardHeaders() uint32 {
	missingFinalityAttestingShardHeaders := uint32(0)

	for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
		missingFinalityAttestingHeaders := mp.requestMissingFinalityAttestingHeaders(
			shardId,
			mp.shardBlockFinality,
			mp.getShardHeaderFromPoolWithNonce)

		missingFinalityAttestingShardHeaders += missingFinalityAttestingHeaders
	}

	return missingFinalityAttestingShardHeaders
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
		mp.hdrsForCurrBlock.missingFinalityAttestingHdrs = mp.requestMissingFinalityAttestingShardHeaders()
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
			missingHeadersHashes[shardData.ShardId] = append(missingHeadersHashes[shardData.ShardId], shardData.HeaderHash)
			mp.hdrsForCurrBlock.missingHdrs++
			continue
		}

		mp.hdrsForCurrBlock.hdrHashAndInfo[string(shardData.HeaderHash)] = &hdrInfo{hdr: hdr, usedInBlock: true}

		if hdr.Nonce > mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardId] {
			mp.hdrsForCurrBlock.highestHdrNonce[shardData.ShardId] = hdr.Nonce
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
	maxItemsInBlock uint32,
	round uint64,
	haveTime func() bool,
) ([]block.ShardData, error) {

	shardInfo := make([]block.ShardData, 0)
	lastPushedHdr := make(map[uint32]data.HeaderHandler, mp.shardCoordinator.NumberOfShards())

	if mp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after entered in createShardInfo method\n"))
		return shardInfo, nil
	}

	mbHdrs := uint32(0)

	timeBefore := time.Now()
	orderedHdrs, orderedHdrHashes, sortedHdrPerShard, err := mp.getOrderedHdrs(round)
	timeAfter := time.Now()

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after ordered %d hdrs in %v sec\n", len(orderedHdrs), timeAfter.Sub(timeBefore).Seconds()))
		return shardInfo, nil
	}

	log.Debug(fmt.Sprintf("time elapsed to ordered %d hdrs: %v sec\n", len(orderedHdrs), timeAfter.Sub(timeBefore).Seconds()))

	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("creating shard info has been started: have %d hdrs in pool\n", len(orderedHdrs)))

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
	for index := range orderedHdrs {
		shId := orderedHdrs[index].ShardId

		lastHdr, ok := lastPushedHdr[shId].(*block.Header)
		if !ok {
			continue
		}

		isFinal, _ := mp.isShardHeaderValidFinal(orderedHdrs[index], lastHdr, sortedHdrPerShard[shId])
		if !isFinal {
			continue
		}

		lastPushedHdr[shId] = orderedHdrs[index]

		shardData := block.ShardData{}
		shardData.ShardMiniBlockHeaders = make([]block.ShardMiniBlockHeader, 0)
		shardData.TxCount = orderedHdrs[index].TxCount
		shardData.ShardId = orderedHdrs[index].ShardId
		shardData.HeaderHash = orderedHdrHashes[index]

		snapshot := mp.accounts.JournalLen()

		for i := 0; i < len(orderedHdrs[index].MiniBlockHeaders); i++ {
			if !haveTime() {
				break
			}

			shardMiniBlockHeader := block.ShardMiniBlockHeader{}
			shardMiniBlockHeader.SenderShardId = orderedHdrs[index].MiniBlockHeaders[i].SenderShardID
			shardMiniBlockHeader.ReceiverShardId = orderedHdrs[index].MiniBlockHeaders[i].ReceiverShardID
			shardMiniBlockHeader.Hash = orderedHdrs[index].MiniBlockHeaders[i].Hash
			shardMiniBlockHeader.TxCount = orderedHdrs[index].MiniBlockHeaders[i].TxCount

			// execute shard miniblock to change the trie root hash
			err := mp.checkAndProcessShardMiniBlockHeader(
				orderedHdrHashes[index],
				&shardMiniBlockHeader,
				round,
				shardData.ShardId,
			)

			if err != nil {
				log.Error(err.Error())
				err = mp.accounts.RevertToSnapshot(snapshot)
				if err != nil {
					log.Error(err.Error())
				}
				break
			}

			shardData.ShardMiniBlockHeaders = append(shardData.ShardMiniBlockHeaders, shardMiniBlockHeader)
			mbHdrs++

			recordsAddedInHeader := mbHdrs + uint32(len(shardInfo))
			spaceRemained := int32(maxItemsInBlock) - int32(recordsAddedInHeader) - 1

			if spaceRemained <= 0 {
				log.Info(fmt.Sprintf("max hdrs accepted in one block is reached: added %d hdrs from %d hdrs\n", mbHdrs, len(orderedHdrs)))

				if len(shardData.ShardMiniBlockHeaders) == len(orderedHdrs[index].MiniBlockHeaders) {
					shardInfo = append(shardInfo, shardData)
					mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrHashes[index])] = &hdrInfo{hdr: orderedHdrs[index], usedInBlock: true}
				}

				log.Info(fmt.Sprintf("creating shard info has been finished: created %d shard data\n", len(shardInfo)))
				mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
				return shardInfo, nil
			}
		}

		if !haveTime() {
			log.Info(fmt.Sprintf("time is up: added %d hdrs from %d hdrs\n", mbHdrs, len(orderedHdrs)))

			if len(shardData.ShardMiniBlockHeaders) == len(orderedHdrs[index].MiniBlockHeaders) {
				shardInfo = append(shardInfo, shardData)
				mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrHashes[index])] = &hdrInfo{hdr: orderedHdrs[index], usedInBlock: true}
			}

			log.Info(fmt.Sprintf("creating shard info has been finished: created %d shard data\n", len(shardInfo)))
			mp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
			return shardInfo, nil
		}

		if len(shardData.ShardMiniBlockHeaders) == len(orderedHdrs[index].MiniBlockHeaders) {
			shardInfo = append(shardInfo, shardData)
			mp.hdrsForCurrBlock.hdrHashAndInfo[string(orderedHdrHashes[index])] = &hdrInfo{hdr: orderedHdrs[index], usedInBlock: true}
		}
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
	// TODO: add PrevRandSeed and RandSeed when BLS signing is completed
	header := &block.MetaBlock{
		ShardInfo:    make([]block.ShardData, 0),
		PeerInfo:     make([]block.PeerData, 0),
		PrevRandSeed: make([]byte, 0),
		RandSeed:     make([]byte, 0),
	}

	defer func() {
		go mp.checkAndRequestIfShardHeadersMissing(round)
	}()

	shardInfo, err := mp.createShardInfo(mp.blockSizeThrottler.MaxItemsToAdd(), round, haveTime)
	if err != nil {
		return nil, err
	}

	peerInfo, err := mp.createPeerInfo()
	if err != nil {
		return nil, err
	}

	header.ShardInfo = shardInfo
	header.PeerInfo = peerInfo
	header.RootHash = mp.getRootHash()
	header.TxCount = getTxCount(shardInfo)

	err = mp.pendingMiniBlocks.AddProcessedHeader(header)
	if err != nil {
		return nil, err
	}

	if mp.endOfEpochTrigger.IsEndOfEpoch() {
		//pendingMiniBlocks := mp.pendingMiniBlocks.PendingMiniBlockHeaders()


	}

	mp.blockSizeThrottler.Add(
		round,
		core.MaxUint32(header.ItemsInBody(), header.ItemsInHeader()))

	return header, nil
}

func (mp *metaProcessor) ApplyValidatorStatistics(header data.HeaderHandler) error {
	metaHdr, ok := header.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	rootHash, err := mp.validatorStatisticsProcessor.UpdatePeerState(header)
	if err != nil {
		return err
	}

	metaHdr.ValidatorStatsRootHash = rootHash

	return nil
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

	mrsData := make(map[uint32][]byte)
	mrsTxs := make(map[string][][]byte)

	// send headers which can validate the current header

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

	var body block.MetaBlockBody

	err := mp.marshalizer.Unmarshal(&body, dta)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	return &body
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

func (mp *metaProcessor) getShardHeaderFromPoolWithNonce(
	nonce uint64,
	shardId uint32,
) (data.HeaderHandler, []byte, error) {

	shardHeader, shardHeaderHash, err := process.GetShardHeaderFromPoolWithNonce(
		nonce,
		shardId,
		mp.dataPool.ShardHeaders(),
		mp.dataPool.HeadersNonces())

	return shardHeader, shardHeaderHash, err
}
