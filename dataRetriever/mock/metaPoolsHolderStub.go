package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type MetaPoolsHolderStub struct {
	MetaChainBlocksCalled func() storage.Cacher
	MiniBlockHashesCalled func() data.ShardedDataCacherNotifier
	ShardHeadersCalled    func() storage.Cacher
	MetaBlockNoncesCalled func() data.Uint64Cacher
}

func (mphs *MetaPoolsHolderStub) MetaChainBlocks() storage.Cacher {
	return mphs.MetaChainBlocksCalled()
}

func (mphs *MetaPoolsHolderStub) MiniBlockHashes() data.ShardedDataCacherNotifier {
	return mphs.MiniBlockHashesCalled()
}

func (mphs *MetaPoolsHolderStub) ShardHeaders() storage.Cacher {
	return mphs.ShardHeadersCalled()
}

func (mphs *MetaPoolsHolderStub) MetaBlockNonces() data.Uint64Cacher {
	return mphs.MetaBlockNoncesCalled()
}
