package components

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/dataRetriever"
)

type poolsHolderWithSyncHeaders struct {
	dataRetriever.PoolsHolder

	headers dataRetriever.HeadersPool
	proofs  dataRetriever.ProofsPool
}

func newPoolsHolderWithSyncHeaders(poolsHolder dataRetriever.PoolsHolder) dataRetriever.PoolsHolder {
	return &poolsHolderWithSyncHeaders{
		PoolsHolder: poolsHolder,
		headers: &syncHeadersPool{
			HeadersPool: poolsHolder.Headers(),
		},
		proofs: &syncProofsPool{
			ProofsPool: poolsHolder.Proofs(),
		},
	}
}

func (holder *poolsHolderWithSyncHeaders) Headers() dataRetriever.HeadersPool {
	return holder.headers
}

func (holder *poolsHolderWithSyncHeaders) Proofs() dataRetriever.ProofsPool {
	return holder.proofs
}

// syncHeadersPool preserves the production pool's storage semantics, but invokes handlers before
// AddHeader returns. Consensus proposal delivery in the simulator is in-process; waiting for the
// handler ensures that a follower has processed the proposal before the drive advances to SIGNATURE.
type syncHeadersPool struct {
	dataRetriever.HeadersPool

	mutAdd      sync.Mutex
	mutHandlers sync.RWMutex
	handlers    []func(headerHandler data.HeaderHandler, headerHash []byte)
}

func (pool *syncHeadersPool) AddHeader(headerHash []byte, header data.HeaderHandler) {
	pool.mutAdd.Lock()
	if _, err := pool.HeadersPool.GetHeaderByHash(headerHash); err == nil {
		pool.mutAdd.Unlock()
		return
	}

	pool.HeadersPool.AddHeader(headerHash, header)

	pool.mutAdd.Unlock()
	pool.callHandlers(header, headerHash)
}

func (pool *syncHeadersPool) callHandlers(header data.HeaderHandler, headerHash []byte) {
	pool.mutHandlers.RLock()
	handlers := append([]func(data.HeaderHandler, []byte){}, pool.handlers...)
	pool.mutHandlers.RUnlock()

	for _, handler := range handlers {
		handler(header, headerHash)
	}
}

func (pool *syncHeadersPool) RegisterHandler(handler func(headerHandler data.HeaderHandler, headerHash []byte)) {
	if handler == nil {
		return
	}

	pool.mutHandlers.Lock()
	pool.handlers = append(pool.handlers, handler)
	pool.mutHandlers.Unlock()
}

// syncProofsPool preserves the production proof pool's storage semantics, but invokes handlers
// before an add returns. The metachain can be waiting for a same-round shard proof while shard
// consensus is driven in another goroutine; making the notification part of the delivery barrier
// prevents the manual drive from observing the proof in the pool before headersForBlock has
// consumed its notification.
type syncProofsPool struct {
	dataRetriever.ProofsPool

	mutHandlers sync.RWMutex
	handlers    []func(headerProof data.HeaderProofHandler)
}

func (pool *syncProofsPool) AddProof(headerProof data.HeaderProofHandler) bool {
	added := pool.ProofsPool.AddProof(headerProof)
	if added {
		pool.callHandlers(headerProof)
	}

	return added
}

func (pool *syncProofsPool) AddProofIfNoneAtNonce(
	headerProof data.HeaderProofHandler,
) (bool, data.HeaderProofHandler) {
	added, existingProof := pool.ProofsPool.AddProofIfNoneAtNonce(headerProof)
	if added {
		pool.callHandlers(headerProof)
	}

	return added, existingProof
}

func (pool *syncProofsPool) UpsertProof(headerProof data.HeaderProofHandler) bool {
	added := pool.ProofsPool.UpsertProof(headerProof)
	if added {
		pool.callHandlers(headerProof)
	}

	return added
}

func (pool *syncProofsPool) RegisterHandler(handler func(headerProof data.HeaderProofHandler)) {
	if handler == nil {
		return
	}

	pool.mutHandlers.Lock()
	pool.handlers = append(pool.handlers, handler)
	pool.mutHandlers.Unlock()
}

func (pool *syncProofsPool) callHandlers(headerProof data.HeaderProofHandler) {
	pool.mutHandlers.RLock()
	handlers := append([]func(data.HeaderProofHandler){}, pool.handlers...)
	pool.mutHandlers.RUnlock()

	for _, handler := range handlers {
		handler(headerProof)
	}
}
