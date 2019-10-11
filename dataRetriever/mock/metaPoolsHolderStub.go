package mock

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type MetaPoolsHolderStub struct {
	MetaBlocksCalled      func() storage.Cacher
	MiniBlockHashesCalled func() dataRetriever.ShardedDataCacherNotifier
	ShardHeadersCalled    func() storage.Cacher
	HeadersNoncesCalled   func() dataRetriever.Uint64SyncMapCacher
}

func (mphs *MetaPoolsHolderStub) MetaBlocks() storage.Cacher {
	return mphs.MetaBlocksCalled()
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

// IsInterfaceNil returns true if there is no value under the interface
func (mphs *MetaPoolsHolderStub) IsInterfaceNil() bool {
	if mphs == nil {
		return true
	}
	return false
}
