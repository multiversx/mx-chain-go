package node

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
)

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
	unitType := dataRetriever.GetHeadersDataUnit(shardId)
	storer := n.dataComponents.StorageService().GetStorer(unitType)

	headerBuffer, err := storer.GetFromEpoch(headerHash, epoch)
	if err != nil {
		return nil, err
	}

	header, err := process.CreateHeader(shardId, n.coreComponents.InternalMarshalizer(), headerBuffer)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (n *Node) getBlockRootHash(headerHash []byte, header data.HeaderHandler) []byte {
	blockRootHash, err := n.processComponents.ScheduledTxsExecutionHandler().GetScheduledRootHashForHeaderWithEpoch(
		headerHash,
		header.GetEpoch())
	if err != nil {
		blockRootHash = header.GetRootHash()
	}

	return blockRootHash
}
