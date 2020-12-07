package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// PoolsHolderStub -
type PoolsHolderStub struct {
	HeadersCalled              func() dataRetriever.HeadersPool
	TransactionsCalled         func() dataRetriever.ShardedDataCacherNotifier
	UnsignedTransactionsCalled func() dataRetriever.ShardedDataCacherNotifier
	RewardTransactionsCalled   func() dataRetriever.ShardedDataCacherNotifier
	MiniBlocksCalled           func() storage.Cacher
	MetaBlocksCalled           func() storage.Cacher
	CurrBlockTxsCalled         func() dataRetriever.TransactionCacher
	TrieNodesCalled            func() storage.Cacher
	PeerChangesBlocksCalled    func() storage.Cacher
	SmartContractsCalled       func() storage.Cacher
}

// NewPoolsHolderStub -
func NewPoolsHolderStub() *PoolsHolderStub {
	return &PoolsHolderStub{}
}

// Headers -
func (holder *PoolsHolderStub) Headers() dataRetriever.HeadersPool {
	if holder.HeadersCalled != nil {
		return holder.HeadersCalled()
	}

	return nil
}

// Transactions -
func (holder *PoolsHolderStub) Transactions() dataRetriever.ShardedDataCacherNotifier {
	if holder.TransactionsCalled != nil {
		return holder.TransactionsCalled()
	}

	return NewShardedDataStub()
}

// UnsignedTransactions -
func (holder *PoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	if holder.UnsignedTransactionsCalled != nil {
		return holder.UnsignedTransactionsCalled()
	}

	return NewShardedDataStub()
}

// RewardTransactions -
func (holder *PoolsHolderStub) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	if holder.RewardTransactionsCalled != nil {
		return holder.RewardTransactionsCalled()
	}

	return NewShardedDataStub()
}

// MiniBlocks -
func (holder *PoolsHolderStub) MiniBlocks() storage.Cacher {
	if holder.MiniBlocksCalled != nil {
		return holder.MiniBlocksCalled()
	}

	return NewCacherStub()
}

// MetaBlocks -
func (holder *PoolsHolderStub) MetaBlocks() storage.Cacher {
	if holder.MetaBlocksCalled != nil {
		return holder.MetaBlocksCalled()
	}

	return NewCacherStub()
}

// CurrentBlockTxs -
func (holder *PoolsHolderStub) CurrentBlockTxs() dataRetriever.TransactionCacher {
	if holder.CurrBlockTxsCalled != nil {
		return holder.CurrBlockTxsCalled()
	}

	return nil
}

// TrieNodes -
func (holder *PoolsHolderStub) TrieNodes() storage.Cacher {
	if holder.TrieNodesCalled != nil {
		return holder.TrieNodesCalled()
	}

	return NewCacherStub()
}

// PeerChangesBlocks -
func (holder *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	if holder.PeerChangesBlocksCalled != nil {
		return holder.PeerChangesBlocksCalled()
	}

	return NewCacherStub()
}

// SmartContracts -
func (holder *PoolsHolderStub) SmartContracts() storage.Cacher {
	if holder.SmartContractsCalled != nil {
		return holder.SmartContractsCalled()
	}

	return NewCacherStub()
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *PoolsHolderStub) IsInterfaceNil() bool {
	return holder == nil
}
