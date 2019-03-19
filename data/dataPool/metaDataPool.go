package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type metaDataPool struct {
	metaBlocks      storage.Cacher
	miniBlockHashes data.ShardedDataCacherNotifier
	shardHeaders    data.ShardedDataCacherNotifier
	metaBlockNonces data.Uint64Cacher
}

// NewMetaDataPool creates a data pools holder object
func NewMetaDataPool(
	metaBlocks storage.Cacher,
	miniBlockHashes data.ShardedDataCacherNotifier,
	shardHeaders data.ShardedDataCacherNotifier,
	metaBlockNonces data.Uint64Cacher,
) (*metaDataPool, error) {

	if metaBlocks == nil {
		return nil, data.ErrNilMetaBlockPool
	}

	if miniBlockHashes == nil {
		return nil, data.ErrNilMiniBlockHashesPool
	}

	if shardHeaders == nil {
		return nil, data.ErrNilShardHeaderPool
	}

	if metaBlockNonces == nil {
		return nil, data.ErrNilMetaBlockNouncesPool
	}

	return &metaDataPool{
		metaBlocks:      metaBlocks,
		miniBlockHashes: miniBlockHashes,
		shardHeaders:    shardHeaders,
		metaBlockNonces: metaBlockNonces,
	}, nil
}

// MetaChainBlocks returns the holder for meta blocks
func (mdp *metaDataPool) MetaChainBlocks() storage.Cacher {
	return mdp.metaBlocks
}

// MiniBlockHashes returns the holder for meta mini block hashes
func (mdp *metaDataPool) MiniBlockHashes() data.ShardedDataCacherNotifier {
	return mdp.miniBlockHashes
}

// ShardHeaders returns the holder for shard headers
func (mdp *metaDataPool) ShardHeaders() data.ShardedDataCacherNotifier {
	return mdp.shardHeaders
}

// MetaBlockNonces returns the holder for meta block nonces
func (mdp *metaDataPool) MetaBlockNonces() data.Uint64Cacher {
	return mdp.metaBlockNonces
}
