package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type MetaPoolsHolderStub struct {
	MetaChainBlocksCalled func() storage.Cacher
	MiniBlockHashesCalled func() dataRetriever.ShardedDataCacherNotifier
	ShardHeadersCalled    func() storage.Cacher
	HeadersNoncesCalled   func() dataRetriever.Uint64SyncMapCacher
}

func (mphs *MetaPoolsHolderStub) MetaChainBlocks() storage.Cacher {
	return mphs.MetaChainBlocksCalled()
}

func (mphs *MetaPoolsHolderStub) MiniBlockHashes() dataRetriever.ShardedDataCacherNotifier {
	return mphs.MiniBlockHashesCalled()
}

func (mphs *MetaPoolsHolderStub) ShardHeaders() storage.Cacher {
	return mphs.ShardHeadersCalled()
}

func (mphs *MetaPoolsHolderStub) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return mphs.HeadersNoncesCalled()
}
