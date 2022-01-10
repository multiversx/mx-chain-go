package node

import (
	"encoding/hex"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
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
func (n *Node) GetInternalMiniBlock(format common.OutportFormat, txHash string) (interface{}, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}

	blockBytes, err := n.getFromStorer(dataRetriever.MiniBlockUnit, hash)
	if err != nil {
		return nil, err
	}

	return n.convertMiniBlockBytesByOutportFormat(format, blockBytes)
}

func (n *Node) convertMiniBlockBytesByOutportFormat(format common.OutportFormat, blockBytes []byte) (interface{}, error) {
	switch format {
	case common.Internal:
		marshalizer := n.coreComponents.InternalMarshalizer()

		miniBlock := &block.MiniBlock{}
		err := marshalizer.Unmarshal(miniBlock, blockBytes)
		if err != nil {
			return nil, err
		}

		return miniBlock, nil
	case common.Proto:
		return blockBytes, nil
	default:
		return nil, ErrInvalidOutportFormat
	}
}

func (n *Node) getFromStorer(unit dataRetriever.UnitType, key []byte) ([]byte, error) {
	historyRepo := n.processComponents.HistoryRepository()
	store := n.dataComponents.StorageService()

	if !historyRepo.IsEnabled() {
		return store.Get(unit, key)
	}

	epoch, err := historyRepo.GetEpochByHash(key)
	if err != nil {
		return nil, err
	}

	storer := store.GetStorer(unit)
	return storer.GetFromEpoch(key, epoch)
}
