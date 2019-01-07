package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type dataPool struct {
	transactions      data.ShardedDataCacherNotifier
	headers           data.ShardedDataCacherNotifier
	hdrNonces         data.Uint64Cacher
	txBlocks          storage.Cacher
	peerChangesBlocks storage.Cacher
	stateBlocks       storage.Cacher
}

// NewDataPool a transient data holder
func NewDataPool(
	transactions data.ShardedDataCacherNotifier,
	headers data.ShardedDataCacherNotifier,
	hdrNonces data.Uint64Cacher,
	txBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	stateBlocks storage.Cacher,
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

	if txBlocks == nil {
		return nil, data.ErrNilTxBlockDataPool
	}

	if peerChangesBlocks == nil {
		return nil, data.ErrNilPeerChangeBlockDataPool
	}

	if stateBlocks == nil {
		return nil, data.ErrNilStateBlockDataPool
	}

	return &dataPool{
		transactions:      transactions,
		headers:           headers,
		hdrNonces:         hdrNonces,
		txBlocks:          txBlocks,
		peerChangesBlocks: peerChangesBlocks,
		stateBlocks:       stateBlocks,
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

// TxBlocks returns the holder for transaction block bodies
func (tdp *dataPool) TxBlocks() storage.Cacher {
	return tdp.txBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (tdp *dataPool) PeerChangesBlocks() storage.Cacher {
	return tdp.peerChangesBlocks
}

// StateBlocks returns the holder for state block bodies
func (tdp *dataPool) StateBlocks() storage.Cacher {
	return tdp.stateBlocks
}
