package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type shardedDataPool struct {
	transactions      data.ShardedDataCacherNotifier
	headers           data.ShardedDataCacherNotifier
	metaBlocks        data.ShardedDataCacherNotifier
	hdrNonces         data.Uint64Cacher
	miniBlocks        storage.Cacher
	peerChangesBlocks storage.Cacher
}

// NewDataPool creates a data pools holder object
func NewShardedDataPool(
	transactions data.ShardedDataCacherNotifier,
	headers data.ShardedDataCacherNotifier,
	hdrNonces data.Uint64Cacher,
	miniBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	metaBlocks data.ShardedDataCacherNotifier,
) (*shardedDataPool, error) {

	if transactions == nil {
		return nil, data.ErrNilTxDataPool
	}

	if headers == nil {
		return nil, data.ErrNilHeadersDataPool
	}

	if hdrNonces == nil {
		return nil, data.ErrNilHeadersNoncesDataPool
	}

	if miniBlocks == nil {
		return nil, data.ErrNilTxBlockDataPool
	}

	if peerChangesBlocks == nil {
		return nil, data.ErrNilPeerChangeBlockDataPool
	}

	if metaBlocks == nil {
		return nil, data.ErrNilMetaBlockPool
	}

	return &shardedDataPool{
		transactions:      transactions,
		headers:           headers,
		hdrNonces:         hdrNonces,
		miniBlocks:        miniBlocks,
		peerChangesBlocks: peerChangesBlocks,
		metaBlocks:        metaBlocks,
	}, nil
}

// Transactions returns the holder for transactions
func (tdp *shardedDataPool) Transactions() data.ShardedDataCacherNotifier {
	return tdp.transactions
}

// Headers returns the holder for headers
func (tdp *shardedDataPool) Headers() data.ShardedDataCacherNotifier {
	return tdp.headers
}

// HeadersNonces returns the holder for (nonce, header hash) pairs
func (tdp *shardedDataPool) HeadersNonces() data.Uint64Cacher {
	return tdp.hdrNonces
}

// MiniBlocks returns the holder for miniblocks
func (tdp *shardedDataPool) MiniBlocks() storage.Cacher {
	return tdp.miniBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (tdp *shardedDataPool) PeerChangesBlocks() storage.Cacher {
	return tdp.peerChangesBlocks
}

// MetaMiniBlockHeaders returns the holder for meta mini block headers
func (tdp *shardedDataPool) MetaBlocks() data.ShardedDataCacherNotifier {
	return tdp.metaBlocks
}
