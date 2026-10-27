package node

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
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
	epoch, err := n.getOptionalEpochByHash(headerHash)
	if err != nil {
		return nil, err
	}

	header, err := n.getBlockHeaderInEpochByHash(headerHash, epoch)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (n *Node) getOptionalEpochByHash(hash []byte) (core.OptionalUint32, error) {
	historyRepository := n.processComponents.HistoryRepository()
	if !historyRepository.IsEnabled() {
		return core.OptionalUint32{}, nil
	}

	epoch, err := historyRepository.GetEpochByHash(hash)
	if err != nil {
		return core.OptionalUint32{}, err
	}

	return core.OptionalUint32{Value: epoch, HasValue: true}, nil
}

func (n *Node) getBlockHeaderInEpochByHash(headerHash []byte, epoch core.OptionalUint32) (data.HeaderHandler, error) {
	shardId := n.processComponents.ShardCoordinator().SelfId()
	unitType := dataRetriever.GetHeadersDataUnit(shardId)
	storer, err := n.dataComponents.StorageService().GetStorer(unitType)
	if err != nil {
		return nil, err
	}

	var headerBuffer []byte

	if epoch.HasValue {
		headerBuffer, err = storer.GetFromEpoch(headerHash, epoch.Value)
	} else {
		headerBuffer, err = storer.Get(headerHash)
	}
	if err != nil {
		return nil, err
	}

	header, err := process.UnmarshalHeader(shardId, n.coreComponents.InternalMarshalizer(), headerBuffer)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// TODO: refactor to remove duplicated code sc query
func (n *Node) getBlockRootHash(headerHash []byte, header data.HeaderHandler) []byte {
	if header.IsHeaderV3() {
		return n.getBlockRootHashV3(headerHash, header)
	}

	blockRootHash, err := n.processComponents.ScheduledTxsExecutionHandler().GetScheduledRootHashForHeaderWithEpoch(
		headerHash,
		header.GetEpoch())
	if err == nil {
		return blockRootHash
	}

	return header.GetRootHash()
}

func (n *Node) getBlockRootHashV3(
	headerHash []byte,
	header data.HeaderHandler,
) []byte {
	rootHash, err := n.getRootHashByExecutionResult(headerHash)
	if err == nil {
		return rootHash
	}

	lastExecutionResult, err := common.ExtractBaseExecutionResultHandler(header.GetLastExecutionResultHandler())
	if err != nil {
		log.Error("getBlockRootHashV3: failed to get root hash for header v3, using root hash directly from header", "error", err)
		return header.GetRootHash()
	}

	return lastExecutionResult.GetRootHash()
}

func (n *Node) getRootHashByExecutionResult(
	currentHeaderHash []byte,
) ([]byte, error) {
	execResStorer, err := n.dataComponents.StorageService().GetStorer(dataRetriever.ExecutionResultsUnit)
	if err != nil {
		return nil, err
	}

	execResBytes, err := execResStorer.Get(currentHeaderHash)
	if err != nil {
		return nil, err
	}

	execRes, err := process.UnmarshalExecutionResult(n.coreComponents.InternalMarshalizer(), execResBytes)
	if err != nil {
		return nil, err
	}

	return execRes.GetRootHash(), nil
}
