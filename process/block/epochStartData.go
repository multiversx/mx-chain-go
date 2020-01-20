package block

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type epochStartData struct {
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	store             dataRetriever.StorageService
	dataPool          dataRetriever.PoolsHolder
	blockTracker      process.BlockTracker
	shardCoordinator  sharding.Coordinator
	epochStartTrigger process.EpochStartTriggerHandler
}

// ArgsNewEpochStartDataCreator defines the input parameters for epoch start data creator
type ArgsNewEpochStartDataCreator struct {
	Marshalizer       marshal.Marshalizer
	Hasher            hashing.Hasher
	Store             dataRetriever.StorageService
	DataPool          dataRetriever.PoolsHolder
	BlockTracker      process.BlockTracker
	ShardCoordinator  sharding.Coordinator
	EpochStartTrigger process.EpochStartTriggerHandler
}

// NewEpochStartDataCreator creates a new epoch start creator
func NewEpochStartDataCreator(args ArgsNewEpochStartDataCreator) (*epochStartData, error) {
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

	e := &epochStartData{
		marshalizer:       args.Marshalizer,
		hasher:            args.Hasher,
		store:             args.Store,
		dataPool:          args.DataPool,
		blockTracker:      args.BlockTracker,
		shardCoordinator:  args.ShardCoordinator,
		epochStartTrigger: args.EpochStartTrigger,
	}

	return e, nil
}

// VerifyEpochStartDataForMetablock verifies if epoch start data given by leader is the same as the one should be created
func (e *epochStartData) VerifyEpochStartDataForMetablock(metaBlock *block.MetaBlock) error {
	if !metaBlock.IsStartOfEpochBlock() {
		return nil
	}

	startData, err := e.CreateEpochStartForMetablock()
	if err != nil {
		return err
	}

	receivedEpochStartHash, err := core.CalculateHash(e.marshalizer, e.hasher, metaBlock.EpochStart)
	if err != nil {
		return err
	}

	createdEpochStartHash, err := core.CalculateHash(e.marshalizer, e.hasher, *startData)
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
			"shardID", shardData.ShardId,
			"num pending miniblocks", len(shardData.PendingMiniBlockHeaders),
			"first pending meta", shardData.FirstPendingMetaBlock,
			"last finished meta", shardData.LastFinishedMetaBlock,
			"rootHash", shardData.RootHash,
			"headerHash", shardData.HeaderHash)
	}
}

// CreateEpochStartForMetablock creates epoch start data if it is needed
func (e *epochStartData) CreateEpochStartForMetablock() (*block.EpochStart, error) {
	if !e.epochStartTrigger.IsEpochStart() {
		return &block.EpochStart{}, nil
	}

	startData, allShardHdrList, err := e.getLastNotarizedAndFinalizedHeaders()
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

func (e *epochStartData) getLastNotarizedAndFinalizedHeaders() (*block.EpochStart, [][]*block.Header, error) {
	startData := &block.EpochStart{
		LastFinalizedHeaders: make([]block.EpochStartShardData, 0),
	}

	allShardHdrList := make([][]*block.Header, e.shardCoordinator.NumberOfShards())
	for shardID := uint32(0); shardID < e.shardCoordinator.NumberOfShards(); shardID++ {
		lastCrossNotarizedHeaderForShard, _, err := e.blockTracker.GetLastCrossNotarizedHeader(shardID)
		if err != nil {
			return nil, nil, err
		}

		shardHeader, ok := lastCrossNotarizedHeaderForShard.(*block.Header)
		if !ok {
			return nil, nil, process.ErrWrongTypeAssertion
		}

		hdrHash, err := core.CalculateHash(e.marshalizer, e.hasher, lastCrossNotarizedHeaderForShard)
		if err != nil {
			return nil, nil, err
		}

		lastMetaHash, lastFinalizedMetaHash, currShardHdrList, err := e.getLastFinalizedMetaHashForShard(shardHeader)
		if err != nil {
			return nil, nil, err
		}

		finalHeader := block.EpochStartShardData{
			ShardId:               lastCrossNotarizedHeaderForShard.GetShardID(),
			HeaderHash:            hdrHash,
			RootHash:              lastCrossNotarizedHeaderForShard.GetRootHash(),
			FirstPendingMetaBlock: lastMetaHash,
			LastFinishedMetaBlock: lastFinalizedMetaHash,
		}

		startData.LastFinalizedHeaders = append(startData.LastFinalizedHeaders, finalHeader)
		allShardHdrList[shardID] = currShardHdrList
	}

	return startData, allShardHdrList, nil
}

func (e *epochStartData) getLastFinalizedMetaHashForShard(shardHdr *block.Header) ([]byte, []byte, []*block.Header, error) {
	var lastMetaHash []byte
	var lastFinalizedMetaHash []byte

	shardHdrList := make([]*block.Header, 0)
	shardHdrList = append(shardHdrList, shardHdr)

	for currentHdr := shardHdr; currentHdr.GetNonce() > 0 && currentHdr.GetEpoch() == shardHdr.GetEpoch(); {
		prevShardHdr, err := process.GetShardHeader(currentHdr.GetPrevHash(), e.dataPool.Headers(), e.marshalizer, e.store)
		if err != nil {
			return nil, nil, nil, err
		}

		if len(currentHdr.MetaBlockHashes) == 0 {
			currentHdr = prevShardHdr
			continue
		}

		shardHdrList = append(shardHdrList, prevShardHdr)
		numAddedMetas := len(currentHdr.MetaBlockHashes)
		if numAddedMetas > 1 {
			if len(lastMetaHash) == 0 {
				lastMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
				lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-2]
				return lastMetaHash, lastFinalizedMetaHash, shardHdrList, nil
			}

			if bytes.Equal(lastMetaHash, currentHdr.MetaBlockHashes[numAddedMetas-1]) {
				lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-2]
				return lastMetaHash, lastFinalizedMetaHash, shardHdrList, nil
			}

			lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
			return lastMetaHash, lastFinalizedMetaHash, shardHdrList, nil
		}

		if len(lastMetaHash) == 0 {
			lastMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
			currentHdr = prevShardHdr
			continue
		}

		lastFinalizedMetaHash = currentHdr.MetaBlockHashes[numAddedMetas-1]
		if !bytes.Equal(lastMetaHash, lastFinalizedMetaHash) {
			return lastMetaHash, lastFinalizedMetaHash, shardHdrList, nil
		}

		currentHdr = prevShardHdr
	}

	lastMetaHash, lastFinalizedMetaHash, err := e.getShardDataFromEpochStartData(shardHdr.Epoch, shardHdr.ShardId, lastMetaHash)
	if err != nil {
		return nil, nil, nil, err
	}

	return lastMetaHash, lastFinalizedMetaHash, shardHdrList, nil
}

func (e *epochStartData) getShardDataFromEpochStartData(
	epoch uint32,
	shId uint32,
	lastMetaHash []byte,
) ([]byte, []byte, error) {

	epochStartIdentifier := core.EpochStartIdentifier(epoch)
	if epoch == 0 {
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
		if shardData.ShardId != shId {
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
) ([]block.ShardMiniBlockHeader, error) {
	allPending := make([]block.ShardMiniBlockHeader, 0)
	for shId, shardData := range startData.LastFinalizedHeaders {
		metaHdr, err := e.getMetaBlockByHash(shardData.FirstPendingMetaBlock)
		if err != nil {
			return nil, err
		}
		allMiniBlockHeaders := getAllMiniBlocksWithDst(metaHdr, uint32(shId))
		stillPending := e.computeStillPending(allShardHdrList[shId], allMiniBlockHeaders)
		allPending = append(allPending, stillPending...)
	}

	return allPending, nil
}

func (e *epochStartData) computeStillPending(
	shardHdrs []*block.Header,
	miniBlockHeaders map[string]block.ShardMiniBlockHeader,
) []block.ShardMiniBlockHeader {

	pendingMiniBlocks := make([]block.ShardMiniBlockHeader, 0)

	for _, shardHdr := range shardHdrs {
		for _, mbHeader := range shardHdr.MiniBlockHeaders {
			delete(miniBlockHeaders, string(mbHeader.Hash))
		}
	}

	for _, mbHeader := range miniBlockHeaders {
		pendingMiniBlocks = append(pendingMiniBlocks, mbHeader)
	}

	return pendingMiniBlocks
}

func getAllMiniBlocksWithDst(m *block.MetaBlock, destId uint32) map[string]block.ShardMiniBlockHeader {
	hashDst := make(map[string]block.ShardMiniBlockHeader)
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
			shardMiniBlockHdr := block.ShardMiniBlockHeader{
				Hash:            val.Hash,
				ReceiverShardID: val.ReceiverShardID,
				SenderShardID:   val.SenderShardID,
				TxCount:         val.TxCount,
			}
			hashDst[string(val.Hash)] = shardMiniBlockHdr
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
