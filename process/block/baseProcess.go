package block

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/block")

type hashAndHdr struct {
	hdr  data.HeaderHandler
	hash []byte
}

type nonceAndHashInfo struct {
	hash  []byte
	nonce uint64
}

type hdrInfo struct {
	usedInBlock bool
	hdr         data.HeaderHandler
}

type hdrForBlock struct {
	missingHdrs                  uint32
	missingFinalityAttestingHdrs uint32
	highestHdrNonce              map[uint32]uint64
	mutHdrsForBlock              sync.RWMutex
	hdrHashAndInfo               map[string]*hdrInfo
}

type baseProcessor struct {
	shardCoordinator             sharding.Coordinator
	nodesCoordinator             sharding.NodesCoordinator
	specialAddressHandler        process.SpecialAddressHandler
	accounts                     state.AccountsAdapter
	forkDetector                 process.ForkDetector
	validatorStatisticsProcessor process.ValidatorStatisticsProcessor
	hasher                       hashing.Hasher
	marshalizer                  marshal.Marshalizer
	store                        dataRetriever.StorageService
	uint64Converter              typeConverters.Uint64ByteSliceConverter
	blockSizeThrottler           process.BlockSizeThrottler
	epochStartTrigger            process.EpochStartTriggerHandler
	headerValidator              process.HeaderConstructionValidator
	blockChainHook               process.BlockChainHookHandler
	txCoordinator                process.TransactionCoordinator
	rounder                      consensus.Rounder
	bootStorer                   process.BootStorer
	requestBlockBodyHandler      process.RequestBlockBodyHandler
	requestHandler               process.RequestHandler
	blockTracker                 process.BlockTracker
	dataPool                     dataRetriever.PoolsHolder

	hdrsForCurrBlock hdrForBlock

	appStatusHandler core.AppStatusHandler
	blockProcessor   blockProcessor
}

func checkForNils(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	if chainHandler == nil || chainHandler.IsInterfaceNil() {
		return process.ErrNilBlockChain
	}
	if headerHandler == nil || headerHandler.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}
	if bodyHandler == nil || bodyHandler.IsInterfaceNil() {
		return process.ErrNilBlockBody
	}
	return nil
}

// SetAppStatusHandler method is used to set appStatusHandler
func (bp *baseProcessor) SetAppStatusHandler(ash core.AppStatusHandler) error {
	if ash == nil || ash.IsInterfaceNil() {
		return process.ErrNilAppStatusHandler
	}

	bp.appStatusHandler = ash
	return nil
}

// checkBlockValidity method checks if the given block is valid
func (bp *baseProcessor) checkBlockValidity(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	err := checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	currentBlockHeader := chainHandler.GetCurrentBlockHeader()

	if currentBlockHeader == nil {
		if headerHandler.GetNonce() == 1 { // first block after genesis
			if bytes.Equal(headerHandler.GetPrevHash(), chainHandler.GetGenesisHeaderHash()) {
				// TODO: add genesis block verification
				return nil
			}

			log.Debug("hash does not match",
				"local block hash", chainHandler.GetGenesisHeaderHash(),
				"received previous hash", headerHandler.GetPrevHash())

			return process.ErrBlockHashDoesNotMatch
		}

		log.Debug("nonce does not match",
			"local block nonce", 0,
			"received nonce", headerHandler.GetNonce())

		return process.ErrWrongNonceInBlock
	}

	if headerHandler.GetRound() <= currentBlockHeader.GetRound() {
		log.Debug("round does not match",
			"local block round", currentBlockHeader.GetRound(),
			"received block round", headerHandler.GetRound())

		return process.ErrLowerRoundInBlock
	}

	if headerHandler.GetNonce() != currentBlockHeader.GetNonce()+1 {
		log.Debug("nonce does not match",
			"local block nonce", currentBlockHeader.GetNonce(),
			"received nonce", headerHandler.GetNonce())

		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, currentBlockHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(headerHandler.GetPrevHash(), prevHeaderHash) {
		log.Debug("hash does not match",
			"local block hash", prevHeaderHash,
			"received previous hash", headerHandler.GetPrevHash())

		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(headerHandler.GetPrevRandSeed(), currentBlockHeader.GetRandSeed()) {
		log.Debug("random seed does not match",
			"local random seed", currentBlockHeader.GetRandSeed(),
			"received previous random seed", headerHandler.GetPrevRandSeed())

		return process.ErrRandSeedDoesNotMatch
	}

	if bodyHandler != nil {
		// TODO: add bodyHandler verification here
	}

	// verification of epoch
	if headerHandler.GetEpoch() < currentBlockHeader.GetEpoch() {
		return process.ErrEpochDoesNotMatch
	}

	// TODO: add signature validation as well, with randomness source and all
	return nil
}

// verifyStateRoot verifies the state root hash given as parameter against the
// Merkle trie root hash stored for accounts and returns if equal or not
func (bp *baseProcessor) verifyStateRoot(rootHash []byte) bool {
	trieRootHash, err := bp.accounts.RootHash()
	if err != nil {
		log.Debug("verify account.RootHash", "error", err.Error())
	}

	return bytes.Equal(trieRootHash, rootHash)
}

// getRootHash returns the accounts merkle tree root hash
func (bp *baseProcessor) getRootHash() []byte {
	rootHash, err := bp.accounts.RootHash()
	if err != nil {
		log.Trace("get account.RootHash", "error", err.Error())
	}

	return rootHash
}

func (bp *baseProcessor) requestHeadersIfMissing(
	sortedHdrs []data.HeaderHandler,
	shardId uint32,
	maxRound uint64,
) error {

	allowedSize := uint64(float64(bp.dataPool.Headers().MaxSize()) * process.MaxOccupancyPercentageAllowed)

	prevHdr, _, err := bp.blockTracker.GetLastCrossNotarizedHeader(shardId)
	if err != nil {
		return err
	}

	lastNotarizedHdrNonce := prevHdr.GetNonce()
	lastNotarizedHdrRound := prevHdr.GetRound()

	missingNonces := make([]uint64, 0)
	for i := 0; i < len(sortedHdrs); i++ {
		currHdr := sortedHdrs[i]
		if currHdr == nil {
			continue
		}

		hdrTooOld := currHdr.GetRound() <= lastNotarizedHdrRound
		if hdrTooOld {
			continue
		}

		hdrTooNew := currHdr.GetRound() > maxRound
		if hdrTooNew {
			break
		}

		if currHdr.GetNonce()-prevHdr.GetNonce() > 1 {
			for j := prevHdr.GetNonce() + 1; j < currHdr.GetNonce(); j++ {
				missingNonces = append(missingNonces, j)
			}
		}

		prevHdr = currHdr
	}

	requested := 0
	for _, nonce := range missingNonces {
		isHeaderOutOfRange := nonce > lastNotarizedHdrNonce+allowedSize
		if isHeaderOutOfRange {
			break
		}

		if requested >= process.MaxHeaderRequestsAllowed {
			break
		}

		requested++
		go bp.requestHeaderByShardAndNonce(shardId, nonce)
	}

	return nil
}

func displayHeader(headerHandler data.HeaderHandler) []*display.LineData {
	return []*display.LineData{
		display.NewLineData(false, []string{
			"",
			"ChainID",
			display.DisplayByteSlice(headerHandler.GetChainID())}),
		display.NewLineData(false, []string{
			"",
			"Epoch",
			fmt.Sprintf("%d", headerHandler.GetEpoch())}),
		display.NewLineData(false, []string{
			"",
			"Round",
			fmt.Sprintf("%d", headerHandler.GetRound())}),
		display.NewLineData(false, []string{
			"",
			"TimeStamp",
			fmt.Sprintf("%d", headerHandler.GetTimeStamp())}),
		display.NewLineData(false, []string{
			"",
			"Nonce",
			fmt.Sprintf("%d", headerHandler.GetNonce())}),
		display.NewLineData(false, []string{
			"",
			"Prev hash",
			display.DisplayByteSlice(headerHandler.GetPrevHash())}),
		display.NewLineData(false, []string{
			"",
			"Prev rand seed",
			display.DisplayByteSlice(headerHandler.GetPrevRandSeed())}),
		display.NewLineData(false, []string{
			"",
			"Rand seed",
			display.DisplayByteSlice(headerHandler.GetRandSeed())}),
		display.NewLineData(false, []string{
			"",
			"Pub keys bitmap",
			core.ToHex(headerHandler.GetPubKeysBitmap())}),
		display.NewLineData(false, []string{
			"",
			"Signature",
			display.DisplayByteSlice(headerHandler.GetSignature())}),
		display.NewLineData(false, []string{
			"",
			"Leader's Signature",
			display.DisplayByteSlice(headerHandler.GetLeaderSignature())}),
		display.NewLineData(false, []string{
			"",
			"Root hash",
			display.DisplayByteSlice(headerHandler.GetRootHash())}),
		display.NewLineData(false, []string{
			"",
			"Validator stats root hash",
			display.DisplayByteSlice(headerHandler.GetValidatorStatsRootHash())}),
		display.NewLineData(true, []string{
			"",
			"Receipts hash",
			display.DisplayByteSlice(headerHandler.GetReceiptsHash())}),
	}
}

// checkProcessorNilParameters will check the imput parameters for nil values
func checkProcessorNilParameters(arguments ArgBaseProcessor) error {

	if check.IfNil(arguments.Accounts) {
		return process.ErrNilAccountsAdapter
	}
	if check.IfNil(arguments.ForkDetector) {
		return process.ErrNilForkDetector
	}
	if check.IfNil(arguments.Hasher) {
		return process.ErrNilHasher
	}
	if check.IfNil(arguments.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arguments.Store) {
		return process.ErrNilStorage
	}
	if check.IfNil(arguments.ShardCoordinator) {
		return process.ErrNilShardCoordinator
	}
	if check.IfNil(arguments.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(arguments.SpecialAddressHandler) {
		return process.ErrNilSpecialAddressHandler
	}
	if check.IfNil(arguments.Uint64Converter) {
		return process.ErrNilUint64Converter
	}
	if check.IfNil(arguments.RequestHandler) {
		return process.ErrNilRequestHandler
	}
	if check.IfNil(arguments.EpochStartTrigger) {
		return process.ErrNilEpochStartTrigger
	}
	if check.IfNil(arguments.Rounder) {
		return process.ErrNilRounder
	}
	if check.IfNil(arguments.BootStorer) {
		return process.ErrNilStorage
	}
	if check.IfNil(arguments.BlockChainHook) {
		return process.ErrNilBlockChainHook
	}
	if check.IfNil(arguments.TxCoordinator) {
		return process.ErrNilTransactionCoordinator
	}
	if check.IfNil(arguments.HeaderValidator) {
		return process.ErrNilHeaderValidator
	}
	if check.IfNil(arguments.BlockTracker) {
		return process.ErrNilBlockTracker
	}

	return nil
}

func (bp *baseProcessor) createBlockStarted() {
	bp.resetMissingHdrs()
	bp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	bp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	bp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64)
	bp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
	bp.txCoordinator.CreateBlockStarted()
}

func (bp *baseProcessor) resetMissingHdrs() {
	bp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	bp.hdrsForCurrBlock.missingHdrs = 0
	bp.hdrsForCurrBlock.missingFinalityAttestingHdrs = 0
	bp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
}

//TODO: remove bool parameter and give instead the set to sort
func (bp *baseProcessor) sortHeadersForCurrentBlockByNonce(usedInBlock bool) map[uint32][]data.HeaderHandler {
	hdrsForCurrentBlock := make(map[uint32][]data.HeaderHandler)

	bp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for _, headerInfo := range bp.hdrsForCurrBlock.hdrHashAndInfo {
		if headerInfo.usedInBlock != usedInBlock {
			continue
		}

		hdrsForCurrentBlock[headerInfo.hdr.GetShardID()] = append(hdrsForCurrentBlock[headerInfo.hdr.GetShardID()], headerInfo.hdr)
	}
	bp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	// sort headers for each shard
	for _, hdrsForShard := range hdrsForCurrentBlock {
		process.SortHeadersByNonce(hdrsForShard)
	}

	return hdrsForCurrentBlock
}

//TODO: remove bool parameter and give instead the set to sort
func (bp *baseProcessor) sortHeaderHashesForCurrentBlockByNonce(usedInBlock bool) map[uint32][][]byte {
	hdrsForCurrentBlockInfo := make(map[uint32][]*nonceAndHashInfo)

	bp.hdrsForCurrBlock.mutHdrsForBlock.RLock()
	for metaBlockHash, headerInfo := range bp.hdrsForCurrBlock.hdrHashAndInfo {
		if headerInfo.usedInBlock != usedInBlock {
			continue
		}

		hdrsForCurrentBlockInfo[headerInfo.hdr.GetShardID()] = append(hdrsForCurrentBlockInfo[headerInfo.hdr.GetShardID()],
			&nonceAndHashInfo{nonce: headerInfo.hdr.GetNonce(), hash: []byte(metaBlockHash)})
	}
	bp.hdrsForCurrBlock.mutHdrsForBlock.RUnlock()

	for _, hdrsForShard := range hdrsForCurrentBlockInfo {
		if len(hdrsForShard) > 1 {
			sort.Slice(hdrsForShard, func(i, j int) bool {
				return hdrsForShard[i].nonce < hdrsForShard[j].nonce
			})
		}
	}

	hdrsHashesForCurrentBlock := make(map[uint32][][]byte, len(hdrsForCurrentBlockInfo))
	for shardId, hdrsForShard := range hdrsForCurrentBlockInfo {
		for _, hdrForShard := range hdrsForShard {
			hdrsHashesForCurrentBlock[shardId] = append(hdrsHashesForCurrentBlock[shardId], hdrForShard.hash)
		}
	}

	return hdrsHashesForCurrentBlock
}

func (bp *baseProcessor) createMiniBlockHeaders(body block.Body) (int, []block.MiniBlockHeader, error) {
	totalTxCount := 0
	miniBlockHeaders := make([]block.MiniBlockHeader, len(body))

	for i := 0; i < len(body); i++ {
		txCount := len(body[i].TxHashes)
		totalTxCount += txCount

		miniBlockHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, body[i])
		if err != nil {
			return 0, nil, err
		}

		miniBlockHeaders[i] = block.MiniBlockHeader{
			Hash:            miniBlockHash,
			SenderShardID:   body[i].SenderShardID,
			ReceiverShardID: body[i].ReceiverShardID,
			TxCount:         uint32(txCount),
			Type:            body[i].Type,
		}
	}

	return totalTxCount, miniBlockHeaders, nil
}

// check if header has the same miniblocks as presented in body
func (bp *baseProcessor) checkHeaderBodyCorrelation(miniBlockHeaders []block.MiniBlockHeader, body block.Body) error {
	mbHashesFromHdr := make(map[string]*block.MiniBlockHeader, len(miniBlockHeaders))
	for i := 0; i < len(miniBlockHeaders); i++ {
		mbHashesFromHdr[string(miniBlockHeaders[i].Hash)] = &miniBlockHeaders[i]
	}

	if len(miniBlockHeaders) != len(body) {
		return process.ErrHeaderBodyMismatch
	}

	for i := 0; i < len(body); i++ {
		miniBlock := body[i]

		mbHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, miniBlock)
		if err != nil {
			return err
		}

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

func (bp *baseProcessor) isHeaderOutOfRange(header data.HeaderHandler) bool {
	if check.IfNil(header) {
		return false
	}

	lastCrossNotarizedHeader, _, err := bp.blockTracker.GetLastCrossNotarizedHeader(header.GetShardID())
	if err != nil {
		log.Debug("isHeaderOutOfRange.GetLastCrossNotarizedHeader",
			"shard", header.GetShardID(),
			"error", err.Error())
		return false
	}

	allowedSize := uint64(float64(bp.dataPool.Headers().MaxSize()) * process.MaxOccupancyPercentageAllowed)
	isHeaderOutOfRange := header.GetNonce() > lastCrossNotarizedHeader.GetNonce()+allowedSize

	return isHeaderOutOfRange
}

// requestMissingFinalityAttestingHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the finality headers greater than the highest header, for given shard,
// related to the block which should be processed
func (bp *baseProcessor) requestMissingFinalityAttestingHeaders(
	shardId uint32,
	finality uint32,
) uint32 {
	requestedHeaders := uint32(0)
	missingFinalityAttestingHeaders := uint32(0)

	highestHdrNonce := bp.hdrsForCurrBlock.highestHdrNonce[shardId]
	if highestHdrNonce == uint64(0) {
		return missingFinalityAttestingHeaders
	}

	lastFinalityAttestingHeader := highestHdrNonce + uint64(finality)
	for i := highestHdrNonce + 1; i <= lastFinalityAttestingHeader; i++ {
		headers, headersHashes := bp.blockTracker.GetTrackedHeadersWithNonce(shardId, i)

		if len(headers) == 0 {
			missingFinalityAttestingHeaders++
			requestedHeaders++
			go bp.requestHeaderByShardAndNonce(shardId, i)
			continue
		}

		for index := range headers {
			bp.hdrsForCurrBlock.hdrHashAndInfo[string(headersHashes[index])] = &hdrInfo{hdr: headers[index], usedInBlock: false}
		}
	}

	if requestedHeaders > 0 {
		log.Debug("requested missing finality attesting headers",
			"num headers", requestedHeaders,
			"shard", shardId)
	}

	return missingFinalityAttestingHeaders
}

func (bp *baseProcessor) requestHeaderByShardAndNonce(targetShardID uint32, nonce uint64) {
	if targetShardID == sharding.MetachainShardId {
		bp.requestHandler.RequestMetaHeaderByNonce(nonce)
	} else {
		bp.requestHandler.RequestShardHeaderByNonce(targetShardID, nonce)
	}
}

func (bp *baseProcessor) cleanupPools(headerHandler data.HeaderHandler) {
	headersPool := bp.dataPool.Headers()
	noncesToFinal := bp.getNoncesToFinal(headerHandler)

	bp.removeHeadersBehindNonceFromPools(
		true,
		headersPool,
		bp.shardCoordinator.SelfId(),
		bp.forkDetector.GetHighestFinalBlockNonce())

	if bp.shardCoordinator.SelfId() == sharding.MetachainShardId {
		for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
			bp.cleanupPoolsForShard(shardID, headersPool, noncesToFinal)
		}
	} else {
		bp.cleanupPoolsForShard(sharding.MetachainShardId, headersPool, noncesToFinal)
	}
}

func (bp *baseProcessor) cleanupPoolsForShard(
	shardID uint32,
	headersPool dataRetriever.HeadersPool,
	noncesToFinal uint64,
) {
	crossNotarizedHeader, _, err := bp.blockTracker.GetCrossNotarizedHeader(shardID, noncesToFinal)
	if err != nil {
		log.Warn("cleanupPoolsForShard",
			"shard", shardID,
			"nonces to final", noncesToFinal,
			"error", err.Error())
		return
	}

	bp.removeHeadersBehindNonceFromPools(
		false,
		headersPool,
		shardID,
		crossNotarizedHeader.GetNonce(),
	)
}

func (bp *baseProcessor) removeHeadersBehindNonceFromPools(
	shouldRemoveBlockBody bool,
	headersPool dataRetriever.HeadersPool,
	shardId uint32,
	nonce uint64,
) {
	if nonce <= 1 {
		return
	}

	if check.IfNil(headersPool) {
		return
	}

	nonces := headersPool.Nonces(shardId)
	for _, nonceFromCache := range nonces {
		if nonceFromCache >= nonce {
			continue
		}

		if shouldRemoveBlockBody {
			bp.removeBlocksBody(nonceFromCache, shardId, headersPool)
		}

		headersPool.RemoveHeaderByNonceAndShardId(nonceFromCache, shardId)
	}
}

func (bp *baseProcessor) removeBlocksBody(nonce uint64, shardId uint32, headersPool dataRetriever.HeadersPool) {
	headers, _, err := headersPool.GetHeadersByNonceAndShardId(nonce, shardId)
	if err != nil {
		return
	}

	for _, header := range headers {
		errNotCritical := bp.removeBlockBodyOfHeader(header)
		if errNotCritical != nil {
			log.Debug("RemoveBlockDataFromPool", "error", errNotCritical.Error())
		}
	}
}

func (bp *baseProcessor) removeBlockBodyOfHeader(headerHandler data.HeaderHandler) error {
	bodyHandler, err := bp.requestBlockBodyHandler.GetBlockBodyFromPool(headerHandler)
	if err != nil {
		return err
	}

	body, ok := bodyHandler.(block.Body)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	err = bp.txCoordinator.RemoveBlockDataFromPool(body)
	if err != nil {
		return err
	}

	return nil
}

func (bp *baseProcessor) cleanupBlockTrackerPools(headerHandler data.HeaderHandler) {
	noncesToFinal := bp.getNoncesToFinal(headerHandler)

	bp.cleanupBlockTrackerPoolsForShard(bp.shardCoordinator.SelfId(), noncesToFinal)

	if bp.shardCoordinator.SelfId() == sharding.MetachainShardId {
		for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
			bp.cleanupBlockTrackerPoolsForShard(shardID, noncesToFinal)
		}
	} else {
		bp.cleanupBlockTrackerPoolsForShard(sharding.MetachainShardId, noncesToFinal)
	}
}

func (bp *baseProcessor) cleanupBlockTrackerPoolsForShard(shardID uint32, noncesToFinal uint64) {
	shardForSelfNotarized := bp.getShardForSelfNotarized(shardID)
	selfNotarizedHeader, _, err := bp.blockTracker.GetLastSelfNotarizedHeader(shardForSelfNotarized)
	if err != nil {
		log.Warn("cleanupBlockTrackerPoolsForShard.GetLastSelfNotarizedHeader",
			"shard", shardForSelfNotarized,
			"error", err.Error())
		return
	}

	selfNotarizedNonce := selfNotarizedHeader.GetNonce()
	crossNotarizedNonce := uint64(0)

	if shardID != bp.shardCoordinator.SelfId() {
		crossNotarizedHeader, _, errNotCritical := bp.blockTracker.GetCrossNotarizedHeader(shardID, noncesToFinal)
		if errNotCritical != nil {
			log.Warn("cleanupBlockTrackerPoolsForShard.GetCrossNotarizedHeader",
				"shard", shardID,
				"nonces to final", noncesToFinal,
				"error", errNotCritical.Error())
			return
		}

		crossNotarizedNonce = crossNotarizedHeader.GetNonce()
	}

	bp.blockTracker.CleanupHeadersBehindNonce(
		shardID,
		selfNotarizedNonce,
		crossNotarizedNonce,
	)

	log.Trace("cleanupBlockTrackerPoolsForShard.CleanupHeadersBehindNonce",
		"shard", shardID,
		"self notarized nonce", selfNotarizedNonce,
		"cross notarized nonce", crossNotarizedNonce)
}

func (bp *baseProcessor) getShardForSelfNotarized(shardID uint32) uint32 {
	isSelfShard := shardID == bp.shardCoordinator.SelfId()
	if isSelfShard && bp.shardCoordinator.SelfId() != sharding.MetachainShardId {
		return sharding.MetachainShardId
	}

	return shardID
}

func (bp *baseProcessor) prepareDataForBootStorer(
	headerInfo bootstrapStorage.BootstrapHeaderInfo,
	round uint64,
	lastSelfNotarizedHeaders []bootstrapStorage.BootstrapHeaderInfo,
	pendingMiniBlocks []bootstrapStorage.PendingMiniBlockInfo,
	highestFinalBlockNonce uint64,
	processedMiniBlocks []bootstrapStorage.MiniBlocksInMeta,
) {
	//TODO add end of epoch stuff

	lastCrossNotarizedHeaders := bp.getLastCrossNotarizedHeaders()

	bootData := bootstrapStorage.BootstrapData{
		LastHeader:                headerInfo,
		LastCrossNotarizedHeaders: lastCrossNotarizedHeaders,
		LastSelfNotarizedHeaders:  lastSelfNotarizedHeaders,
		PendingMiniBlocks:         pendingMiniBlocks,
		HighestFinalBlockNonce:    highestFinalBlockNonce,
		ProcessedMiniBlocks:       processedMiniBlocks,
	}

	err := bp.bootStorer.Put(int64(round), bootData)
	if err != nil {
		log.Warn("cannot save boot data in storage",
			"error", err.Error())
	}
}

func (bp *baseProcessor) getLastCrossNotarizedHeaders() []bootstrapStorage.BootstrapHeaderInfo {
	lastCrossNotarizedHeaders := make([]bootstrapStorage.BootstrapHeaderInfo, 0, bp.shardCoordinator.NumberOfShards()+1)

	for shardID := uint32(0); shardID < bp.shardCoordinator.NumberOfShards(); shardID++ {
		bootstrapHeaderInfo := bp.getLastCrossNotarizedHeadersForShard(shardID)
		if bootstrapHeaderInfo != nil {
			lastCrossNotarizedHeaders = append(lastCrossNotarizedHeaders, *bootstrapHeaderInfo)
		}
	}

	bootstrapHeaderInfo := bp.getLastCrossNotarizedHeadersForShard(sharding.MetachainShardId)
	if bootstrapHeaderInfo != nil {
		lastCrossNotarizedHeaders = append(lastCrossNotarizedHeaders, *bootstrapHeaderInfo)
	}

	return bootstrapStorage.TrimHeaderInfoSlice(lastCrossNotarizedHeaders)
}

func (bp *baseProcessor) getLastCrossNotarizedHeadersForShard(shardID uint32) *bootstrapStorage.BootstrapHeaderInfo {
	lastCrossNotarizedHeader, lastCrossNotarizedHeaderHash, err := bp.blockTracker.GetLastCrossNotarizedHeader(shardID)
	if err != nil {
		log.Warn("getLastCrossNotarizedHeadersForShard",
			"shard", shardID,
			"error", err.Error())
		return nil
	}

	if lastCrossNotarizedHeader.GetNonce() == 0 {
		return nil
	}

	headerInfo := &bootstrapStorage.BootstrapHeaderInfo{
		ShardId: lastCrossNotarizedHeader.GetShardID(),
		Nonce:   lastCrossNotarizedHeader.GetNonce(),
		Hash:    lastCrossNotarizedHeaderHash,
	}

	return headerInfo
}

func deleteSelfReceiptsMiniBlocks(body block.Body) block.Body {
	for i := 0; i < len(body); {
		mb := body[i]
		if mb.ReceiverShardID != mb.SenderShardID {
			i++
			continue
		}

		if mb.Type != block.ReceiptBlock && mb.Type != block.SmartContractResultBlock {
			i++
			continue
		}

		body[i] = body[len(body)-1]
		body = body[:len(body)-1]
		if i == len(body)-1 {
			break
		}
	}

	return body
}

func (bp *baseProcessor) getNoncesToFinal(headerHandler data.HeaderHandler) uint64 {
	currentBlockNonce := uint64(0)
	if !check.IfNil(headerHandler) {
		currentBlockNonce = headerHandler.GetNonce()
	}

	noncesToFinal := uint64(0)
	finalBlockNonce := bp.forkDetector.GetHighestFinalBlockNonce()
	if currentBlockNonce > finalBlockNonce {
		noncesToFinal = currentBlockNonce - finalBlockNonce
	}

	return noncesToFinal
}

// DecodeBlockBody method decodes block body from a given byte array
func (bp *baseProcessor) DecodeBlockBody(dta []byte) data.BodyHandler {
	if dta == nil {
		return nil
	}

	var body block.Body

	err := bp.marshalizer.Unmarshal(&body, dta)
	if err != nil {
		log.Debug("DecodeBlockBody.Unmarshal", "error", err.Error())
		return nil
	}

	return body
}

// DecodeBlockHeader method decodes block header from a given byte array
func (bp *baseProcessor) DecodeBlockHeader(dta []byte) data.HeaderHandler {
	if dta == nil {
		return nil
	}

	header := bp.blockProcessor.CreateNewHeader()

	err := bp.marshalizer.Unmarshal(&header, dta)
	if err != nil {
		log.Debug("DecodeBlockHeader.Unmarshal", "error", err.Error())
		return nil
	}

	return header
}

// DecodeBlockBodyAndHeader method decodes block body and header from a given byte array
func (bp *baseProcessor) DecodeBlockBodyAndHeader(dta []byte) (data.BodyHandler, data.HeaderHandler) {
	if dta == nil {
		return nil, nil
	}

	//TODO refactor this hack when data.Body will be an actual structure (protobuf feat)
	bodyAndHeader := struct {
		Body   block.Body
		Header data.HeaderHandler
	}{
		Body:   make(block.Body, 0),
		Header: bp.blockProcessor.CreateNewHeader(),
	}

	err := bp.marshalizer.Unmarshal(&bodyAndHeader, dta)
	if err != nil {
		log.Debug("DecodeBlockBodyAndHeader.Unmarshal: dta", "error", err.Error())
		return nil, nil
	}

	return bodyAndHeader.Body, bodyAndHeader.Header
}

func (bp *baseProcessor) saveBody(body block.Body) {
	err := bp.txCoordinator.SaveBlockDataToStorage(body)
	if err != nil {
		log.Warn("saveBody.SaveBlockDataToStorage", "error", err.Error())
	}

	for i := 0; i < len(body); i++ {
		marshalizedMiniBlock, errNotCritical := bp.marshalizer.Marshal(body[i])
		if errNotCritical != nil {
			log.Warn("saveBody.Marshal", "error", errNotCritical.Error())
			continue
		}

		miniBlockHash := bp.hasher.Compute(string(marshalizedMiniBlock))
		errNotCritical = bp.store.Put(dataRetriever.MiniBlockUnit, miniBlockHash, marshalizedMiniBlock)
		if errNotCritical != nil {
			log.Warn("saveBody.Put -> MiniBlockUnit", "error", errNotCritical.Error())
		}
	}
}

func (bp *baseProcessor) saveShardHeader(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
	nonceToByteSlice := bp.uint64Converter.ToByteSlice(header.GetNonce())
	hdrNonceHashDataUnit := dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(header.GetShardID())

	errNotCritical := bp.store.Put(hdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	if errNotCritical != nil {
		log.Warn(fmt.Sprintf("saveHeader.Put -> ShardHdrNonceHashDataUnit_%d", header.GetShardID()),
			"error", errNotCritical.Error(),
		)
	}

	errNotCritical = bp.store.Put(dataRetriever.BlockHeaderUnit, headerHash, marshalizedHeader)
	if errNotCritical != nil {
		log.Warn("saveHeader.Put -> BlockHeaderUnit", "error", errNotCritical.Error())
	}
}

func (bp *baseProcessor) saveMetaHeader(header data.HeaderHandler, headerHash []byte, marshalizedHeader []byte) {
	nonceToByteSlice := bp.uint64Converter.ToByteSlice(header.GetNonce())

	errNotCritical := bp.store.Put(dataRetriever.MetaHdrNonceHashDataUnit, nonceToByteSlice, headerHash)
	if errNotCritical != nil {
		log.Warn("saveMetaHeader.Put -> MetaHdrNonceHashDataUnit", "error", errNotCritical.Error())
	}

	errNotCritical = bp.store.Put(dataRetriever.MetaBlockUnit, headerHash, marshalizedHeader)
	if errNotCritical != nil {
		log.Warn("saveMetaHeader.Put -> MetaBlockUnit", "error", errNotCritical.Error())
	}
}

func getLastSelfNotarizedHeaderByItself(chainHandler data.ChainHandler) (data.HeaderHandler, []byte) {
	if check.IfNil(chainHandler.GetCurrentBlockHeader()) {
		return chainHandler.GetGenesisHeader(), chainHandler.GetGenesisHeaderHash()
	}

	return chainHandler.GetCurrentBlockHeader(), chainHandler.GetCurrentBlockHeaderHash()
}
