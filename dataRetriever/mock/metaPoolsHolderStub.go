package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type MetaPoolsHolderStub struct {
	MetaChainBlocksCalled    func() storage.Cacher
	MiniBlockHashesCalled    func() dataRetriever.ShardedDataCacherNotifier
	ShardHeadersCalled       func() storage.Cacher
	ShardHeadersNoncesCalled func() dataRetriever.Uint64Cacher
	MetaBlockNoncesCalled    func() dataRetriever.Uint64Cacher
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

func (mphs *MetaPoolsHolderStub) MetaBlockNonces() dataRetriever.Uint64Cacher {
	return mphs.MetaBlockNoncesCalled()
}

func (mphs *MetaPoolsHolderStub) ShardHeadersNonces() dataRetriever.Uint64Cacher {
	return mphs.ShardHeadersNoncesCalled()
}
