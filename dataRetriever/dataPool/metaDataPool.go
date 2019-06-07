package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type metaDataPool struct {
	metaBlocks         storage.Cacher
	miniBlockHashes    dataRetriever.ShardedDataCacherNotifier
	shardHeaders       storage.Cacher
	shardHeadersNonces dataRetriever.Uint64Cacher
	metaBlockNonces    dataRetriever.Uint64Cacher
}

// NewMetaDataPool creates a data pools holder object
func NewMetaDataPool(
	metaBlocks storage.Cacher,
	miniBlockHashes dataRetriever.ShardedDataCacherNotifier,
	shardHeaders storage.Cacher,
	metaBlockNonces dataRetriever.Uint64Cacher,
	shardHeadersNonces dataRetriever.Uint64Cacher,
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
	if metaBlockNonces == nil {
		return nil, dataRetriever.ErrNilMetaBlockNoncesPool
	}
	if shardHeadersNonces == nil {
		return nil, dataRetriever.ErrNilHeadersNoncesDataPool
	}

	return &metaDataPool{
		metaBlocks:         metaBlocks,
		miniBlockHashes:    miniBlockHashes,
		shardHeaders:       shardHeaders,
		metaBlockNonces:    metaBlockNonces,
		shardHeadersNonces: shardHeadersNonces,
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

// MetaBlockNonces returns the holder for meta block nonces
func (mdp *metaDataPool) MetaBlockNonces() dataRetriever.Uint64Cacher {
	return mdp.metaBlockNonces
}

// ShardHeadersNonces returns the holder for shard headers nonces
func (mdp *metaDataPool) ShardHeadersNonces() dataRetriever.Uint64Cacher {
	return mdp.shardHeadersNonces
}
