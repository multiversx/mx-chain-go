package block

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/display"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

var shardMBHeaderCounterMutex = sync.RWMutex{}
var shardMBHeadersCurrentBlockProcessed = 0
var shardMBHeadersTotalProcessed = 0

const maxHeadersInBlock = 256

// metaProcessor implements metaProcessor interface and actually it tries to execute block
type metaProcessor struct {
	*baseProcessor

	OnRequestHeader           func(shardId uint32, mbHash []byte)
	requestedHeadersHashes    map[string]bool
	mutRequestedHeadersHahses sync.RWMutex

	ChRcvAllHeaders chan bool
}

// NewMetaProcessor creates a new metaProcessor object
func NewMetaProcessor(
	accounts state.AccountsAdapter,
	dataPool dataRetriever.PoolsHolder,
	forkDetector process.ForkDetector,
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	store dataRetriever.StorageService,
	requestHeaderHandler func(shardId uint32, txHash []byte),
) (*metaProcessor, error) {

	if requestHeaderHandler == nil {
		return nil, process.ErrNilRequestHeaderHandler
	}

	err := checkProcessorNilParameters(
		accounts,
		dataPool,
		forkDetector,
		hasher,
		marshalizer,
		store)
	if err != nil {
		return nil, err
	}

	base := &baseProcessor{
		accounts:     accounts,
		dataPool:     dataPool,
		forkDetector: forkDetector,
		hasher:       hasher,
		marshalizer:  marshalizer,
		store:        store,
	}

	bp := metaProcessor{
		baseProcessor: base,
	}

	bp.requestedHeadersHashes = make(map[string]bool)

	bp.OnRequestHeader = requestHeaderHandler
	headerPool := bp.dataPool.Headers()
	if headerPool == nil {
		return nil, process.ErrNilHeadersDataPool
	}
	headerPool.RegisterHandler(bp.receivedHeader)

	bp.ChRcvAllHeaders = make(chan bool)

	return &bp, nil
}

// RevertAccountState reverts the account state for cleanup failed process
func (mp *metaProcessor) RevertAccountState() {
	err := mp.accounts.RevertToSnapshot(0)
	if err != nil {
		log.Error(err.Error())
	}
}

// ProcessBlock processes a block. It returns nil if all ok or the specific error
func (mp *metaProcessor) ProcessBlock(
	chainHandler data.ChainHandler,
	headerHandler data.HeaderHandler,
	bodyHandler data.BodyHandler,
	haveTime func() time.Duration,
) error {
	err := checkForNils(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	header, ok := headerHandler.(*block.MetaBlock)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	err = mp.checkBlockValidity(chainHandler, headerHandler, bodyHandler)
	if err != nil {
		return err
	}

	requestedBlockHeaders := mp.requestBlockHeaders(header)

	if haveTime() < 0 {
		return process.ErrTimeIsOut
	}

	if requestedBlockHeaders > 0 {
		log.Info(fmt.Sprintf("requested %d missing block headers\n", requestedBlockHeaders))
		err = mp.waitForBlockHeaders(haveTime())
		log.Info(fmt.Sprintf("received %d missing block headers\n", requestedBlockHeaders-len(mp.requestedHeadersHashes)))
		if err != nil {
			return err
		}
	}

	if mp.accounts.JournalLen() != 0 {
		return process.ErrAccountStateDirty
	}

	defer func() {
		if err != nil {
			mp.RevertAccountState()
		}
	}()

	err = mp.processBlockHeaders(header, int32(header.Round), haveTime)
	if err != nil {
		return err
	}

	if !mp.verifyStateRoot(header.GetRootHash()) {
		err = process.ErrRootStateMissmatch
		return err
	}

	return nil
}

// RestoreBlockIntoPools restores the block into associated pools
func (mp *metaProcessor) RestoreBlockIntoPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error {
	if headerHandler == nil {
		return process.ErrNilMetaBlockHeader
	}
	return nil
}

// removeBlockInfoFromPool removes the block info from associated pools
func (mp *metaProcessor) removeBlockInfoFromPool(header *block.MetaBlock) error {
	return nil
}

// CreateBlockBody creates block body of metachain
func (mp *metaProcessor) CreateBlockBody(round int32, haveTime func() bool) (data.BodyHandler, error) {
	return &block.MetaBlockBody{}, nil
}

// CreateGenesisBlock creates the genesis block body from map of account balances
func (mp *metaProcessor) CreateGenesisBlock(balances map[string]*big.Int) (rootHash []byte, err error) {
	return []byte("metachain genesis block root hash"), nil
}

func (mp *metaProcessor) processBlockHeaders(header *block.MetaBlock, round int32, haveTime func() time.Duration) error {
	hdrPool := mp.dataPool.Headers()

	msg := "The following miniblock hashes weere successfully processed: "

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		for j := 0; j < len(shardData.ShardMiniBlockHeaders); j++ {
			if haveTime() < 0 {
				return process.ErrTimeIsOut
			}

			headerHash := shardData.HeaderHash
			shardMiniBlockHeader := &shardData.ShardMiniBlockHeaders[j]
			err := mp.processAndRemoveBadShardMiniBlockHeader(
				headerHash,
				shardMiniBlockHeader,
				hdrPool,
				round,
				shardData.ShardId,
			)
			if err != nil {
				return err
			}

			msg = fmt.Sprintf("%s\n%s", msg, toB64(shardMiniBlockHeader.Hash))
		}
	}

	log.Info(fmt.Sprintf("%s\n", msg))

	return nil
}

func (mp *metaProcessor) processAndRemoveBadShardMiniBlockHeader(
	headerHash []byte,
	shardMiniBlockHeader *block.ShardMiniBlockHeader,
	hdrPool storage.Cacher,
	round int32,
	shardId uint32,
) error {
	if hdrPool == nil {
		return process.ErrNilHeadersDataPool
	}
	// TODO: Here should be add shard miniblocks processing mechanism which should change the state of metachain
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

	buff, err := mp.marshalizer.Marshal(headerHandler)
	if err != nil {
		return err
	}

	headerHash := mp.hasher.Compute(string(buff))
	err = mp.store.Put(dataRetriever.MetaBlockUnit, headerHash, buff)
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

	for i := 0; i < len(header.ShardInfo); i++ {
		buff, err = mp.marshalizer.Marshal(header.ShardInfo[i])
		if err != nil {
			return err
		}

		shardDataHash := mp.hasher.Compute(string(buff))
		err = mp.store.Put(dataRetriever.MetaShardDataUnit, shardDataHash, buff)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(header.PeerInfo); i++ {
		buff, err = mp.marshalizer.Marshal(header.PeerInfo[i])
		if err != nil {
			return err
		}

		peerDataHash := mp.hasher.Compute(string(buff))
		err = mp.store.Put(dataRetriever.MetaPeerDataUnit, peerDataHash, buff)
		if err != nil {
			return err
		}
	}

	headerNoncePool := mp.dataPool.HeadersNonces()
	if headerNoncePool == nil {
		err = process.ErrNilDataPoolHolder
		return err
	}

	_ = headerNoncePool.Put(headerHandler.GetNonce(), headerHash)

	_, err = mp.accounts.Commit()
	if err != nil {
		return err
	}

	errNotCritical := mp.removeBlockInfoFromPool(header)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
	}

	errNotCritical = mp.forkDetector.AddHeader(header, headerHash, true)
	if errNotCritical != nil {
		log.Info(errNotCritical.Error())
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

	// write data to log
	go mp.displayMetaBlock(header)

	return nil
}

// receivedHeader is a call back function which is called when a new header
// is added in the headers pool
func (mp *metaProcessor) receivedHeader(headerHash []byte) {
	mp.mutRequestedHeadersHahses.Lock()
	if len(mp.requestedHeadersHashes) > 0 {
		if mp.requestedHeadersHashes[string(headerHash)] {
			delete(mp.requestedHeadersHashes, string(headerHash))
		}
		lenReqHeadersHashes := len(mp.requestedHeadersHashes)
		mp.mutRequestedHeadersHahses.Unlock()

		if lenReqHeadersHashes == 0 {
			mp.ChRcvAllHeaders <- true
		}
		return
	}
	mp.mutRequestedHeadersHahses.Unlock()
}

func (mp *metaProcessor) requestBlockHeaders(header *block.MetaBlock) int {
	mp.mutRequestedHeadersHahses.Lock()
	requestedHeaders := 0
	missingHeaderHashes := mp.computeMissingHeaders(header)
	mp.requestedHeadersHashes = make(map[string]bool)
	if mp.OnRequestHeader != nil {
		for shardId, headerHash := range missingHeaderHashes {
			requestedHeaders++
			mp.requestedHeadersHashes[string(headerHash)] = true
			go mp.OnRequestHeader(shardId, headerHash)
		}
	}
	mp.mutRequestedHeadersHahses.Unlock()
	return requestedHeaders
}

func (mp *metaProcessor) computeMissingHeaders(header *block.MetaBlock) map[uint32][]byte {
	missingHeaders := make(map[uint32][]byte)
	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]
		header := mp.getHeaderFromPool(shardData.ShardId, shardData.HeaderHash)
		if header == nil {
			missingHeaders[shardData.ShardId] = shardData.HeaderHash
		}
	}
	return missingHeaders
}

// getHeaderFromPool gets the header from a given shard id and a given header hash
func (mp *metaProcessor) getHeaderFromPool(shardID uint32, headerHash []byte) *block.Header {
	headerPool := mp.dataPool.Headers()
	if headerPool == nil {
		log.Error(process.ErrNilHeadersDataPool.Error())
		return nil
	}

	val, ok := headerPool.Peek(headerHash)
	if !ok {
		return nil
	}

	header, ok := val.(*block.Header)
	if !ok {
		return nil
	}

	return header
}

// CreateBlockHeader creates a miniblock header list given a block body
func (mp *metaProcessor) CreateBlockHeader(bodyHandler data.BodyHandler, round int32, haveTime func() bool) (data.HeaderHandler, error) {
	header := &block.MetaBlock{}

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

	return header, nil
}

func (mp *metaProcessor) createShardInfo(
	maxHdrInBlock int,
	round int32,
	haveTime func() bool,
) ([]block.ShardData, error) {
	shardInfo := make([]block.ShardData, 0)

	if mp.accounts.JournalLen() != 0 {
		return nil, process.ErrAccountStateDirty
	}

	if !haveTime() {
		log.Info(fmt.Sprintf("time is up after entered in createShardInfo method\n"))
		return shardInfo, nil
	}

	hdrPool := mp.dataPool.Headers()
	if hdrPool == nil {
		return nil, process.ErrNilHeadersDataPool
	}

	hdrs := uint32(0)

	timeBefore := time.Now()
	orderedHdrs, orderedHdrHashes, err := getHdrs(hdrPool)
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

	for index := range orderedHdrs {
		shardData := block.ShardData{}
		shardData.ShardMiniBlockHeaders = make([]block.ShardMiniBlockHeader, 0)
		shardData.TxCount = orderedHdrs[index].TxCount
		shardData.ShardId = orderedHdrs[index].ShardId
		shardData.HeaderHash = orderedHdrHashes[index]

		for i := 0; i < len(orderedHdrs[index].MiniBlockHeaders); i++ {
			if !haveTime() {
				break
			}

			shardMiniBlockHeader := block.ShardMiniBlockHeader{}
			shardMiniBlockHeader.SenderShardId = orderedHdrs[index].MiniBlockHeaders[i].SenderShardID
			shardMiniBlockHeader.ReceiverShardId = orderedHdrs[index].MiniBlockHeaders[i].ReceiverShardID
			shardMiniBlockHeader.Hash = orderedHdrs[index].MiniBlockHeaders[i].Hash
			shardMiniBlockHeader.TxCount = orderedHdrs[index].MiniBlockHeaders[i].TxCount

			snapshot := mp.accounts.JournalLen()

			// execute shard miniblock to change the trie root hash
			err := mp.processAndRemoveBadShardMiniBlockHeader(
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
			hdrs++

			if hdrs >= uint32(maxHdrInBlock) { // max transactions count in one block was reached
				log.Info(fmt.Sprintf("max hdrs accepted in one block is reached: added %d hdrs from %d hdrs\n", hdrs, len(orderedHdrs)))

				if len(shardData.ShardMiniBlockHeaders) == len(orderedHdrs[index].MiniBlockHeaders) {
					shardInfo = append(shardInfo, shardData)
				}

				log.Info(fmt.Sprintf("creating shard info has been finished: created %d shard data\n", len(shardInfo)))
				return shardInfo, nil
			}
		}

		if !haveTime() {
			log.Info(fmt.Sprintf("time is up: added %d hdrs from %d hdrs\n", hdrs, len(orderedHdrs)))

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

func (mp *metaProcessor) waitForBlockHeaders(waitTime time.Duration) error {
	select {
	case <-mp.ChRcvAllHeaders:
		return nil
	case <-time.After(waitTime):
		return process.ErrTimeIsOut
	}
}

func (mp *metaProcessor) getHeadersInPool() int {
	headerPool := mp.dataPool.Headers()
	if headerPool == nil {
		log.Error(process.ErrNilHeadersDataPool.Error())
		return -1
	}

	return headerPool.Len()
}

func getHdrs(hdrStore storage.Cacher) ([]*block.Header, [][]byte, error) {
	if hdrStore == nil {
		return nil, nil, process.ErrNilCacher
	}

	headers := make([]*block.Header, 0)
	hdrHashes := make([][]byte, 0)

	for _, key := range hdrStore.Keys() {
		val, _ := hdrStore.Peek(key)
		if val == nil {
			continue
		}

		hdr, ok := val.(*block.Header)
		if !ok {
			continue
		}

		hdrHashes = append(hdrHashes, key)
		headers = append(headers, hdr)
	}

	return headers, hdrHashes, nil
}

// MarshalizedDataForCrossShard prepares underlying data into a marshalized object according to destination
func (mp *metaProcessor) MarshalizedDataForCrossShard(
	bodyHandler data.BodyHandler,
) (map[uint32][]byte, map[uint32][][]byte, error) {
	mrsData := make(map[uint32][]byte)
	mrsTxs := make(map[uint32][][]byte)
	return mrsData, mrsTxs, nil
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

func (mp *metaProcessor) displayMetaBlock(header *block.MetaBlock) {
	if header == nil {
		return
	}

	headerHash, err := mp.computeHeaderHash(header)
	if err != nil {
		log.Error(err.Error())
		return
	}

	dispHeader := []string{"Part", "Parameter", "Value"}
	dispLines := displayHeader(header)

	metaLines := make([]*display.LineData, 0)
	metaLines = append(metaLines, display.NewLineData(false, []string{
		"Header",
		"Block type",
		"MetaBlock"}))
	metaLines = append(metaLines, dispLines...)

	metaLines = displayShardInfo(metaLines, header)

	tblString, err := display.CreateTableString(dispHeader, metaLines)
	if err != nil {
		log.Error(err.Error())
		return
	}

	shardMBHeaderCounterMutex.Lock()
	tblString = tblString + fmt.Sprintf("\nHeader hash: %s\n\nTotal shard MB headers "+
		"processed until now: %d. Total shard MB headers processed for this block: %d. Total shard headers remained in pool: %d\n",
		toB64(headerHash),
		shardMBHeadersTotalProcessed,
		shardMBHeadersCurrentBlockProcessed,
		mp.getHeadersInPool())
	shardMBHeaderCounterMutex.Unlock()

	log.Info(tblString)
}

func displayShardInfo(lines []*display.LineData, header *block.MetaBlock) []*display.LineData {
	shardMBHeaderCounterMutex.RLock()
	shardMBHeadersCurrentBlockProcessed = 0
	shardMBHeaderCounterMutex.RUnlock()

	for i := 0; i < len(header.ShardInfo); i++ {
		shardData := header.ShardInfo[i]

		part := fmt.Sprintf("ShardData_%d", shardData.ShardId)

		if shardData.ShardMiniBlockHeaders == nil || len(shardData.ShardMiniBlockHeaders) == 0 {
			lines = append(lines, display.NewLineData(false, []string{
				part, "", "<NIL> or <EMPTY>"}))
		}

		shardMBHeaderCounterMutex.Lock()
		shardMBHeadersCurrentBlockProcessed += len(shardData.ShardMiniBlockHeaders)
		shardMBHeadersTotalProcessed += len(shardData.ShardMiniBlockHeaders)
		shardMBHeaderCounterMutex.Unlock()

		for j := 0; j < len(shardData.ShardMiniBlockHeaders); j++ {
			if j == 0 || j >= len(shardData.ShardMiniBlockHeaders)-1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("ShardMiniBlockHeaderHash %d", j+1),
					toB64(shardData.ShardMiniBlockHeaders[j].Hash)}))

				part = ""
			} else if j == 1 {
				lines = append(lines, display.NewLineData(false, []string{
					part,
					fmt.Sprintf("..."),
					fmt.Sprintf("...")}))

				part = ""
			}
		}

		lines[len(lines)-1].HorizontalRuleAfter = true
	}

	return lines
}
