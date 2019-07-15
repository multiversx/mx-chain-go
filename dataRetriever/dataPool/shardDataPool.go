package dataPool

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type shardedDataPool struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier
	headers              storage.Cacher
	metaBlocks           storage.Cacher
	headersNonces        dataRetriever.Uint64SyncMapCacher
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
}

// NewShardedDataPool creates a data pools holder object
func NewShardedDataPool(
	transactions dataRetriever.ShardedDataCacherNotifier,
	unsignedTransactions dataRetriever.ShardedDataCacherNotifier,
	headers storage.Cacher,
	headersNonces dataRetriever.Uint64SyncMapCacher,
	miniBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	metaBlocks storage.Cacher,
) (*shardedDataPool, error) {

	if transactions == nil {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if unsignedTransactions == nil {
		return nil, dataRetriever.ErrNilUnsignedTransactionPool
	}
	if headers == nil {
		return nil, dataRetriever.ErrNilHeadersDataPool
	}
	if headersNonces == nil {
		return nil, dataRetriever.ErrNilHeadersNoncesDataPool
	}
	if miniBlocks == nil {
		return nil, dataRetriever.ErrNilTxBlockDataPool
	}
	if peerChangesBlocks == nil {
		return nil, dataRetriever.ErrNilPeerChangeBlockDataPool
	}
	if metaBlocks == nil {
		return nil, dataRetriever.ErrNilMetaBlockPool
	}

	return &shardedDataPool{
		transactions:         transactions,
		unsignedTransactions: unsignedTransactions,
		headers:              headers,
		headersNonces:        headersNonces,
		miniBlocks:           miniBlocks,
		peerChangesBlocks:    peerChangesBlocks,
		metaBlocks:           metaBlocks,
	}, nil
}

// Transactions returns the holder for transactions
func (tdp *shardedDataPool) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return tdp.transactions
}

// UnsignedTransactions returns the holder for unsigned transactions (cross shard result entities)
func (tdp *shardedDataPool) UnsignedTransactions() dataRetriever.ShardedDataCacherNotifier {
	return tdp.unsignedTransactions
}

// Headers returns the holder for headers
func (tdp *shardedDataPool) Headers() storage.Cacher {
	return tdp.headers
}

// HeadersNonces returns the holder nonce-block hash pairs. It will hold both shard headers nonce-hash pairs
// also metachain header nonce-hash pairs
func (tdp *shardedDataPool) HeadersNonces() dataRetriever.Uint64SyncMapCacher {
	return tdp.headersNonces
}

// MiniBlocks returns the holder for miniblocks
func (tdp *shardedDataPool) MiniBlocks() storage.Cacher {
	return tdp.miniBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (tdp *shardedDataPool) PeerChangesBlocks() storage.Cacher {
	return tdp.peerChangesBlocks
}

// MetaBlocks returns the holder for meta blocks
func (tdp *shardedDataPool) MetaBlocks() storage.Cacher {
	return tdp.metaBlocks
}
