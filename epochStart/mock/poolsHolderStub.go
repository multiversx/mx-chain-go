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
	TrieNodesCalled            func() storage.Cacher
	CurrBlockTxsCalled         func() dataRetriever.TransactionCacher
}

// CurrentBlockTxs -
func (phs *PoolsHolderStub) CurrentBlockTxs() dataRetriever.TransactionCacher {
	if phs.CurrBlockTxsCalled != nil {
		return phs.CurrBlockTxsCalled()
	}
	return nil
}

// Headers -
func (phs *PoolsHolderStub) Headers() dataRetriever.HeadersPool {
	if phs.HeadersCalled != nil {
		return phs.HeadersCalled()
	}
	return nil
}

// PeerChangesBlocks -
func (phs *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	if phs.PeerChangesBlocksCalled != nil {
		return phs.PeerChangesBlocksCalled()
	}
	return nil
}

// Transactions -
func (phs *PoolsHolderStub) Transactions() dataRetriever.ShardedDataCacherNotifier {
	if phs.TransactionsCalled != nil {
		return phs.TransactionsCalled()
	}
	return nil
}

// MiniBlocks -
func (phs *PoolsHolderStub) MiniBlocks() storage.Cacher {
	if phs.MiniBlocksCalled != nil {
		return phs.MiniBlocksCalled()
	}
	return nil
}

// UnsignedTransactions -
func (phs *PoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	if phs.UnsignedTransactionsCalled != nil {
		return phs.UnsignedTransactionsCalled()
	}
	return nil
}

// RewardTransactions -
func (phs *PoolsHolderStub) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	if phs.RewardTransactionsCalled != nil {
		return phs.RewardTransactionsCalled()
	}
	return nil
}

// TrieNodes -
func (phs *PoolsHolderStub) TrieNodes() storage.Cacher {
	if phs.TrieNodesCalled != nil {
		return phs.TrieNodesCalled()
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (phs *PoolsHolderStub) IsInterfaceNil() bool {
	return phs == nil
}
