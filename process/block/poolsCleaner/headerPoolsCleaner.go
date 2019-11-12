package poolsCleaner

import (
	"sync/atomic"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// headerPoolsCleaner represents a pools cleaner for headers which should be already final
type headerPoolsCleaner struct {
	shardCoordinator     sharding.Coordinator
	headersNoncesPool    dataRetriever.Uint64SyncMapCacher
	headersPool          storage.Cacher
	notarizedHeadersPool storage.Cacher

	numRemovedHeaders uint64
}

// NewHeaderPoolsCleaner creates an object for cleaning headers which should be already final
func NewHeaderPoolsCleaner(
	shardCoordinator sharding.Coordinator,
	headersNoncesPool dataRetriever.Uint64SyncMapCacher,
	headersPool storage.Cacher,
	notarizedHeadersPool storage.Cacher,
) (*headerPoolsCleaner, error) {

	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(headersNoncesPool) {
		return nil, process.ErrNilHeadersNoncesDataPool
	}
	if check.IfNil(headersPool) {
		return nil, process.ErrNilHeadersDataPool
	}
	if check.IfNil(notarizedHeadersPool) {
		return nil, process.ErrNilNotarizedHeadersDataPool
	}

	return &headerPoolsCleaner{numRemovedHeaders: 0}, nil
}

// Clean removes from pools headers which should be already final
func (hpc *headerPoolsCleaner) Clean(
	finalNonceInSelfShard uint64,
	finalNoncesInNotarizedShards map[uint32]uint64,
) {
	hpc.removeHeadersBehindNonceFromPools(
		hpc.headersPool,
		hpc.headersNoncesPool,
		hpc.shardCoordinator.SelfId(),
		finalNonceInSelfShard)

	for shardId, nonce := range finalNoncesInNotarizedShards {
		hpc.removeHeadersBehindNonceFromPools(
			hpc.notarizedHeadersPool,
			hpc.headersNoncesPool,
			shardId,
			nonce)
	}

	return
}

func (hpc *headerPoolsCleaner) removeHeadersBehindNonceFromPools(
	cacher storage.Cacher,
	uint64SyncMapCacher dataRetriever.Uint64SyncMapCacher,
	shardId uint32,
	nonce uint64,
) {

	if nonce <= 1 {
		return
	}

	if check.IfNil(cacher) {
		return
	}

	for _, key := range cacher.Keys() {
		val, _ := cacher.Peek(key)
		if val == nil {
			continue
		}

		hdr, ok := val.(data.HeaderHandler)
		if !ok {
			continue
		}

		if hdr.GetShardID() != shardId || hdr.GetNonce() >= nonce {
			continue
		}

		atomic.AddUint64(&hpc.numRemovedHeaders, 1)

		cacher.Remove(key)

		if check.IfNil(uint64SyncMapCacher) {
			continue
		}

		uint64SyncMapCacher.Remove(hdr.GetNonce(), hdr.GetShardID())
	}
}

// NumRemovedHeaders returns the number of removed headers from pools
func (hpc *headerPoolsCleaner) NumRemovedHeaders() uint64 {
	return atomic.LoadUint64(&hpc.numRemovedHeaders)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hpc *headerPoolsCleaner) IsInterfaceNil() bool {
	return hpc == nil
}
