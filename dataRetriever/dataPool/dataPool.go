package dataPool

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ dataRetriever.PoolsHolder = (*dataPool)(nil)

type dataPool struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	rewardTransactions   dataRetriever.ShardedDataCacherNotifier
	headers              dataRetriever.HeadersPool
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
	trieNodes            storage.Cacher
	currBlockTxs         dataRetriever.TransactionCacher
	smartContracts       storage.Cacher
}

// NewDataPool creates a data pools holder object
func NewDataPool(
	transactions dataRetriever.ShardedDataCacherNotifier,
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier,
	rewardTransactions dataRetriever.ShardedDataCacherNotifier,
	headers dataRetriever.HeadersPool,
	miniBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	trieNodes storage.Cacher,
	currBlockTxs dataRetriever.TransactionCacher,
	smartContracts storage.Cacher,
) (*dataPool, error) {

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
	if check.IfNil(trieNodes) {
		return nil, dataRetriever.ErrNilTrieNodesPool
	}
	if check.IfNil(smartContracts) {
		return nil, dataRetriever.ErrNilSmartContractsPool
	}

	return &dataPool{
		transactions:         transactions,
		unsignedTransactions: unsignedTransactions,
		rewardTransactions:   rewardTransactions,
		headers:              headers,
		miniBlocks:           miniBlocks,
		peerChangesBlocks:    peerChangesBlocks,
		trieNodes:            trieNodes,
		currBlockTxs:         currBlockTxs,
		smartContracts:       smartContracts,
	}, nil
}

// CurrentBlockTxs returns the holder for current block transactions
func (dp *dataPool) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return dp.currBlockTxs
}

// Transactions returns the holder for transactions
func (dp *dataPool) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return dp.transactions
}

// UnsignedTransactions returns the holder for unsigned transactions (cross shard result entities)
func (dp *dataPool) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return dp.unsignedTransactions
}

// RewardTransactions returns the holder for reward transactions (cross shard result entities)
func (dp *dataPool) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return dp.rewardTransactions
}

// Headers returns the holder for headers
func (dp *dataPool) Headers() dataRetriever.HeadersPool {
	return dp.headers
}

// MiniBlocks returns the holder for miniblocks
func (dp *dataPool) MiniBlocks() storage.Cacher {
	return dp.miniBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (dp *dataPool) PeerChangesBlocks() storage.Cacher {
	return dp.peerChangesBlocks
}

// TrieNodes returns the holder for trie nodes
func (dp *dataPool) TrieNodes() storage.Cacher {
	return dp.trieNodes
}

// SmartContracts returns the holder for smart contracts
func (dp *dataPool) SmartContracts() storage.Cacher {
	return dp.smartContracts
}

// IsInterfaceNil returns true if there is no value under the interface
func (dp *dataPool) IsInterfaceNil() bool {
	return dp == nil
}
