package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// GetRawMiniBlock will return MiniBlock as byte array by hash
func (n *Node) GetRawMiniBlock(txHash string) ([]byte, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	blockBytes, err := n.getFromStorer(dataRetriever.MiniBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return blockBytes, nil
}

// GetInternalMiniBlock will MiniBlock by hash
func (n *Node) GetInternalMiniBlock(txHash string) (*block.MiniBlock, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	blockBytes, err := n.getFromStorer(dataRetriever.MiniBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	marshalizer := n.coreComponents.InternalMarshalizer()

	miniBlock := &block.MiniBlock{}
	err = marshalizer.Unmarshal(miniBlock, blockBytes)
	if err != nil {
		return nil, err
	}

	return miniBlock, nil
}

func (n *Node) getFromStorer(unit dataRetriever.UnitType, key []byte) ([]byte, error) {
	historyRepo := n.processComponents.HistoryRepository()
	store := n.dataComponents.StorageService()

	hasDbLookupExtensions := historyRepo.IsEnabled()

	if !hasDbLookupExtensions {
		return store.Get(unit, key)
	}

	epoch, err := historyRepo.GetEpochByHash(key)
	if err != nil {
		return nil, err
	}

	storer := store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}
