package node

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

func (n *Node) getBlockHeaderByHash(hash []byte) (data.HeaderHandler, error) {
	return process.GetHeaderFromStorage(
		n.processComponents.ShardCoordinator().SelfId(),
		hash,
		n.coreComponents.InternalMarshalizer(),
		n.dataComponents.StorageService(),
	)
}

func (n *Node) getBlockRootHash(headerHash []byte, header data.HeaderHandler) []byte {
	blockRootHash, err := n.processComponents.ScheduledTxsExecutionHandler().GetScheduledRootHashForHeader(headerHash)
	if err != nil {
		blockRootHash = header.GetRootHash()
	}

	return blockRootHash
}

func (n *Node) getBlockHeaderByNonce(nonce uint64) (data.HeaderHandler, []byte, error) {
	return process.GetHeaderFromStorageWithNonce(
		nonce,
		n.processComponents.ShardCoordinator().SelfId(),
		n.dataComponents.StorageService(),
		n.coreComponents.Uint64ByteSliceConverter(),
		n.coreComponents.InternalMarshalizer(),
	)
}
