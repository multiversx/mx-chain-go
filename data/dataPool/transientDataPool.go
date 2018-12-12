package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type transientDataPool struct {
	transactions      data.ShardedDataCacherNotifier
	headers           data.ShardedDataCacherNotifier
	hdrNonces         data.Uint64Cacher
	txBlocks          storage.Cacher
	peerChangesBlocks storage.Cacher
	stateBlocks       storage.Cacher
}

// NewTransientDataPool a transient data holder
func NewTransientDataPool(
	transactions data.ShardedDataCacherNotifier,
	headers data.ShardedDataCacherNotifier,
	hdrNonces data.Uint64Cacher,
	txBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	stateBlocks storage.Cacher,
) (*transientDataPool, error) {

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

	return &transientDataPool{
		transactions:      transactions,
		headers:           headers,
		hdrNonces:         hdrNonces,
		txBlocks:          txBlocks,
		peerChangesBlocks: peerChangesBlocks,
		stateBlocks:       stateBlocks,
	}, nil
}

// Transactions returns the holder for transactions
func (tdp *transientDataPool) Transactions() data.ShardedDataCacherNotifier {
	return tdp.transactions
}

// Headers returns the holder for headers
func (tdp *transientDataPool) Headers() data.ShardedDataCacherNotifier {
	return tdp.headers
}

// HeadersNonces returns the holder for (nonce, header hash) pairs
func (tdp *transientDataPool) HeadersNonces() data.Uint64Cacher {
	return tdp.hdrNonces
}

// TxBlocks returns the holder for transaction block bodies
func (tdp *transientDataPool) TxBlocks() storage.Cacher {
	return tdp.txBlocks
}

// PeerChangesBlocks returns the holder for peer changes block bodies
func (tdp *transientDataPool) PeerChangesBlocks() storage.Cacher {
	return tdp.peerChangesBlocks
}

// StateBlocks returns the holder for state block bodies
func (tdp *transientDataPool) StateBlocks() storage.Cacher {
	return tdp.stateBlocks
}
