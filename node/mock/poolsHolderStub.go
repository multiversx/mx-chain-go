package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type PoolsHolderStub struct {
	HeadersCalled           func() data.ShardedDataCacherNotifier
	HeadersNoncesCalled     func() data.Uint64Cacher
	PeerChangesBlocksCalled func() storage.Cacher
	TransactionsCalled      func() data.ShardedDataCacherNotifier
	MiniBlocksCalled        func() storage.Cacher
	MetaBlocksCalled        func() storage.Cacher
}

func (phs *PoolsHolderStub) Headers() data.ShardedDataCacherNotifier {
	return phs.HeadersCalled()
}

func (phs *PoolsHolderStub) HeadersNonces() data.Uint64Cacher {
	return phs.HeadersNoncesCalled()
}

func (phs *PoolsHolderStub) PeerChangesBlocks() storage.Cacher {
	return phs.PeerChangesBlocksCalled()
}

func (phs *PoolsHolderStub) Transactions() data.ShardedDataCacherNotifier {
	return phs.TransactionsCalled()
}

func (phs *PoolsHolderStub) MiniBlocks() storage.Cacher {
	return phs.MiniBlocksCalled()
}

func (phs *PoolsHolderStub) MetaBlocks() storage.Cacher {
	return phs.MetaBlocksCalled()
}
