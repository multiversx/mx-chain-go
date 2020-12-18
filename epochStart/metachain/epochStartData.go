package metachain

import (
	"bytes"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.EpochStartDataCreator = (*epochStartData)(nil)

type epochStartData struct {
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	store             dataRetriever.StorageService
	dataPool          dataRetriever.PoolsHolder
	blockTracker      process.BlockTracker
	shardCoordinator  sharding.Coordinator
	epochStartTrigger process.EpochStartTriggerHandler
	requestHandler    epochStart.RequestHandler
	genesisEpoch      uint32
}

// ArgsNewEpochStartData defines the input parameters for epoch start data creator
type ArgsNewEpochStartData struct {
	Marshalizer       marshal.Marshalizer
	Hasher            hashing.Hasher
	Store             dataRetriever.StorageService
	DataPool          dataRetriever.PoolsHolder
	BlockTracker      process.BlockTracker
	ShardCoordinator  sharding.Coordinator
	EpochStartTrigger process.EpochStartTriggerHandler
	RequestHandler    epochStart.RequestHandler
	GenesisEpoch      uint32
}

// NewEpochStartData creates a new epoch start creator
func NewEpochStartData(args ArgsNewEpochStartData) (*epochStartData, error) {
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Store) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(args.DataPool) {
		return nil, process.ErrNilDataPoolHolder
	}
	if check.IfNil(args.BlockTracker) {
		return nil, process.ErrNilBlockTracker
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.RequestHandler) {
		return nil, process.ErrNilRequestHandler
	}

	e := &epochStartData{
		marshalizer:       args.Marshalizer,
		hasher:            args.Hasher,
		store:             args.Store,
		dataPool:          args.DataPool,
		blockTracker:      args.BlockTracker,
		shardCoordinator:  args.ShardCoordinator,
		epochStartTrigger: args.EpochStartTrigger,
		requestHandler:    args.RequestHandler,
		genesisEpoch:      args.GenesisEpoch,
	}

	return e, nil
}

// VerifyEpochStartDataForMetablock verifies if epoch start data given by leader is the same as the one should be created
func (e *epochStartData) VerifyEpochStartDataForMetablock(metaBlock *block.MetaBlock) error {
	if !metaBlock.IsStartOfEpochBlock() {
		return nil
	}

	startData, err := e.CreateEpochStartData()
	if err != nil {
		return err
	}

	epochStartDataWithoutEconomics := metaBlock.EpochStart
	epochStartDataWithoutEconomics.Economics = block.Economics{}
	receivedEpochStartHash, err := core.CalculateHash(e.marshalizer, e.hasher, &epochStartDataWithoutEconomics)
	if err != nil {
		return err
	}

	createdEpochStartHash, err := core.CalculateHash(e.marshalizer, e.hasher, startData)
	if err != nil {
		return err
	}

	if !bytes.Equal(receivedEpochStartHash, createdEpochStartHash) {
		log.Warn("RECEIVED epoch start data")
		displayEpochStartData(&metaBlock.EpochStart)

		log.Warn("CREATED epoch start data")
		displayEpochStartData(startData)

		return process.ErrEpochStartDataDoesNotMatch
	}

	return nil
}

func displayEpochStartData(startData *block.EpochStart) {
	for _, shardData := range startData.LastFinalizedHeaders {
		log.Debug("epoch start shard data",
			"shardID", shardData.ShardID,
			"num pending miniblocks", len(shardData.PendingMiniBlockHeaders),
			"first pending meta", shardData.FirstPendingMetaBlock,
			"last finished meta", shardData.LastFinishedMetaBlock,
			"rootHash", shardData.RootHash,
			"headerHash", shardData.HeaderHash)
	}
}

// CreateEpochStartData creates epoch start data if it is needed
func (e *epochStartData) CreateEpochStartData() (*block.EpochStart, error) {
	if !e.epochStartTrigger.IsEpochStart() {
		return &block.EpochStart{}, nil
	}

	startData, allShardHdrList, err := e.createShardStartDataAndLastProcessedHeaders()
	if err != nil {
		return nil, err
	}

	pendingMiniBlocks, err := e.computePendingMiniBlockList(startData, allShardHdrList)
	if err != nil {
		return nil, err
	}

	for _, pendingMiniBlock := range pendingMiniBlocks {
		recvShId := pendingMiniBlock.ReceiverShardID

		startData.LastFinalizedHeaders[recvShId].PendingMiniBlockHeaders =
			append(startData.LastFinalizedHeaders[recvShId].PendingMiniBlockHeaders, pendingMiniBlock)
	}

	return startData, nil
}

func (e *epochStartData) createShardStartDataAndLastProcessedHeaders() (*block.EpochStart, [][]*block.Header, error) {
	startData := &block.EpochStart{
		LastFinalizedHeaders: make([]block.EpochStartShardData, 0),
	}

	allShardHdrList := make([][]*block.Header, e.shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < e.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeaderForShard, hdrHash, err := e.blockTracker.GetLastCrossNotarizedHeader(shardID)
		if err != nil {
			return nil, nil, err
		}

		shardHeader, ok := lastCrossNotarizedHeaderForShard.(*block.Header)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		firstPendingMetaHash, lastFinalizedMetaHash, currShardHdrList, err := e.lastFinalizedFirstPendingListHeadersForShard(shardHeader)
		if err != nil {
			return nil, nil, err
		}

		finalHeader := block.EpochStartShardData{
			ShardID:               lastCrossNotarizedHeaderForShard.GetShardID(),
			Epoch:                 lastCrossNotarizedHeaderForShard.GetEpoch(),
			Round:                 lastCrossNotarizedHeaderForShard.GetRound(),
			Nonce:                 lastCrossNotarizedHeaderForShard.GetNonce(),
			HeaderHash:            hdrHash,
			RootHash:              lastCrossNotarizedHeaderForShard.GetRootHash(),
			FirstPendingMetaBlock: firstPendingMetaHash,
			LastFinishedMetaBlock: lastFinalizedMetaHash,
		}

		startData.LastFinalizedHeaders = append(startData.LastFinalizedHeaders, finalHeader)
		allShardHdrList[shardID] = currShardHdrList
	}

	return startData, allShardHdrList, nil
}

func (e *epochStartData) lastFinalizedFirstPendingListHeadersForShard(shardHdr *block.Header) ([]byte, []byte, []*block.Header, error) {
	var firstPendingMetaHash []byte
	var lastFinalizedMetaHash []byte

	shardHdrList := make([]*block.Header, 0)
	shardHdrList = append(shardHdrList, shardHdr)

	for currentHdr := shardHdr; currentHdr.GetNonce() > 0 && currentHdr.GetEpoch() == shardHdr.GetEpoch(); {
		prevShardHdr, err := process.GetShardHeader(currentHdr.GetPrevHash(), e.dataPool.Headers(), e.marshalizer, e.store)
		if err != nil {
			go e.requestHandler.RequestShardHeader(currentHdr.ShardID, currentHdr.GetPrevHash())
			log.Warn("shard remained in an epoch that is too old",
				"shardID", currentHdr.ShardID,
				"shard Epoch", currentHdr.Epoch,
				"meta Epoch", e.epochStartTrigger.MetaEpoch(),
				"err", err)
			break
		}

		shardHdrList = append(shardHdrList, prevShardHdr)
		if len(currentHdr.MetaBlockHashes) == 0 {
			currentHdr = prevShardHdr
			continue
		}

		numAddedMetas := len(currentHdr.MetaBlockHashes)
		if numAddedMetas > 1 {
			if len(firstPendingMetaHash) == 0 {
				firstPendingMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
				lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-2]
				return firstPendingMetaHash, lastFinalizedMetaHash, shardHdrList, nil
			}

			if bytes.Equal(firstPendingMetaHash, currentHdr.MetaBlockHashes[numAddedMetas-1]) {
				lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-2]
				return firstPendingMetaHash, lastFinalizedMetaHash, shardHdrList, nil
			}

			lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
			return firstPendingMetaHash, lastFinalizedMetaHash, shardHdrList, nil
		}

		if len(firstPendingMetaHash) == 0 {
			firstPendingMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
			currentHdr = prevShardHdr
			continue
		}

		lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
		if !bytes.Equal(firstPendingMetaHash, lastFinalizedMetaHash) {
			return firstPendingMetaHash, lastFinalizedMetaHash, shardHdrList, nil
		}

		currentHdr = prevShardHdr
	}

	firstPendingMetaHash, lastFinalizedMetaHash, err := e.getShardDataFromEpochStartData(shardHdr.ShardID, firstPendingMetaHash)
	if err != nil {
		return nil, nil, nil, err
	}

	return firstPendingMetaHash, lastFinalizedMetaHash, shardHdrList, nil
}

func (e *epochStartData) getShardDataFromEpochStartData(
	shId uint32,
	lastMetaHash []byte,
) ([]byte, []byte, error) {

	prevEpoch := e.genesisEpoch
	if e.epochStartTrigger.Epoch() > e.genesisEpoch {
		prevEpoch = e.epochStartTrigger.Epoch() - 1
	}

	epochStartIdentifier := core.EpochStartIdentifier(prevEpoch)
	if prevEpoch == 0 {
		return lastMetaHash, []byte(epochStartIdentifier), nil
	}

	previousEpochStartMeta, err := process.GetMetaHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)
	if err != nil {
		return nil, nil, err
	}

	if !previousEpochStartMeta.IsStartOfEpochBlock() {
		return nil, nil, process.ErrNotEpochStartBlock
	}

	for _, shardData := range previousEpochStartMeta.EpochStart.LastFinalizedHeaders {
		if shardData.ShardID != shId {
			continue
		}

		if len(lastMetaHash) == 0 || bytes.Equal(lastMetaHash, shardData.FirstPendingMetaBlock) {
			return shardData.FirstPendingMetaBlock, shardData.LastFinishedMetaBlock, nil
		}

		return lastMetaHash, shardData.FirstPendingMetaBlock, nil
	}

	return nil, nil, process.ErrGettingShardDataFromEpochStartData
}

func (e *epochStartData) computePendingMiniBlockList(
	startData *block.EpochStart,
	allShardHdrList [][]*block.Header,
) ([]block.MiniBlockHeader, error) {

	prevEpoch := e.genesisEpoch
	if e.epochStartTrigger.Epoch() > e.genesisEpoch {
		prevEpoch = e.epochStartTrigger.Epoch() - 1
	}

	epochStartIdentifier := core.EpochStartIdentifier(prevEpoch)
	previousEpochStartMeta, _ := process.GetMetaHeaderFromStorage([]byte(epochStartIdentifier), e.marshalizer, e.store)

	allPending := make([]block.MiniBlockHeader, 0)
	for shId, shardData := range startData.LastFinalizedHeaders {
		if shardData.Nonce == 0 {
			//shard has only the genesis block
			continue
		}
		if len(shardData.FirstPendingMetaBlock) == 0 {
			continue
		}

		firstPending, mapPendingMiniBlocks := getEpochStartDataForShard(previousEpochStartMeta, uint32(shId))
		if bytes.Equal(firstPending, shardData.FirstPendingMetaBlock) {
			stillPending := e.computeStillPending(uint32(shId), allShardHdrList[shId], mapPendingMiniBlocks)
			allPending = append(allPending, stillPending...)
			continue
		}

		metaHdr, err := e.getMetaBlockByHash(shardData.FirstPendingMetaBlock)
		if err != nil {
			go e.requestHandler.RequestMetaHeader(shardData.FirstPendingMetaBlock)
			return nil, err
		}

		allMiniBlockHeaders := getAllMiniBlocksWithDst(metaHdr, uint32(shId))
		stillPending := e.computeStillPending(uint32(shId), allShardHdrList[shId], allMiniBlockHeaders)
		allPending = append(allPending, stillPending...)
	}

	return allPending, nil
}

func getEpochStartDataForShard(epochStartMetaHdr *block.MetaBlock, shardID uint32) ([]byte, map[string]block.MiniBlockHeader) {
	if check.IfNil(epochStartMetaHdr) {
		return nil, nil
	}

	for _, header := range epochStartMetaHdr.EpochStart.LastFinalizedHeaders {
		if header.ShardID != shardID {
			continue
		}

		mapPendingMiniBlocks := make(map[string]block.MiniBlockHeader, len(header.PendingMiniBlockHeaders))
		for _, mbHdr := range header.PendingMiniBlockHeaders {
			mapPendingMiniBlocks[string(mbHdr.Hash)] = mbHdr
		}

		return header.FirstPendingMetaBlock, mapPendingMiniBlocks
	}

	return nil, nil
}

func (e *epochStartData) computeStillPending(
	shardID uint32,
	shardHdrs []*block.Header,
	miniBlockHeaders map[string]block.MiniBlockHeader,
) []block.MiniBlockHeader {

	pendingMiniBlocks := make([]block.MiniBlockHeader, 0)

	for _, shardHdr := range shardHdrs {
		for _, mbHeader := range shardHdr.MiniBlockHeaders {
			delete(miniBlockHeaders, string(mbHeader.Hash))
		}
	}

	for _, mbHeader := range miniBlockHeaders {
		log.Debug("pending miniblock for shard ", "id", shardID, "hash", mbHeader.Hash)
		pendingMiniBlocks = append(pendingMiniBlocks, mbHeader)
	}

	sort.Slice(pendingMiniBlocks, func(i, j int) bool {
		return bytes.Compare(pendingMiniBlocks[i].Hash, pendingMiniBlocks[j].Hash) < 0
	})

	return pendingMiniBlocks
}

func getAllMiniBlocksWithDst(m *block.MetaBlock, destId uint32) map[string]block.MiniBlockHeader {
	hashDst := make(map[string]block.MiniBlockHeader)
	for i := 0; i < len(m.ShardInfo); i++ {
		if m.ShardInfo[i].ShardID == destId {
			continue
		}

		for _, val := range m.ShardInfo[i].ShardMiniBlockHeaders {
			if val.ReceiverShardID == destId && val.SenderShardID != destId {
				hashDst[string(val.Hash)] = val
			}
		}
	}

	for _, val := range m.MiniBlockHeaders {
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			hashDst[string(val.Hash)] = val
		}
	}

	return hashDst
}

func (e *epochStartData) getMetaBlockByHash(metaHash []byte) (*block.MetaBlock, error) {
	return process.GetMetaHeader(metaHash, e.dataPool.Headers(), e.marshalizer, e.store)
}

// IsInterfaceNil returns true if underlying object is nil
func (e *epochStartData) IsInterfaceNil() bool {
	return e == nil
}
