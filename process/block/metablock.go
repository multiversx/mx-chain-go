package block

import (
	"encoding/base64"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/serviceContainer"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var shardMBHeaderCounterMutex = sync.RWMutex{}
var shardMBHeadersCurrentBlockProcessed = 0
var shardMBHeadersTotalProcessed = 0

const maxHeadersInBlock = 256

// metaProcessor implements metaProcessor interface and actually it tries to execute block
type metaProcessor struct {
	*baseProcessor
	core     serviceContainer.Core
	dataPool dataRetriever.MetaPoolsHolder

	currHighestShardHdrNonces     map[uint32]uint64
	requestedShardHeaderHashes    map[string]bool
	allNeededShardHdrsFound       bool
	mutRequestedShardHeaderHashes sync.RWMutex

	nextKValidity uint32

	chRcvAllHdrs chan bool
}

// NewMetaProcessor creates a new metaProcessor object
func NewMetaProcessor(
	core serviceContainer.Core,
	accounts state.AccountsAdapter,
	dataPool dataRetriever.MetaPoolsHolder,
	forkDetector process.ForkDetector,
	shardCoordinator sharding.Coordinator,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
	startHeaders map[uint32]data.HeaderHandler,
	requestHandler process.RequestHandler,
) (*metaProcessor, error) {

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
	if dataPool.ShardHeaders() == nil {
		return nil, process.ErrNilHeadersDataPool
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
		onRequestHeaderHandler:        requestHandler.RequestHeader,
		onRequestHeaderHandlerByNonce: requestHandler.RequestHeaderByNonce,
	}

	err = base.setLastNotarizedHeadersSlice(startHeaders, true)
	if err != nil {
		return nil, err
	}

	mp := metaProcessor{
		core:          core,
		baseProcessor: base,
		dataPool:      dataPool,
	}

	mp.requestedShardHeaderHashes = make(map[string]bool)

	headerPool := mp.dataPool.ShardHeaders()
	headerPool.RegisterHandler(mp.receivedHeader)

	mp.chRcvAllHdrs = make(chan bool)

	mp.nextKValidity = process.ShardBlockFinality

	//TODO: This should be injected when BlockProcessor will be refactored
	mp.uint64Converter = uint64ByteSlice.NewBigEndianConverter()

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
		return err
	}

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	requestedBlockHeaders := mp.requestBlockHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	if requestedBlockHeaders > 0 {
		log.Info(fmt.Sprintf("requested %d missing block headers\n", requestedBlockHeaders))
		err = mp.waitForBlockHeaders(haveTime())
		log.Info(fmt.Sprintf("received %d missing block headers\n", requestedBlockHeaders-len(mp.requestedShardHeaderHashes)))
		if err != nil {
			return err
		}
	}

	if mp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	highestNonceHdrs, err := mp.checkShardHeadersValidity(header)
	if err != nil {
		return err
	}

	err = mp.checkShardHeadersFinality(header, highestNonceHdrs)
	if err != nil {
		return err
	}

	if haveTime() < 0 {
		return process.ErrTimeIsOut
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
		err = process.ErrRootStateMissmatch
		return err
	}

	go mp.checkAndRequestIfShardHeadersMissing(header.Round)

	return nil
}

func (mp *metaProcessor) checkAndRequestIfShardHeadersMissing(round uint32) error {
	_, _, sortedHdrPerShard, err := mp.getOrderedHdrs(round)
	if err != nil {
		return err
	}

	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		// map from *block.Header to dataHandler
		sortedHdrs := make([]data.HeaderHandler, 0)
		for j := 0; j < len(sortedHdrPerShard[i]); j++ {
			sortedHdrs = append(sortedHdrs, sortedHdrPerShard[i][j])
		}

		err := mp.requestHeadersIfMissing(sortedHdrs, i, round)
		if err != nil {
			log.Debug(err.Error())
			continue
		}
	}

	return nil
}

func (mp *metaProcessor) indexBlock(metaBlock *block.MetaBlock, headerPool map[string]*block.Header) {
	if mp.core == nil || mp.core.Indexer() == nil {
		return
	}

	// Update tps benchmarks in the DB
	tpsBenchmark := mp.core.TPSBenchmark()
	if tpsBenchmark != nil {
		go mp.core.Indexer().UpdateTPS(tpsBenchmark)
	}

	//TODO: maybe index metablocks also?
}

// removeBlockInfoFromPool removes the block info from associated pools
func (mp *metaProcessor) removeBlockInfoFromPool(header *block.MetaBlock) error {
	if header == nil {
		return process.ErrNilMetaBlockHeader
	}

	headerPool := mp.dataPool.ShardHeaders()
	if headerPool == nil {
		return process.ErrNilHeadersDataPool
	}

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		headerPool.Remove(shardData.HeaderHash)
	}

	return nil
}

// RestoreBlockIntoPools restores the block into associated pools
func (mp *metaProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if headerHandler == nil {
		return process.ErrNilMetaBlockHeader
	}

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	headerPool := mp.dataPool.ShardHeaders()
	if headerPool == nil {
		return process.ErrNilHeadersDataPool
	}

	hdrHashes := make([][]byte, 0)
	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		hdrHashes = append(hdrHashes, shardData.HeaderHash)
	}

	hdrsBuff, err := mp.store.GetAll(dataRetriever.BlockHeaderUnit, hdrHashes)
	if err != nil {
		return err
	}

	for hdrHash, hdrBuff := range hdrsBuff {
		hdr := block.Header{}
		err = mp.marshalizer.Unmarshal(&hdr, hdrBuff)
		if err != nil {
			return err
		}

		headerPool.Put([]byte(hdrHash), &hdr)

		shardMBHeaderCounterMutex.Lock()
		shardMBHeadersTotalProcessed -= len(hdr.MiniBlockHeaders)
		shardMBHeaderCounterMutex.Unlock()
	}

	mp.restoreLastNotarized()

	return nil
}

// CreateBlockBody creates block body of metachain
func (mp *metaProcessor) CreateBlockBody(round uint32, haveTime func() bool) (data.BodyHandler, error) {
	return &block.MetaBlockBody{}, nil
}

func (mp *metaProcessor) processBlockHeaders(header *block.MetaBlock, round uint32, haveTime func() time.Duration) error {
	hdrPool := mp.dataPool.ShardHeaders()

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
				hdrPool,
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
		log.Info(fmt.Sprintf("the following miniblocks hashes were successfully processed:%s\n", msg))
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

	tempHeaderPool := make(map[string]*block.Header)

	err = checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	buff, err := mp.marshalizer.Marshal(headerHandler)
	if err != nil {
		return err
	}

	headerHash := mp.hasher.Compute(string(buff))
	errNotCritical := mp.store.Put(dataRetriever.MetaBlockUnit, headerHash, buff)
	if errNotCritical != nil {
		log.Error(errNotCritical.Error())
	}

	nonceToByteSlice := mp.uint64Converter.ToByteSlice(headerHandler.GetNonce())
	err = mp.store.Put(dataRetriever.ShardHdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	body, ok := bodyHandler.(*block.MetaBlockBody)
	if !ok {
		err = process.ErrWrongTypeAssertion
		return err
	}

	headerNoncePool := mp.dataPool.MetaBlockNonces()
	if headerNoncePool == nil {
		err = process.ErrNilDataPoolHolder
		return err
	}

	_ = headerNoncePool.Put(headerHandler.GetNonce(), headerHash)

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		header, err := process.GetShardHeaderFromPool(shardData.HeaderHash, mp.dataPool.ShardHeaders())
		if header == nil {
			return err
		}

		tempHeaderPool[string(shardData.HeaderHash)] = header

		buff, err = mp.marshalizer.Marshal(header)
		if err != nil {
			return err
		}

		errNotCritical = mp.store.Put(dataRetriever.BlockHeaderUnit, shardData.HeaderHash, buff)
		if errNotCritical != nil {
			log.Error(errNotCritical.Error())
		}
	}

	_, err = mp.accounts.Commit()
	if err != nil {
		return err
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

	err = mp.createLastNotarizedHdrs(header)
	if err != nil {
		return err
	}

	errNotCritical = mp.removeBlockInfoFromPool(header)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	errNotCritical = mp.forkDetector.AddHeader(header, headerHash, process.BHProcessed)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	if mp.core != nil && mp.core.TPSBenchmark() != nil {
		mp.core.TPSBenchmark().Update(header)
	}

	mp.indexBlock(header, tempHeaderPool)

	go mp.displayMetaBlock(header)

	return nil
}

func (mp *metaProcessor) createLastNotarizedHdrs(header *block.MetaBlock) error {
	mp.mutNotarizedHdrs.Lock()
	defer mp.mutNotarizedHdrs.Unlock()

	if mp.lastNotarizedHdrs == nil || mp.finalNotarizedHdrs == nil {
		return process.ErrNotarizedHdrsSliceIsNil
	}

	// save the last headers with the highest round per shard
	tmpNotedHdrs := make(mapShardLastHeaders, mp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		tmpNotedHdrs[i] = mp.lastNotarizedHdrs[i]
	}

	var err error
	defer func() {
		if err != nil {
			for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
				mp.lastNotarizedHdrs[i] = tmpNotedHdrs[i]
			}
		}
	}()

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		header, err := process.GetShardHeaderFromPool(shardData.HeaderHash, mp.dataPool.ShardHeaders())
		if header == nil {
			return err
		}

		if mp.lastNotarizedHdrs[header.ShardId].GetNonce() < header.Nonce {
			mp.lastNotarizedHdrs[header.ShardId] = header
		}
	}

	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		mp.finalNotarizedHdrs[i] = tmpNotedHdrs[i]
	}

	return nil
}

// gets all the headers from the metablock in sorted order per shard
func (mp *metaProcessor) getSortedShardHdrsFromMetablock(header *block.MetaBlock) (map[uint32][]*block.Header, error) {
	sortedShardHdrs := make(map[uint32][]*block.Header, mp.shardCoordinator.NumberOfShards())

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		header, err := process.GetShardHeaderFromPool(shardData.HeaderHash, mp.dataPool.ShardHeaders())
		if header == nil {
			go mp.onRequestHeaderHandler(shardData.ShardId, shardData.HeaderHash)
			return nil, err
		}

		sortedShardHdrs[header.ShardId] = append(sortedShardHdrs[header.ShardId], header)
	}

	for shId := uint32(0); shId < mp.shardCoordinator.NumberOfShards(); shId++ {
		hdrsForShard := sortedShardHdrs[shId]
		if len(hdrsForShard) <= 1 {
			continue
		}

		sort.Slice(hdrsForShard, func(i, j int) bool {
			return hdrsForShard[i].GetNonce() < hdrsForShard[j].GetNonce()
		})
	}

	return sortedShardHdrs, nil
}

// check if shard headers were signed and constructed correctly and returns headers which has to be
// checked for finality
func (mp *metaProcessor) checkShardHeadersValidity(header *block.MetaBlock) (mapShardLastHeaders, error) {
	mp.mutNotarizedHdrs.RLock()
	if mp.lastNotarizedHdrs == nil {
		mp.mutNotarizedHdrs.RUnlock()
		return nil, process.ErrNotarizedHdrsSliceIsNil
	}

	tmpNotedHdrs := make(mapShardLastHeaders, mp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		tmpNotedHdrs[i] = mp.lastNotarizedHdrs[i]
	}
	mp.mutNotarizedHdrs.RUnlock()

	sortedShardHdrs, err := mp.getSortedShardHdrsFromMetablock(header)
	if err != nil {
		return nil, err
	}

	highestNonceHdrs := make(mapShardLastHeaders)
	for shId := uint32(0); shId < mp.shardCoordinator.NumberOfShards(); shId++ {
		hdrsForShard := sortedShardHdrs[shId]
		if len(hdrsForShard) == 0 {
			continue
		}

		for i := 0; i < len(hdrsForShard); i++ {
			err := mp.isHdrConstructionValid(hdrsForShard[i], tmpNotedHdrs[shId])
			if err != nil {
				return nil, err
			}
			tmpNotedHdrs[shId] = hdrsForShard[i]
			highestNonceHdrs[shId] = hdrsForShard[i]
		}
	}

	return highestNonceHdrs, nil
}

// check if shard headers are final by checking if newer headers were constructed upon them
func (mp *metaProcessor) checkShardHeadersFinality(header *block.MetaBlock, highestNonceHdrs mapShardLastHeaders) error {
	if header == nil {
		return process.ErrNilBlockHeader
	}

	//TODO: change this to look at the pool where values are saved by prevHash. can be done after resolver is done
	_, _, sortedHdrPerShard, err := mp.getOrderedHdrs(header.GetRound())
	if err != nil {
		return err
	}

	for index, lastVerifiedHdr := range highestNonceHdrs {
		if index != lastVerifiedHdr.GetShardID() {
			return process.ErrShardIdMissmatch
		}

		// verify if there are "K" block after current to make this one final
		nextBlocksVerified := uint32(0)
		shId := lastVerifiedHdr.GetShardID()
		for i := 0; i < len(sortedHdrPerShard[shId]); i++ {
			if nextBlocksVerified >= mp.nextKValidity {
				break
			}

			// found a header with the next nonce
			tmpHdr := sortedHdrPerShard[shId][i]
			if tmpHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
				err := mp.isHdrConstructionValid(tmpHdr, lastVerifiedHdr)
				if err != nil {
					continue
				}

				lastVerifiedHdr = tmpHdr
				nextBlocksVerified += 1
			}
		}

		if nextBlocksVerified < mp.nextKValidity {
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
		if nextBlocksVerified >= mp.nextKValidity {
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

	if nextBlocksVerified >= mp.nextKValidity {
		return true, hdrIds
	}

	return false, nil
}

// receivedHeader is a call back function which is called when a new header
// is added in the headers pool
func (mp *metaProcessor) receivedHeader(headerHash []byte) {
	shardHeaderCache := mp.dataPool.ShardHeaders()
	if shardHeaderCache == nil {
		return
	}

	shardHdrsNoncesCache := mp.dataPool.ShardHeadersNonces()
	if shardHdrsNoncesCache == nil && mp.nextKValidity > 0 {
		return
	}

	value, ok := shardHeaderCache.Peek(headerHash)
	if !ok {
		return
	}

	hdr, ok := value.(data.HeaderHandler)
	if !ok {
		return
	}

	mp.mutRequestedShardHeaderHashes.Lock()

	if !mp.allNeededShardHdrsFound {
		if mp.requestedShardHeaderHashes[string(headerHash)] {
			delete(mp.requestedShardHeaderHashes, string(headerHash))

			if hdr.GetNonce() > mp.currHighestShardHdrNonces[hdr.GetShardID()] {
				mp.currHighestShardHdrNonces[hdr.GetShardID()] = hdr.GetNonce()
			}
		}

		lenReqHeadersHashes := len(mp.requestedShardHeaderHashes)
		areFinalityAttestingHdrsInCache := true
		if lenReqHeadersHashes == 0 {
			for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
				// ask for finality attesting hdrs if it is not yet in cache
				for i := mp.currHighestShardHdrNonces[shardId] + 1; i <= mp.currHighestShardHdrNonces[shardId]+uint64(mp.nextKValidity); i++ {
					if mp.currHighestShardHdrNonces[shardId] == uint64(0) {
						continue
					}

					value, okPeek := shardHdrsNoncesCache.Peek(i)
					mapOfHashes, okTypeAssertion := value.(sync.Map)
					_, okLoad := mapOfHashes.Load(shardId)

					finalityAttestingHeaderForShardNotFound := !okPeek || !okTypeAssertion || !okLoad
					if finalityAttestingHeaderForShardNotFound {
						go mp.onRequestHeaderHandlerByNonce(shardId, i)
						areFinalityAttestingHdrsInCache = false
					}
				}
			}
		}

		mp.allNeededShardHdrsFound = lenReqHeadersHashes == 0 && areFinalityAttestingHdrsInCache

		if mp.allNeededShardHdrsFound {
			mp.mutRequestedShardHeaderHashes.Unlock()
			mp.chRcvAllHdrs <- true
			return
		}
	}

	mp.mutRequestedShardHeaderHashes.Unlock()
}

func (mp *metaProcessor) requestBlockHeaders(header *block.MetaBlock) int {
	mp.mutRequestedShardHeaderHashes.Lock()

	mp.currHighestShardHdrNonces = make(map[uint32]uint64, mp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < mp.shardCoordinator.NumberOfShards(); i++ {
		mp.currHighestShardHdrNonces[i] = uint64(0)
	}

	missingHeaderHashes := mp.computeMissingHeaders(header)
	mp.requestedShardHeaderHashes = make(map[string]bool)

	for shardId, headerHash := range missingHeaderHashes {
		mp.requestedShardHeaderHashes[string(headerHash)] = true
		go mp.onRequestHeaderHandler(shardId, headerHash)
	}

	if len(missingHeaderHashes) > 0 {
		mp.allNeededShardHdrsFound = false
	}

	mp.mutRequestedShardHeaderHashes.Unlock()

	return len(missingHeaderHashes)
}

func (mp *metaProcessor) computeMissingHeaders(header *block.MetaBlock) map[uint32][]byte {
	missingHeaders := make(map[uint32][]byte)

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		header, _ := process.GetShardHeaderFromPool(shardData.HeaderHash, mp.dataPool.ShardHeaders())
		if header == nil {
			missingHeaders[shardData.ShardId] = shardData.HeaderHash
			continue
		}

		if header.GetNonce() > mp.currHighestShardHdrNonces[shardData.ShardId] {
			mp.currHighestShardHdrNonces[shardData.ShardId] = header.GetNonce()
		}
	}

	return missingHeaders
}

func (mp *metaProcessor) checkAndProcessShardMiniBlockHeader(
	headerHash []byte,
	shardMiniBlockHeader *block.ShardMiniBlockHeader,
	hdrPool storage.Cacher,
	round uint32,
	shardId uint32,
) error {

	if hdrPool == nil {
		return process.ErrNilHeadersDataPool
	}
	// TODO: real processing has to be done here, using metachain state
	return nil
}

func (mp *metaProcessor) createShardInfo(
	maxMiniBlockHdrsInBlock uint32,
	round uint32,
	haveTime func() bool,
) ([]block.ShardData, error) {

	shardInfo := make([]block.ShardData, 0)
	lastPushedHdr := make(mapShardLastHeaders, mp.shardCoordinator.NumberOfShards())

	if mp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after entered in createShardInfo method\n"))
		return shardInfo, nil
	}

	hdrPool := mp.dataPool.ShardHeaders()
	if hdrPool == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	mbHdrs := uint32(0)

	timeBefore := time.Now()
	orderedHdrs, orderedHdrHashes, sortedHdrPerShard, err := mp.getOrderedHdrs(round)
	timeAfter := time.Now()

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after ordered %d hdrs in %v sec\n", len(orderedHdrs), timeAfter.Sub(timeBefore).Seconds()))
		return shardInfo, nil
	}

	log.Info(fmt.Sprintf("time elapsed to ordered %d hdrs: %v sec\n", len(orderedHdrs), timeAfter.Sub(timeBefore).Seconds()))

	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("creating shard info has been started: have %d hdrs in pool\n", len(orderedHdrs)))

	// save last committed hdr for verification
	mp.mutNotarizedHdrs.RLock()
	if mp.lastNotarizedHdrs == nil {
		mp.mutNotarizedHdrs.RUnlock()
		return nil, process.ErrNotarizedHdrsSliceIsNil
	}
	for shardId := uint32(0); shardId < mp.shardCoordinator.NumberOfShards(); shardId++ {
		lastPushedHdr[shardId] = mp.lastNotarizedHdrs[shardId]
	}
	mp.mutNotarizedHdrs.RUnlock()

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
				hdrPool,
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

			if mbHdrs >= maxMiniBlockHdrsInBlock { // max mini block headers count in one block was reached
				log.Info(fmt.Sprintf("max hdrs accepted in one block is reached: added %d hdrs from %d hdrs\n", mbHdrs, len(orderedHdrs)))

				if len(shardData.ShardMiniBlockHeaders) == len(orderedHdrs[index].MiniBlockHeaders) {
					shardInfo = append(shardInfo, shardData)
				}

				log.Info(fmt.Sprintf("creating shard info has been finished: created %d shard data\n", len(shardInfo)))
				return shardInfo, nil
			}
		}

		if !haveTime() {
			log.Info(fmt.Sprintf("time is up: added %d hdrs from %d hdrs\n", mbHdrs, len(orderedHdrs)))

			if len(shardData.ShardMiniBlockHeaders) == len(orderedHdrs[index].MiniBlockHeaders) {
				shardInfo = append(shardInfo, shardData)
			}

			log.Info(fmt.Sprintf("creating shard info has been finished: created %d shard data\n", len(shardInfo)))
			return shardInfo, nil
		}

		if len(shardData.ShardMiniBlockHeaders) == len(orderedHdrs[index].MiniBlockHeaders) {
			shardInfo = append(shardInfo, shardData)
		}
	}

	log.Info(fmt.Sprintf("creating shard info has been finished: created %d shard data\n", len(shardInfo)))
	return shardInfo, nil
}

func (mp *metaProcessor) createPeerInfo() ([]block.PeerData, error) {
	// TODO: to be implemented
	peerInfo := make([]block.PeerData, 0)
	return peerInfo, nil
}

// CreateBlockHeader creates a miniblock header list given a block body
func (mp *metaProcessor) CreateBlockHeader(bodyHandler data.BodyHandler, round uint32, haveTime func() bool) (data.HeaderHandler, error) {
	// TODO: add PrevRandSeed and RandSeed when BLS signing is completed
	header := &block.MetaBlock{
		ShardInfo:    make([]block.ShardData, 0),
		PeerInfo:     make([]block.PeerData, 0),
		PrevRandSeed: make([]byte, 0),
		RandSeed:     make([]byte, 0),
	}

	shardInfo, err := mp.createShardInfo(maxHeadersInBlock, round, haveTime)
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

	go mp.checkAndRequestIfShardHeadersMissing(header.Round)

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

func (mp *metaProcessor) displayMetaBlock(header *block.MetaBlock) {
	if header == nil {
		return
	}

	headerHash, err := mp.computeHeaderHash(header)
	if err != nil {
		log.Error(err.Error())
		return
	}

	mp.displayLogInfo(header, headerHash)
}

func (mp *metaProcessor) displayLogInfo(
	header *block.MetaBlock,
	headerHash []byte,
) {
	dispHeader, dispLines := createDisplayableMetaHeader(header)

	tblString, err := display.CreateTableString(dispHeader, dispLines)
	if err != nil {
		log.Error(err.Error())
		return
	}

	shardMBHeaderCounterMutex.RLock()
	tblString = tblString + fmt.Sprintf("\nHeader hash: %s\n\nTotal shard MB headers "+
		"processed until now: %d. Total shard MB headers processed for this block: %d. Total shard headers remained in pool: %d\n",
		core.ToB64(headerHash),
		shardMBHeadersTotalProcessed,
		shardMBHeadersCurrentBlockProcessed,
		mp.getHeadersCountInPool())
	shardMBHeaderCounterMutex.RUnlock()

	log.Info(tblString)
}

func createDisplayableMetaHeader(
	header *block.MetaBlock,
) ([]string, []*display.LineData) {

	tableHeader := []string{"Part", "Parameter", "Value"}

	lines := displayHeader(header)

	metaLines := make([]*display.LineData, 0)
	metaLines = append(metaLines, display.NewLineData(false, []string{
		"Header",
		"Block type",
		"MetaBlock"}))
	metaLines = append(metaLines, lines...)

	metaLines = displayShardInfo(metaLines, header)
	return tableHeader, metaLines
}

func displayShardInfo(lines []*display.LineData, header *block.MetaBlock) []*display.LineData {
	shardMBHeaderCounterMutex.Lock()
	shardMBHeadersCurrentBlockProcessed = 0
	shardMBHeaderCounterMutex.Unlock()

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]

		lines = append(lines, display.NewLineData(false, []string{
			fmt.Sprintf("ShardData_%d", shardData.ShardId),
			"Header hash",
			base64.StdEncoding.EncodeToString(shardData.HeaderHash)}))

		if shardData.ShardMiniBlockHeaders == nil || len(shardData.ShardMiniBlockHeaders) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				"", "ShardMiniBlockHeaders", "<EMPTY>"}))
		}

		shardMBHeaderCounterMutex.Lock()
		shardMBHeadersCurrentBlockProcessed += len(shardData.ShardMiniBlockHeaders)
		shardMBHeadersTotalProcessed += len(shardData.ShardMiniBlockHeaders)
		shardMBHeaderCounterMutex.Unlock()

		for j := 0; j < len(shardData.ShardMiniBlockHeaders); j++ {
			if j == 0 || j >= len(shardData.ShardMiniBlockHeaders)-1 {
				lines = append(lines, display.NewLineData(false, []string{
					"",
					fmt.Sprintf("ShardMiniBlockHeaderHash_%d", j+1),
					core.ToB64(shardData.ShardMiniBlockHeaders[j].Hash)}))
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					"",
					fmt.Sprintf("..."),
					fmt.Sprintf("...")}))
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}

// MarshalizedDataToBroadcast prepares underlying data into a marshalized object according to destination
func (mp *metaProcessor) MarshalizedDataToBroadcast(
	header data.HeaderHandler,
	bodyHandler data.BodyHandler,
) (map[uint32][]byte, map[uint32][][]byte, error) {

	mrsData := make(map[uint32][]byte)
	mrsTxs := make(map[uint32][][]byte)

	// send headers which can validate the current header

	return mrsData, mrsTxs, nil
}

func (mp *metaProcessor) getOrderedHdrs(round uint32) ([]*block.Header, [][]byte, map[uint32][]*block.Header, error) {
	hdrStore := mp.dataPool.ShardHeaders()
	if hdrStore == nil {
		return nil, nil, nil, process.ErrNilCacher
	}

	hashAndBlockMap := make(map[uint32][]*hashAndHdr, mp.shardCoordinator.NumberOfShards())
	headersMap := make(map[uint32][]*block.Header)
	headers := make([]*block.Header, 0)
	hdrHashes := make([][]byte, 0)

	mp.mutNotarizedHdrs.RLock()
	if mp.lastNotarizedHdrs == nil {
		mp.mutNotarizedHdrs.RUnlock()
		return nil, nil, nil, process.ErrNotarizedHdrsSliceIsNil
	}

	// get keys and arrange them into shards
	for _, key := range hdrStore.Keys() {
		val, _ := hdrStore.Peek(key)
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
		if mp.lastNotarizedHdrs[currShardId] == nil {
			continue
		}

		if hdr.GetRound() <= mp.lastNotarizedHdrs[currShardId].GetRound() {
			continue
		}

		if hdr.GetNonce() <= mp.lastNotarizedHdrs[currShardId].GetNonce() {
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

func (mp *metaProcessor) getHeadersCountInPool() int {
	headerPool := mp.dataPool.ShardHeaders()
	if headerPool == nil {
		log.Error(process.ErrNilHeadersDataPool.Error())
		return -1
	}

	return headerPool.Len()
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
