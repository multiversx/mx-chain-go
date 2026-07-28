package components

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	proofscache "github.com/multiversx/mx-chain-go/dataRetriever/dataPool/proofsCache"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

func TestSyncHeadersPool_AddHeaderCallsHandlerSynchronouslyOnce(t *testing.T) {
	t.Parallel()

	var mutStored sync.Mutex
	stored := make(map[string]data.HeaderHandler)
	inner := &pool.HeadersPoolStub{
		AddCalled: func(hash []byte, header data.HeaderHandler) {
			mutStored.Lock()
			stored[string(hash)] = header
			mutStored.Unlock()
		},
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			mutStored.Lock()
			defer mutStored.Unlock()

			header, ok := stored[string(hash)]
			if !ok {
				return nil, errors.New("missing header")
			}

			return header, nil
		},
	}
	headers := &syncHeadersPool{HeadersPool: inner}

	var callbackCalls atomic.Int32
	headers.RegisterHandler(func(_ data.HeaderHandler, _ []byte) {
		callbackCalls.Add(1)
	})

	header := &block.Header{Nonce: 1}
	headers.AddHeader([]byte("hash"), header)
	headers.AddHeader([]byte("hash"), header)

	require.Equal(t, int32(1), callbackCalls.Load())
	require.Same(t, header, stored["hash"])
}

func TestSyncHeadersPool_ConcurrentSameHeaderCallsHandlerOnce(t *testing.T) {
	t.Parallel()

	var mutStored sync.Mutex
	stored := make(map[string]data.HeaderHandler)
	inner := &pool.HeadersPoolStub{
		AddCalled: func(hash []byte, header data.HeaderHandler) {
			mutStored.Lock()
			stored[string(hash)] = header
			mutStored.Unlock()
		},
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			mutStored.Lock()
			defer mutStored.Unlock()

			header, ok := stored[string(hash)]
			if !ok {
				return nil, errors.New("missing header")
			}

			return header, nil
		},
	}
	headers := &syncHeadersPool{HeadersPool: inner}

	var callbackCalls atomic.Int32
	headers.RegisterHandler(func(_ data.HeaderHandler, _ []byte) {
		callbackCalls.Add(1)
	})

	const numCallers = 20
	var wg sync.WaitGroup
	wg.Add(numCallers)
	for range numCallers {
		go func() {
			headers.AddHeader([]byte("hash"), &block.Header{Nonce: 1})
			wg.Done()
		}()
	}
	wg.Wait()

	require.Equal(t, int32(1), callbackCalls.Load())
}

func TestSyncProofsPool_AddAndUpsertCallHandlersSynchronously(t *testing.T) {
	t.Parallel()

	proofs := &syncProofsPool{
		ProofsPool: proofscache.NewProofsPool(3, 100),
	}

	var callbackCalls atomic.Int32
	proofs.RegisterHandler(func(_ data.HeaderProofHandler) {
		callbackCalls.Add(1)
	})

	proof := &block.HeaderProof{
		HeaderHash:    []byte("hash"),
		HeaderNonce:   1,
		HeaderShardId: 0,
	}
	require.True(t, proofs.AddProof(proof))
	require.Equal(t, int32(1), callbackCalls.Load())

	require.False(t, proofs.AddProof(proof))
	require.Equal(t, int32(1), callbackCalls.Load())

	require.True(t, proofs.UpsertProof(proof))
	require.Equal(t, int32(2), callbackCalls.Load())
}
