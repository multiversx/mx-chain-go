package node

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (n *Node) getBlockHashByNonce(nonce uint64) ([]byte, error) {
	shardId := n.processComponents.ShardCoordinator().SelfId()
	hashByNonceUnit := dataRetriever.GetHdrNonceHashDataUnit(shardId)

	return process.GetHeaderHashFromStorageWithNonce(
		nonce,
		n.dataComponents.StorageService(),
		n.coreComponents.Uint64ByteSliceConverter(),
		n.coreComponents.InternalMarshalizer(),
		hashByNonceUnit,
	)
}

func (n *Node) getBlockHeaderByHash(headerHash []byte) (data.HeaderHandler, error) {
	epoch, err := n.getEpochByHash(headerHash)
	if err != nil {
		return nil, err
	}

	header, err := n.getBlockHeaderInEpochByHash(epoch, headerHash)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (n *Node) getEpochByHash(hash []byte) (uint32, error) {
	historyRepository := n.processComponents.HistoryRepository()
	if !historyRepository.IsEnabled() {
		return 0, fmt.Errorf("%w: history repository (dblookupext) is not enabled", ErrNotSupported)
	}

	return historyRepository.GetEpochByHash(hash)
}

func (n *Node) getBlockHeaderInEpochByHash(epoch uint32, headerHash []byte) (data.HeaderHandler, error) {
	shardId := n.processComponents.ShardCoordinator().SelfId()

	var blockHeader data.HeaderHandler
	var unitType dataRetriever.UnitType

	if shardId == core.MetachainShardId {
		unitType = dataRetriever.MetaBlockUnit
	} else {
		unitType = dataRetriever.BlockHeaderUnit
	}

	storer := n.dataComponents.StorageService().GetStorer(unitType)
	data, err := storer.GetFromEpoch(headerHash, epoch)
	if err != nil {
		return nil, err
	}

	if shardId == core.MetachainShardId {
		blockHeader = &block.MetaBlock{}
		err = n.coreComponents.InternalMarshalizer().Unmarshal(blockHeader, data)
		if err != nil {
			return nil, err
		}
	} else {
		blockHeader, err = process.CreateShardHeader(n.coreComponents.InternalMarshalizer(), data)
		if err != nil {
			return nil, err
		}
	}

	return blockHeader, nil
}

func (n *Node) getBlockRootHash(headerHash []byte, header data.HeaderHandler) []byte {
	blockRootHash, err := n.getScheduledRootHash(header.GetEpoch(), headerHash)
	if err != nil {
		blockRootHash = header.GetRootHash()
	}

	return blockRootHash
}

func (n *Node) getScheduledRootHash(epoch uint32, headerHash []byte) ([]byte, error) {
	storer := n.dataComponents.StorageService().GetStorer(dataRetriever.ScheduledSCRsUnit)
	data, err := storer.GetFromEpoch(headerHash, epoch)
	if err != nil {
		return nil, err
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{}
	err = n.coreComponents.InternalMarshalizer().Unmarshal(scheduledSCRs, data)
	if err != nil {
		return nil, err
	}

	return scheduledSCRs.GetRootHash(), nil
}

func (n *Node) getBlockHeaderByNonce(nonce uint64) (data.HeaderHandler, []byte, error) {
	headerHash, err := n.getBlockHashByNonce(nonce)
	if err != nil {
		return nil, nil, err
	}

	header, err := n.getBlockHeaderByHash(headerHash)
	if err != nil {
		return nil, nil, err
	}

	return header, headerHash, nil
}
