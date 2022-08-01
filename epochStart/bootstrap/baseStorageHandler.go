package bootstrap

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
)

type miniBlockInfo struct {
	miniBlockHashes              [][]byte
	fullyProcessed               []bool
	indexOfLastTxProcessed       []int32
	pendingMiniBlocksMap         map[string]struct{}
	pendingMiniBlocksPerShardMap map[uint32][][]byte
}

type processedIndexes struct {
	firstIndex int32
	lastIndex  int32
}

// baseStorageHandler handles the storage functions for saving bootstrap data
type baseStorageHandler struct {
	storageService   dataRetriever.StorageService
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	currentEpoch     uint32
	uint64Converter  typeConverters.Uint64ByteSliceConverter
}

func (bsh *baseStorageHandler) groupMiniBlocksByShard(miniBlocks map[string]*block.MiniBlock) ([]bootstrapStorage.PendingMiniBlocksInfo, error) {
	pendingMBsMap := make(map[uint32][][]byte)
	for hash, miniBlock := range miniBlocks {
		receiverShId := miniBlock.ReceiverShardID // we need the receiver only on meta to properly load the pendingMiniBlocks structure
		pendingMBsMap[receiverShId] = append(pendingMBsMap[receiverShId], []byte(hash))
	}

	sliceToRet := make([]bootstrapStorage.PendingMiniBlocksInfo, 0)
	for shardID, hashes := range pendingMBsMap {
		sliceToRet = append(sliceToRet, bootstrapStorage.PendingMiniBlocksInfo{
			ShardID:          shardID,
			MiniBlocksHashes: hashes,
		})
	}

	return sliceToRet, nil
}

func (bsh *baseStorageHandler) saveNodesCoordinatorRegistry(
	metaBlock data.HeaderHandler,
	nodesConfig *nodesCoordinator.NodesCoordinatorRegistry,
) ([]byte, error) {
	key := append([]byte(common.NodesCoordinatorRegistryKeyPrefix), metaBlock.GetPrevRandSeed()...)

	// TODO: replace hardcoded json - although it is hardcoded in nodesCoordinator as well.
	registryBytes, err := json.Marshal(nodesConfig)
	if err != nil {
		return nil, err
	}

	bootstrapUnit := bsh.storageService.GetStorer(dataRetriever.BootstrapUnit)
	err = bootstrapUnit.Put(key, registryBytes)
	if err != nil {
		return nil, err
	}

	log.Debug("saving nodes coordinator config", "key", key)

	return metaBlock.GetPrevRandSeed(), nil
}

func (bsh *baseStorageHandler) saveMetaHdrToStorage(metaBlock data.HeaderHandler) ([]byte, error) {
	headerBytes, err := bsh.marshalizer.Marshal(metaBlock)
	if err != nil {
		return nil, err
	}

	headerHash := bsh.hasher.Compute(string(headerBytes))

	metaHdrStorage := bsh.storageService.GetStorer(dataRetriever.MetaBlockUnit)
	err = metaHdrStorage.Put(headerHash, headerBytes)
	if err != nil {
		return nil, err
	}

	nonceToByteSlice := bsh.uint64Converter.ToByteSlice(metaBlock.GetNonce())
	metaHdrNonceStorage := bsh.storageService.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	err = metaHdrNonceStorage.Put(nonceToByteSlice, headerHash)
	if err != nil {
		return nil, err
	}

	return headerHash, nil
}

func (bsh *baseStorageHandler) saveShardHdrToStorage(hdr data.HeaderHandler) ([]byte, error) {
	headerBytes, err := bsh.marshalizer.Marshal(hdr)
	if err != nil {
		return nil, err
	}

	headerHash := bsh.hasher.Compute(string(headerBytes))

	shardHdrStorage := bsh.storageService.GetStorer(dataRetriever.BlockHeaderUnit)
	err = shardHdrStorage.Put(headerHash, headerBytes)
	if err != nil {
		return nil, err
	}

	nonceToByteSlice := bsh.uint64Converter.ToByteSlice(hdr.GetNonce())
	shardHdrNonceStorage := bsh.storageService.GetStorer(dataRetriever.ShardHdrNonceHashDataUnit + dataRetriever.UnitType(hdr.GetShardID()))
	err = shardHdrNonceStorage.Put(nonceToByteSlice, headerHash)
	if err != nil {
		return nil, err
	}

	return headerHash, nil
}

func (bsh *baseStorageHandler) saveMetaHdrForEpochTrigger(metaBlock data.HeaderHandler) error {
	lastHeaderBytes, err := bsh.marshalizer.Marshal(metaBlock)
	if err != nil {
		return err
	}

	epochStartIdentifier := core.EpochStartIdentifier(metaBlock.GetEpoch())
	metaHdrStorage := bsh.storageService.GetStorer(dataRetriever.MetaBlockUnit)
	err = metaHdrStorage.Put([]byte(epochStartIdentifier), lastHeaderBytes)
	if err != nil {
		return err
	}

	triggerStorage := bsh.storageService.GetStorer(dataRetriever.BootstrapUnit)
	err = triggerStorage.Put([]byte(epochStartIdentifier), lastHeaderBytes)
	if err != nil {
		return err
	}

	return nil
}

func (bsh *baseStorageHandler) saveMiniblocks(miniblocks map[string]*block.MiniBlock) {
	hashes := make([]string, 0, len(miniblocks))
	for hash, mb := range miniblocks {
		errNotCritical := bsh.saveMiniblock([]byte(hash), mb)
		if errNotCritical != nil {
			log.Warn("baseStorageHandler.saveMiniblocks - not a critical error", "error", errNotCritical)
		}

		hashes = append(hashes, hex.EncodeToString([]byte(hash)))
	}

	log.Debug("baseStorageHandler.saveMiniblocks", "saved miniblocks", strings.Join(hashes, ", "))
}

func (bsh *baseStorageHandler) saveMiniblock(hash []byte, mb *block.MiniBlock) error {
	mbBytes, err := bsh.marshalizer.Marshal(mb)
	if err != nil {
		return err
	}

	return bsh.storageService.Put(dataRetriever.MiniBlockUnit, hash, mbBytes)
}

func (bsh *baseStorageHandler) saveMiniblocksFromComponents(components *ComponentsNeededForBootstrap) {
	log.Debug("saving pending miniblocks", "num pending miniblocks", len(components.PendingMiniBlocks))
	bsh.saveMiniblocks(components.PendingMiniBlocks)

	peerMiniblocksMap := bsh.convertPeerMiniblocks(components.PeerMiniBlocks)
	log.Debug("saving peer miniblocks",
		"num peer miniblocks in slice", len(components.PeerMiniBlocks),
		"num peer miniblocks in map", len(peerMiniblocksMap))
	bsh.saveMiniblocks(peerMiniblocksMap)
}

func (bsh *baseStorageHandler) convertPeerMiniblocks(slice []*block.MiniBlock) map[string]*block.MiniBlock {
	result := make(map[string]*block.MiniBlock)
	for _, mb := range slice {
		hash, errNotCritical := core.CalculateHash(bsh.marshalizer, bsh.hasher, mb)
		if errNotCritical != nil {
			log.Error("internal error computing hash in baseStorageHandler.convertPeerMiniblocks",
				"miniblock", mb, "error", errNotCritical)
			continue
		}

		log.Debug("computed peer miniblock hash", "hash", hash)
		result[string(hash)] = mb
	}

	return result
}
