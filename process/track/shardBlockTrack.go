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

const maxPendingMetaSources = 4

type shardBlockTrack struct {
	*baseBlockTrack
	metaFinalityView process.MetaFinalityView

	mutPendingSelfHeaders   sync.Mutex
	mutPendingNotifications sync.Mutex
	pendingNotifications    sync.WaitGroup
	pendingSelfHeadersView  atomic.Uint64
	pendingSelfHeaders      map[string]*pendingSelfHeader
	resolvedSelfHeaders     map[string]uint64
	numPendingSelfHeaders   atomic.Int64
	numResolvedSelfHeaders  atomic.Int64
}

type pendingSelfHeader struct {
	hash              []byte
	shardID           uint32
	nonce             uint64
	epoch             uint32
	lastRequest       time.Time
	lastSourceRequest time.Time
	view              uint64
	sourcesVersion    uint64
	sources           []pendingMetaSource
	sourceOverflow    bool
	lastOverflowScan  time.Time
	nextSourceToAsk   int
	deliveredHeader   data.HeaderHandler
}

type pendingMetaSource struct {
	header                   data.MetaHeaderHandler
	hash                     []byte
	nonce                    uint64
	crossedByCanonicalAnchor bool
}

type pendingSourceRequest struct {
	hash  []byte
	epoch uint32
}

type sourceMetaVerdict uint8

const (
	sourceMetaUnknown sourceMetaVerdict = iota
	sourceMetaHeldFinal
	sourceMetaDead
)

type metaAuthoritySnapshot struct {
	hashes  map[string]struct{}
	sources []pendingMetaSource
	valid   bool
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

	metaFinalityView, err := NewMetaFinalityView(ArgsMetaFinalityView{
		HeadersPool: bbt.headersPool,
		ProofsPool:  bbt.proofsPool,
	})
	if err != nil {
		return nil, err
	}
	err = bbt.initNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	sbt := shardBlockTrack{
		baseBlockTrack:      bbt,
		metaFinalityView:    metaFinalityView,
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
	if !check.IfNil(header) && header.GetShardID() != core.MetachainShardId {
		sbt.receivedPendingSelfHeader(header, hash)
	}
	sbt.baseBlockTrack.receivedHeader(header, hash)
	if !check.IfNil(header) && header.GetShardID() == core.MetachainShardId {
		sbt.retryPendingSelfHeaders()
	}
}

func (sbt *shardBlockTrack) receivedProof(proof data.HeaderProofHandler) {
	sbt.baseBlockTrack.receivedProof(proof)
	if !check.IfNil(proof) && proof.GetHeaderShardId() == core.MetachainShardId {
		sbt.retryPendingSelfHeaders()
	}
}

// GetSelfHeaders gets a slice of self headers from a given metablock
func (sbt *shardBlockTrack) GetSelfHeaders(headerHandler data.HeaderHandler) []*HeaderInfo {
	return unwrapSelfHeaderInfo(sbt.getSelfHeadersWithSource(headerHandler, nil))
}

func (sbt *shardBlockTrack) getSelfHeadersWithSource(headerHandler data.HeaderHandler, headerHash []byte) []*selfHeaderInfo {
	selfHeadersInfo := make([]*selfHeaderInfo, 0)

	metaBlock, ok := headerHandler.(data.MetaHeaderHandler)
	if !ok {
		log.Debug("GetSelfHeaders", "error", process.ErrWrongTypeAssertion)
		return selfHeadersInfo
	}
	pendingView := sbt.pendingSelfHeadersView.Load()

	for _, shardInfo := range process.GetShardHeadersReferencedByMeta(metaBlock) {
		if shardInfo.GetShardID() != sbt.shardCoordinator.SelfId() {
			continue
		}
		if metaBlock.IsHeaderV3() {
			headerInfo, handled := sbt.getHeaderFromPendingState(shardInfo, pendingView, metaBlock, headerHash)
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

				sbt.addOrRefreshPendingSelfHeader(shardInfo, pendingView, metaBlock, headerHash)
				header, err = sbt.getPendingSelfHeader(shardInfo)
				if err != nil {
					continue
				}
			}
		}

		if !sbt.prepareSelfHeaderForReturn(
			header,
			shardInfo.GetHeaderHash(),
			shardInfo.GetNonce(),
			pendingView,
			metaBlock.IsHeaderV3(),
		) {
			continue
		}

		selfHeadersInfo = append(selfHeadersInfo, sbt.newSelfHeaderInfo(
			header,
			shardInfo.GetHeaderHash(),
			metaBlock,
			headerHash,
			pendingView,
		))
	}
	if metaBlock.IsHeaderV3() && !sbt.isPendingSelfHeadersViewCurrent(pendingView) {
		clear(selfHeadersInfo)
		return selfHeadersInfo[:0]
	}

	return selfHeadersInfo
}

func (sbt *shardBlockTrack) getHeaderFromPendingState(
	shardInfo process.ShardInfoHandler,
	pendingView uint64,
	sourceMetaHeader data.MetaHeaderHandler,
	sourceMetaHash []byte,
) (*selfHeaderInfo, bool) {
	if !sbt.isPendingSelfHeadersViewCurrent(pendingView) {
		return nil, true
	}
	if sbt.numPendingSelfHeaders.Load() == 0 && sbt.numResolvedSelfHeaders.Load() == 0 {
		return nil, false
	}

	key := string(shardInfo.GetHeaderHash())
	sbt.mutPendingSelfHeaders.Lock()
	if !sbt.isPendingSelfHeadersViewCurrent(pendingView) {
		sbt.mutPendingSelfHeaders.Unlock()
		return nil, true
	}
	_, resolved := sbt.resolvedSelfHeaders[key]
	pendingHeader, pending := sbt.pendingSelfHeaders[key]
	pending = pending && pendingHeader.view == pendingView
	sourceOverflowed := false
	if pending {
		sourceOverflowed = sbt.updatePendingSource(pendingHeader, sourceMetaHeader, sourceMetaHash)
	}
	sbt.mutPendingSelfHeaders.Unlock()
	if resolved {
		return nil, true
	}
	if !pending {
		return nil, false
	}
	sbt.addOrRefreshPendingSelfHeader(shardInfo, pendingView, sourceMetaHeader, sourceMetaHash)

	header, err := sbt.getPendingSelfHeader(shardInfo)
	if err != nil {
		return nil, true
	}
	if sourceOverflowed {
		sbt.resolvePendingSelfHeaderWithSnapshot(key, header, nil, 0, &pendingMetaSource{
			header: sourceMetaHeader,
			hash:   append([]byte(nil), sourceMetaHash...),
			nonce:  sourceMetaHeader.GetNonce(),
		})
		return nil, true
	}
	if !sbt.prepareSelfHeaderForReturn(header, shardInfo.GetHeaderHash(), shardInfo.GetNonce(), pendingView, true) {
		return nil, true
	}

	return sbt.newSelfHeaderInfo(
		header,
		shardInfo.GetHeaderHash(),
		sourceMetaHeader,
		sourceMetaHash,
		pendingView,
	), true
}

func (sbt *shardBlockTrack) newSelfHeaderInfo(
	header data.HeaderHandler,
	headerHash []byte,
	sourceMetaHeader data.MetaHeaderHandler,
	sourceMetaHash []byte,
	sourceView uint64,
) *selfHeaderInfo {
	info := &selfHeaderInfo{
		Hash:   headerHash,
		Header: header,
	}
	if check.IfNil(sourceMetaHeader) || !sourceMetaHeader.IsHeaderV3() {
		return info
	}

	info.sourceMetaHeader = sourceMetaHeader
	info.sourceMetaHash = append([]byte(nil), sourceMetaHash...)
	info.sourceView = sourceView

	return info
}

func unwrapSelfHeaderInfo(headersInfo []*selfHeaderInfo) []*HeaderInfo {
	unwrapped := make([]*HeaderInfo, 0, len(headersInfo))
	for _, headerInfo := range headersInfo {
		unwrapped = append(unwrapped, &HeaderInfo{
			Hash:   headerInfo.Hash,
			Header: headerInfo.Header,
		})
	}

	return unwrapped
}

func (sbt *shardBlockTrack) updatePendingSource(
	pending *pendingSelfHeader,
	sourceMetaHeader data.MetaHeaderHandler,
	sourceMetaHash []byte,
) bool {
	if check.IfNil(sourceMetaHeader) || len(sourceMetaHash) == 0 {
		return false
	}
	for index := range pending.sources {
		if bytes.Equal(pending.sources[index].hash, sourceMetaHash) {
			pending.sources[index].header = sourceMetaHeader
			return false
		}
	}
	if len(pending.sources) >= maxPendingMetaSources {
		if !pending.sourceOverflow {
			pending.sourceOverflow = true
			pending.sourcesVersion++
		}
		return true
	}

	updatedSources, changed := addOrUpdateMetaSource(pending.sources, sourceMetaHeader, sourceMetaHash)
	pending.sources = updatedSources
	if changed {
		pending.sourcesVersion++
	}

	return false
}

func addOrUpdateMetaSource(
	sources []pendingMetaSource,
	sourceMetaHeader data.MetaHeaderHandler,
	sourceMetaHash []byte,
) ([]pendingMetaSource, bool) {
	if check.IfNil(sourceMetaHeader) || len(sourceMetaHash) == 0 {
		return sources, false
	}

	sourceNonce := sourceMetaHeader.GetNonce()
	for index := range sources {
		if bytes.Equal(sources[index].hash, sourceMetaHash) {
			sources[index].header = sourceMetaHeader
			return sources, false
		}
	}

	sources = append(sources, pendingMetaSource{
		header: sourceMetaHeader,
		hash:   append([]byte(nil), sourceMetaHash...),
		nonce:  sourceNonce,
	})

	return sources, true
}

func (sbt *shardBlockTrack) setInitialPendingSource(
	pending *pendingSelfHeader,
	sourceMetaHeader data.MetaHeaderHandler,
	sourceMetaHash []byte,
) {
	if check.IfNil(sourceMetaHeader) {
		return
	}

	pending.sources = []pendingMetaSource{{
		header: sourceMetaHeader,
		hash:   append([]byte(nil), sourceMetaHash...),
		nonce:  sourceMetaHeader.GetNonce(),
	}}
	pending.sourcesVersion = 1
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

func (sbt *shardBlockTrack) addOrRefreshPendingSelfHeader(
	shardInfo process.ShardInfoHandler,
	pendingView uint64,
	sourceMetaHeader data.MetaHeaderHandler,
	sourceMetaHash []byte,
) {
	now := time.Now()
	requestInterval := sbt.requestHandler.RequestInterval()
	key := string(shardInfo.GetHeaderHash())
	request := false

	sbt.mutPendingSelfHeaders.Lock()
	if !sbt.isPendingSelfHeadersViewCurrent(pendingView) {
		sbt.mutPendingSelfHeaders.Unlock()
		return
	}
	if _, resolved := sbt.resolvedSelfHeaders[key]; resolved {
		sbt.mutPendingSelfHeaders.Unlock()
		return
	}

	pending, exists := sbt.pendingSelfHeaders[key]
	if exists {
		if pending.view != pendingView {
			sbt.mutPendingSelfHeaders.Unlock()
			return
		}
		sbt.updatePendingSource(pending, sourceMetaHeader, sourceMetaHash)
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
		view:        pendingView,
	}
	sbt.setInitialPendingSource(pending, sourceMetaHeader, sourceMetaHash)
	hasRoom, evicted := sbt.makeRoomForPendingSelfHeader(pending.nonce)
	if !hasRoom {
		sbt.mutPendingSelfHeaders.Unlock()
		return
	}

	sbt.numPendingSelfHeaders.Store(int64(len(sbt.pendingSelfHeaders) + 1))
	sbt.pendingSelfHeaders[key] = pending
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

func (sbt *shardBlockTrack) prepareSelfHeaderForReturn(
	header data.HeaderHandler,
	hash []byte,
	_ uint64,
	pendingView uint64,
	checkPendingView bool,
) bool {
	if checkPendingView && !sbt.isPendingSelfHeadersViewCurrent(pendingView) {
		return false
	}
	if sbt.numPendingSelfHeaders.Load() == 0 && sbt.numResolvedSelfHeaders.Load() == 0 {
		return true
	}

	key := string(hash)

	sbt.mutPendingSelfHeaders.Lock()
	if _, resolved := sbt.resolvedSelfHeaders[key]; resolved {
		sbt.mutPendingSelfHeaders.Unlock()
		return false
	}
	_, exists := sbt.pendingSelfHeaders[key]
	if !exists {
		sbt.mutPendingSelfHeaders.Unlock()
		return true
	}
	sbt.mutPendingSelfHeaders.Unlock()

	sbt.resolvePendingSelfHeader(key, header)

	return false
}

func (sbt *shardBlockTrack) receivedPendingSelfHeader(header data.HeaderHandler, hash []byte) {
	if sbt.numPendingSelfHeaders.Load() == 0 || check.IfNil(header) {
		return
	}

	sbt.resolvePendingSelfHeader(string(hash), header)
}

func (sbt *shardBlockTrack) resolvePendingSelfHeader(key string, deliveredHeader data.HeaderHandler) {
	sbt.resolvePendingSelfHeaderWithSnapshot(key, deliveredHeader, nil, 0, nil)
}

func (sbt *shardBlockTrack) resolvePendingSelfHeaderWithSnapshot(
	key string,
	deliveredHeader data.HeaderHandler,
	providedSnapshot *metaAuthoritySnapshot,
	snapshotView uint64,
	transientSource *pendingMetaSource,
) {
	sbt.mutPendingNotifications.Lock()
	sbt.pendingNotifications.Wait()

	sbt.mutPendingSelfHeaders.Lock()
	pending, exists := sbt.pendingSelfHeaders[key]
	if !exists || !sbt.isPendingSelfHeadersViewCurrent(pending.view) {
		sbt.mutPendingSelfHeaders.Unlock()
		sbt.mutPendingNotifications.Unlock()
		return
	}
	if !check.IfNil(deliveredHeader) {
		if !sbt.pendingSelfHeaderMatches(pending, deliveredHeader) {
			sbt.mutPendingSelfHeaders.Unlock()
			sbt.mutPendingNotifications.Unlock()
			return
		}
		pending.deliveredHeader = deliveredHeader
	}
	header := pending.deliveredHeader
	if check.IfNil(header) {
		sbt.mutPendingSelfHeaders.Unlock()
		sbt.mutPendingNotifications.Unlock()
		return
	}
	sources := clonePendingMetaSources(pending.sources)
	sourcesVersion := pending.sourcesVersion
	sourceOverflow := pending.sourceOverflow
	view := pending.view
	sbt.mutPendingSelfHeaders.Unlock()

	snapshot := metaAuthoritySnapshot{}
	if providedSnapshot != nil && snapshotView == view && sbt.isPendingSelfHeadersViewCurrent(snapshotView) {
		snapshot = *providedSnapshot
	} else {
		snapshot = sbt.currentMetaAuthoritySnapshot()
	}
	hasHeldFinalSource := false
	var unknownSources []pendingMetaSource
	for _, source := range sources {
		verdict := sbt.sourceMetaVerdict(
			source.header,
			source.hash,
			source.nonce,
			snapshot,
			source.crossedByCanonicalAnchor,
		)
		if verdict == sourceMetaHeldFinal {
			hasHeldFinalSource = true
			break
		}
		if verdict == sourceMetaUnknown {
			unknownSources = append(unknownSources, source)
		}
	}
	if !hasHeldFinalSource && transientSource != nil {
		hasHeldFinalSource = sbt.sourceMetaVerdict(
			transientSource.header,
			transientSource.hash,
			transientSource.nonce,
			snapshot,
			false,
		) == sourceMetaHeldFinal
	}
	if !hasHeldFinalSource && len(unknownSources) != len(sources) {
		sbt.mutPendingSelfHeaders.Lock()
		current, currentExists := sbt.pendingSelfHeaders[key]
		if !currentExists || current != pending || current.sourcesVersion != sourcesVersion || current.view != view ||
			!sbt.isPendingSelfHeadersViewCurrent(view) {
			sbt.mutPendingSelfHeaders.Unlock()
			sbt.mutPendingNotifications.Unlock()
			return
		}
		current.sources = unknownSources
		current.sourcesVersion++
		sourcesVersion = current.sourcesVersion
		sbt.mutPendingSelfHeaders.Unlock()
	}
	if !hasHeldFinalSource && sourceOverflow &&
		sbt.reservePendingOverflowScan(key, pending, sourcesVersion, view) {
		hasHeldFinalSource = sbt.snapshotAuthorizesPending(snapshot, pending)
	}
	if !hasHeldFinalSource && (sourceOverflow || len(unknownSources) > 0) {
		sbt.mutPendingNotifications.Unlock()
		requests := sbt.preparePendingSourceRequests(key, pending, sourcesVersion, unknownSources, snapshot.hashes)
		sbt.requestPendingSources(requests)
		return
	}

	sbt.mutPendingSelfHeaders.Lock()
	currentPending, exists := sbt.pendingSelfHeaders[key]
	if !exists || currentPending != pending || currentPending.sourcesVersion != sourcesVersion || currentPending.view != view ||
		!sbt.isPendingSelfHeadersViewCurrent(view) {
		sbt.mutPendingSelfHeaders.Unlock()
		sbt.mutPendingNotifications.Unlock()
		return
	}
	delete(sbt.pendingSelfHeaders, key)
	sbt.numPendingSelfHeaders.Store(int64(len(sbt.pendingSelfHeaders)))
	if hasHeldFinalSource {
		sbt.rememberResolvedSelfHeader(key, pending.nonce)
	}
	sbt.mutPendingSelfHeaders.Unlock()
	if !hasHeldFinalSource {
		log.Debug("discarded shard header from dead meta authority", "nonce", pending.nonce, "hash", pending.hash)
		sbt.mutPendingNotifications.Unlock()
		return
	}
	sbt.notifyPendingSelfHeader(header, pending)
	sbt.mutPendingNotifications.Unlock()
}

func (sbt *shardBlockTrack) reservePendingOverflowScan(
	key string,
	pending *pendingSelfHeader,
	sourcesVersion uint64,
	view uint64,
) bool {
	now := time.Now()
	requestInterval := sbt.requestHandler.RequestInterval()

	sbt.mutPendingSelfHeaders.Lock()
	defer sbt.mutPendingSelfHeaders.Unlock()

	current, exists := sbt.pendingSelfHeaders[key]
	if !exists || current != pending || current.sourcesVersion != sourcesVersion || current.view != view ||
		!current.sourceOverflow || !sbt.isPendingSelfHeadersViewCurrent(view) {
		return false
	}
	if !current.lastOverflowScan.IsZero() && now.Sub(current.lastOverflowScan) < requestInterval {
		return false
	}

	current.lastOverflowScan = now
	return true
}

func clonePendingMetaSources(sources []pendingMetaSource) []pendingMetaSource {
	cloned := make([]pendingMetaSource, len(sources))
	for index, source := range sources {
		cloned[index] = source
		cloned[index].hash = append([]byte(nil), source.hash...)
	}

	return cloned
}

func (sbt *shardBlockTrack) pendingSelfHeaderMatches(pending *pendingSelfHeader, header data.HeaderHandler) bool {
	return header.GetShardID() == pending.shardID && header.GetNonce() == pending.nonce && header.GetEpoch() == pending.epoch
}

func (sbt *shardBlockTrack) sourceMetaVerdict(
	sourceMetaHeader data.MetaHeaderHandler,
	sourceMetaHash []byte,
	sourceMetaNonce uint64,
	snapshot metaAuthoritySnapshot,
	crossedByCanonicalAnchor bool,
) sourceMetaVerdict {
	if !check.IfNil(sourceMetaHeader) && len(sourceMetaHash) == 0 {
		if sourceMetaHeader.IsHeaderV3() {
			return sourceMetaUnknown
		}
		return sourceMetaHeldFinal
	}
	if check.IfNil(sourceMetaHeader) || sourceMetaHeader.GetNonce() != sourceMetaNonce {
		return sourceMetaUnknown
	}
	if sbt.metaFinalityView.IsDeadMetaBlock(sourceMetaHash, sourceMetaNonce) {
		return sourceMetaDead
	}
	if snapshot.valid && sbt.metaFinalityView.IsMetaHeaderSettlementReady(sourceMetaHeader, sourceMetaHash) {
		_, inCurrentSnapshot := snapshot.hashes[string(sourceMetaHash)]
		if crossedByCanonicalAnchor || inCurrentSnapshot {
			return sourceMetaHeldFinal
		}
	}

	return sourceMetaUnknown
}

func (sbt *shardBlockTrack) currentMetaAuthoritySnapshot() metaAuthoritySnapshot {
	snapshot := metaAuthoritySnapshot{hashes: make(map[string]struct{})}
	anchor, anchorHash, err := sbt.GetLastCrossNotarizedHeader(core.MetachainShardId)
	if err != nil || check.IfNil(anchor) || len(anchorHash) == 0 ||
		sbt.metaFinalityView.IsDeadMetaBlock(anchorHash, anchor.GetNonce()) {
		return snapshot
	}
	snapshot.valid = true

	snapshot.hashes[string(anchorHash)] = struct{}{}
	if meta, ok := anchor.(data.MetaHeaderHandler); ok {
		snapshot.sources = append(snapshot.sources, pendingMetaSource{header: meta, hash: anchorHash, nonce: anchor.GetNonce()})
	}
	continuation, hashes := sbt.ComputeLongestChain(core.MetachainShardId, anchor)
	for index, header := range continuation {
		if index >= len(hashes) || check.IfNil(header) ||
			sbt.metaFinalityView.IsDeadMetaBlock(hashes[index], header.GetNonce()) {
			break
		}
		snapshot.hashes[string(hashes[index])] = struct{}{}
		if meta, ok := header.(data.MetaHeaderHandler); ok {
			snapshot.sources = append(snapshot.sources, pendingMetaSource{header: meta, hash: hashes[index], nonce: header.GetNonce()})
		}
	}

	current := anchor
	for scanned := 0; scanned < maxMetaBlocksScannedForInclusion && current.GetNonce() > 0; scanned++ {
		parentHash := current.GetPrevHash()
		parent := sbt.getTrackedOrPooledMetaHeader(parentHash, current.GetNonce()-1)
		if check.IfNil(parent) || parent.GetShardID() != core.MetachainShardId ||
			parent.GetNonce()+1 != current.GetNonce() {
			break
		}
		snapshot.hashes[string(parentHash)] = struct{}{}
		if meta, ok := parent.(data.MetaHeaderHandler); ok {
			snapshot.sources = append(snapshot.sources, pendingMetaSource{header: meta, hash: parentHash, nonce: parent.GetNonce()})
		}
		current = parent
	}

	return snapshot
}

func (sbt *shardBlockTrack) snapshotAuthorizesPending(snapshot metaAuthoritySnapshot, pending *pendingSelfHeader) bool {
	for _, source := range snapshot.sources {
		if sbt.metaFinalityView.IsMetaHeaderSettlementReady(source.header, source.hash) &&
			sbt.metaFinalityView.IsShardHeaderIncluded(source.header, pending.shardID, pending.hash, pending.nonce) {
			return true
		}
	}

	return false
}

// AddCrossNotarizedHeader invalidates pending snapshots when the meta anchor identity changes.
func (sbt *shardBlockTrack) AddCrossNotarizedHeader(
	shardID uint32,
	crossNotarizedHeader data.HeaderHandler,
	crossNotarizedHeaderHash []byte,
) {
	if shardID != core.MetachainShardId || sbt.numPendingSelfHeaders.Load() == 0 || check.IfNil(crossNotarizedHeader) {
		sbt.baseBlockTrack.AddCrossNotarizedHeader(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
		return
	}

	sbt.mutPendingNotifications.Lock()
	current, currentHash, err := sbt.GetLastCrossNotarizedHeader(core.MetachainShardId)
	identityChanged := err != nil || check.IfNil(current) ||
		current.GetNonce() != crossNotarizedHeader.GetNonce() || !bytes.Equal(currentHash, crossNotarizedHeaderHash)
	if !identityChanged {
		sbt.baseBlockTrack.AddCrossNotarizedHeader(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
		sbt.mutPendingNotifications.Unlock()
		return
	}

	sbt.pendingSelfHeadersView.Add(1)
	sbt.pendingNotifications.Wait()
	sourceHashes := sbt.pendingSourceHashes()
	crossedSources, connected := sbt.crossedPendingSources(
		current,
		currentHash,
		crossNotarizedHeader,
		crossNotarizedHeaderHash,
		sourceHashes,
	)
	sbt.baseBlockTrack.AddCrossNotarizedHeader(shardID, crossNotarizedHeader, crossNotarizedHeaderHash)
	sbt.mutPendingSelfHeaders.Lock()
	newView := sbt.pendingSelfHeadersView.Load() + 1
	for _, pending := range sbt.pendingSelfHeaders {
		for index := range pending.sources {
			if !connected {
				pending.sources[index].crossedByCanonicalAnchor = false
			}
			if _, crossed := crossedSources[string(pending.sources[index].hash)]; crossed {
				pending.sources[index].crossedByCanonicalAnchor = true
			}
		}
		pending.view = newView
		pending.sourcesVersion++
	}
	sbt.pendingSelfHeadersView.Add(1)
	sbt.mutPendingSelfHeaders.Unlock()
	sbt.mutPendingNotifications.Unlock()
}

func (sbt *shardBlockTrack) pendingSourceHashes() map[string]struct{} {
	sbt.mutPendingSelfHeaders.Lock()
	defer sbt.mutPendingSelfHeaders.Unlock()

	hashes := make(map[string]struct{})
	for _, pending := range sbt.pendingSelfHeaders {
		for _, source := range pending.sources {
			if len(source.hash) > 0 {
				hashes[string(source.hash)] = struct{}{}
			}
		}
	}

	return hashes
}

func (sbt *shardBlockTrack) crossedPendingSources(
	current data.HeaderHandler,
	currentHash []byte,
	next data.HeaderHandler,
	nextHash []byte,
	pendingHashes map[string]struct{},
) (map[string]struct{}, bool) {
	crossed := make(map[string]struct{})
	if check.IfNil(current) || check.IfNil(next) || len(currentHash) == 0 || len(nextHash) == 0 ||
		next.GetNonce() <= current.GetNonce() || next.GetShardID() != core.MetachainShardId {
		return crossed, false
	}

	header := next
	hash := nextHash
	for scanned := 0; scanned <= sbt.maxNumHeadersToKeepPerShard; scanned++ {
		if _, tracked := pendingHashes[string(hash)]; tracked {
			crossed[string(hash)] = struct{}{}
		}
		if header.GetNonce() == current.GetNonce() {
			if !bytes.Equal(hash, currentHash) {
				return nil, false
			}

			return crossed, true
		}
		if header.GetNonce() <= current.GetNonce() || header.GetNonce() == 0 {
			break
		}

		hash = header.GetPrevHash()
		header = sbt.getTrackedOrPooledMetaHeader(hash, header.GetNonce()-1)
		if check.IfNil(header) || header.GetShardID() != core.MetachainShardId {
			break
		}
	}

	return nil, false
}

func (sbt *shardBlockTrack) preparePendingSourceRequests(
	key string,
	pending *pendingSelfHeader,
	sourcesVersion uint64,
	unknownSources []pendingMetaSource,
	canonicalSources map[string]struct{},
) []pendingSourceRequest {
	if len(unknownSources) == 0 {
		return nil
	}

	now := time.Now()
	requestInterval := sbt.requestHandler.RequestInterval()
	sbt.mutPendingSelfHeaders.Lock()
	defer sbt.mutPendingSelfHeaders.Unlock()

	current, exists := sbt.pendingSelfHeaders[key]
	if !exists || current != pending || current.sourcesVersion != sourcesVersion ||
		!sbt.isPendingSelfHeadersViewCurrent(current.view) {
		return nil
	}
	if !current.lastSourceRequest.IsZero() && now.Sub(current.lastSourceRequest) < requestInterval {
		return nil
	}

	for _, preferCanonical := range []bool{true, false} {
		for offset := 0; offset < len(current.sources); offset++ {
			index := (current.nextSourceToAsk + offset) % len(current.sources)
			source := &current.sources[index]
			_, inCurrentSnapshot := canonicalSources[string(source.hash)]
			isCanonical := source.crossedByCanonicalAnchor || inCurrentSnapshot
			if isCanonical != preferCanonical || !containsPendingMetaSource(unknownSources, source) ||
				check.IfNil(source.header) || len(source.hash) == 0 {
				continue
			}

			current.lastSourceRequest = now
			current.nextSourceToAsk = (index + 1) % len(current.sources)
			return []pendingSourceRequest{{
				hash:  append([]byte(nil), source.hash...),
				epoch: source.header.GetEpoch(),
			}}
		}
	}

	return nil
}

func containsPendingMetaSource(sources []pendingMetaSource, candidate *pendingMetaSource) bool {
	for _, source := range sources {
		if source.nonce == candidate.nonce && bytes.Equal(source.hash, candidate.hash) {
			return true
		}
	}

	return false
}

func (sbt *shardBlockTrack) requestPendingSources(requests []pendingSourceRequest) {
	for _, request := range requests {
		sbt.requestHandler.RequestMetaHeaderForEpoch(request.hash, request.epoch)
		sbt.requestHandler.RequestEquivalentProofByHashForEpoch(core.MetachainShardId, request.hash, request.epoch)
	}
}

func (sbt *shardBlockTrack) retryPendingSelfHeaders() {
	if sbt.numPendingSelfHeaders.Load() == 0 {
		return
	}

	sbt.mutPendingSelfHeaders.Lock()
	keys := make([]string, 0, len(sbt.pendingSelfHeaders))
	for key, pending := range sbt.pendingSelfHeaders {
		if !check.IfNil(pending.deliveredHeader) {
			keys = append(keys, key)
		}
	}
	sbt.mutPendingSelfHeaders.Unlock()
	if len(keys) == 0 {
		return
	}

	snapshotView := sbt.pendingSelfHeadersView.Load()
	snapshot := sbt.currentMetaAuthoritySnapshot()
	if !sbt.isPendingSelfHeadersViewCurrent(snapshotView) {
		return
	}

	for _, key := range keys {
		sbt.resolvePendingSelfHeaderWithSnapshot(key, nil, &snapshot, snapshotView, nil)
	}
}

func (sbt *shardBlockTrack) publishSelfNotarizedFromCrossHeaders(shardID uint32, headersInfo []*selfHeaderInfo) {
	hasV3Source := false
	for _, headerInfo := range headersInfo {
		if !check.IfNil(headerInfo.sourceMetaHeader) && headerInfo.sourceMetaHeader.IsHeaderV3() {
			hasV3Source = true
			break
		}
	}
	if !hasV3Source {
		headers, hashes := selfHeaderInfoToSlices(headersInfo)
		sbt.selfNotarizedFromCrossHeadersNotifier.CallHandlers(shardID, headers, hashes)
		return
	}

	sbt.mutPendingNotifications.Lock()
	sbt.pendingNotifications.Wait()
	snapshot := sbt.currentMetaAuthoritySnapshot()

	admitted := make([]*selfHeaderInfo, 0, len(headersInfo))
	retainedKeys := make([]string, 0)
	for _, headerInfo := range headersInfo {
		if check.IfNil(headerInfo.sourceMetaHeader) || !headerInfo.sourceMetaHeader.IsHeaderV3() {
			admitted = append(admitted, headerInfo)
			continue
		}
		if len(headerInfo.sourceMetaHash) == 0 {
			continue
		}
		if !sbt.isPendingSelfHeadersViewCurrent(headerInfo.sourceView) {
			continue
		}
		verdict := sbt.sourceMetaVerdict(
			headerInfo.sourceMetaHeader,
			headerInfo.sourceMetaHash,
			headerInfo.sourceMetaHeader.GetNonce(),
			snapshot,
			false,
		)
		if verdict == sourceMetaHeldFinal {
			sbt.mutPendingSelfHeaders.Lock()
			key := string(headerInfo.Hash)
			_, wasResolved := sbt.resolvedSelfHeaders[key]
			if !wasResolved {
				sbt.rememberResolvedSelfHeader(key, headerInfo.Header.GetNonce())
			}
			sbt.mutPendingSelfHeaders.Unlock()
			if wasResolved {
				continue
			}
			admitted = append(admitted, headerInfo)
			continue
		}
		if verdict == sourceMetaUnknown {
			if key := sbt.retainDeliveredSelfHeader(headerInfo); key != "" {
				retainedKeys = append(retainedKeys, key)
			}
		}
	}

	if len(admitted) > 0 {
		headers, hashes := selfHeaderInfoToSlices(admitted)
		sbt.pendingNotifications.Add(1)
		go func() {
			defer sbt.pendingNotifications.Done()
			sbt.selfNotarizedFromCrossHeadersNotifier.callHandlersAndWait(shardID, headers, hashes)
		}()
	}
	sbt.mutPendingNotifications.Unlock()

	for _, key := range retainedKeys {
		sbt.resolvePendingSelfHeader(key, nil)
	}
}

func (sbt *shardBlockTrack) retainDeliveredSelfHeader(headerInfo *selfHeaderInfo) string {
	if check.IfNil(headerInfo.Header) || check.IfNil(headerInfo.sourceMetaHeader) ||
		len(headerInfo.Hash) == 0 || len(headerInfo.sourceMetaHash) == 0 {
		return ""
	}

	key := string(headerInfo.Hash)
	sbt.mutPendingSelfHeaders.Lock()
	defer sbt.mutPendingSelfHeaders.Unlock()

	if !sbt.isPendingSelfHeadersViewCurrent(headerInfo.sourceView) {
		return ""
	}
	if _, resolved := sbt.resolvedSelfHeaders[key]; resolved {
		return ""
	}
	if pending, exists := sbt.pendingSelfHeaders[key]; exists {
		sbt.updatePendingSource(pending, headerInfo.sourceMetaHeader, headerInfo.sourceMetaHash)
		pending.deliveredHeader = headerInfo.Header
		return key
	}

	pending := &pendingSelfHeader{
		hash:            append([]byte(nil), headerInfo.Hash...),
		shardID:         headerInfo.Header.GetShardID(),
		nonce:           headerInfo.Header.GetNonce(),
		epoch:           headerInfo.Header.GetEpoch(),
		view:            headerInfo.sourceView,
		deliveredHeader: headerInfo.Header,
	}
	sbt.setInitialPendingSource(pending, headerInfo.sourceMetaHeader, headerInfo.sourceMetaHash)
	hasRoom, evicted := sbt.makeRoomForPendingSelfHeader(pending.nonce)
	if !hasRoom {
		return ""
	}

	sbt.pendingSelfHeaders[key] = pending
	sbt.numPendingSelfHeaders.Store(int64(len(sbt.pendingSelfHeaders)))
	if evicted != nil {
		log.Warn("evicted unresolved held-final shard header", "nonce", evicted.nonce, "hash", evicted.hash)
	}

	return key
}

func (sbt *shardBlockTrack) isPendingSelfHeadersViewCurrent(view uint64) bool {
	// Odd views mark a reset in progress.
	return view%2 == 0 && view == sbt.pendingSelfHeadersView.Load()
}

func (sbt *shardBlockTrack) notifyPendingSelfHeader(header data.HeaderHandler, pending *pendingSelfHeader) {
	log.Debug("resolved held-final shard header", "nonce", pending.nonce, "hash", pending.hash)
	sbt.pendingNotifications.Add(1)
	go func() {
		defer sbt.pendingNotifications.Done()

		sbt.selfNotarizedFromCrossHeadersNotifier.callHandlersAndWait(
			core.MetachainShardId,
			[]data.HeaderHandler{header},
			[][]byte{pending.hash},
		)
	}()
}

// RemoveLastNotarizedHeaders removes the reverted tracker checkpoints and their pending source.
func (sbt *shardBlockTrack) RemoveLastNotarizedHeaders() {
	sbt.mutPendingNotifications.Lock()
	defer sbt.mutPendingNotifications.Unlock()

	var removedMetaHash []byte
	_, _, err := sbt.crossNotarizer.GetNotarizedHeader(core.MetachainShardId, 1)
	if err != nil {
		sbt.pendingSelfHeadersView.Add(1)
		sbt.pendingNotifications.Wait()
		sbt.mutPendingSelfHeaders.Lock()
		newView := sbt.pendingSelfHeadersView.Load() + 1
		for _, pending := range sbt.pendingSelfHeaders {
			pending.view = newView
			pending.sourcesVersion++
		}
		sbt.pendingSelfHeadersView.Add(1)
		sbt.mutPendingSelfHeaders.Unlock()
		sbt.baseBlockTrack.RemoveLastNotarizedHeaders()
		return
	}
	_, removedMetaHash, _ = sbt.crossNotarizer.GetLastNotarizedHeader(core.MetachainShardId)

	sbt.pendingSelfHeadersView.Add(1)
	sbt.pendingNotifications.Wait()

	sbt.baseBlockTrack.RemoveLastNotarizedHeaders()
	_, currentMetaHash, currentErr := sbt.crossNotarizer.GetLastNotarizedHeader(core.MetachainShardId)
	if currentErr == nil && bytes.Equal(currentMetaHash, removedMetaHash) {
		removedMetaHash = nil
	}

	sbt.mutPendingSelfHeaders.Lock()
	newView := sbt.pendingSelfHeadersView.Load() + 1
	for key, pending := range sbt.pendingSelfHeaders {
		retainedSources := pending.sources[:0]
		for _, source := range pending.sources {
			if len(removedMetaHash) == 0 || !bytes.Equal(source.hash, removedMetaHash) {
				retainedSources = append(retainedSources, source)
			}
		}
		if len(retainedSources) == 0 && !pending.sourceOverflow {
			delete(sbt.pendingSelfHeaders, key)
			continue
		}
		pending.sources = retainedSources
		pending.lastSourceRequest = time.Time{}
		pending.sourcesVersion++
		pending.view = newView
	}
	sbt.numPendingSelfHeaders.Store(int64(len(sbt.pendingSelfHeaders)))
	sbt.pendingSelfHeadersView.Add(1)
	sbt.mutPendingSelfHeaders.Unlock()
}

func (sbt *shardBlockTrack) rememberResolvedSelfHeader(key string, nonce uint64) {
	if sbt.maxNumHeadersToKeepPerShard <= 0 {
		return
	}
	if _, exists := sbt.resolvedSelfHeaders[key]; exists {
		sbt.resolvedSelfHeaders[key] = nonce
		sbt.numResolvedSelfHeaders.Store(int64(len(sbt.resolvedSelfHeaders)))
		return
	}

	if len(sbt.resolvedSelfHeaders) >= sbt.maxNumHeadersToKeepPerShard {
		oldestKey := key
		oldestNonce := nonce
		for resolvedKey, resolvedNonce := range sbt.resolvedSelfHeaders {
			isOlder := resolvedNonce < oldestNonce
			sameNonceLowerHash := resolvedNonce == oldestNonce && resolvedKey < oldestKey
			if isOlder || sameNonceLowerHash {
				oldestKey = resolvedKey
				oldestNonce = resolvedNonce
			}
		}
		if oldestKey == key {
			return
		}
		delete(sbt.resolvedSelfHeaders, oldestKey)
	}

	sbt.resolvedSelfHeaders[key] = nonce
	sbt.numResolvedSelfHeaders.Store(int64(len(sbt.resolvedSelfHeaders)))
}

// RestoreToGenesis resets all tracked shard state.
func (sbt *shardBlockTrack) RestoreToGenesis() {
	sbt.mutPendingNotifications.Lock()
	defer sbt.mutPendingNotifications.Unlock()
	sbt.pendingSelfHeadersView.Add(1)
	sbt.pendingNotifications.Wait()

	sbt.baseBlockTrack.RestoreToGenesis()

	sbt.mutPendingSelfHeaders.Lock()
	sbt.pendingSelfHeaders = make(map[string]*pendingSelfHeader)
	sbt.resolvedSelfHeaders = make(map[string]uint64)
	sbt.numPendingSelfHeaders.Store(0)
	sbt.numResolvedSelfHeaders.Store(0)
	sbt.pendingSelfHeadersView.Add(1)
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
