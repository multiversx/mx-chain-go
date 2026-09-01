package track

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process"
)

type shardBlockTrack struct {
	*baseBlockTrack

	mutPendingSelfHeaders  sync.Mutex
	pendingSelfHeaders     map[string]*pendingSelfHeader
	resolvedSelfHeaders    map[string]uint64
	numPendingSelfHeaders  atomic.Int64
	numResolvedSelfHeaders atomic.Int64
}

type pendingSelfHeader struct {
	hash        []byte
	shardID     uint32
	nonce       uint64
	epoch       uint32
	lastRequest time.Time
}

// NewShardBlockTrack creates an object for tracking the received shard blocks
func NewShardBlockTrack(arguments ArgShardTracker) (*shardBlockTrack, error) {
	err := checkTrackerNilParameters(arguments.ArgBaseTracker)
	if err != nil {
		return nil, err
	}

	bbt, err := createBaseBlockTrack(arguments.ArgBaseTracker)
	if err != nil {
		return nil, err
	}

	err = bbt.initNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	sbt := shardBlockTrack{
		baseBlockTrack:      bbt,
		pendingSelfHeaders:  make(map[string]*pendingSelfHeader),
		resolvedSelfHeaders: make(map[string]uint64),
	}

	argBlockProcessor := ArgBlockProcessor{
		HeaderValidator:                       arguments.HeaderValidator,
		RequestHandler:                        arguments.RequestHandler,
		ShardCoordinator:                      arguments.ShardCoordinator,
		BlockTracker:                          &sbt,
		CrossNotarizer:                        bbt.crossNotarizer,
		SelfNotarizer:                         bbt.selfNotarizer,
		CrossNotarizedHeadersNotifier:         bbt.crossNotarizedHeadersNotifier,
		SelfNotarizedFromCrossHeadersNotifier: bbt.selfNotarizedFromCrossHeadersNotifier,
		SelfNotarizedHeadersNotifier:          bbt.selfNotarizedHeadersNotifier,
		FinalMetachainHeadersNotifier:         bbt.finalMetachainHeadersNotifier,
		RoundHandler:                          arguments.RoundHandler,
		ProcessConfigsHandler:                 arguments.ProcessConfigsHandler,
		EnableEpochsHandler:                   arguments.EnableEpochsHandler,
		EnableRoundsHandler:                   arguments.EnableRoundsHandler,
		ProofsPool:                            arguments.ProofsPool,
		Marshaller:                            arguments.Marshalizer,
		Hasher:                                arguments.Hasher,
		HeadersPool:                           arguments.PoolsHolder.Headers(),
		IsImportDBMode:                        arguments.IsImportDBMode,
	}

	blockProcessorObject, err := NewBlockProcessor(argBlockProcessor)
	if err != nil {
		return nil, err
	}

	sbt.blockProcessor = blockProcessorObject
	sbt.headers = make(map[uint32]map[uint64][]*HeaderInfo)
	sbt.headersPool.RegisterHandler(sbt.receivedHeader)
	sbt.proofsPool.RegisterHandler(sbt.receivedProof)
	sbt.headersPool.Clear()

	return &sbt, nil
}

func (sbt *shardBlockTrack) receivedHeader(header data.HeaderHandler, hash []byte) {
	sbt.receivedPendingSelfHeader(header, hash)
	sbt.baseBlockTrack.receivedHeader(header, hash)
}

// GetSelfHeaders gets a slice of self headers from a given metablock
func (sbt *shardBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*HeaderInfo {
	selfHeadersInfo := make([]*HeaderInfo, 0)

	metaBlock, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		log.Debug("GetSelfHeaders", "error", process.ErrWrongTypeAssertion)
		return selfHeadersInfo
	}

	for _, shardInfo := range process.GetShardHeadersReferencedByMeta(metaBlock) {
		if shardInfo.GetShardID() != sbt.shardCoordinator.SelfId() {
			continue
		}
		if metaBlock.IsHeaderV3() {
			headerInfo, handled := sbt.getHeaderFromPendingState(shardInfo)
			if handled {
				if headerInfo != nil {
					selfHeadersInfo = append(selfHeadersInfo, headerInfo)
				}
				continue
			}
		}

		header, err := process.GetShardHeader(shardInfo.GetHeaderHash(), sbt.headersPool, sbt.marshalizer, sbt.store)
		if err != nil {
			log.Trace("GetSelfHeaders.GetShardHeader", "error", err.Error())

			header, err = sbt.getTrackedShardHeaderWithNonceAndHash(shardInfo.GetShardID(), shardInfo.GetNonce(), shardInfo.GetHeaderHash())
			if err != nil {
				log.Trace("GetSelfHeaders.getTrackedShardHeaderWithNonceAndHash", "error", err.Error())
				if !metaBlock.IsHeaderV3() {
					continue
				}

				sbt.addOrRefreshPendingSelfHeader(shardInfo)
				header, err = sbt.getPendingSelfHeader(shardInfo)
				if err != nil {
					continue
				}
			}
		}

		if !sbt.claimPendingSelfHeaderForReturn(shardInfo.GetHeaderHash(), shardInfo.GetNonce()) {
			continue
		}

		selfHeadersInfo = append(selfHeadersInfo, &HeaderInfo{Hash: shardInfo.GetHeaderHash(), Header: header})
	}

	return selfHeadersInfo
}

func (sbt *shardBlockTrack) getHeaderFromPendingState(shardInfo process.ShardInfoHandler) (*HeaderInfo, bool) {
	if sbt.numPendingSelfHeaders.Load() == 0 && sbt.numResolvedSelfHeaders.Load() == 0 {
		return nil, false
	}

	key := string(shardInfo.GetHeaderHash())
	sbt.mutPendingSelfHeaders.Lock()
	sbt.pruneResolvedSelfHeadersBefore(shardInfo.GetNonce())
	_, resolved := sbt.resolvedSelfHeaders[key]
	_, pending := sbt.pendingSelfHeaders[key]
	sbt.mutPendingSelfHeaders.Unlock()
	if resolved {
		return nil, true
	}
	if !pending {
		return nil, false
	}

	header, err := sbt.getPendingSelfHeader(shardInfo)
	if err != nil {
		sbt.addOrRefreshPendingSelfHeader(shardInfo)
		return nil, true
	}
	if !sbt.claimPendingSelfHeaderForReturn(shardInfo.GetHeaderHash(), shardInfo.GetNonce()) {
		return nil, true
	}

	return &HeaderInfo{Hash: shardInfo.GetHeaderHash(), Header: header}, true
}

func (sbt *shardBlockTrack) getPendingSelfHeader(shardInfo process.ShardInfoHandler) (data.ShardHeaderHandler, error) {
	header, err := sbt.headersPool.GetHeaderByHash(shardInfo.GetHeaderHash())
	if err == nil && !check.IfNil(header) {
		shardHeader, ok := header.(data.ShardHeaderHandler)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		return shardHeader, nil
	}

	return sbt.getTrackedShardHeaderWithNonceAndHash(
		shardInfo.GetShardID(),
		shardInfo.GetNonce(),
		shardInfo.GetHeaderHash(),
	)
}

func (sbt *shardBlockTrack) addOrRefreshPendingSelfHeader(shardInfo process.ShardInfoHandler) {
	now := time.Now()
	requestInterval := sbt.requestHandler.RequestInterval()
	key := string(shardInfo.GetHeaderHash())
	request := false

	sbt.mutPendingSelfHeaders.Lock()
	sbt.pruneResolvedSelfHeadersBefore(shardInfo.GetNonce())
	if _, resolved := sbt.resolvedSelfHeaders[key]; resolved {
		sbt.mutPendingSelfHeaders.Unlock()
		return
	}

	pending, exists := sbt.pendingSelfHeaders[key]
	if exists {
		if now.Sub(pending.lastRequest) >= requestInterval {
			pending.lastRequest = now
			request = true
		}
		sbt.mutPendingSelfHeaders.Unlock()
		if request {
			sbt.requestPendingSelfHeader(pending)
		}
		return
	}

	pending = &pendingSelfHeader{
		hash:        append([]byte(nil), shardInfo.GetHeaderHash()...),
		shardID:     shardInfo.GetShardID(),
		nonce:       shardInfo.GetNonce(),
		epoch:       shardInfo.GetEpoch(),
		lastRequest: now,
	}
	hasRoom, evicted := sbt.makeRoomForPendingSelfHeader(pending.nonce)
	if !hasRoom {
		sbt.mutPendingSelfHeaders.Unlock()
		return
	}

	sbt.pendingSelfHeaders[key] = pending
	sbt.numPendingSelfHeaders.Store(int64(len(sbt.pendingSelfHeaders)))
	sbt.mutPendingSelfHeaders.Unlock()

	if evicted != nil {
		log.Warn("evicted unresolved held-final shard header", "nonce", evicted.nonce, "hash", evicted.hash)
	}
	log.Debug("retained unresolved held-final shard header", "nonce", pending.nonce, "hash", pending.hash)
	sbt.requestPendingSelfHeader(pending)
}

func (sbt *shardBlockTrack) makeRoomForPendingSelfHeader(newNonce uint64) (bool, *pendingSelfHeader) {
	if sbt.maxNumHeadersToKeepPerShard <= 0 {
		return false, nil
	}
	if len(sbt.pendingSelfHeaders) < sbt.maxNumHeadersToKeepPerShard {
		return true, nil
	}

	var highestKey string
	var highestNonce uint64
	for key, pending := range sbt.pendingSelfHeaders {
		if highestKey == "" || pending.nonce > highestNonce {
			highestKey = key
			highestNonce = pending.nonce
		}
	}
	if newNonce >= highestNonce {
		return false, nil
	}

	evicted := sbt.pendingSelfHeaders[highestKey]
	delete(sbt.pendingSelfHeaders, highestKey)

	return true, evicted
}

func (sbt *shardBlockTrack) requestPendingSelfHeader(pending *pendingSelfHeader) {
	sbt.requestHandler.RequestShardHeaderForEpoch(pending.shardID, pending.hash, pending.epoch)
	sbt.requestHandler.RequestEquivalentProofByHashForEpoch(pending.shardID, pending.hash, pending.epoch)
}

func (sbt *shardBlockTrack) claimPendingSelfHeaderForReturn(hash []byte, nonce uint64) bool {
	if sbt.numPendingSelfHeaders.Load() == 0 && sbt.numResolvedSelfHeaders.Load() == 0 {
		return true
	}

	key := string(hash)

	sbt.mutPendingSelfHeaders.Lock()
	defer sbt.mutPendingSelfHeaders.Unlock()
	sbt.pruneResolvedSelfHeadersBefore(nonce)

	if _, resolved := sbt.resolvedSelfHeaders[key]; resolved {
		return false
	}
	if _, pending := sbt.pendingSelfHeaders[key]; !pending {
		return true
	}

	delete(sbt.pendingSelfHeaders, key)
	sbt.rememberResolvedSelfHeader(key, nonce)
	sbt.numPendingSelfHeaders.Store(int64(len(sbt.pendingSelfHeaders)))

	return true
}

func (sbt *shardBlockTrack) receivedPendingSelfHeader(header data.HeaderHandler, hash []byte) {
	if sbt.numPendingSelfHeaders.Load() == 0 || check.IfNil(header) {
		return
	}

	key := string(hash)
	sbt.mutPendingSelfHeaders.Lock()
	pending, exists := sbt.pendingSelfHeaders[key]
	if !exists || header.GetShardID() != pending.shardID || header.GetNonce() != pending.nonce || header.GetEpoch() != pending.epoch {
		sbt.mutPendingSelfHeaders.Unlock()
		return
	}

	delete(sbt.pendingSelfHeaders, key)
	sbt.rememberResolvedSelfHeader(key, pending.nonce)
	sbt.numPendingSelfHeaders.Store(int64(len(sbt.pendingSelfHeaders)))
	sbt.mutPendingSelfHeaders.Unlock()

	log.Debug("resolved held-final shard header", "nonce", pending.nonce, "hash", pending.hash)
	sbt.selfNotarizedFromCrossHeadersNotifier.CallHandlers(
		core.MetachainShardId,
		[]data.HeaderHandler{header},
		[][]byte{pending.hash},
	)
}

func (sbt *shardBlockTrack) rememberResolvedSelfHeader(key string, nonce uint64) {
	if len(sbt.resolvedSelfHeaders) >= sbt.maxNumHeadersToKeepPerShard {
		for resolvedKey := range sbt.resolvedSelfHeaders {
			delete(sbt.resolvedSelfHeaders, resolvedKey)
			break
		}
	}

	sbt.resolvedSelfHeaders[key] = nonce
	sbt.numResolvedSelfHeaders.Store(int64(len(sbt.resolvedSelfHeaders)))
}

func (sbt *shardBlockTrack) pruneResolvedSelfHeadersBefore(nonce uint64) {
	for key, resolvedNonce := range sbt.resolvedSelfHeaders {
		if resolvedNonce < nonce {
			delete(sbt.resolvedSelfHeaders, key)
		}
	}
	sbt.numResolvedSelfHeaders.Store(int64(len(sbt.resolvedSelfHeaders)))
}

// RestoreToGenesis resets all tracked shard state.
func (sbt *shardBlockTrack) RestoreToGenesis() {
	sbt.baseBlockTrack.RestoreToGenesis()

	sbt.mutPendingSelfHeaders.Lock()
	sbt.pendingSelfHeaders = make(map[string]*pendingSelfHeader)
	sbt.resolvedSelfHeaders = make(map[string]uint64)
	sbt.numPendingSelfHeaders.Store(0)
	sbt.numResolvedSelfHeaders.Store(0)
	sbt.mutPendingSelfHeaders.Unlock()
}

func (sbt *shardBlockTrack) getTrackedShardHeaderWithNonceAndHash(
	shardID uint32,
	nonce uint64,
	hash []byte,
) (data.ShardHeaderHandler, error) {

	headers, headersHashes := sbt.GetTrackedHeadersWithNonce(shardID, nonce)
	for i := 0; i < len(headers); i++ {
		if !bytes.Equal(headersHashes[i], hash) {
			continue
		}

		header, ok := headers[i].(data.ShardHeaderHandler)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}

		return header, nil
	}

	return nil, process.ErrMissingHeader
}

// CleanupInvalidCrossHeaders cleans headers added to the block tracker that have become invalid after processing
func (sbt *shardBlockTrack) CleanupInvalidCrossHeaders(_ uint32, _ uint64) {
	// no rule for shard
}

// ComputeLongestSelfChain computes the longest chain from self shard
func (sbt *shardBlockTrack) ComputeLongestSelfChain() (data.HeaderHandler, []byte, []data.HeaderHandler, [][]byte) {
	lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, err := sbt.selfNotarizer.GetLastNotarizedHeader(core.MetachainShardId)
	if err != nil {
		log.Warn("ComputeLongestSelfChain.GetLastNotarizedHeader", "error", err.Error())
		return nil, nil, nil, nil
	}

	headers, hashes := sbt.ComputeLongestChain(sbt.shardCoordinator.SelfId(), lastSelfNotarizedHeader)
	return lastSelfNotarizedHeader, lastSelfNotarizedHeaderHash, headers, hashes
}

// ComputeCrossInfo computes the cross info from a given slice of metablocks
func (sbt *shardBlockTrack) ComputeCrossInfo(headers []data.HeaderHandler) {
	lenHeaders := len(headers)
	if lenHeaders == 0 {
		return
	}

	metaBlock, ok := headers[lenHeaders-1].(data.MetaHeaderHandler)
	if !ok {
		log.Debug("ComputeCrossInfo", "error", process.ErrWrongTypeAssertion)
		return
	}

	sbt.setShardInfoData(metaBlock)

	log.Debug("compute cross info from meta block",
		"epoch", metaBlock.GetEpoch(),
		"round", metaBlock.GetRound(),
		"nonce", metaBlock.GetNonce(),
	)

	for shardID := uint32(0); shardID < sbt.shardCoordinator.NumberOfShards(); shardID++ {
		log.Debug("cross info",
			"shard", shardID,
			"pending miniblocks", sbt.blockBalancer.GetNumPendingMiniBlocks(shardID),
			"last meta nonce processed", sbt.blockBalancer.GetLastShardProcessedMetaNonce(shardID),
			"shard is stuck", sbt.IsShardStuck(shardID),
			"global chain stuck", sbt.ShouldSkipMiniBlocksCreationFromSelf())
	}
}

func (sbt *shardBlockTrack) setShardInfoData(
	metaBlock data.MetaHeaderHandler,
) {
	for _, shardInfo := range metaBlock.GetShardInfoHandlers() {
		if !metaBlock.IsHeaderV3() {
			sbt.blockBalancer.SetNumPendingMiniBlocks(shardInfo.GetShardID(), shardInfo.GetNumPendingMiniBlocks())
		}

		sbt.blockBalancer.SetLastShardProcessedMetaNonce(shardInfo.GetShardID(), shardInfo.GetLastIncludedMetaNonce())
	}

	for _, shardInfo := range metaBlock.GetShardInfoProposalHandlers() {
		if metaBlock.IsHeaderV3() {
			sbt.blockBalancer.SetNumPendingMiniBlocks(shardInfo.GetShardID(), shardInfo.GetNumPendingMiniBlocks())
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (sbt *shardBlockTrack) IsInterfaceNil() bool {
	return sbt == nil
}
