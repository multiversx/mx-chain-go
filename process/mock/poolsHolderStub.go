package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// PoolsHolderStub -
type PoolsHolderStub struct {
	HeadersCalled              func() dataRetriever.HeadersPool
	PeerChangesBlocksCalled    func() storage.Cacher
	TransactionsCalled         func() dataRetriever.ShardedDataCacherNotifier
	UnsignedTransactionsCalled func() dataRetriever.ShardedDataCacherNotifier
	RewardTransactionsCalled   func() dataRetriever.ShardedDataCacherNotifier
	MiniBlocksCalled           func() storage.Cacher
	MetaBlocksCalled           func() storage.Cacher
	TrieNodesCalled            func() storage.Cacher
	CurrBlockTxsCalled         func() dataRetriever.TransactionCacher
}

// CurrentBlockTxs -
func (phs *PoolsHolderStub) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return phs.CurrBlockTxsCalled()
}

// Headers -
func (phs *PoolsHolderStub) Headers() dataRetriever.HeadersPool {
	return phs.HeadersCalled()
}

// PeerChangesBlocks -
func (phs *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	return phs.PeerChangesBlocksCalled()
}

// Transactions -
func (phs *PoolsHolderStub) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.TransactionsCalled()
}

// MiniBlocks -
func (phs *PoolsHolderStub) MiniBlocks() storage.Cacher {
	return phs.MiniBlocksCalled()
}

// MetaBlocks -
func (phs *PoolsHolderStub) MetaBlocks() storage.Cacher {
	return phs.MetaBlocksCalled()
}

// UnsignedTransactions -
func (phs *PoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.UnsignedTransactionsCalled()
}

// RewardTransactions -
func (phs *PoolsHolderStub) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.RewardTransactionsCalled()
}

// TrieNodes -
func (phs *PoolsHolderStub) TrieNodes() storage.Cacher {
	return phs.TrieNodesCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (phs *PoolsHolderStub) IsInterfaceNil() bool {
	if phs == nil {
		return true
	}
	return false
}
