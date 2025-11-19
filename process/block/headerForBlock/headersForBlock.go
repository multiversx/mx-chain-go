package headerForBlock

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

var log = logger.GetOrCreate("headersForBlock")

// ArgHeadersForBlock is the DTO that holds data needed to create a new instance of headersForBlock
type ArgHeadersForBlock struct {
	DataPool                                    dataRetriever.PoolsHolder
	RequestHandler                              process.RequestHandler
	EnableEpochsHandler                         common.EnableEpochsHandler
	ShardCoordinator                            sharding.Coordinator
	BlockTracker                                process.BlockTracker
	TxCoordinator                               process.TransactionCoordinator
	RoundHandler                                consensus.RoundHandler
	ExtraDelayForRequestBlockInfoInMilliseconds int
	GenesisNonce                                uint64
}

type headersForBlock struct {
	dataPool                     dataRetriever.PoolsHolder
	requestHandler               process.RequestHandler
	enableEpochsHandler          common.EnableEpochsHandler
	shardCoordinator             sharding.Coordinator
	blockTracker                 process.BlockTracker
	txCoordinator                process.TransactionCoordinator
	roundHandler                 process.RoundHandler
	extraDelayRequestBlockInfo   time.Duration
	chRcvAllHdrs                 chan bool
	genesisNonce                 uint64
	blockFinality                uint32
	missingHdrs                  uint32
	missingFinalityAttestingHdrs uint32
	missingProofs                uint32
	highestHdrNonce              map[uint32]uint64
	mutHdrsForBlock              sync.RWMutex
	hdrHashAndInfo               map[string]HeaderInfo
	lastNotarizedShardHeaders    map[uint32]LastNotarizedHeaderInfoHandler
}

type crossShardMetaData interface {
	GetNonce() uint64
	GetShardID() uint32
	GetHeaderHash() []byte
}

// NewHeadersForBlock returns a new instance of headersForBlock
func NewHeadersForBlock(args ArgHeadersForBlock) (*headersForBlock, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	instance := &headersForBlock{
		dataPool:                   args.DataPool,
		requestHandler:             args.RequestHandler,
		enableEpochsHandler:        args.EnableEpochsHandler,
		shardCoordinator:           args.ShardCoordinator,
		blockTracker:               args.BlockTracker,
		txCoordinator:              args.TxCoordinator,
		roundHandler:               args.RoundHandler,
		extraDelayRequestBlockInfo: time.Duration(args.ExtraDelayForRequestBlockInfoInMilliseconds) * time.Millisecond,
		genesisNonce:               args.GenesisNonce,
		blockFinality:              process.BlockFinality,
		chRcvAllHdrs:               make(chan bool),
		hdrHashAndInfo:             make(map[string]HeaderInfo),
		highestHdrNonce:            make(map[uint32]uint64),
		lastNotarizedShardHeaders:  make(map[uint32]LastNotarizedHeaderInfoHandler),
	}

	instance.dataPool.Headers().RegisterHandler(instance.receivedMetaBlock)
	instance.dataPool.Headers().RegisterHandler(instance.receivedShardBlock)
	instance.dataPool.Proofs().RegisterHandler(instance.checkReceivedProofIfAttestingIsNeeded)

	return instance, nil
}

func checkArgs(args ArgHeadersForBlock) error {
	if check.IfNil(args.DataPool) {
		return process.ErrNilDataPoolHolder
	}
	if check.IfNil(args.DataPool.Headers()) {
		return process.ErrNilHeadersDataPool
	}
	if check.IfNil(args.DataPool.Proofs()) {
		return process.ErrNilProofsPool
	}
	if check.IfNil(args.RequestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(args.EnableEpochsHandler) {
		return process.ErrNilEnableEpochsHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(args.BlockTracker) {
		return process.ErrNilBlockTracker
	}
	if check.IfNil(args.TxCoordinator) {
		return process.ErrNilTransactionCoordinator
	}
	if check.IfNil(args.RoundHandler) {
		return process.ErrNilRoundHandler
	}

	return nil
}

// AddHeaderUsedInBlock adds the provided header info for the provided hash as used in block
func (hfb *headersForBlock) AddHeaderUsedInBlock(
	hash string,
	header data.HeaderHandler,
) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	hfb.hdrHashAndInfo[hash] = newHeaderInfo(header, true, false, false)
}

// AddHeaderNotUsedInBlock adds the provided header info for the provided hash as not used in block
func (hfb *headersForBlock) AddHeaderNotUsedInBlock(
	hash string,
	header data.HeaderHandler,
) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	hfb.hdrHashAndInfo[hash] = newHeaderInfo(header, false, false, false)
}

// GetHeaderInfo returns the header info for the provided hash
func (hfb *headersForBlock) GetHeaderInfo(hash string) (HeaderInfo, bool) {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	hi, found := hfb.hdrHashAndInfo[hash]
	return hi, found
}

// GetHeadersInfoMap returns the internal header hash map
func (hfb *headersForBlock) GetHeadersInfoMap() map[string]HeaderInfo {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	return hfb.getHeadersInfoMapUnprotected()
}

func (hfb *headersForBlock) getHeadersInfoMapUnprotected() map[string]HeaderInfo {
	mapCopy := make(map[string]HeaderInfo, len(hfb.hdrHashAndInfo))
	for hash, hi := range hfb.hdrHashAndInfo {
		mapCopy[hash] = hi
	}

	return mapCopy
}

// GetHeadersMap returns a map of headers with their hashes
func (hfb *headersForBlock) GetHeadersMap() map[string]data.HeaderHandler {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	headersMap := make(map[string]data.HeaderHandler, len(hfb.hdrHashAndInfo))
	for hash, hi := range hfb.hdrHashAndInfo {
		headersMap[hash] = hi.GetHeader()
	}

	return headersMap
}

// ComputeHeadersForCurrentBlock returns a map of header handlers for current block
func (hfb *headersForBlock) ComputeHeadersForCurrentBlock(usedInBlock bool) (map[uint32][]data.HeaderHandler, error) {
	hdrsForCurrentBlock := make(map[uint32][]data.HeaderHandler)

	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	hdrHashAndInfo, err := hfb.filterHeadersWithoutProofs()
	if err != nil {
		return nil, err
	}

	for hdrHash, hi := range hdrHashAndInfo {
		if hi.UsedInBlock() != usedInBlock {
			continue
		}

		hdr := hi.GetHeader()
		if hfb.hasMissingProof(hdr, hdrHash) {
			return nil, fmt.Errorf("%w for header with hash %s", process.ErrMissingHeaderProof, hex.EncodeToString([]byte(hdrHash)))
		}

		hdrsForCurrentBlock[hdr.GetShardID()] = append(hdrsForCurrentBlock[hdr.GetShardID()], hdr)
	}

	return hdrsForCurrentBlock, nil
}

func (hfb *headersForBlock) filterHeadersWithoutProofs() (map[string]HeaderInfo, error) {
	removedNonces := make(map[uint32]map[uint64]struct{})
	noncesWithProofs := make(map[uint32]map[uint64]struct{})
	shardIDs := common.GetShardIDs(hfb.shardCoordinator.NumberOfShards())
	for shard := range shardIDs {
		removedNonces[shard] = make(map[uint64]struct{})
		noncesWithProofs[shard] = make(map[uint64]struct{})
	}
	filteredHeadersInfo := make(map[string]HeaderInfo)

	hdrHashAndInfo := hfb.getHeadersInfoMapUnprotected()
	for hdrHash, hi := range hdrHashAndInfo {
		hdr := hi.GetHeader()
		if hfb.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, hdr.GetEpoch()) {
			if hfb.hasMissingProof(hdr, hdrHash) {
				removedNonces[hdr.GetShardID()][hdr.GetNonce()] = struct{}{}
				continue
			}

			noncesWithProofs[hdr.GetShardID()][hdr.GetNonce()] = struct{}{}
			filteredHeadersInfo[hdrHash] = hi
			continue
		}

		filteredHeadersInfo[hdrHash] = hi
	}

	for shard, nonces := range removedNonces {
		for nonce := range nonces {
			if _, ok := noncesWithProofs[shard][nonce]; !ok {
				return nil, fmt.Errorf("%w for shard %d and nonce %d", process.ErrMissingHeaderProof, shard, nonce)
			}
		}
	}

	return filteredHeadersInfo, nil
}

// ComputeHeadersForCurrentBlockInfo returns a map of nonce and hash infos for the current block
func (hfb *headersForBlock) ComputeHeadersForCurrentBlockInfo(usedInBlock bool) (map[uint32][]NonceAndHashInfo, error) {
	hdrsForCurrentBlockInfo := make(map[uint32][]NonceAndHashInfo)

	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	hdrHashAndInfo := hfb.getHeadersInfoMapUnprotected()
	for metaBlockHash, hi := range hdrHashAndInfo {
		if hi.UsedInBlock() != usedInBlock {
			continue
		}

		hdr := hi.GetHeader()
		if hfb.hasMissingProof(hdr, metaBlockHash) {
			return nil, fmt.Errorf("%w for header with hash %s", process.ErrMissingHeaderProof, hex.EncodeToString([]byte(metaBlockHash)))
		}

		hdrsForCurrentBlockInfo[hdr.GetShardID()] = append(hdrsForCurrentBlockInfo[hdr.GetShardID()],
			&nonceAndHashInfo{nonce: hdr.GetNonce(), hash: []byte(metaBlockHash)})
	}

	return hdrsForCurrentBlockInfo, nil
}

func (hfb *headersForBlock) hasMissingProof(header data.HeaderHandler, hdrHash string) bool {
	isFlagEnabledForHeader := hfb.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) && header.GetNonce() >= 1
	if !isFlagEnabledForHeader {
		return false
	}

	return !hfb.dataPool.Proofs().HasProof(header.GetShardID(), []byte(hdrHash))
}

// Reset resets the internal state of the component
func (hfb *headersForBlock) Reset() {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	hfb.resetMissingHeaders()
	hfb.initMaps()
}

func (hfb *headersForBlock) initMaps() {
	hfb.hdrHashAndInfo = make(map[string]HeaderInfo)
	hfb.highestHdrNonce = make(map[uint32]uint64)
	hfb.lastNotarizedShardHeaders = make(map[uint32]LastNotarizedHeaderInfoHandler)
}

func (hfb *headersForBlock) resetMissingHeaders() {
	hfb.missingHdrs = 0
	hfb.missingFinalityAttestingHdrs = 0
	hfb.missingProofs = 0
}

func (hfb *headersForBlock) setHasProof(hash string) {
	_, ok := hfb.hdrHashAndInfo[hash]
	if !ok {
		hfb.hdrHashAndInfo[hash] = newEmptyHeaderInfo()
	}

	hfb.hdrHashAndInfo[hash].SetHasProof(true)
}

func (hfb *headersForBlock) setHasProofRequested(hash string) {
	_, ok := hfb.hdrHashAndInfo[hash]
	if !ok {
		hfb.hdrHashAndInfo[hash] = newEmptyHeaderInfo()
	}

	hfb.hdrHashAndInfo[hash].SetHasProofRequested(true)
	hfb.missingProofs++
}

// RequestShardHeaders requests the needed shard headers for the provided header
func (hfb *headersForBlock) RequestShardHeaders(metaBlock data.MetaHeaderHandler) {
	_ = core.EmptyChannel(hfb.chRcvAllHdrs)

	if len(metaBlock.GetShardInfoHandlers()) == 0 && len(metaBlock.GetShardInfoProposalHandlers()) == 0 {
		return
	}

	hfb.computeExistingAndRequestMissingShardHeaders(metaBlock)
}

// RequestMetaHeaders requests the needed meta headers for the provided header
func (hfb *headersForBlock) RequestMetaHeaders(shardHeader data.ShardHeaderHandler) {
	_ = core.EmptyChannel(hfb.chRcvAllHdrs)

	if len(shardHeader.GetMetaBlockHashes()) == 0 {
		return
	}

	hfb.computeExistingAndRequestMissingMetaHeaders(shardHeader)
}

// GetMissingData returns the missing data
func (hfb *headersForBlock) GetMissingData() (uint32, uint32, uint32) {
	hfb.mutHdrsForBlock.RLock()
	defer hfb.mutHdrsForBlock.RUnlock()

	return hfb.missingHdrs, hfb.missingProofs, hfb.missingFinalityAttestingHdrs
}

// WaitForHeadersIfNeeded waits for any missing headers
func (hfb *headersForBlock) WaitForHeadersIfNeeded(haveTime func() time.Duration) error {
	hfb.mutHdrsForBlock.RLock()
	requestedHdrs := hfb.missingHdrs
	requestedFinalityAttestingHdrs := hfb.missingFinalityAttestingHdrs
	requestedProofs := hfb.missingProofs
	hfb.mutHdrsForBlock.RUnlock()
	haveMissingMetaHeaders := requestedHdrs > 0 || requestedFinalityAttestingHdrs > 0 || requestedProofs > 0
	if !haveMissingMetaHeaders {
		return nil
	}

	if requestedHdrs > 0 {
		log.Debug("requested missing headers",
			"num headers", requestedHdrs,
		)
	}
	if requestedFinalityAttestingHdrs > 0 {
		log.Debug("requested missing finality attesting headers",
			"num finality meta headers", requestedFinalityAttestingHdrs,
		)
	}
	if requestedProofs > 0 {
		log.Debug("requested missing header proofs",
			"num proofs", requestedProofs,
		)
	}

	err := hfb.waitForHdrHashes(haveTime())
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()
	missingHdrs := hfb.missingHdrs
	missingProofs := hfb.missingProofs
	if requestedHdrs > 0 {
		log.Debug("received missing headers",
			"num headers", requestedHdrs-missingHdrs,
		)
	}

	if requestedProofs > 0 {
		log.Debug("received missing header proofs",
			"num proofs", requestedProofs-missingProofs,
		)
	}

	hfb.resetMissingHeaders()

	return err
}

func (hfb *headersForBlock) waitForHdrHashes(waitTime time.Duration) error {
	select {
	case <-hfb.chRcvAllHdrs:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

func (hfb *headersForBlock) computeExistingAndRequestMissingMetaHeaders(header data.ShardHeaderHandler) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	metaBlockHashes := header.GetMetaBlockHashes()
	lastMetaBlockNonceWithProof := uint64(0)
	for i := 0; i < len(metaBlockHashes); i++ {
		hdr, err := process.GetMetaHeaderFromPool(
			metaBlockHashes[i],
			hfb.dataPool.Headers())

		if err != nil {
			hfb.missingHdrs++
			hfb.hdrHashAndInfo[string(metaBlockHashes[i])] = newHeaderInfo(nil, true, false, false)

			go hfb.requestHandler.RequestMetaHeader(metaBlockHashes[i])
			continue
		}

		hfb.hdrHashAndInfo[string(metaBlockHashes[i])] = newHeaderInfo(hdr, true, false, false)

		if hdr.GetNonce() > hfb.highestHdrNonce[core.MetachainShardId] {
			hfb.highestHdrNonce[core.MetachainShardId] = hdr.GetNonce()
		}

		shouldConsiderProofsForNotarization := hfb.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, hdr.GetEpoch())
		if !shouldConsiderProofsForNotarization {
			continue
		}

		hfb.requestProofIfNeeded(metaBlockHashes[i], hdr)

		if hdr.GetNonce() > lastMetaBlockNonceWithProof {
			lastMetaBlockNonceWithProof = hdr.GetNonce()
		}
	}

	if hfb.missingHdrs == 0 && lastMetaBlockNonceWithProof == 0 {
		hfb.missingFinalityAttestingHdrs = hfb.requestMissingFinalityAttestingHeaders(
			core.MetachainShardId,
			hfb.blockFinality,
		)
	}
}

func (hfb *headersForBlock) requestMissingAndUpdateBasedOnCrossShardData(cd crossShardMetaData) {
	if cd.GetNonce() == hfb.genesisNonce {
		lastCrossNotarizedHeaderForShard, found := hfb.lastNotarizedShardHeaders[cd.GetShardID()]
		if !found {
			log.Warn("requestMissingAndUpdateBasedOnCrossShardData could not find last notarized", "shard", cd.GetShardID())
			return
		}
		if !bytes.Equal(lastCrossNotarizedHeaderForShard.GetHash(), cd.GetHeaderHash()) {
			log.Warn("genesis hash mismatch",
				"last notarized hash", lastCrossNotarizedHeaderForShard.GetHash(),
				"genesis nonce", hfb.genesisNonce,
				"genesis hash", cd.GetHeaderHash())
		}
		return
	}

	hdr, err := process.GetShardHeaderFromPool(
		cd.GetHeaderHash(),
		hfb.dataPool.Headers())

	if err != nil {
		hfb.missingHdrs++
		hfb.hdrHashAndInfo[string(cd.GetHeaderHash())] = newHeaderInfo(nil, true, false, false)

		go hfb.requestHandler.RequestShardHeader(cd.GetShardID(), cd.GetHeaderHash())
		return
	}

	hfb.hdrHashAndInfo[string(cd.GetHeaderHash())] = newHeaderInfo(hdr, true, false, false)

	hfb.requestProofIfNeeded(cd.GetHeaderHash(), hdr)

	if common.IsEpochChangeBlockForFlagActivation(hdr, hfb.enableEpochsHandler, common.AndromedaFlag) {
		return
	}

	if hdr.GetNonce() > hfb.highestHdrNonce[cd.GetShardID()] {
		hfb.highestHdrNonce[cd.GetShardID()] = hdr.GetNonce()
	}

	hfb.updateLastNotarizedBlockForShard(hdr, cd.GetHeaderHash())
}

func (hfb *headersForBlock) computeExistingAndRequestMissingBasedOnShardData(metaBlock data.MetaHeaderHandler) {
	for _, shardData := range metaBlock.GetShardInfoHandlers() {
		hfb.requestMissingAndUpdateBasedOnCrossShardData(shardData)
	}
}

func (hfb *headersForBlock) computeExistingAndRequestMissingBasedOnShardDataProposal(metaBlock data.MetaHeaderHandler) {
	for _, shardDataProposal := range metaBlock.GetShardInfoProposalHandlers() {
		hfb.requestMissingAndUpdateBasedOnCrossShardData(shardDataProposal)
	}
}

func (hfb *headersForBlock) computeExistingAndRequestMissingShardHeaders(metaBlock data.MetaHeaderHandler) {
	hfb.mutHdrsForBlock.Lock()
	defer hfb.mutHdrsForBlock.Unlock()

	if metaBlock.IsHeaderV3() {
		hfb.computeExistingAndRequestMissingBasedOnShardDataProposal(metaBlock)
		return
	}

	hfb.computeExistingAndRequestMissingBasedOnShardData(metaBlock)
	if hfb.missingHdrs == 0 {
		hfb.missingFinalityAttestingHdrs = hfb.requestMissingFinalityAttestingShardHeaders()
	}
}

func (hfb *headersForBlock) requestProofIfNeeded(currentHeaderHash []byte, header data.HeaderHandler) bool {
	if !hfb.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch()) {
		return false
	}
	if hfb.dataPool.Proofs().HasProof(header.GetShardID(), currentHeaderHash) {
		hfb.setHasProof(string(currentHeaderHash))

		return true
	}

	hfb.setHasProofRequested(string(currentHeaderHash))
	go hfb.requestHandler.RequestEquivalentProofByHash(header.GetShardID(), currentHeaderHash)

	return false
}

// requestMissingFinalityAttestingHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the finality headers greater than the highest header, for given shard,
// related to the block which should be processed
// this method should be called only under the mutex protection: hdrsForCurrBlock.mutHdrsForBlock
func (hfb *headersForBlock) requestMissingFinalityAttestingHeaders(
	shardID uint32,
	finality uint32,
) uint32 {
	requestedHeaders := uint32(0)
	missingFinalityAttestingHeaders := uint32(0)

	highestHdrNonce := hfb.highestHdrNonce[shardID]
	if highestHdrNonce == uint64(0) {
		return missingFinalityAttestingHeaders
	}

	headersPool := hfb.dataPool.Headers()
	lastFinalityAttestingHeader := highestHdrNonce + uint64(finality)
	for i := highestHdrNonce + 1; i <= lastFinalityAttestingHeader; i++ {
		headers, headersHashes, err := headersPool.GetHeadersByNonceAndShardId(i, shardID)
		if err != nil {
			missingFinalityAttestingHeaders++
			requestedHeaders++
			go hfb.requestHeaderByShardAndNonce(shardID, i)
			continue
		}

		for index := range headers {
			hfb.hdrHashAndInfo[string(headersHashes[index])] = newHeaderInfo(
				headers[index],
				false,
				false,
				false,
			)
		}
	}

	if requestedHeaders > 0 {
		log.Debug("requested missing finality attesting headers",
			"num headers", requestedHeaders,
			"shard", shardID)
	}

	return missingFinalityAttestingHeaders
}

func (hfb *headersForBlock) requestHeaderByShardAndNonce(shardID uint32, nonce uint64) {
	if shardID == core.MetachainShardId {
		hfb.requestHandler.RequestMetaHeaderByNonce(nonce)
	} else {
		hfb.requestHandler.RequestShardHeaderByNonce(shardID, nonce)
	}
}

func (hfb *headersForBlock) checkReceivedProofIfAttestingIsNeeded(proof data.HeaderProofHandler) {
	hfb.mutHdrsForBlock.Lock()
	hInfo, ok := hfb.hdrHashAndInfo[string(proof.GetHeaderHash())]
	if !ok {
		hfb.mutHdrsForBlock.Unlock()
		return // proof not missing
	}

	isWaitingForProofs := hInfo.HasProofRequested()
	if !isWaitingForProofs {
		hfb.mutHdrsForBlock.Unlock()
		return
	}

	if hfb.missingProofs > 0 {
		hfb.missingProofs--
	}

	missingHdrs := hfb.missingHdrs
	missingFinalityAttestingHdrs := hfb.missingFinalityAttestingHdrs
	missingProofs := hfb.missingProofs
	hfb.mutHdrsForBlock.Unlock()

	allMissingHeadersReceived := missingHdrs == 0 && missingFinalityAttestingHdrs == 0 && missingProofs == 0
	if allMissingHeadersReceived {
		hfb.chRcvAllHdrs <- true
	}
}

// receivedShardBlock is a call back function which is called when a new header
// is added in the headers pool
func (hfb *headersForBlock) receivedShardBlock(headerHandler data.HeaderHandler, shardHeaderHash []byte) {
	shardHeader, ok := headerHandler.(data.ShardHeaderHandler)
	if !ok {
		return
	}

	log.Trace("received shard header from network",
		"shard", shardHeader.GetShardID(),
		"round", shardHeader.GetRound(),
		"nonce", shardHeader.GetNonce(),
		"hash", shardHeaderHash,
	)

	hfb.mutHdrsForBlock.Lock()

	missingHdrs := hfb.missingHdrs
	missingFinalityAttestingHdrs := hfb.missingFinalityAttestingHdrs
	haveMissingShardHeaders := missingHdrs > 0 || missingFinalityAttestingHdrs > 0
	if haveMissingShardHeaders {
		hdrInfoForHash, found := hfb.hdrHashAndInfo[string(shardHeaderHash)]
		headerInfoIsNotNil := found && !check.IfNil(hdrInfoForHash)
		headerIsMissing := headerInfoIsNotNil && !check.IfNil(hdrInfoForHash) && check.IfNil(hdrInfoForHash.GetHeader())
		hasProof := headerInfoIsNotNil && hdrInfoForHash.HasProof()
		hasProofRequested := headerInfoIsNotNil && hdrInfoForHash.HasProofRequested()
		if headerIsMissing {
			hdrInfoForHash.SetHeader(shardHeader)
			hfb.missingHdrs--

			if shardHeader.GetNonce() > hfb.highestHdrNonce[shardHeader.GetShardID()] {
				hfb.highestHdrNonce[shardHeader.GetShardID()] = shardHeader.GetNonce()
			}
			hfb.updateLastNotarizedBlockForShard(shardHeader, shardHeaderHash)

			if !hasProof && !hasProofRequested {
				hfb.requestProofIfNeeded(shardHeaderHash, shardHeader)
			}
		}

		if hfb.missingHdrs == 0 {
			hfb.missingFinalityAttestingHdrs = hfb.requestMissingFinalityAttestingShardHeaders()
			if hfb.missingFinalityAttestingHdrs == 0 {
				log.Debug("received all missing finality attesting shard headers")
			}
		}

		var missingProofs uint32
		missingHdrs = hfb.missingHdrs
		missingProofs = hfb.missingProofs
		missingFinalityAttestingHdrs = hfb.missingFinalityAttestingHdrs
		hfb.mutHdrsForBlock.Unlock()

		allMissingShardHeadersReceived := missingHdrs == 0 && missingFinalityAttestingHdrs == 0 && missingProofs == 0
		if allMissingShardHeadersReceived {
			hfb.chRcvAllHdrs <- true
		}
	} else {
		hfb.mutHdrsForBlock.Unlock()
	}

	go hfb.requestMiniBlocksIfNeeded(headerHandler)
}

// requestMissingFinalityAttestingShardHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the shardBlockFinality headers greater than the highest shard header,
// for the given shard, related to the block which should be processed
// this method should be called only under the mutex protection: hdrsForCurrBlock.mutHdrsForBlock
func (hfb *headersForBlock) requestMissingFinalityAttestingShardHeaders() uint32 {
	missingFinalityAttestingShardHeaders := uint32(0)

	for shardId := uint32(0); shardId < hfb.shardCoordinator.NumberOfShards(); shardId++ {
		lastNotarizedShardHeader, found := hfb.lastNotarizedShardHeaders[shardId]
		missingFinalityAttestingHeaders := uint32(0)
		if found && lastNotarizedShardHeader != nil && !lastNotarizedShardHeader.NotarizedBasedOnProof() {
			missingFinalityAttestingHeaders = hfb.requestMissingFinalityAttestingHeaders(
				shardId,
				hfb.blockFinality,
			)
		}
		missingFinalityAttestingShardHeaders += missingFinalityAttestingHeaders
	}

	return missingFinalityAttestingShardHeaders
}

// receivedMetaBlock is a callback function when a new metablock was received
// upon receiving, it parses the new metablock and requests miniblocks and transactions
// which destination is the current shard
func (hfb *headersForBlock) receivedMetaBlock(headerHandler data.HeaderHandler, metaBlockHash []byte) {
	metaBlock, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return
	}

	log.Trace("received meta block from network",
		"round", metaBlock.Round,
		"nonce", metaBlock.Nonce,
		"hash", metaBlockHash,
	)

	hfb.mutHdrsForBlock.Lock()

	missingHdrs := hfb.missingHdrs
	missingFinalityAttestingHdrs := hfb.missingFinalityAttestingHdrs
	haveMissingMetaHeaders := missingHdrs > 0 || missingFinalityAttestingHdrs > 0
	if haveMissingMetaHeaders {
		hdrInfoForHash, found := hfb.hdrHashAndInfo[string(metaBlockHash)]
		headerInfoIsNotNil := found && !check.IfNil(hdrInfoForHash)
		headerIsMissing := headerInfoIsNotNil && !check.IfNil(hdrInfoForHash) && check.IfNil(hdrInfoForHash.GetHeader())
		hasProof := headerInfoIsNotNil && hdrInfoForHash.HasProof()
		hasProofRequested := headerInfoIsNotNil && hdrInfoForHash.HasProofRequested()
		if headerIsMissing {
			hdrInfoForHash.SetHeader(metaBlock)
			hfb.missingHdrs--

			if metaBlock.Nonce > hfb.highestHdrNonce[core.MetachainShardId] {
				hfb.highestHdrNonce[core.MetachainShardId] = metaBlock.Nonce
			}

			if !hasProof && !hasProofRequested {
				hfb.requestProofIfNeeded(metaBlockHash, metaBlock)
			}
		}

		if hfb.missingHdrs == 0 {
			hfb.checkFinalityRequestingMissing(metaBlock)

			if hfb.missingFinalityAttestingHdrs == 0 {
				log.Debug("received all missing finality attesting meta headers")
			}
		}

		missingHdrs = hfb.missingHdrs
		missingFinalityAttestingHdrs = hfb.missingFinalityAttestingHdrs
		missingProofs := hfb.missingProofs
		hfb.mutHdrsForBlock.Unlock()

		allMissingMetaHeadersReceived := missingHdrs == 0 && missingFinalityAttestingHdrs == 0 && missingProofs == 0
		if allMissingMetaHeadersReceived {
			hfb.chRcvAllHdrs <- true
		}
	} else {
		hfb.mutHdrsForBlock.Unlock()
	}

	go hfb.requestMiniBlocksIfNeeded(headerHandler)
}

func (hfb *headersForBlock) checkFinalityRequestingMissing(metaBlock *block.MetaBlock) {
	shouldConsiderProofsForNotarization := hfb.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, metaBlock.Epoch)
	if !shouldConsiderProofsForNotarization {
		hfb.missingFinalityAttestingHdrs = hfb.requestMissingFinalityAttestingHeaders(
			core.MetachainShardId,
			hfb.blockFinality,
		)
	}
}

func (hfb *headersForBlock) updateLastNotarizedBlockForShard(hdr data.ShardHeaderHandler, headerHash []byte) {
	lastNotarizedForShard, found := hfb.lastNotarizedShardHeaders[hdr.GetShardID()]
	if !found || lastNotarizedForShard == nil {
		lastNotarizedForShard = newLastNotarizedHeaderInfo(hdr, headerHash, false, false)
		hfb.lastNotarizedShardHeaders[hdr.GetShardID()] = lastNotarizedForShard
	}

	if hdr.GetNonce() >= lastNotarizedForShard.GetHeader().GetNonce() {
		notarizedBasedOnProofs := hfb.enableEpochsHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, hdr.GetEpoch())
		hasProof := false
		if notarizedBasedOnProofs {
			hasProof = hfb.dataPool.Proofs().HasProof(hdr.GetShardID(), headerHash)
		}

		lastNotarizedForShard.SetHeader(hdr)
		lastNotarizedForShard.SetHash(headerHash)
		lastNotarizedForShard.SetNotarizedBasedOnProof(notarizedBasedOnProofs)
		lastNotarizedForShard.SetHasProof(hasProof)
	}
}

func (hfb *headersForBlock) requestMiniBlocksIfNeeded(headerHandler data.HeaderHandler) {
	lastCrossNotarizedHeader, _, err := hfb.blockTracker.GetLastCrossNotarizedHeader(headerHandler.GetShardID())
	if err != nil {
		log.Debug("requestMiniBlocksIfNeeded.GetLastCrossNotarizedHeader",
			"shard", headerHandler.GetShardID(),
			"error", err.Error())
		return
	}

	isHeaderOutOfRequestRange := headerHandler.GetNonce() > lastCrossNotarizedHeader.GetNonce()+process.MaxHeadersToRequestInAdvance
	if isHeaderOutOfRequestRange {
		return
	}

	waitTime := hfb.extraDelayRequestBlockInfo
	roundDifferences := hfb.roundHandler.Index() - int64(headerHandler.GetRound())
	if roundDifferences > 1 {
		waitTime = 0
	}

	// waiting for late broadcast of mini blocks and transactions to be done and received
	time.Sleep(waitTime)

	hfb.txCoordinator.RequestMiniBlocksAndTransactions(headerHandler)
}

// IsInterfaceNil returns true if underlying object is nil
func (hfb *headersForBlock) IsInterfaceNil() bool {
	return hfb == nil
}
