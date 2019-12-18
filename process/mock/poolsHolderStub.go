package mock

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type PoolsHolderStub struct {
	HeadersCalled           func() storage.Cacher
	HeadersNoncesCalled     func() dataRetriever.Uint64SyncMapCacher
	PeerChangesBlocksCalled func() storage.Cacher
	MiniBlocksCalled        func() storage.Cacher
	MetaBlocksCalled        func() storage.Cacher
	CurrBlockTxsCalled      func() dataRetriever.TransactionCacher

	TransactionsTxPool         dataRetriever.TxPool
	UnsignedTransactionsTxPool dataRetriever.TxPool
	RewardTransactionsTxPool   dataRetriever.TxPool
}

func (stub *PoolsHolderStub) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return stub.CurrBlockTxsCalled()
}

func (stub *PoolsHolderStub) Headers() storage.Cacher {
	return stub.HeadersCalled()
}

func (stub *PoolsHolderStub) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return stub.HeadersNoncesCalled()
}

func (stub *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	return stub.PeerChangesBlocksCalled()
}

func (stub *PoolsHolderStub) MiniBlocks() storage.Cacher {
	return stub.MiniBlocksCalled()
}

func (stub *PoolsHolderStub) MetaBlocks() storage.Cacher {
	return stub.MetaBlocksCalled()
}

func (stub *PoolsHolderStub) Transactions() dataRetriever.TxPool {
	check.AssertNotNil(stub.TransactionsTxPool, "stub.TransactionsTxPool")
	return stub.TransactionsTxPool
}

func (stub *PoolsHolderStub) UnsignedTransactions() dataRetriever.TxPool {
	check.AssertNotNil(stub.UnsignedTransactionsTxPool, "stub.UnsignedTransactionsTxPool")
	return stub.UnsignedTransactionsTxPool
}

func (stub *PoolsHolderStub) RewardTransactions() dataRetriever.TxPool {
	check.AssertNotNil(stub.RewardTransactionsTxPool, "stub.RewardTransactionsTxPool")
	return stub.RewardTransactionsTxPool
}

// IsInterfaceNil returns true if there is no value under the interface
func (stub *PoolsHolderStub) IsInterfaceNil() bool {
	if stub == nil {
		return true
	}
	return false
}
