package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type PoolsHolderStub struct {
	TransactionsCalled         func() dataRetriever.TxPool
	UnsignedTransactionsCalled func() dataRetriever.ShardedDataCacherNotifier
	RewardTransactionsCalled   func() dataRetriever.ShardedDataCacherNotifier

	HeadersCalled           func() storage.Cacher
	HeadersNoncesCalled     func() dataRetriever.Uint64SyncMapCacher
	PeerChangesBlocksCalled func() storage.Cacher
	MiniBlocksCalled        func() storage.Cacher
	MetaBlocksCalled        func() storage.Cacher
	CurrBlockTxsCalled      func() dataRetriever.TransactionCacher

	TransactionsTxPool         dataRetriever.TxPool
	UnsignedTransactionsTxPool dataRetriever.ShardedDataCacherNotifier
	RewardTransactionsTxPool   dataRetriever.ShardedDataCacherNotifier
}

func (phs *PoolsHolderStub) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return phs.CurrBlockTxsCalled()
}

func (phs *PoolsHolderStub) Headers() storage.Cacher {
	return phs.HeadersCalled()
}

func (phs *PoolsHolderStub) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return phs.HeadersNoncesCalled()
}

func (phs *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	return phs.PeerChangesBlocksCalled()
}

func (phs *PoolsHolderStub) MiniBlocks() storage.Cacher {
	return phs.MiniBlocksCalled()
}

func (phs *PoolsHolderStub) MetaBlocks() storage.Cacher {
	return phs.MetaBlocksCalled()
}

func (phs *PoolsHolderStub) Transactions() dataRetriever.TxPool {
	if phs.TransactionsCalled != nil {
		return phs.TransactionsCalled()
	}

	check.AssertNotNil(phs.TransactionsTxPool, "phs.TransactionsTxPool")
	return phs.TransactionsTxPool
}

func (phs *PoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	if phs.UnsignedTransactionsCalled != nil {
		return phs.UnsignedTransactionsCalled()
	}

	check.AssertNotNil(phs.UnsignedTransactionsTxPool, "phs.UnsignedTransactionsTxPool")
	return phs.UnsignedTransactionsTxPool
}

func (phs *PoolsHolderStub) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	if phs.RewardTransactionsCalled != nil {
		return phs.RewardTransactionsCalled()
	}

	check.AssertNotNil(phs.RewardTransactionsTxPool, "phs.RewardTransactionsTxPool")
	return phs.RewardTransactionsTxPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (phs *PoolsHolderStub) IsInterfaceNil() bool {
	if phs == nil {
		return true
	}
	return false
}
