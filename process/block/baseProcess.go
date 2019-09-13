package block

import (
	"bytes"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
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
)

var log = logger.DefaultLogger()

type hashAndHdr struct {
	hdr  data.HeaderHandler
	hash []byte
}

type mapShardHeaders map[uint32][]data.HeaderHandler

type baseProcessor struct {
	shardCoordinator   sharding.Coordinator
	accounts           state.AccountsAdapter
	forkDetector       process.ForkDetector
	hasher             hashing.Hasher
	marshalizer        marshal.Marshalizer
	store              dataRetriever.StorageService
	uint64Converter    typeConverters.Uint64ByteSliceConverter
	blockSizeThrottler process.BlockSizeThrottler

	mutNotarizedHdrs sync.RWMutex
	notarizedHdrs    mapShardHeaders

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

// RevertAccountState reverts the account state for cleanup failed process
func (bp *baseProcessor) RevertAccountState() {
	err := bp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Error(err.Error())
	}
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

			log.Info(fmt.Sprintf("hash not match: local block hash is %s and node received block with previous hash %s\n",
				core.ToB64(chainHandler.GetGenesisHeaderHash()),
				core.ToB64(headerHandler.GetPrevHash())))

			return process.ErrBlockHashDoesNotMatch
		}

		log.Info(fmt.Sprintf("nonce not match: local block nonce is 0 and node received block with nonce %d\n",
			headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	if headerHandler.GetRound() <= currentBlockHeader.GetRound() {
		log.Info(fmt.Sprintf("round not match: local block round is %d and node received block with round %d\n",
			currentBlockHeader.GetRound(), headerHandler.GetRound()))

		return process.ErrLowerRoundInBlock
	}

	if headerHandler.GetNonce() != currentBlockHeader.GetNonce()+1 {
		log.Info(fmt.Sprintf("nonce not match: local block nonce is %d and node received block with nonce %d\n",
			currentBlockHeader.GetNonce(), headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, currentBlockHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(headerHandler.GetPrevRandSeed(), currentBlockHeader.GetRandSeed()) {
		log.Info(fmt.Sprintf("random seed not match: local block random seed is %s and node received block with previous random seed %s\n",
			core.ToB64(currentBlockHeader.GetRandSeed()), core.ToB64(headerHandler.GetPrevRandSeed())))

		return process.ErrRandSeedMismatch
	}

	if !bytes.Equal(headerHandler.GetPrevHash(), prevHeaderHash) {
		log.Info(fmt.Sprintf("hash not match: local block hash is %s and node received block with previous hash %s\n",
			core.ToB64(prevHeaderHash), core.ToB64(headerHandler.GetPrevHash())))

		return process.ErrBlockHashDoesNotMatch
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
			return process.ErrRootStateMissmatch
		}
		return nil
	}

	//TODO: add verification if rand seed was correctly computed add other verification
	//TODO: check here if the 2 header blocks were correctly signed and the consensus group was correctly elected
	if prevHdr.GetRound() >= currHdr.GetRound() {
		return process.ErrLowerRoundInOtherChainBlock
	}

	if currHdr.GetNonce() != prevHdr.GetNonce()+1 {
		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(bp.marshalizer, bp.hasher, prevHdr)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHdr.GetPrevRandSeed(), prevHdr.GetRandSeed()) {
		return process.ErrRandSeedMismatch
	}

	if !bytes.Equal(currHdr.GetPrevHash(), prevHeaderHash) {
		return process.ErrHashDoesNotMatchInOtherChainBlock
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

func (bp *baseProcessor) removeNotarizedHdrsBehindFinal(hdrsToAttestFinality uint32) {
	bp.mutNotarizedHdrs.Lock()
	for shardId := range bp.notarizedHdrs {
		notarizedHdrsCount := uint32(len(bp.notarizedHdrs[shardId]))
		if notarizedHdrsCount > hdrsToAttestFinality {
			finalIndex := notarizedHdrsCount - 1 - hdrsToAttestFinality
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

func (bp *baseProcessor) requestHeadersIfMissing(sortedHdrs []data.HeaderHandler, shardId uint32, maxRound uint64) error {
	prevHdr, err := bp.getLastNotarizedHdr(shardId)
	if err != nil {
		return err
	}

	isLastNotarizedCloseToOurRound := maxRound-prevHdr.GetRound() <= process.MaxHeaderRequestsAllowed
	if len(sortedHdrs) == 0 && isLastNotarizedCloseToOurRound {
		return process.ErrNoSortedHdrsForShard
	}

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
	}

	// ask for headers, if there most probably should be
	if len(missingNonces) == 0 && !isLastNotarizedCloseToOurRound {
		startNonce := prevHdr.GetNonce() + 1
		for nonce := startNonce; nonce < startNonce+process.MaxHeaderRequestsAllowed; nonce++ {
			missingNonces = append(missingNonces, nonce)
		}
	}

	requested := 0
	for _, nonce := range missingNonces {
		// do the request here
		if bp.onRequestHeaderHandlerByNonce == nil {
			return process.ErrNilRequestHeaderHandlerByNonce
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
func checkProcessorNilParameters(
	accounts state.AccountsAdapter,
	forkDetector process.ForkDetector,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
	shardCoordinator sharding.Coordinator,
	uint64Converter typeConverters.Uint64ByteSliceConverter,
) error {

	if accounts == nil || accounts.IsInterfaceNil() {
		return process.ErrNilAccountsAdapter
	}
	if forkDetector == nil || forkDetector.IsInterfaceNil() {
		return process.ErrNilForkDetector
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return process.ErrNilHasher
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return process.ErrNilMarshalizer
	}
	if store == nil || store.IsInterfaceNil() {
		return process.ErrNilStorage
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return process.ErrNilShardCoordinator
	}
	if uint64Converter == nil || uint64Converter.IsInterfaceNil() {
		return process.ErrNilUint64Converter
	}

	return nil
}
