package dataPool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type shardedDataPool struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              storage.Cacher
	metaBlocks           storage.Cacher
	headersNonces        dataRetriever.Uint64SyncMapCacher
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
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64SyncMapCacher,
	miniBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	metaBlocks storage.Cacher,
	trieNodes storage.Cacher,
	currBlockTxs dataRetriever.TransactionCacher,
) (*shardedDataPool, error) {

	if transactions == nil || transactions.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if unsignedTransactions == nil || unsignedTransactions.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilUnsignedTransactionPool
	}
	if rewardTransactions == nil || rewardTransactions.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilRewardTransactionPool
	}
	if headers == nil || headers.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilHeadersDataPool
	}
	if headersNonces == nil || headersNonces.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilHeadersNoncesDataPool
	}
	if miniBlocks == nil || miniBlocks.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilTxBlockDataPool
	}
	if peerChangesBlocks == nil || peerChangesBlocks.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilPeerChangeBlockDataPool
	}
	if metaBlocks == nil || metaBlocks.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilMetaBlockPool
	}
	if trieNodes == nil || trieNodes.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilTrieNodesPool
	}
	if currBlockTxs == nil || currBlockTxs.IsInterfaceNil() {
		return nil, dataRetriever.ErrNilCurrBlockTxs
	}

	return &shardedDataPool{
		transactions:         transactions,
		unsignedTransactions: unsignedTransactions,
		rewardTransactions:   rewardTransactions,
		headers:              headers,
		headersNonces:        headersNonces,
		miniBlocks:           miniBlocks,
		peerChangesBlocks:    peerChangesBlocks,
		metaBlocks:           metaBlocks,
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
func (tdp *shardedDataPool) Headers() storage.Cacher {
	return tdp.headers
}

// HeadersNonces returns the holder nonce-block hash pairs. It will hold both shard headers nonce-hash pairs
// also metachain header nonce-hash pairs
func (tdp *shardedDataPool) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return tdp.headersNonces
}

// MiniBlocks returns the holder for miniblocks
func (tdp *shardedDataPool) MiniBlocks() storage.Cacher {
	return tdp.miniBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (tdp *shardedDataPool) PeerChangesBlocks() storage.Cacher {
	return tdp.peerChangesBlocks
}

// MetaBlocks returns the holder for meta blocks
func (tdp *shardedDataPool) MetaBlocks() storage.Cacher {
	return tdp.metaBlocks
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
