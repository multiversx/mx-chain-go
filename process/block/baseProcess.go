package block

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

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
	missingHdrs                    uint32
	missingFinalityAttestingHdrs   uint32
	requestedFinalityAttestingHdrs map[uint32][]uint64
	highestHdrNonce                map[uint32]uint64
	mutHdrsForBlock                sync.RWMutex
	hdrHashAndInfo                 map[string]*hdrInfo
}

type mapShardHeaders map[uint32][]data.HeaderHandler
type mapShardHeader map[uint32]data.HeaderHandler

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
	rounder                      consensus.Rounder

	hdrsForCurrBlock hdrForBlock

	mutNotarizedHdrs sync.RWMutex
	notarizedHdrs    mapShardHeaders

	mutLastHdrs sync.RWMutex
	lastHdrs    mapShardHeader

	onRequestHeaderHandlerByNonce func(shardId uint32, nonce uint64)
	onRequestHeaderHandler        func(shardId uint32, hash []byte)

	appStatusHandler core.AppStatusHandler
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

// AddLastNotarizedHdr adds the last notarized header
func (bp *baseProcessor) AddLastNotarizedHdr(shardId uint32, processedHdr data.HeaderHandler) {
	bp.mutNotarizedHdrs.Lock()
	bp.notarizedHdrs[shardId] = append(bp.notarizedHdrs[shardId], processedHdr)
	bp.mutNotarizedHdrs.Unlock()
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

			log.Info(fmt.Sprintf("hash does not match: local block hash is %s and node received block with previous hash %s\n",
				core.ToB64(chainHandler.GetGenesisHeaderHash()),
				core.ToB64(headerHandler.GetPrevHash())))

			return process.ErrBlockHashDoesNotMatch
		}

		log.Info(fmt.Sprintf("nonce does not match: local block nonce is 0 and node received block with nonce %d\n",
			headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	if headerHandler.GetRound() <= currentBlockHeader.GetRound() {
		log.Info(fmt.Sprintf("round does not match: local block round is %d and node received block with round %d\n",
			currentBlockHeader.GetRound(), headerHandler.GetRound()))

		return process.ErrLowerRoundInBlock
	}

	if headerHandler.GetNonce() != currentBlockHeader.GetNonce()+1 {
		log.Info(fmt.Sprintf("nonce does not match: local block nonce is %d and node received block with nonce %d\n",
			currentBlockHeader.GetNonce(), headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, currentBlockHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(headerHandler.GetPrevHash(), prevHeaderHash) {
		log.Info(fmt.Sprintf("hash does not match: local block hash is %s and node received block with previous hash %s\n",
			core.ToB64(prevHeaderHash), core.ToB64(headerHandler.GetPrevHash())))

		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(headerHandler.GetPrevRandSeed(), currentBlockHeader.GetRandSeed()) {
		log.Info(fmt.Sprintf("random seed does not match: local block random seed is %s and node received block with previous random seed %s\n",
			core.ToB64(currentBlockHeader.GetRandSeed()), core.ToB64(headerHandler.GetPrevRandSeed())))

		return process.ErrRandSeedDoesNotMatch
	}

	if bodyHandler != nil {
		// TODO: add bodyHandler verification here
	}

	// TODO: add signature validation as well, with randomness source and all
	return nil
}

// verifyStateRoot verifies the state root hash given as parameter against the
// Merkle trie root hash stored for accounts and returns if equal or not
func (bp *baseProcessor) verifyStateRoot(rootHash []byte) bool {
	trieRootHash, err := bp.accounts.RootHash()
	if err != nil {
		log.Debug(err.Error())
	}

	return bytes.Equal(trieRootHash, rootHash)
}

// getRootHash returns the accounts merkle tree root hash
func (bp *baseProcessor) getRootHash() []byte {
	rootHash, err := bp.accounts.RootHash()
	if err != nil {
		log.Debug(err.Error())
	}

	return rootHash
}

func (bp *baseProcessor) isHdrConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	if prevHdr == nil || prevHdr.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}
	if currHdr == nil || currHdr.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}

	// special case with genesis nonce - 0
	if currHdr.GetNonce() == 0 {
		if prevHdr.GetNonce() != 0 {
			return process.ErrWrongNonceInBlock
		}
		// block with nonce 0 was already saved
		if prevHdr.GetRootHash() != nil {
			return process.ErrRootStateDoesNotMatch
		}
		return nil
	}

	//TODO: add verification if rand seed was correctly computed add other verification
	//TODO: check here if the 2 header blocks were correctly signed and the consensus group was correctly elected
	if prevHdr.GetRound() >= currHdr.GetRound() {
		log.Debug(fmt.Sprintf("round does not match in shard %d: local block round is %d and node received block with round %d\n",
			currHdr.GetShardID(), prevHdr.GetRound(), currHdr.GetRound()))
		return process.ErrLowerRoundInBlock
	}

	if currHdr.GetNonce() != prevHdr.GetNonce()+1 {
		log.Debug(fmt.Sprintf("nonce does not match in shard %d: local block nonce is %d and node received block with nonce %d\n",
			currHdr.GetShardID(), prevHdr.GetNonce(), currHdr.GetNonce()))
		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, prevHdr)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHdr.GetPrevHash(), prevHeaderHash) {
		log.Debug(fmt.Sprintf("block hash does not match in shard %d: local block hash is %s and node received block with previous hash %s\n",
			currHdr.GetShardID(), core.ToB64(prevHeaderHash), core.ToB64(currHdr.GetPrevHash())))
		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(currHdr.GetPrevRandSeed(), prevHdr.GetRandSeed()) {
		log.Debug(fmt.Sprintf("random seed does not match in shard %d: local block random seed is %s and node received block with previous random seed %s\n",
			currHdr.GetShardID(), core.ToB64(prevHdr.GetRandSeed()), core.ToB64(currHdr.GetPrevRandSeed())))
		return process.ErrRandSeedDoesNotMatch
	}

	return nil
}

func (bp *baseProcessor) checkHeaderTypeCorrect(shardId uint32, hdr data.HeaderHandler) error {
	if shardId >= bp.shardCoordinator.NumberOfShards() && shardId != sharding.MetachainShardId {
		return process.ErrShardIdMissmatch
	}

	if shardId < bp.shardCoordinator.NumberOfShards() {
		_, ok := hdr.(*block.Header)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
	}

	if shardId == sharding.MetachainShardId {
		_, ok := hdr.(*block.MetaBlock)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
	}

	return nil
}

func (bp *baseProcessor) removeNotarizedHdrsBehindPreviousFinal(hdrsToPreservedBehindFinal uint32) {
	bp.mutNotarizedHdrs.Lock()
	for shardId := range bp.notarizedHdrs {
		notarizedHdrsCount := uint32(len(bp.notarizedHdrs[shardId]))
		if notarizedHdrsCount > hdrsToPreservedBehindFinal {
			finalIndex := notarizedHdrsCount - 1 - hdrsToPreservedBehindFinal
			bp.notarizedHdrs[shardId] = bp.notarizedHdrs[shardId][finalIndex:]
		}
	}
	bp.mutNotarizedHdrs.Unlock()
}

func (bp *baseProcessor) removeLastNotarized() {
	bp.mutNotarizedHdrs.Lock()
	for shardId := range bp.notarizedHdrs {
		notarizedHdrsCount := len(bp.notarizedHdrs[shardId])
		if notarizedHdrsCount > 1 {
			bp.notarizedHdrs[shardId] = bp.notarizedHdrs[shardId][:notarizedHdrsCount-1]
		}

		//TODO: Remove this print later
		if notarizedHdrsCount == 1 {
			if bp.notarizedHdrs[shardId][0].GetNonce() > 0 {
				log.Info(fmt.Sprintf("critical error: remove last notarized header for shard %d with round %d and nonce %d has been failed\n",
					bp.notarizedHdrs[shardId][0].GetShardID(),
					bp.notarizedHdrs[shardId][0].GetRound(),
					bp.notarizedHdrs[shardId][0].GetNonce(),
				))
			}
		}
	}
	bp.mutNotarizedHdrs.Unlock()
}

func (bp *baseProcessor) lastNotarizedHdrForShard(shardId uint32) data.HeaderHandler {
	notarizedHdrsCount := len(bp.notarizedHdrs[shardId])
	if notarizedHdrsCount > 0 {
		return bp.notarizedHdrs[shardId][notarizedHdrsCount-1]
	}

	return nil
}

func (bp *baseProcessor) lastHdrForShard(shardId uint32) data.HeaderHandler {
	bp.mutLastHdrs.RLock()
	defer bp.mutLastHdrs.RUnlock()

	return bp.lastHdrs[shardId]
}

func (bp *baseProcessor) setLastHdrForShard(shardId uint32, header data.HeaderHandler) {
	if check.IfNil(header) {
		return
	}

	bp.mutLastHdrs.Lock()
	defer bp.mutLastHdrs.Unlock()

	lastHeader, ok := bp.lastHdrs[shardId]
	if ok && lastHeader.GetRound() > header.GetRound() {
		return
	}

	bp.lastHdrs[shardId] = header
}

func (bp *baseProcessor) saveLastNotarizedHeader(shardId uint32, processedHdrs []data.HeaderHandler) error {
	bp.mutNotarizedHdrs.Lock()
	defer bp.mutNotarizedHdrs.Unlock()

	if bp.notarizedHdrs == nil {
		return process.ErrNotarizedHdrsSliceIsNil
	}

	err := bp.checkHeaderTypeCorrect(shardId, bp.lastNotarizedHdrForShard(shardId))
	if err != nil {
		return err
	}

	sort.Slice(processedHdrs, func(i, j int) bool {
		return processedHdrs[i].GetNonce() < processedHdrs[j].GetNonce()
	})

	tmpLastNotarizedHdrForShard := bp.lastNotarizedHdrForShard(shardId)

	for i := 0; i < len(processedHdrs); i++ {
		err = bp.checkHeaderTypeCorrect(shardId, processedHdrs[i])
		if err != nil {
			return err
		}

		err = bp.isHdrConstructionValid(processedHdrs[i], tmpLastNotarizedHdrForShard)
		if err != nil {
			return err
		}

		tmpLastNotarizedHdrForShard = processedHdrs[i]
	}

	bp.notarizedHdrs[shardId] = append(bp.notarizedHdrs[shardId], tmpLastNotarizedHdrForShard)
	DisplayLastNotarized(bp.marshalizer, bp.hasher, tmpLastNotarizedHdrForShard, shardId)

	return nil
}

func (bp *baseProcessor) getLastNotarizedHdr(shardId uint32) (data.HeaderHandler, error) {
	bp.mutNotarizedHdrs.RLock()
	defer bp.mutNotarizedHdrs.RUnlock()

	if bp.notarizedHdrs == nil {
		return nil, process.ErrNotarizedHdrsSliceIsNil
	}

	hdr := bp.lastNotarizedHdrForShard(shardId)

	err := bp.checkHeaderTypeCorrect(shardId, hdr)
	if err != nil {
		return nil, err
	}

	return hdr, nil
}

// SetLastNotarizedHeadersSlice sets the headers blocks in notarizedHdrs for every shard
// This is done when starting a new epoch so metachain can use it when validating next shard header blocks
// and shard can validate the next meta header
func (bp *baseProcessor) setLastNotarizedHeadersSlice(startHeaders map[uint32]data.HeaderHandler) error {
	//TODO: protect this to be called only once at genesis time
	//TODO: do this on constructor as it is a must to for blockprocessor to work
	bp.mutNotarizedHdrs.Lock()
	defer bp.mutNotarizedHdrs.Unlock()

	if startHeaders == nil {
		return process.ErrNotarizedHdrsSliceIsNil
	}

	bp.notarizedHdrs = make(mapShardHeaders, bp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < bp.shardCoordinator.NumberOfShards(); i++ {
		hdr, ok := startHeaders[i].(*block.Header)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
		bp.notarizedHdrs[i] = append(bp.notarizedHdrs[i], hdr)
	}

	hdr, ok := startHeaders[sharding.MetachainShardId].(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}
	bp.notarizedHdrs[sharding.MetachainShardId] = append(bp.notarizedHdrs[sharding.MetachainShardId], hdr)

	return nil
}

func (bp *baseProcessor) requestHeadersIfMissing(
	sortedHdrs []data.HeaderHandler,
	shardId uint32,
	maxRound uint64,
	cacher storage.Cacher,
) error {

	allowedSize := uint64(float64(cacher.MaxSize()) * process.MaxOccupancyPercentageAllowed)

	prevHdr, err := bp.getLastNotarizedHdr(shardId)
	if err != nil {
		return err
	}

	lastNotarizedHdrNonce := prevHdr.GetNonce()
	highestHdr := prevHdr

	missingNonces := make([]uint64, 0)
	for i := 0; i < len(sortedHdrs); i++ {
		currHdr := sortedHdrs[i]
		if currHdr == nil {
			continue
		}

		if i > 0 {
			prevHdr = sortedHdrs[i-1]
		}

		hdrTooNew := currHdr.GetRound() > maxRound || prevHdr.GetRound() > maxRound
		if hdrTooNew {
			continue
		}

		if currHdr.GetNonce()-prevHdr.GetNonce() > 1 {
			for j := prevHdr.GetNonce() + 1; j < currHdr.GetNonce(); j++ {
				missingNonces = append(missingNonces, j)
			}
		}

		highestHdr = currHdr
	}

	// ask for headers, if there most probably should be
	roundDiff := int64(maxRound) - int64(highestHdr.GetRound())
	if roundDiff > 0 {
		nonceDiff := int64(lastNotarizedHdrNonce) + process.MaxHeadersToRequestInAdvance - int64(highestHdr.GetNonce())
		if nonceDiff > 0 {
			nbHeadersToRequestInAdvance := core.MinUint64(uint64(roundDiff), uint64(nonceDiff))
			startNonce := highestHdr.GetNonce() + 1
			for nonce := startNonce; nonce < startNonce+nbHeadersToRequestInAdvance; nonce++ {
				missingNonces = append(missingNonces, nonce)
			}
		}
	}

	requested := 0
	for _, nonce := range missingNonces {
		// do the request here
		if bp.onRequestHeaderHandlerByNonce == nil {
			return process.ErrNilRequestHeaderHandlerByNonce
		}

		isHeaderOutOfRange := nonce > lastNotarizedHdrNonce+allowedSize
		if isHeaderOutOfRange {
			break
		}

		if requested >= process.MaxHeaderRequestsAllowed {
			break
		}

		requested++
		go bp.onRequestHeaderHandlerByNonce(shardId, nonce)
	}

	return nil
}

func displayHeader(headerHandler data.HeaderHandler) []*display.LineData {
	lines := make([]*display.LineData, 0)

	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Epoch",
		fmt.Sprintf("%d", headerHandler.GetEpoch())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Round",
		fmt.Sprintf("%d", headerHandler.GetRound())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"TimeStamp",
		fmt.Sprintf("%d", headerHandler.GetTimeStamp())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Nonce",
		fmt.Sprintf("%d", headerHandler.GetNonce())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev hash",
		core.ToB64(headerHandler.GetPrevHash())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev rand seed",
		core.ToB64(headerHandler.GetPrevRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Rand seed",
		core.ToB64(headerHandler.GetRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Pub keys bitmap",
		core.ToHex(headerHandler.GetPubKeysBitmap())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Signature",
		core.ToB64(headerHandler.GetSignature())}))
	lines = append(lines, display.NewLineData(true, []string{
		"",
		"Root hash",
		core.ToB64(headerHandler.GetRootHash())}))
	return lines
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
	if check.IfNil(arguments.Rounder) {
		return process.ErrNilRounder
	}

	return nil
}

func (bp *baseProcessor) createBlockStarted() {
	bp.resetMissingHdrs()
	bp.hdrsForCurrBlock.mutHdrsForBlock.Lock()
	bp.hdrsForCurrBlock.hdrHashAndInfo = make(map[string]*hdrInfo)
	bp.hdrsForCurrBlock.highestHdrNonce = make(map[uint32]uint64)
	bp.hdrsForCurrBlock.requestedFinalityAttestingHdrs = make(map[uint32][]uint64)
	bp.hdrsForCurrBlock.mutHdrsForBlock.Unlock()
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

	hdrsHashesForCurrentBlock := make(map[uint32][][]byte)
	for shardId, hdrsForShard := range hdrsForCurrentBlockInfo {
		for _, hdrForShard := range hdrsForShard {
			hdrsHashesForCurrentBlock[shardId] = append(hdrsHashesForCurrentBlock[shardId], hdrForShard.hash)
		}
	}

	return hdrsHashesForCurrentBlock
}

func (bp *baseProcessor) isHeaderOutOfRange(header data.HeaderHandler, cacher storage.Cacher) bool {
	lastNotarizedHdr, err := bp.getLastNotarizedHdr(header.GetShardID())
	if err != nil {
		return false
	}

	allowedSize := uint64(float64(cacher.MaxSize()) * process.MaxOccupancyPercentageAllowed)
	isHeaderOutOfRange := header.GetNonce() > lastNotarizedHdr.GetNonce()+allowedSize

	return isHeaderOutOfRange
}

func (bp *baseProcessor) wasHeaderRequested(shardId uint32, nonce uint64) bool {
	requestedNonces, ok := bp.hdrsForCurrBlock.requestedFinalityAttestingHdrs[shardId]
	if !ok {
		return false
	}

	for _, requestedNonce := range requestedNonces {
		if requestedNonce == nonce {
			return true
		}
	}

	return false
}

// requestMissingFinalityAttestingHeaders requests the headers needed to accept the current selected headers for
// processing the current block. It requests the finality headers greater than the highest header, for given shard,
// related to the block which should be processed
func (bp *baseProcessor) requestMissingFinalityAttestingHeaders(
	shardId uint32,
	finality uint32,
	getHeaderFromPoolWithNonce func(uint64, uint32) (data.HeaderHandler, []byte, error),
	cacher storage.Cacher,
) uint32 {
	requestedHeaders := uint32(0)
	missingFinalityAttestingHeaders := uint32(0)

	highestHdrNonce := bp.hdrsForCurrBlock.highestHdrNonce[shardId]
	if highestHdrNonce == uint64(0) {
		return missingFinalityAttestingHeaders
	}

	lastFinalityAttestingHeader := highestHdrNonce + uint64(finality)
	for i := highestHdrNonce + 1; i <= lastFinalityAttestingHeader; i++ {
		headers, headersHashes := bp.getHeadersFromPools(getHeaderFromPoolWithNonce, cacher, shardId, i)

		if len(headers) == 0 {
			missingFinalityAttestingHeaders++
			wasHeaderRequested := bp.wasHeaderRequested(shardId, i)
			if !wasHeaderRequested {
				requestedHeaders++
				bp.hdrsForCurrBlock.requestedFinalityAttestingHdrs[shardId] = append(bp.hdrsForCurrBlock.requestedFinalityAttestingHdrs[shardId], i)
				go bp.onRequestHeaderHandlerByNonce(shardId, i)
			}

			continue
		}

		for index := range headers {
			bp.hdrsForCurrBlock.hdrHashAndInfo[string(headersHashes[index])] = &hdrInfo{hdr: headers[index], usedInBlock: false}
		}
	}

	if requestedHeaders > 0 {
		log.Info(fmt.Sprintf("requested %d missing finality attesting headers for shard %d\n",
			requestedHeaders,
			shardId))
	}

	return missingFinalityAttestingHeaders
}

func (bp *baseProcessor) isShardStuck(shardId uint32) bool {
	header := bp.lastHdrForShard(shardId)
	if check.IfNil(header) {
		return false
	}

	isShardStuck := bp.rounder.Index()-int64(header.GetRound()) > process.MaxRoundsWithoutCommittedBlock
	return isShardStuck
}

func (bp *baseProcessor) cleanupPools(
	headersNoncesPool dataRetriever.Uint64SyncMapCacher,
	headersPool storage.Cacher,
	notarizedHeadersPool storage.Cacher,
) {
	bp.removeHeadersBehindNonceFromPools(
		headersPool,
		headersNoncesPool,
		bp.shardCoordinator.SelfId(),
		bp.forkDetector.GetHighestFinalBlockNonce())

	for shardId := range bp.notarizedHdrs {
		lastNotarizedHdr := bp.lastNotarizedHdrForShard(shardId)
		if check.IfNil(lastNotarizedHdr) {
			continue
		}

		bp.removeHeadersBehindNonceFromPools(
			notarizedHeadersPool,
			headersNoncesPool,
			shardId,
			lastNotarizedHdr.GetNonce())
	}

	return
}

func (bp *baseProcessor) removeHeadersBehindNonceFromPools(
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	shardId uint32,
	nonce uint64,
) {

	if nonce <= 1 {
		return
	}

	if check.IfNil(cacher) {
		return
	}

	for _, key := range cacher.Keys() {
		val, _ := cacher.Peek(key)
		if val == nil {
			continue
		}

		hdr, ok := val.(data.HeaderHandler)
		if !ok {
			continue
		}

		if hdr.GetShardID() != shardId || hdr.GetNonce() >= nonce {
			continue
		}

		cacher.Remove(key)

		if check.IfNil(uint64SyncMapCacher) {
			continue
		}

		uint64SyncMapCacher.Remove(hdr.GetNonce(), hdr.GetShardID())
	}
}

func (bp *baseProcessor) removeHeaderFromPools(
	header data.HeaderHandler,
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
) {

	if check.IfNil(header) {
		return
	}

	headerHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, header)
	if err != nil {
		return
	}

	if !check.IfNil(cacher) {
		cacher.Remove(headerHash)
	}

	if !check.IfNil(uint64SyncMapCacher) {
		syncMap, ok := uint64SyncMapCacher.Get(header.GetNonce())
		if !ok {
			return
		}

		hash, ok := syncMap.Load(header.GetShardID())
		if hash == nil || !ok {
			return
		}

		if bytes.Equal(headerHash, hash) {
			uint64SyncMapCacher.Remove(header.GetNonce(), header.GetShardID())
		}
	}
}

func (bp *baseProcessor) getHeadersFromPools(
	getHeaderFromPoolWithNonce func(uint64, uint32) (data.HeaderHandler, []byte, error),
	cacher storage.Cacher,
	shardId uint32,
	nonce uint64,
) ([]data.HeaderHandler, [][]byte) {

	headers := make([]data.HeaderHandler, 0)
	headersHashes := make([][]byte, 0)

	for _, headerHash := range cacher.Keys() {
		val, _ := cacher.Peek(headerHash)
		if val == nil {
			continue
		}

		header, ok := val.(data.HeaderHandler)
		if !ok {
			continue
		}

		if header.GetShardID() == shardId && header.GetNonce() == nonce {
			headers = append(headers, header)
			headersHashes = append(headersHashes, headerHash)
		}
	}

	header, headerHash, err := getHeaderFromPoolWithNonce(nonce, shardId)
	if err != nil {
		return headers, headersHashes
	}

	headers = append(headers, header)
	headersHashes = append(headersHashes, headerHash)

	return headers, headersHashes
}
