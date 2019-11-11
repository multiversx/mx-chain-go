package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type PoolsHolderStub struct {
	HeadersCalled              func() storage.Cacher
	HeadersNoncesCalled        func() dataRetriever.Uint64SyncMapCacher
	PeerChangesBlocksCalled    func() storage.Cacher
	TransactionsCalled         func() dataRetriever.ShardedDataCacherNotifier
	UnsignedTransactionsCalled func() dataRetriever.ShardedDataCacherNotifier
	RewardTransactionsCalled   func() dataRetriever.ShardedDataCacherNotifier
	MiniBlocksCalled           func() storage.Cacher
	MetaBlocksCalled           func() storage.Cacher
	CurrBlockTxsCalled         func() dataRetriever.TransactionCacher
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

func (phs *PoolsHolderStub) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.TransactionsCalled()
}

func (phs *PoolsHolderStub) MiniBlocks() storage.Cacher {
	return phs.MiniBlocksCalled()
}

func (phs *PoolsHolderStub) MetaBlocks() storage.Cacher {
	return phs.MetaBlocksCalled()
}

func (phs *PoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.UnsignedTransactionsCalled()
}

func (phs *PoolsHolderStub) RewardTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.RewardTransactionsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (phs *PoolsHolderStub) IsInterfaceNil() bool {
	if phs == nil {
		return true
	}
	return false
}
