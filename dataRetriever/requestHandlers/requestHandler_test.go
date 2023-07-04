package requestHandlers

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/storage/cache"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timeoutSendRequests = time.Second * 2
var errExpected = errors.New("expected error")

func createRequestersFinderStubThatShouldNotBeCalled(tb testing.TB) *dataRetrieverMocks.RequestersFinderStub {
	return &dataRetrieverMocks.RequestersFinderStub{
		IntraShardRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
			assert.Fail(tb, "IntraShardRequesterCalled should not have been called")
			return nil, nil
		},
		MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
			assert.Fail(tb, "MetaChainRequesterCalled should not have been called")
			return nil, nil
		},
		CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, err error) {
			assert.Fail(tb, "CrossShardRequesterCalled should not have been called")
			return nil, nil
		},
	}
}

func TestNewResolverRequestHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil finder should error", func(t *testing.T) {
		t.Parallel()

		rrh, err := NewResolverRequestHandler(
			nil,
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		assert.Nil(t, rrh)
		assert.Equal(t, dataRetriever.ErrNilRequestersFinder, err)
	})
	t.Run("nil requested items handler should error", func(t *testing.T) {
		t.Parallel()

		rrh, err := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{},
			nil,
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		assert.Nil(t, rrh)
		assert.Equal(t, dataRetriever.ErrNilRequestedItemsHandler, err)
	})
	t.Run("nil whitelist handler should error", func(t *testing.T) {
		t.Parallel()

		rrh, err := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{},
			&mock.RequestedItemsHandlerStub{},
			nil,
			1,
			0,
			time.Second,
		)

		assert.Nil(t, rrh)
		assert.Equal(t, dataRetriever.ErrNilWhiteListHandler, err)
	})
	t.Run("invalid max txs to request should error", func(t *testing.T) {
		t.Parallel()

		rrh, err := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			0,
			0,
			time.Second,
		)

		assert.Nil(t, rrh)
		assert.Equal(t, dataRetriever.ErrInvalidMaxTxRequest, err)
	})
	t.Run("invalid request interval should error", func(t *testing.T) {
		t.Parallel()

		rrh, err := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Millisecond-time.Nanosecond,
		)

		assert.Nil(t, rrh)
		assert.True(t, errors.Is(err, dataRetriever.ErrRequestIntervalTooSmall))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		rrh, err := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		assert.Nil(t, err)
		assert.NotNil(t, rrh)
	})
}

func TestResolverRequestHandler_RequestTransaction(t *testing.T) {
	t.Parallel()

	t.Run("no hash should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					require.Fail(t, "should have not been called")
					return nil, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTransaction(0, make([][]byte, 0))
	})
	t.Run("error when getting cross shard requester should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})
	})
	t.Run("uncastable requester should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		wrongTxRequester := &dataRetrieverMocks.NonceRequesterStub{}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return wrongTxRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})
	})
	t.Run("should request", func(t *testing.T) {
		t.Parallel()

		chTxRequested := make(chan struct{})
		txRequester := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				chTxRequested <- struct{}{}
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return txRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})

		select {
		case <-chTxRequested:
		case <-time.After(timeoutSendRequests):
			assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
		}

		time.Sleep(time.Second)
	})
	t.Run("should request 4 times if different shards", func(t *testing.T) {
		t.Parallel()

		numRequests := uint32(0)
		txRequester := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				atomic.AddUint32(&numRequests, 1)
				return nil
			},
		}

		timeSpan := time.Second
		timeCache := cache.NewTimeCache(timeSpan)
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return txRequester, nil
				},
			},
			timeCache,
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})
		rrh.RequestTransaction(1, [][]byte{[]byte("txHash")})
		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})
		rrh.RequestTransaction(1, [][]byte{[]byte("txHash")})

		time.Sleep(time.Second) // let the go routines finish
		assert.Equal(t, uint32(2), atomic.LoadUint32(&numRequests))
		time.Sleep(time.Second) // sweep will take effect

		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})
		rrh.RequestTransaction(1, [][]byte{[]byte("txHash")})
		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})
		rrh.RequestTransaction(1, [][]byte{[]byte("txHash")})

		time.Sleep(time.Second) // let the go routines finish
		assert.Equal(t, uint32(4), atomic.LoadUint32(&numRequests))
	})
	t.Run("errors on request should not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		chTxRequested := make(chan struct{})
		txRequester := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				chTxRequested <- struct{}{}
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return txRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})

		select {
		case <-chTxRequested:
		case <-time.After(timeoutSendRequests):
			assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
		}

		time.Sleep(time.Second)
	})
}

func TestResolverRequestHandler_RequestMiniBlock(t *testing.T) {
	t.Parallel()

	t.Run("hash already requested", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					require.Fail(t, "should not have been called")
					return nil, nil
				},
			},
			&mock.RequestedItemsHandlerStub{
				HasCalled: func(key string) bool {
					return true
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMiniBlock(0, make([]byte, 0))
	})
	t.Run("CrossShardRequester returns error", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMiniBlock(0, make([]byte, 0))
	})
	t.Run("RequestDataFromHash error", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		mbRequester := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMiniBlock(0, []byte("mbHash"))
	})
	t.Run("should request", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		mbRequester := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMiniBlock(0, []byte("mbHash"))

		assert.True(t, wasCalled)
	})
	t.Run("should call with the correct epoch", func(t *testing.T) {
		t.Parallel()

		expectedEpoch := uint32(7)
		mbRequester := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				assert.Equal(t, expectedEpoch, epoch)
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.SetEpoch(expectedEpoch)

		rrh.RequestMiniBlock(0, []byte("mbHash"))
	})
}

func TestResolverRequestHandler_RequestShardHeader(t *testing.T) {
	t.Parallel()

	t.Run("hash already requested should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			createRequestersFinderStubThatShouldNotBeCalled(t),
			&mock.RequestedItemsHandlerStub{
				HasCalled: func(key string) bool {
					return true
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeader(0, make([]byte, 0))
	})
	t.Run("no hash should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			createRequestersFinderStubThatShouldNotBeCalled(t),
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeader(1, make([]byte, 0))
	})
	t.Run("RequestDataFromHash returns error should work", func(t *testing.T) {
		t.Parallel()

		mbRequester := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeader(0, []byte("hdrHash"))
	})
	t.Run("should call request", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		mbRequester := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeader(0, []byte("hdrHash"))

		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestMetaHeader(t *testing.T) {
	t.Parallel()

	t.Run("header already requested should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{},
			&mock.RequestedItemsHandlerStub{
				HasCalled: func(key string) bool {
					return true
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMetaHeader([]byte("hdrHash"))
	})
	t.Run("cast fail should work", func(t *testing.T) {
		t.Parallel()

		req := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				require.Fail(t, "should have not been called")
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return req, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMetaHeader([]byte("hdrHash"))
	})
	t.Run("MetaChainRequester returns error should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMetaHeader([]byte("hdrHash"))
	})
	t.Run("RequestDataFromHash returns error should work", func(t *testing.T) {
		t.Parallel()

		req := &dataRetrieverMocks.HeaderRequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return req, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMetaHeader([]byte("hdrHash"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		mbRequester := &dataRetrieverMocks.HeaderRequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMetaHeader([]byte("hdrHash"))

		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestShardHeaderByNonce(t *testing.T) {
	t.Parallel()

	t.Run("nonce already requested should work", func(t *testing.T) {
		t.Parallel()

		called := false
		rrh, _ := NewResolverRequestHandler(
			createRequestersFinderStubThatShouldNotBeCalled(t),
			&mock.RequestedItemsHandlerStub{
				HasCalled: func(key string) bool {
					called = true
					return true
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeaderByNonce(0, 0)
		require.True(t, called)
	})
	t.Run("invalid nonce should work", func(t *testing.T) {
		t.Parallel()

		called := false
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, err error) {
					called = true
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			core.MetachainShardId,
			time.Second,
		)

		rrh.RequestShardHeaderByNonce(1, 0)
		require.True(t, called)
	})
	t.Run("finder returns error should work and not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, shardID uint32) (requester dataRetriever.Requester, e error) {
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeaderByNonce(0, 0)
	})
	t.Run("cast fails should work and not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		hdrRequester := &dataRetrieverMocks.RequesterStub{}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, shardID uint32) (requester dataRetriever.Requester, e error) {
					return hdrRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeaderByNonce(0, 0)
	})
	t.Run("resolver fails should work and not panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		hdrRequester := &dataRetrieverMocks.NonceRequesterStub{
			RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, shardID uint32) (requester dataRetriever.Requester, e error) {
					return hdrRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeaderByNonce(0, 0)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		hdrRequester := &dataRetrieverMocks.NonceRequesterStub{
			RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, shardID uint32) (requester dataRetriever.Requester, e error) {
					return hdrRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestShardHeaderByNonce(0, 0)

		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestMetaHeaderByNonce(t *testing.T) {
	t.Parallel()

	t.Run("nonce already requested should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			createRequestersFinderStubThatShouldNotBeCalled(t),
			&mock.RequestedItemsHandlerStub{
				HasCalled: func(key string) bool {
					return true
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestMetaHeaderByNonce(0)
	})
	t.Run("MetaChainRequester returns error should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{
				AddCalled: func(keys [][]byte) {
					require.Fail(t, "should not have been called")
				},
			},
			100,
			0,
			time.Second,
		)

		rrh.RequestMetaHeaderByNonce(0)
	})
	t.Run("RequestDataFromNonce returns error should work", func(t *testing.T) {
		t.Parallel()

		hdrRequester := &dataRetrieverMocks.HeaderRequesterStub{
			RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return hdrRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestMetaHeaderByNonce(0)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		hdrRequester := &dataRetrieverMocks.HeaderRequesterStub{
			RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return hdrRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestMetaHeaderByNonce(0)

		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestScrErrorWhenGettingCrossShardRequesterShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestUnsignedTransactions(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestScrWrongResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	wrongTxRequester := &dataRetrieverMocks.NonceRequesterStub{}

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
				return wrongTxRequester, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestUnsignedTransactions(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestScrShouldRequestScr(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txRequester := &dataRetrieverMocks.HashSliceRequesterStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
				return txRequester, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestUnsignedTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

func TestResolverRequestHandler_RequestScrErrorsOnRequestShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	chTxRequested := make(chan struct{})
	txRequester := &dataRetrieverMocks.HashSliceRequesterStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return errExpected
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
				return txRequester, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestUnsignedTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

func TestResolverRequestHandler_RequestRewardShouldRequestReward(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txRequester := &dataRetrieverMocks.HashSliceRequesterStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, e error) {
				return txRequester, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestRewardTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

func TestRequestTrieNodes(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		chTxRequested := make(chan struct{})
		requesterMock := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hash [][]byte, epoch uint32) error {
				chTxRequested <- struct{}{}
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaCrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
					return requesterMock, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTrieNodes(0, [][]byte{[]byte("hash")}, "topic")
		select {
		case <-chTxRequested:
		case <-time.After(timeoutSendRequests):
			assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
		}

		time.Sleep(time.Second)
	})
	t.Run("nil resolver", func(t *testing.T) {
		t.Parallel()

		localError := errors.New("test error")
		called := false
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaCrossShardRequesterCalled: func(baseTopic string, shId uint32) (requester dataRetriever.Requester, err error) {
					called = true
					return nil, localError
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTrieNodes(core.MetachainShardId, [][]byte{[]byte("hash")}, "topic")
		assert.True(t, called)
	})
	t.Run("no hash", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaCrossShardRequesterCalled: func(baseTopic string, shId uint32) (requester dataRetriever.Requester, err error) {
					require.Fail(t, "should have not been called")
					return nil, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestTrieNodes(core.MetachainShardId, [][]byte{}, "topic")
	})
}

func TestResolverRequestHandler_RequestStartOfEpochMetaBlock(t *testing.T) {
	t.Parallel()

	t.Run("epoch already requested", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
					require.Fail(t, "should not have been called")
					return nil, nil
				},
			},
			&mock.RequestedItemsHandlerStub{
				HasCalled: func(key string) bool {
					return true
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestStartOfEpochMetaBlock(0)
	})
	t.Run("missing resolver", func(t *testing.T) {
		t.Parallel()

		called := false
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
					called = true
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestStartOfEpochMetaBlock(0)
		assert.True(t, called)
	})
	t.Run("wrong resolver", func(t *testing.T) {
		t.Parallel()

		called := false
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
					called = true
					return &dataRetrieverMocks.RequesterStub{}, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestStartOfEpochMetaBlock(0)
		assert.True(t, called)
	})
	t.Run("RequestDataFromEpoch fails", func(t *testing.T) {
		t.Parallel()

		called := false
		requesterMock := &dataRetrieverMocks.EpochRequesterStub{
			RequestDataFromEpochCalled: func(identifier []byte) error {
				called = true
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
					return requesterMock, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestStartOfEpochMetaBlock(0)
		assert.True(t, called)
	})
	t.Run("add error", func(t *testing.T) {
		t.Parallel()

		called := false
		requesterMock := &dataRetrieverMocks.EpochRequesterStub{
			RequestDataFromEpochCalled: func(identifier []byte) error {
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
					return requesterMock, nil
				},
			},
			&mock.RequestedItemsHandlerStub{
				AddCalled: func(key string) error {
					called = true
					return errExpected
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestStartOfEpochMetaBlock(0)
		assert.True(t, called)
	})
}

func TestResolverRequestHandler_RequestTrieNodeRequestFails(t *testing.T) {
	chTxRequested := make(chan struct{})
	localErr := errors.New("local error")
	requesterMock := &dataRetrieverMocks.ChunkRequesterStub{
		RequestDataFromReferenceAndChunkCalled: func(hash []byte, chunkIndex uint32) error {
			chTxRequested <- struct{}{}
			return localErr
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			MetaChainRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
				return requesterMock, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestTrieNode([]byte("hash"), "topic", 1)
	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestTrieNode")
	}

	time.Sleep(time.Second)
}

func TestResolverRequestHandler_RequestTrieNodeShouldWork(t *testing.T) {
	chTxRequested := make(chan struct{})
	requesterMock := &dataRetrieverMocks.ChunkRequesterStub{
		RequestDataFromReferenceAndChunkCalled: func(hash []byte, chunkIndex uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			MetaChainRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
				return requesterMock, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestTrieNode([]byte("hash"), "topic", 1)
	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestTrieNode")
	}
}

func TestResolverRequestHandler_RequestTrieNodeNilResolver(t *testing.T) {
	t.Parallel()

	localError := errors.New("test error")
	called := false
	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
				called = true
				return nil, localError
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestTrieNode([]byte("hash"), "topic", 1)
	assert.True(t, called)
}

func TestResolverRequestHandler_RequestTrieNodeNotAValidResolver(t *testing.T) {
	t.Parallel()

	called := false
	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, err error) {
				called = true
				return &dataRetrieverMocks.HashSliceRequesterStub{}, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestTrieNode([]byte("hash"), "topic", 1)
	assert.True(t, called)
}

func TestResolverRequestHandler_RequestPeerAuthenticationsByHashes(t *testing.T) {
	t.Parallel()

	providedHashes := [][]byte{[]byte("h1"), []byte("h2")}
	providedShardId := uint32(15)
	t.Run("MetaChainRequester returns error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		paRequester := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paRequester, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsByHashes(providedShardId, providedHashes)
		assert.False(t, wasCalled)
	})
	t.Run("cast fails", func(t *testing.T) {
		t.Parallel()

		req := &dataRetrieverMocks.NonceRequesterStub{}
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return req, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsByHashes(providedShardId, providedHashes)
	})
	t.Run("RequestDataFromHashArray returns error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		paRequester := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				wasCalled = true
				assert.Equal(t, providedHashes, hashes)
				return errExpected
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{
				AddCalled: func(key string) error {
					require.Fail(t, "should not have been called")
					return nil
				},
			},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsByHashes(providedShardId, providedHashes)
		assert.True(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		wasCalled := false
		paRequester := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				wasCalled = true
				assert.Equal(t, providedHashes, hashes)
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsByHashes(providedShardId, providedHashes)
		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestValidatorInfo(t *testing.T) {
	t.Parallel()

	t.Run("hash already requested should work", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					require.Fail(t, "should not have been called")
					return nil, nil
				},
			},
			&mock.RequestedItemsHandlerStub{
				HasCalled: func(key string) bool {
					return true
				},
			},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorInfo(providedHash)
	})
	t.Run("MetaChainRequester returns error", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		wasCalled := false
		res := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return res, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorInfo(providedHash)
		assert.False(t, wasCalled)
	})
	t.Run("RequestDataFromHash returns error", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		res := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return res, nil
				},
			},
			&mock.RequestedItemsHandlerStub{
				AddCalled: func(key string) error {
					require.Fail(t, "should not have been called")
					return nil
				},
			},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorInfo(providedHash)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		wasCalled := false
		res := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				assert.True(t, bytes.Equal(providedHash, hash))
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					assert.Equal(t, common.ValidatorInfoTopic, baseTopic)
					return res, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorInfo(providedHash)
		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestValidatorsInfo(t *testing.T) {
	t.Parallel()

	t.Run("no hash", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					require.Fail(t, "should not have been called")
					return nil, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorsInfo([][]byte{})
	})
	t.Run("MetaChainRequester returns error", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		wasCalled := false
		res := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return res, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorsInfo([][]byte{providedHash})
		assert.False(t, wasCalled)
	})
	t.Run("RequestDataFromHashArray returns error", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		res := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				return errExpected
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return res, nil
				},
			},
			&mock.RequestedItemsHandlerStub{
				AddCalled: func(key string) error {
					require.Fail(t, "should not have been called")
					return nil
				},
			},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorsInfo([][]byte{providedHash})
	})
	t.Run("cast fails", func(t *testing.T) {
		t.Parallel()

		providedHash := []byte("provided hash")
		mbRequester := &dataRetrieverMocks.NonceRequesterStub{} // uncastable to HashSliceRequester
		wasCalled := false
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{
				AddCalled: func(keys [][]byte) {
					wasCalled = true
				},
			},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorsInfo([][]byte{providedHash})
		assert.False(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedHashes := [][]byte{[]byte("provided hash 1"), []byte("provided hash 2")}
		wasCalled := false
		res := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				assert.Equal(t, providedHashes, hashes)
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (requester dataRetriever.Requester, e error) {
					assert.Equal(t, common.ValidatorInfoTopic, baseTopic)
					return res, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorsInfo(providedHashes)
		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestMiniblocks(t *testing.T) {
	t.Parallel()

	t.Run("no hash should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
					require.Fail(t, "should have not been called")
					return nil, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestMiniBlocks(0, [][]byte{})
	})
	t.Run("CrossShardRequester fails should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestMiniBlocks(0, [][]byte{[]byte("mbHash")})
	})
	t.Run("cast fails should work", func(t *testing.T) {
		t.Parallel()

		nonceRequester := &dataRetrieverMocks.NonceRequesterStub{} // uncastable to HashSliceRequester
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
					return nonceRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{
				AddCalled: func(keys [][]byte) {
					require.Fail(t, "should have not been called")
				},
			},
			100,
			0,
			time.Second,
		)

		rrh.RequestMiniBlocks(0, [][]byte{[]byte("mbHash")})
	})
	t.Run("request data fails should work", func(t *testing.T) {
		t.Parallel()

		mbRequester := &dataRetrieverMocks.HashSliceRequesterStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				return errExpected
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
					return mbRequester, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestMiniBlocks(0, [][]byte{[]byte("mbHash")})
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Requester, error) {
					return &dataRetrieverMocks.HashSliceRequesterStub{}, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestMiniBlocks(0, [][]byte{[]byte("mbHash")})
	})
}

func TestResolverRequestHandler_RequestInterval(t *testing.T) {
	t.Parallel()

	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		100,
		0,
		time.Second,
	)
	require.Equal(t, time.Second, rrh.RequestInterval())
}

func TestResolverRequestHandler_NumPeersToQuery(t *testing.T) {
	t.Parallel()

	t.Run("get returns error", func(t *testing.T) {
		t.Parallel()

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				GetCalled: func(key string) (dataRetriever.Requester, error) {
					return nil, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		_, _, err := rrh.GetNumPeersToQuery("key")
		require.Equal(t, errExpected, err)

		err = rrh.SetNumPeersToQuery("key", 1, 1)
		require.Equal(t, errExpected, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		req := &dataRetrieverMocks.RequesterStub{
			SetNumPeersToQueryCalled: func(intra int, cross int) {
				require.Equal(t, 1, intra)
				require.Equal(t, 1, cross)
			},
			NumPeersToQueryCalled: func() (int, int) {
				return 10, 10
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				GetCalled: func(key string) (dataRetriever.Requester, error) {
					return req, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		intra, cross, err := rrh.GetNumPeersToQuery("key")
		require.NoError(t, err)
		require.Equal(t, 10, intra)
		require.Equal(t, 10, cross)

		err = rrh.SetNumPeersToQuery("key", 1, 1)
		require.NoError(t, err)
	})
}

func TestResolverRequestHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var rrh *resolverRequestHandler
	require.True(t, rrh.IsInterfaceNil())

	rrh, _ = NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		100,
		0,
		time.Second,
	)
	require.False(t, rrh.IsInterfaceNil())
}
