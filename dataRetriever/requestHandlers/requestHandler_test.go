package requestHandlers

import (
	"bytes"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/storage/cache"
	dataRetrieverMocks "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
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

func TestNewResolverRequestHandlerNilFinder(t *testing.T) {
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
}

func TestNewResolverRequestHandlerNilRequestedItemsHandler(t *testing.T) {
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
}

func TestNewResolverRequestHandlerMaxTxRequestTooSmall(t *testing.T) {
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
}

func TestNewResolverRequestHandler(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestTransactionErrorWhenGettingCrossShardRequesterShouldNotPanic(t *testing.T) {
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

	rrh.RequestTransaction(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestTransactionWrongResolverShouldNotPanic(t *testing.T) {
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

	rrh.RequestTransaction(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestTransactionShouldRequestTransactions(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestTransactionShouldRequest4TimesIfDifferentShardsAndEnoughTime(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestTransactionErrorsOnRequestShouldNotPanic(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestMiniBlockErrorWhenGettingCrossShardRequesterShouldNotPanic(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestMiniBlockErrorsOnRequestShouldNotPanic(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestMiniBlockShouldCallRequestOnResolver(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestMiniBlockShouldCallWithTheCorrectEpoch(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestShardHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestShardHeaderHashBadRequest(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestShardHeaderShouldCallRequestOnResolver(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mbRequester := &mock.HeaderResolverStub{
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
}

func TestResolverRequestHandler_RequestMetadHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
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

	rrh.RequestMetaHeader(make([]byte, 0))
}

func TestResolverRequestHandler_RequestMetadHeaderHashNotHeaderResolverShouldNotRequest(t *testing.T) {
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

	assert.False(t, wasCalled)
}

func TestResolverRequestHandler_RequestMetaHeaderShouldCallRequestOnResolver(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestShardHeaderByNonceAlreadyRequestedShouldNotRequest(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestShardHeaderByNonceBadRequest(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	called := false
	rrh, _ := NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{
			CrossShardRequesterCalled: func(baseTopic string, crossShard uint32) (requester dataRetriever.Requester, err error) {
				called = true
				return nil, localErr
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
}

func TestResolverRequestHandler_RequestShardHeaderByNonceFinderReturnsErrorShouldNotPanic(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestShardHeaderByNonceFinderReturnsAWrongResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	hdrRequester := &dataRetrieverMocks.RequesterStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
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
}

func TestResolverRequestHandler_RequestShardHeaderByNonceResolverFailsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	hdrRequester := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
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
}

func TestResolverRequestHandler_RequestShardHeaderByNonceShouldRequest(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hdrRequester := &mock.HeaderResolverStub{
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
}

func TestResolverRequestHandler_RequestMetaHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
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
}

func TestResolverRequestHandler_RequestMetaHeaderByNonceShouldRequest(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hdrRequester := &mock.HeaderResolverStub{
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

func TestRequestTrieNodes_ShouldWork(t *testing.T) {
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
}

func TestRequestTrieNodes_NilResolver(t *testing.T) {
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
}

func TestRequestStartOfEpochMetaBlock_MissingResolver(t *testing.T) {
	t.Parallel()

	called := false
	localError := errors.New("test error")
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

	rrh.RequestStartOfEpochMetaBlock(0)
	assert.True(t, called)
}

func TestRequestStartOfEpochMetaBlock_WrongResolver(t *testing.T) {
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
}

func TestRequestStartOfEpochMetaBlock_RequestDataFromEpochError(t *testing.T) {
	t.Parallel()

	called := false
	localError := errors.New("test error")
	requesterMock := &dataRetrieverMocks.EpochRequesterStub{
		RequestDataFromEpochCalled: func(identifier []byte) error {
			called = true
			return localError
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
}

func TestRequestStartOfEpochMetaBlock_AddError(t *testing.T) {
	t.Parallel()

	called := false
	localError := errors.New("test error")
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
				return localError
			},
		},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	rrh.RequestStartOfEpochMetaBlock(0)
	assert.True(t, called)
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

//------- RequestPeerAuthentications

func TestResolverRequestHandler_RequestPeerAuthenticationsByHashes(t *testing.T) {
	t.Parallel()

	providedHashes := [][]byte{[]byte("h1"), []byte("h2")}
	providedShardId := uint32(15)
	t.Run("CrossShardRequester returns error", func(t *testing.T) {
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

		wasCalled := false
		mbRequester := &dataRetrieverMocks.RequesterStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&dataRetrieverMocks.RequestersFinderStub{
				MetaChainRequesterCalled: func(baseTopic string) (dataRetriever.Requester, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return mbRequester, errExpected
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
			&mock.RequestedItemsHandlerStub{},
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

	t.Run("MetaChainRequester returns error", func(t *testing.T) {
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
					return res, errors.New("provided err")
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
	t.Run("should work", func(t *testing.T) {
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

	t.Run("MetaChainRequester returns error", func(t *testing.T) {
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
					return res, errors.New("provided err")
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
	t.Run("cast fails", func(t *testing.T) {
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
