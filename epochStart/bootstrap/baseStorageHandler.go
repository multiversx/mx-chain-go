package bootstrap

import (
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/block/bootstrapStorage"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// HighestRoundFromBootStorage is the key for the highest round that is saved in storage
const highestRoundFromBootStorage = "highestRoundFromBootStorage"

const triggerRegistrykeyPrefix = "epochStartTrigger_"

const nodesCoordinatorRegistrykeyPrefix = "indexHashed_"

// baseStorageHandler handles the storage functions for saving bootstrap data
type baseStorageHandler struct {
	storageService   dataRetriever.StorageService
	shardCoordinator sharding.Coordinator
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	currentEpoch     uint32
}

func (bsh *baseStorageHandler) getAndSavePendingMiniBlocks(miniBlocks map[string]*block.MiniBlock) ([]bootstrapStorage.PendingMiniBlocksInfo, error) {
	pendingMBsMap := make(map[uint32][][]byte)
	for hash, miniBlock := range miniBlocks {
		if _, ok := pendingMBsMap[miniBlock.SenderShardID]; !ok {
			pendingMBsMap[miniBlock.SenderShardID] = make([][]byte, 0)
		}
		pendingMBsMap[miniBlock.SenderShardID] = append(pendingMBsMap[miniBlock.SenderShardID], []byte(hash))
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

func (bsh *baseStorageHandler) getAndSaveNodesCoordinatorKey(
	metaBlock *block.MetaBlock,
	nodesConfig *sharding.NodesCoordinatorRegistry,
) ([]byte, error) {
	key := append([]byte(nodesCoordinatorRegistrykeyPrefix), metaBlock.RandSeed...)

	registryBytes, err := json.Marshal(nodesConfig)
	if err != nil {
		return nil, err
	}

	err = bsh.storageService.GetStorer(dataRetriever.BootstrapUnit).Put(key, registryBytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (bsh *baseStorageHandler) saveTries(components *ComponentsNeededForBootstrap) error {
	for _, trie := range components.UserAccountTries {
		err := trie.Commit()
		if err != nil {
			return err
		}
	}

	for _, trie := range components.PeerAccountTries {
		err := trie.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}
