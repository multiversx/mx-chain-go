package dataPool

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type shardedDataPool struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              dataRetriever.HeadersPool
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	trieNodes            storage.Cacher
	currBlockTxs         dataRetriever.TransactionCacher
}

// NewShardedDataPool creates a data pools holder object
func NewShardedDataPool(
	transactions dataRetriever.ShardedDataCacherNotifier,
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier,
	rewardTransactions dataRetriever.ShardedDataCacherNotifier,
	headers dataRetriever.HeadersPool,
	miniBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	trieNodes storage.Cacher,
	currBlockTxs dataRetriever.TransactionCacher,
) (*shardedDataPool, error) {

	if check.IfNil(transactions) {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if check.IfNil(unsignedTransactions) {
		return nil, dataRetriever.ErrNilUnsignedTransactionPool
	}
	if check.IfNil(rewardTransactions) {
		return nil, dataRetriever.ErrNilRewardTransactionPool
	}
	if check.IfNil(headers) {
		return nil, dataRetriever.ErrNilHeadersDataPool
	}
	if check.IfNil(miniBlocks) {
		return nil, dataRetriever.ErrNilTxBlockDataPool
	}
	if check.IfNil(peerChangesBlocks) {
		return nil, dataRetriever.ErrNilPeerChangeBlockDataPool
	}
	if check.IfNil(currBlockTxs) {
		return nil, dataRetriever.ErrNilCurrBlockTxs
	}
	if trieNodes == nil || trieNodes.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilTrieNodesPool
	}

	return &shardedDataPool{
		transactions:         transactions,
		unsignedTransactions: unsignedTransactions,
		rewardTransactions:   rewardTransactions,
		headers:              headers,
		miniBlocks:           miniBlocks,
		peerChangesBlocks:    peerChangesBlocks,
		trieNodes:            trieNodes,
		currBlockTxs:         currBlockTxs,
	}, nil
}

// CurrentBlockTxs returns the holder for current block transactions
func (tdp *shardedDataPool) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return tdp.currBlockTxs
}

// Transactions returns the holder for transactions
func (tdp *shardedDataPool) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return tdp.transactions
}

// UnsignedTransactions returns the holder for unsigned transactions (cross shard result entities)
func (tdp *shardedDataPool) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return tdp.unsignedTransactions
}

// RewardTransactions returns the holder for reward transactions (cross shard result entities)
func (tdp *shardedDataPool) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return tdp.rewardTransactions
}

// Headers returns the holder for headers
func (tdp *shardedDataPool) Headers() dataRetriever.HeadersPool {
	return tdp.headers
}

// MiniBlocks returns the holder for miniblocks
func (tdp *shardedDataPool) MiniBlocks() storage.Cacher {
	return tdp.miniBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (tdp *shardedDataPool) PeerChangesBlocks() storage.Cacher {
	return tdp.peerChangesBlocks
}

// TrieNodes returns the holder for trie nodes
func (tdp *shardedDataPool) TrieNodes() storage.Cacher {
	return tdp.trieNodes
}

// IsInterfaceNil returns true if there is no value under the interface
func (tdp *shardedDataPool) IsInterfaceNil() bool {
	if tdp == nil {
		return true
	}
	return false
}
