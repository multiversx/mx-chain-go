package block

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

var log = logger.DefaultLogger()

type hashAndHdr struct {
	hdr  data.HeaderHandler
	hash []byte
}

type mapShardLastHeaders map[uint32]data.HeaderHandler

type baseProcessor struct {
	shardCoordinator sharding.Coordinator
	accounts         state.AccountsAdapter
	forkDetector     process.ForkDetector
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	store            dataRetriever.StorageService

	mutLastNotarizedHdrs sync.RWMutex
	lastNotarizedHdrs    mapShardLastHeaders

	onRequestHeaderHandlerByNonce func(shardId uint32, nonce uint64)
}

func checkForNils(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
) error {

	if chainHandler == nil {
		return process.ErrNilBlockChain
	}
	if headerHandler == nil {
		return process.ErrNilBlockHeader
	}
	if bodyHandler == nil {
		return process.ErrNilBlockBody
	}
	return nil
}

// SetOnRequestHeaderHandlerByNonce sets request handler to ask for missing headers by nonce
func (bp *baseProcessor) SetOnRequestHeaderHandlerByNonce(requestHandler func(shardId uint32, nonce uint64)) error {
	if requestHandler == nil {
		return process.ErrNilRequestHeaderHandlerByNonce
	}
	bp.onRequestHeaderHandlerByNonce = requestHandler
	return nil
}

// RevertAccountState reverts the account state for cleanup failed process
func (bp *baseProcessor) RevertAccountState() {
	err := bp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Error(err.Error())
	}
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

	if chainHandler.GetCurrentBlockHeader() == nil {
		if headerHandler.GetNonce() == 1 { // first block after genesis
			if bytes.Equal(headerHandler.GetPrevHash(), chainHandler.GetGenesisHeaderHash()) {
				// TODO: add genesis block verification
				return nil
			}

			log.Info(fmt.Sprintf("hash not match: local block hash is empty and node received block with previous hash %s\n",
				toB64(headerHandler.GetPrevHash())))

			return process.ErrInvalidBlockHash
		}

		log.Info(fmt.Sprintf("nonce not match: local block nonce is 0 and node received block with nonce %d\n",
			headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	if headerHandler.GetNonce() != chainHandler.GetCurrentBlockHeader().GetNonce()+1 {
		log.Info(fmt.Sprintf("nonce not match: local block nonce is %d and node received block with nonce %d\n",
			chainHandler.GetCurrentBlockHeader().GetNonce(), headerHandler.GetNonce()))

		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := bp.computeHeaderHash(chainHandler.GetCurrentBlockHeader())
	if err != nil {
		return err
	}

	if !bytes.Equal(headerHandler.GetPrevHash(), prevHeaderHash) {
		log.Info(fmt.Sprintf("hash not match: local block hash is %s and node received block with previous hash %s\n",
			toB64(prevHeaderHash), toB64(headerHandler.GetPrevHash())))

		return process.ErrInvalidBlockHash
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
	return bytes.Equal(bp.accounts.RootHash(), rootHash)
}

// getRootHash returns the accounts merkle tree root hash
func (bp *baseProcessor) getRootHash() []byte {
	return bp.accounts.RootHash()
}

func (bp *baseProcessor) computeHeaderHash(headerHandler data.HeaderHandler) ([]byte, error) {
	headerMarsh, err := bp.marshalizer.Marshal(headerHandler)
	if err != nil {
		return nil, err
	}

	headerHash := bp.hasher.Compute(string(headerMarsh))

	return headerHash, nil
}

func (bp *baseProcessor) isHdrConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	if prevHdr == nil {
		return process.ErrNilBlockHeader
	}
	if currHdr == nil {
		return process.ErrNilBlockHeader
	}

	// special case with genesis nonce - 0
	if currHdr.GetNonce() == 0 {
		if prevHdr.GetNonce() != 0 {
			return process.ErrWrongNonceInBlock
		}
		// block with nonce 0 was already saved
		if prevHdr.GetRootHash() != nil {
			return process.ErrWrongNonceInBlock
		}
		return nil
	}

	//TODO: add verification if rand seed was calculated good add other verification
	//TODO: check here if the 2 header blocks were correctly signed and the consensus group was correctly elected
	if prevHdr.GetRound() > currHdr.GetRound() {
		return process.ErrLowShardHeaderRound
	}

	if currHdr.GetNonce() != prevHdr.GetNonce()+1 {
		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := bp.computeHeaderHash(prevHdr)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHdr.GetPrevRandSeed(), prevHdr.GetRandSeed()) {
		return process.ErrRandSeedMismatch
	}

	if !bytes.Equal(currHdr.GetPrevHash(), prevHeaderHash) {
		return process.ErrInvalidBlockHash
	}

	return nil
}

func (bp *baseProcessor) saveLastNotarizedHeader(shardId uint32, processedHdrs []data.HeaderHandler) error {
	bp.mutLastNotarizedHdrs.Lock()
	defer bp.mutLastNotarizedHdrs.Unlock()

	if bp.lastNotarizedHdrs == nil {
		return process.ErrLastNotarizedHdrsSliceIsNil
	}

	hdr := bp.lastNotarizedHdrs[shardId]
	if hdr == nil {
		return process.ErrLastNotarizedHdrsSliceIsNil
	}

	if shardId == sharding.MetachainShardId {
		_, ok := hdr.(*block.MetaBlock)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
	}

	tmpLastNotarized := hdr

	var err error
	defer func() {
		if err != nil {
			bp.lastNotarizedHdrs[shardId] = tmpLastNotarized
		}
	}()

	sort.Slice(processedHdrs, func(i, j int) bool {
		return processedHdrs[i].GetNonce() < processedHdrs[i].GetNonce()
	})

	for i := 0; i < len(processedHdrs); i++ {
		err := bp.isHdrConstructionValid(processedHdrs[i], bp.lastNotarizedHdrs[shardId])
		if err != nil {
			continue
		}
		bp.lastNotarizedHdrs[shardId] = processedHdrs[i]
	}

	return nil
}

func (bp *baseProcessor) getLastNotarizedHdr(shardId uint32) (data.HeaderHandler, error) {
	bp.mutLastNotarizedHdrs.RLock()
	defer bp.mutLastNotarizedHdrs.RUnlock()

	if bp.lastNotarizedHdrs == nil {
		return nil, process.ErrLastNotarizedHdrsSliceIsNil
	}

	hdr := bp.lastNotarizedHdrs[shardId]
	if hdr == nil {
		return nil, process.ErrLastNotarizedHdrsSliceIsNil
	}

	if shardId == sharding.MetachainShardId {
		_, ok := hdr.(*block.MetaBlock)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}
	}

	if shardId < bp.shardCoordinator.NumberOfShards() {
		_, ok := hdr.(*block.Header)
		if !ok {
			return nil, process.ErrWrongTypeAssertion
		}
	}

	return hdr, nil

}

// SetLastNotarizedHeadersSlice sets the headers blocks in lastNotarizedHdrs for every shard
// This is done when starting a new epoch so metachain can use it when validating next shard header blocks
// and shard can validate the next meta header
func (bp *baseProcessor) SetLastNotarizedHeadersSlice(startHeaders map[uint32]data.HeaderHandler, metaChainActive bool) error {
	//TODO: protect this to be called only once at genesis time
	bp.mutLastNotarizedHdrs.Lock()
	defer bp.mutLastNotarizedHdrs.Unlock()

	if startHeaders == nil {
		return process.ErrLastNotarizedHdrsSliceIsNil
	}

	bp.lastNotarizedHdrs = make(mapShardLastHeaders, bp.shardCoordinator.NumberOfShards())
	for i := uint32(0); i < bp.shardCoordinator.NumberOfShards(); i++ {
		hdr, ok := startHeaders[i].(*block.Header)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
		bp.lastNotarizedHdrs[i] = hdr
	}

	if metaChainActive {
		hdr, ok := startHeaders[sharding.MetachainShardId].(*block.MetaBlock)
		if !ok {
			return process.ErrWrongTypeAssertion
		}
		bp.lastNotarizedHdrs[sharding.MetachainShardId] = hdr
	}

	return nil
}

func (bp *baseProcessor) requestHeadersIfMissing(sortedHdrs []data.HeaderHandler, shardId uint32, maxNonce uint64) error {
	prevHdr, err := bp.getLastNotarizedHdr(shardId)
	if err != nil {
		return err
	}

	if len(sortedHdrs) == 0 {
		return process.ErrNoSortedHdrsForShard
	}

	missingNonces := make([]uint64, 0)
	for i := 0; i < len(sortedHdrs); i++ {
		currHdr := sortedHdrs[i]

		if i > 0 {
			prevHdr = sortedHdrs[i-1]
		}

		if currHdr.GetNonce()-prevHdr.GetNonce() > 1 {
			for j := prevHdr.GetNonce(); j < currHdr.GetNonce(); j++ {
				missingNonces = append(missingNonces, j)
			}
		}
	}

	for _, nonce := range missingNonces {
		if nonce > maxNonce {
			return nil
		}

		// do the request here
		if bp.onRequestHeaderHandlerByNonce == nil {
			return process.ErrNilRequestHeaderHandlerByNonce
		}

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
		toB64(headerHandler.GetPrevHash())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Prev rand seed",
		toB64(headerHandler.GetPrevRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Rand seed",
		toB64(headerHandler.GetRandSeed())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Pub keys bitmap",
		toHex(headerHandler.GetPubKeysBitmap())}))
	lines = append(lines, display.NewLineData(false, []string{
		"",
		"Signature",
		toB64(headerHandler.GetSignature())}))
	lines = append(lines, display.NewLineData(true, []string{
		"",
		"Root hash",
		toB64(headerHandler.GetRootHash())}))
	return lines
}

func toHex(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return "0x" + hex.EncodeToString(buff)
}

func toB64(buff []byte) string {
	if buff == nil {
		return "<NIL>"
	}
	return base64.StdEncoding.EncodeToString(buff)
}

// checkProcessorNilParameters will check the imput parameters for nil values
func checkProcessorNilParameters(
	accounts state.AccountsAdapter,
	forkDetector process.ForkDetector,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
) error {

	if accounts == nil {
		return process.ErrNilAccountsAdapter
	}
	if forkDetector == nil {
		return process.ErrNilForkDetector
	}
	if hasher == nil {
		return process.ErrNilHasher
	}
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}
	if store == nil {
		return process.ErrNilStorage
	}

	return nil
}
