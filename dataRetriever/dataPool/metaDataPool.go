package dataPool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type metaDataPool struct {
	metaBlocks      storage.Cacher
	miniBlockHashes dataRetriever.ShardedDataCacherNotifier
	shardHeaders    storage.Cacher
	headersNonces   dataRetriever.Uint64SyncMapCacher
}

// NewMetaDataPool creates a data pools holder object
func NewMetaDataPool(
	metaBlocks storage.Cacher,
	miniBlockHashes dataRetriever.ShardedDataCacherNotifier,
	shardHeaders storage.Cacher,
	headersNonces dataRetriever.Uint64SyncMapCacher,
) (*metaDataPool, error) {

	if metaBlocks == nil {
		return nil, dataRetriever.ErrNilMetaBlockPool
	}
	if miniBlockHashes == nil {
		return nil, dataRetriever.ErrNilMiniBlockHashesPool
	}
	if shardHeaders == nil {
		return nil, dataRetriever.ErrNilShardHeaderPool
	}
	if headersNonces == nil {
		return nil, dataRetriever.ErrNilMetaBlockNoncesPool
	}

	return &metaDataPool{
		metaBlocks:      metaBlocks,
		miniBlockHashes: miniBlockHashes,
		shardHeaders:    shardHeaders,
		headersNonces:   headersNonces,
	}, nil
}

// MetaChainBlocks returns the holder for meta blocks
func (mdp *metaDataPool) MetaChainBlocks() storage.Cacher {
	return mdp.metaBlocks
}

// MiniBlockHashes returns the holder for meta mini block hashes
func (mdp *metaDataPool) MiniBlockHashes() dataRetriever.ShardedDataCacherNotifier {
	return mdp.miniBlockHashes
}

// ShardHeaders returns the holder for shard headers
func (mdp *metaDataPool) ShardHeaders() storage.Cacher {
	return mdp.shardHeaders
}

// HeadersNonces returns the holder nonce-block hash pairs. It will hold both shard headers nonce-hash pairs
// also metachain header nonce-hash pairs
func (mdp *metaDataPool) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return mdp.headersNonces
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdp *metaDataPool) IsInterfaceNil() bool {
	if mdp == nil {
		return true
	}
	return false
}
