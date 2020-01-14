package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type MetaPoolsHolderStub struct {
	MiniBlocksCalled           func() storage.Cacher
	HeadersCalled              func() dataRetriever.HeadersPool
	TrieNodesCalled            func() storage.Cacher
	TransactionsCalled         func() dataRetriever.ShardedDataCacherNotifier
	UnsignedTransactionsCalled func() dataRetriever.ShardedDataCacherNotifier
	CurrBlockTxsCalled         func() dataRetriever.TransactionCacher
}

func (mphs *MetaPoolsHolderStub) CurrentBlockTxs() dataRetriever.TransactionCacher {
	return mphs.CurrBlockTxsCalled()
}

func (mphs *MetaPoolsHolderStub) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return mphs.TransactionsCalled()
}

func (mphs *MetaPoolsHolderStub) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return mphs.UnsignedTransactionsCalled()
}

func (mphs *MetaPoolsHolderStub) MiniBlocks() storage.Cacher {
	return mphs.MiniBlocksCalled()
}

func (mphs *MetaPoolsHolderStub) Headers() dataRetriever.HeadersPool {
	return mphs.HeadersCalled()
}

func (mphs *MetaPoolsHolderStub) TrieNodes() storage.Cacher {
	return mphs.TrieNodesCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mphs *MetaPoolsHolderStub) IsInterfaceNil() bool {
	if mphs == nil {
		return true
	}
	return false
}
