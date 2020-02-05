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
	MiniBlocksCalled           func() storage.Cacher
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

// UnsignedTransactions -
func (phs *PoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return phs.UnsignedTransactionsCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (phs *PoolsHolderStub) IsInterfaceNil() bool {
	if phs == nil {
		return true
	}
	return false
}
