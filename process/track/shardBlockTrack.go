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
	metaFinalityView process.MetaFinalityView

	mutPendingSelfHeaders   sync.Mutex
	mutPendingNotifications sync.Mutex
	pendingNotifications    sync.WaitGroup
	pendingSelfHeadersView  atomic.Uint64
	pendingSelfHeaders      map[string]*pendingSelfHeader
	resolvedSelfHeaders     map[string]uint64
	publishedSelfHeaders    map[string]*publishedSelfHeader
	publishedBySourceNonce  map[uint64]map[string]*publishedSelfHeader
	numPendingSelfHeaders   atomic.Int64
	numResolvedSelfHeaders  atomic.Int64
	numPublishedSelfHeaders atomic.Int64

	invalidatedSelfNotarizedFromCrossHeadersNotifier *blockNotifier
}

type pendingSelfHeader struct {
	hash            []byte
	shardID         uint32
	nonce           uint64
	epoch           uint32
	lastRequest     time.Time
	view            uint64
	sourcesVersion  uint64
	sources         []pendingMetaSource
	deliveredHeader data.HeaderHandler
}

type pendingMetaSource struct {
	header      data.MetaHeaderHandler
	hash        []byte
	nonce       uint64
	lastRequest time.Time
}

type pendingSourceRequest struct {
	hash  []byte
	epoch uint32
}

type publishedSelfHeader struct {
	header  data.HeaderHandler
	hash    []byte
	nonce   uint64
	sources []pendingMetaSource
}

type sourceMetaVerdict uint8

const (
	sourceMetaUnknown sourceMetaVerdict = iota
	sourceMetaHeldFinal
	sourceMetaDead
)

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
	invalidatedNotifier, err := NewBlockNotifier()
	if err != nil {
		return nil, err
	}

	err = bbt.initNotarizedHeaders(arguments.StartHeaders)
	if err != nil {
		return nil, err
	}

	sbt := shardBlockTrack{
		baseBlockTrack:         bbt,
		metaFinalityView:       metaFinalityView,
		pendingSelfHeaders:     make(map[string]*pendingSelfHeader),
		resolvedSelfHeaders:    make(map[string]uint64),
		publishedSelfHeaders:   make(map[string]*publishedSelfHeader),
		publishedBySourceNonce: make(map[uint64]map[string]*publishedSelfHeader),
		invalidatedSelfNotarizedFromCrossHeadersNotifier: invalidatedNotifier,
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
		sbt.invalidatePublishedSelfHeadersFromDeadSources(header.GetNonce())
	}
}

func (sbt *shardBlockTrack) receivedProof(proof data.HeaderProofHandler) {
	sbt.baseBlockTrack.receivedProof(proof)
	if !check.IfNil(proof) && proof.GetHeaderShardId() == core.MetachainShardId {
		sbt.retryPendingSelfHeaders()
		sbt.invalidatePublishedSelfHeadersFromDeadSources(proof.GetHeaderNonce())
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
	if sbt.numPendingSelfHeaders.Load() == 0 && sbt.numResolvedSelfHeaders.Load() == 0 &&
		sbt.numPublishedSelfHeaders.Load() == 0 {
		return nil, false
	}

	key := string(shardInfo.GetHeaderHash())
	sbt.mutPendingSelfHeaders.Lock()
	if !sbt.isPendingSelfHeadersViewCurrent(pendingView) {
		sbt.mutPendingSelfHeaders.Unlock()
		return nil, true
	}
	_, resolved := sbt.resolvedSelfHeaders[key]
	published := sbt.publishedSelfHeaders[key]
	pendingHeader, pending := sbt.pendingSelfHeaders[key]
	pending = pending && pendingHeader.view == pendingView
	if pending {
		sbt.updatePendingSource(pendingHeader, sourceMetaHeader, sourceMetaHash)
	}
	sbt.mutPendingSelfHeaders.Unlock()
	if published != nil && !check.IfNil(published.header) {
		return sbt.newSelfHeaderInfo(
			published.header,
			shardInfo.GetHeaderHash(),
			sourceMetaHeader,
			sourceMetaHash,
			pendingView,
		), true
	}
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
) {
	updatedSources, changed := addOrUpdateMetaSource(pending.sources, sourceMetaHeader, sourceMetaHash)
	pending.sources = updatedSources
	if changed {
		pending.sourcesVersion++
	}
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
	if len(sources) > 0 && sourceNonce < sources[0].nonce {
		return sources, false
	}
	if len(sources) > 0 && sourceNonce > sources[0].nonce {
		clear(sources)
		sources = sources[:0]
	}
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
	view := pending.view
	sbt.mutPendingSelfHeaders.Unlock()

	hasHeldFinalSource := false
	var unknownSources []pendingMetaSource
	for _, source := range sources {
		verdict := sbt.sourceMetaVerdict(source.header, source.hash, source.nonce)
		if verdict == sourceMetaHeldFinal {
			hasHeldFinalSource = true
			break
		}
		if verdict == sourceMetaUnknown {
			unknownSources = append(unknownSources, source)
		}
	}
	if !hasHeldFinalSource && (len(sources) == 0 || len(unknownSources) > 0) {
		sbt.mutPendingNotifications.Unlock()
		requests := sbt.preparePendingSourceRequests(key, pending, sourcesVersion, unknownSources)
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
		sbt.rememberPublishedSelfHeaderLocked(header, pending.hash, sources)
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
) sourceMetaVerdict {
	if !check.IfNil(sourceMetaHeader) && len(sourceMetaHash) == 0 {
		return sourceMetaHeldFinal
	}
	if check.IfNil(sourceMetaHeader) || sourceMetaHeader.GetNonce() != sourceMetaNonce {
		return sourceMetaUnknown
	}
	if sbt.metaFinalityView.IsDeadMetaBlock(sourceMetaHash, sourceMetaNonce) {
		return sourceMetaDead
	}
	if sbt.metaFinalityView.IsMetaHeaderHeldFinal(sourceMetaHeader, sourceMetaHash) {
		return sourceMetaHeldFinal
	}

	return sourceMetaUnknown
}

func (sbt *shardBlockTrack) preparePendingSourceRequests(
	key string,
	pending *pendingSelfHeader,
	sourcesVersion uint64,
	unknownSources []pendingMetaSource,
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

	requests := make([]pendingSourceRequest, 0, len(unknownSources))
	for _, unknownSource := range unknownSources {
		for index := range current.sources {
			source := &current.sources[index]
			if source.nonce != unknownSource.nonce || !bytes.Equal(source.hash, unknownSource.hash) ||
				check.IfNil(source.header) || len(source.hash) == 0 || now.Sub(source.lastRequest) < requestInterval {
				continue
			}

			source.lastRequest = now
			requests = append(requests, pendingSourceRequest{
				hash:  append([]byte(nil), source.hash...),
				epoch: source.header.GetEpoch(),
			})
			break
		}
	}

	return requests
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

	for _, key := range keys {
		sbt.resolvePendingSelfHeader(key, nil)
	}
}

func (sbt *shardBlockTrack) publishSelfNotarizedFromCrossHeaders(shardID uint32, headersInfo []*selfHeaderInfo) {
	hasSourceAwareHeader := false
	for _, headerInfo := range headersInfo {
		if !check.IfNil(headerInfo.sourceMetaHeader) && len(headerInfo.sourceMetaHash) > 0 {
			hasSourceAwareHeader = true
			break
		}
	}
	if !hasSourceAwareHeader {
		headers, hashes := selfHeaderInfoToSlices(headersInfo)
		sbt.selfNotarizedFromCrossHeadersNotifier.CallHandlers(shardID, headers, hashes)
		return
	}

	sbt.mutPendingNotifications.Lock()
	sbt.pendingNotifications.Wait()

	admitted := make([]*selfHeaderInfo, 0, len(headersInfo))
	retainedKeys := make([]string, 0)
	for _, headerInfo := range headersInfo {
		if check.IfNil(headerInfo.sourceMetaHeader) || len(headerInfo.sourceMetaHash) == 0 {
			admitted = append(admitted, headerInfo)
			continue
		}
		if !sbt.isPendingSelfHeadersViewCurrent(headerInfo.sourceView) {
			continue
		}
		verdict := sbt.sourceMetaVerdict(
			headerInfo.sourceMetaHeader,
			headerInfo.sourceMetaHash,
			headerInfo.sourceMetaHeader.GetNonce(),
		)
		if verdict == sourceMetaHeldFinal {
			sbt.mutPendingSelfHeaders.Lock()
			_, wasPublished := sbt.publishedSelfHeaders[string(headerInfo.Hash)]
			sbt.rememberPublishedSelfHeaderLocked(
				headerInfo.Header,
				headerInfo.Hash,
				[]pendingMetaSource{{
					header: headerInfo.sourceMetaHeader,
					hash:   append([]byte(nil), headerInfo.sourceMetaHash...),
					nonce:  headerInfo.sourceMetaHeader.GetNonce(),
				}},
			)
			sbt.mutPendingSelfHeaders.Unlock()
			if wasPublished {
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

// RemoveLastNotarizedHeaders removes the reverted tracker checkpoints and invalidates pending V3 claims.
func (sbt *shardBlockTrack) RemoveLastNotarizedHeaders() {
	sbt.mutPendingNotifications.Lock()
	defer sbt.mutPendingNotifications.Unlock()
	sbt.pendingSelfHeadersView.Add(1)
	sbt.pendingNotifications.Wait()

	sbt.mutPendingSelfHeaders.Lock()
	defer sbt.mutPendingSelfHeaders.Unlock()

	sbt.baseBlockTrack.RemoveLastNotarizedHeaders()
	sbt.pendingSelfHeaders = make(map[string]*pendingSelfHeader)
	sbt.resolvedSelfHeaders = make(map[string]uint64)
	sbt.numPendingSelfHeaders.Store(0)
	sbt.numResolvedSelfHeaders.Store(0)
	sbt.pendingSelfHeadersView.Add(1)
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

func (sbt *shardBlockTrack) rememberPublishedSelfHeaderLocked(
	header data.HeaderHandler,
	hash []byte,
	sources []pendingMetaSource,
) {
	if check.IfNil(header) || !header.IsHeaderV3() || len(hash) == 0 || len(sources) == 0 ||
		sbt.maxNumHeadersToKeepPerShard <= 0 {
		return
	}

	key := string(hash)
	published, exists := sbt.publishedSelfHeaders[key]
	if exists {
		oldSourceNonce := published.sources[0].nonce
		for _, source := range sources {
			sbt.updatePublishedSource(published, source)
		}
		if published.sources[0].nonce != oldSourceNonce {
			sbt.removePublishedSourceIndex(key, oldSourceNonce)
			sbt.addPublishedSourceIndex(key, published)
		}
		return
	}

	if len(sbt.publishedSelfHeaders) >= sbt.maxNumHeadersToKeepPerShard {
		oldestKey := key
		oldestNonce := header.GetNonce()
		for publishedKey, publishedHeader := range sbt.publishedSelfHeaders {
			if publishedHeader.nonce < oldestNonce ||
				(publishedHeader.nonce == oldestNonce && publishedKey < oldestKey) {
				oldestKey = publishedKey
				oldestNonce = publishedHeader.nonce
			}
		}
		if oldestKey == key {
			return
		}
		oldest := sbt.publishedSelfHeaders[oldestKey]
		sbt.removePublishedSourceIndex(oldestKey, oldest.sources[0].nonce)
		delete(sbt.publishedSelfHeaders, oldestKey)
	}

	published = &publishedSelfHeader{
		header: header,
		hash:   append([]byte(nil), hash...),
		nonce:  header.GetNonce(),
	}
	for _, source := range sources {
		sbt.updatePublishedSource(published, source)
	}
	if len(published.sources) == 0 {
		return
	}

	sbt.publishedSelfHeaders[key] = published
	sbt.addPublishedSourceIndex(key, published)
	sbt.numPublishedSelfHeaders.Store(int64(len(sbt.publishedSelfHeaders)))
}

func (sbt *shardBlockTrack) addPublishedSourceIndex(key string, published *publishedSelfHeader) {
	if len(published.sources) == 0 {
		return
	}

	sourceNonce := published.sources[0].nonce
	indexed := sbt.publishedBySourceNonce[sourceNonce]
	if indexed == nil {
		indexed = make(map[string]*publishedSelfHeader)
		sbt.publishedBySourceNonce[sourceNonce] = indexed
	}
	indexed[key] = published
}

func (sbt *shardBlockTrack) removePublishedSourceIndex(key string, sourceNonce uint64) {
	indexed := sbt.publishedBySourceNonce[sourceNonce]
	delete(indexed, key)
	if len(indexed) == 0 {
		delete(sbt.publishedBySourceNonce, sourceNonce)
	}
}

func (sbt *shardBlockTrack) updatePublishedSource(published *publishedSelfHeader, source pendingMetaSource) {
	published.sources, _ = addOrUpdateMetaSource(published.sources, source.header, source.hash)
}

func (sbt *shardBlockTrack) invalidatePublishedSelfHeadersFromDeadSources(evidenceNonce uint64) {
	if sbt.numPublishedSelfHeaders.Load() == 0 {
		return
	}

	sbt.mutPendingNotifications.Lock()
	sbt.pendingNotifications.Wait()

	sbt.mutPendingSelfHeaders.Lock()
	publishedHeaders := make([]*publishedSelfHeader, 0)
	for depth := uint64(1); depth <= metaReconciliationEvidenceDepth && evidenceNonce >= depth; depth++ {
		for _, published := range sbt.publishedBySourceNonce[evidenceNonce-depth] {
			publishedHeaders = append(publishedHeaders, published)
		}
	}
	sbt.mutPendingSelfHeaders.Unlock()

	invalidated := make([]*publishedSelfHeader, 0, len(publishedHeaders))
	for _, published := range publishedHeaders {
		allSourcesDead := len(published.sources) > 0
		for _, source := range published.sources {
			if !sbt.metaFinalityView.IsDeadMetaBlock(source.hash, source.nonce) {
				allSourcesDead = false
				break
			}
		}
		if allSourcesDead {
			invalidated = append(invalidated, published)
		}
	}

	if len(invalidated) == 0 {
		sbt.mutPendingNotifications.Unlock()
		return
	}

	headers := make([]data.HeaderHandler, 0, len(invalidated))
	hashes := make([][]byte, 0, len(invalidated))
	sbt.mutPendingSelfHeaders.Lock()
	for _, published := range invalidated {
		key := string(published.hash)
		if sbt.publishedSelfHeaders[key] != published {
			continue
		}

		sbt.removePublishedSourceIndex(key, published.sources[0].nonce)
		delete(sbt.publishedSelfHeaders, key)
		delete(sbt.resolvedSelfHeaders, key)
		headers = append(headers, published.header)
		hashes = append(hashes, published.hash)
	}
	sbt.numPublishedSelfHeaders.Store(int64(len(sbt.publishedSelfHeaders)))
	sbt.numResolvedSelfHeaders.Store(int64(len(sbt.resolvedSelfHeaders)))
	sbt.mutPendingSelfHeaders.Unlock()

	if len(headers) > 0 {
		sbt.pendingNotifications.Add(1)
		go func() {
			defer sbt.pendingNotifications.Done()
			sbt.invalidatedSelfNotarizedFromCrossHeadersNotifier.callHandlersAndWait(
				core.MetachainShardId,
				headers,
				hashes,
			)
		}()
	}
	sbt.mutPendingNotifications.Unlock()
}

// RegisterInvalidatedSelfNotarizedFromCrossHeadersHandler registers a handler for revoked V3 meta authority.
func (sbt *shardBlockTrack) RegisterInvalidatedSelfNotarizedFromCrossHeadersHandler(
	handler func(shardID uint32, headers []data.HeaderHandler, headersHashes [][]byte),
) {
	sbt.invalidatedSelfNotarizedFromCrossHeadersNotifier.RegisterHandler(handler)
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
	sbt.publishedSelfHeaders = make(map[string]*publishedSelfHeader)
	sbt.publishedBySourceNonce = make(map[uint64]map[string]*publishedSelfHeader)
	sbt.numPendingSelfHeaders.Store(0)
	sbt.numResolvedSelfHeaders.Store(0)
	sbt.numPublishedSelfHeaders.Store(0)
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
