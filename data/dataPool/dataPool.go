package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type dataPool struct {
	transactions      data.ShardedDataCacherNotifier
	headers           data.ShardedDataCacherNotifier
	hdrNonces         data.Uint64Cacher
	miniBlocks        storage.Cacher
	peerChangesBlocks storage.Cacher
}

// NewDataPool creates a data pools holder object
func NewDataPool(
	transactions data.ShardedDataCacherNotifier,
	headers data.ShardedDataCacherNotifier,
	hdrNonces data.Uint64Cacher,
	miniBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
) (*dataPool, error) {

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

	return &dataPool{
		transactions:      transactions,
		headers:           headers,
		hdrNonces:         hdrNonces,
		miniBlocks:        miniBlocks,
		peerChangesBlocks: peerChangesBlocks,
	}, nil
}

// Transactions returns the holder for transactions
func (tdp *dataPool) Transactions() data.ShardedDataCacherNotifier {
	return tdp.transactions
}

// Headers returns the holder for headers
func (tdp *dataPool) Headers() data.ShardedDataCacherNotifier {
	return tdp.headers
}

// HeadersNonces returns the holder for (nonce, header hash) pairs
func (tdp *dataPool) HeadersNonces() data.Uint64Cacher {
	return tdp.hdrNonces
}

// MiniBlocks returns the holder for miniblocks
func (tdp *dataPool) MiniBlocks() storage.Cacher {
	return tdp.miniBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (tdp *dataPool) PeerChangesBlocks() storage.Cacher {
	return tdp.peerChangesBlocks
}
