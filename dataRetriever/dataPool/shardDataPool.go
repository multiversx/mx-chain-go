package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

type shardedDataPool struct {
	transactions         dataRetriever.ShardedDataCacherNotifier
	smartContractResults dataRetriever.ShardedDataCacherNotifier
	headers              storage.Cacher
	metaBlocks           storage.Cacher
	hdrNonces            dataRetriever.Uint64Cacher
	metaHdrNonces        dataRetriever.Uint64Cacher
	miniBlocks           storage.Cacher
	peerChangesBlocks    storage.Cacher
}

// NewShardedDataPool creates a data pools holder object
func NewShardedDataPool(
	transactions dataRetriever.ShardedDataCacherNotifier,
	smartContractResults dataRetriever.ShardedDataCacherNotifier,
	headers storage.Cacher,
	hdrNonces dataRetriever.Uint64Cacher,
	miniBlocks storage.Cacher,
	peerChangesBlocks storage.Cacher,
	metaBlocks storage.Cacher,
	metaHdrNonces dataRetriever.Uint64Cacher,
) (*shardedDataPool, error) {

	if transactions == nil {
		return nil, dataRetriever.ErrNilTxDataPool
	}
	if smartContractResults == nil {
		return nil, dataRetriever.ErrNilSmartContractResultsPool
	}
	if headers == nil {
		return nil, dataRetriever.ErrNilHeadersDataPool
	}
	if hdrNonces == nil {
		return nil, dataRetriever.ErrNilHeadersNoncesDataPool
	}
	if metaHdrNonces == nil {
		return nil, dataRetriever.ErrNilMetaBlockNoncesPool
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
		smartContractResults: smartContractResults,
		headers:              headers,
		hdrNonces:            hdrNonces,
		metaHdrNonces:        metaHdrNonces,
		miniBlocks:           miniBlocks,
		peerChangesBlocks:    peerChangesBlocks,
		metaBlocks:           metaBlocks,
	}, nil
}

// Transactions returns the holder for transactions
func (tdp *shardedDataPool) Transactions() dataRetriever.ShardedDataCacherNotifier {
	return tdp.transactions
}

// SmartContractResults returns the holder for transactions
func (tdp *shardedDataPool) SmartContractResults() dataRetriever.ShardedDataCacherNotifier {
	return tdp.smartContractResults
}

// Headers returns the holder for headers
func (tdp *shardedDataPool) Headers() storage.Cacher {
	return tdp.headers
}

// HeadersNonces returns the holder for (nonce, header hash) pairs
func (tdp *shardedDataPool) HeadersNonces() dataRetriever.Uint64Cacher {
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

// MetaBlocks returns the holder for meta blocks
func (tdp *shardedDataPool) MetaBlocks() storage.Cacher {
	return tdp.metaBlocks
}

// MetaHeadersNonces returns the holder for (nonce, meta header hash) pairs
func (tdp *shardedDataPool) MetaHeadersNonces() dataRetriever.Uint64Cacher {
	return tdp.metaHdrNonces
}
