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
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var timeoutSendRequests = time.Second * 2
var errExpected = errors.New("expected error")

func createResolversFinderStubThatShouldNotBeCalled(tb testing.TB) *mock.ResolversFinderStub {
	return &mock.ResolversFinderStub{
		IntraShardResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			assert.Fail(tb, "IntraShardResolverCalled should not have been called")
			return nil, nil
		},
		MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
			assert.Fail(tb, "MetaChainResolverCalled should not have been called")
			return nil, nil
		},
		CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
			assert.Fail(tb, "CrossShardResolverCalled should not have been called")
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
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewResolverRequestHandlerNilRequestedItemsHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewResolverRequestHandler(
		&mock.ResolversFinderStub{},
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
		&mock.ResolversFinderStub{},
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
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	assert.Nil(t, err)
	assert.NotNil(t, rrh)
}

// ------- RequestTransaction

func TestResolverRequestHandler_RequestTransactionErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
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

	wrongTxResolver := &mock.HeaderResolverStub{}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return wrongTxResolver, nil
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
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
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
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			atomic.AddUint32(&numRequests, 1)
			return nil
		},
	}

	timeSpan := time.Second
	cache := timecache.NewTimeCache(timeSpan)
	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		cache,
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
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return errExpected
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
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

// ------- RequestMiniBlock

func TestResolverRequestHandler_RequestMiniBlockErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
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

	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
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
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
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
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			assert.Equal(t, expectedEpoch, epoch)
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
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

// ------- RequestShardHeader

func TestResolverRequestHandler_RequestShardHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
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
		createResolversFinderStubThatShouldNotBeCalled(t),
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
	mbResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
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

// ------- RequestMetaHeader

func TestResolverRequestHandler_RequestMetadHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
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
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
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
	mbResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
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

// ------- RequestShardHeaderByNonce

func TestResolverRequestHandler_RequestShardHeaderByNonceAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	called := false
	rrh, _ := NewResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
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
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, err error) {
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
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
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

	hdrResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
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

	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
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
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
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

// ------- RequestMetaHeaderByNonce

func TestResolverRequestHandler_RequestMetaHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
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
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromNonceCalled: func(nonce uint64, epoch uint32) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
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

// ------- RequestSmartContractResult

func TestResolverRequestHandler_RequestScrErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
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

	wrongTxResolver := &mock.HeaderResolverStub{}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return wrongTxResolver, nil
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
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
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
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return errExpected
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
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

// ------- RequestRewardTransaction

func TestResolverRequestHandler_RequestRewardShouldRequestReward(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
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
	resolverMock := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hash [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaCrossShardResolverCalled: func(baseTopic string, crossShard uint32) (dataRetriever.Resolver, error) {
				return resolverMock, nil
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
		&mock.ResolversFinderStub{
			MetaCrossShardResolverCalled: func(baseTopic string, shId uint32) (resolver dataRetriever.Resolver, err error) {
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
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
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
	resolverMock := &mock.HashSliceResolverStub{}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				called = true
				return resolverMock, nil
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
	resolverMock := &mock.HeaderResolverStub{
		RequestDataFromEpochCalled: func(identifier []byte) error {
			called = true
			return localError
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				return resolverMock, nil
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
	resolverMock := &mock.HeaderResolverStub{
		RequestDataFromEpochCalled: func(identifier []byte) error {
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				return resolverMock, nil
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
	resolverMock := &mock.ChunkResolverStub{
		RequestDataFromReferenceAndChunkCalled: func(hash []byte, chunkIndex uint32) error {
			chTxRequested <- struct{}{}
			return localErr
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
				return resolverMock, nil
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
	resolverMock := &mock.ChunkResolverStub{
		RequestDataFromReferenceAndChunkCalled: func(hash []byte, chunkIndex uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
				return resolverMock, nil
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
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
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
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, err error) {
				called = true
				return &mock.HashSliceResolverStub{}, nil
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

func TestResolverRequestHandler_RequestPeerAuthenticationsChunk(t *testing.T) {
	t.Parallel()

	providedChunkId := uint32(123)
	providedShardId := uint32(15)
	t.Run("CrossShardResolver returns error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		paResolver := &mock.PeerAuthenticationResolverStub{
			RequestDataFromChunkCalled: func(chunkIndex uint32, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paResolver, errExpected
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsChunk(providedShardId, providedChunkId)
		assert.False(t, wasCalled)
	})
	t.Run("cast fails", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		mbResolver := &mock.ResolverStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return mbResolver, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsChunk(providedShardId, providedChunkId)
		assert.False(t, wasCalled)
	})
	t.Run("RequestDataFromChunk returns error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		paResolver := &mock.PeerAuthenticationResolverStub{
			RequestDataFromChunkCalled: func(chunkIndex uint32, epoch uint32) error {
				wasCalled = true
				assert.Equal(t, providedChunkId, chunkIndex)
				return errExpected
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paResolver, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsChunk(providedShardId, providedChunkId)
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
		paResolver := &mock.PeerAuthenticationResolverStub{
			RequestDataFromChunkCalled: func(chunkIndex uint32, epoch uint32) error {
				wasCalled = true
				assert.Equal(t, providedChunkId, chunkIndex)
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paResolver, nil
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			1,
			0,
			time.Second,
		)

		rrh.RequestPeerAuthenticationsChunk(providedShardId, providedChunkId)
		assert.True(t, wasCalled)
	})
}

func TestResolverRequestHandler_RequestPeerAuthenticationsByHashes(t *testing.T) {
	t.Parallel()

	providedHashes := [][]byte{[]byte("h1"), []byte("h2")}
	providedShardId := uint32(15)
	t.Run("CrossShardResolver returns error", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		paResolver := &mock.PeerAuthenticationResolverStub{
			RequestDataFromChunkCalled: func(chunkIndex uint32, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paResolver, errExpected
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
		mbResolver := &mock.ResolverStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return mbResolver, errExpected
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
		paResolver := &mock.PeerAuthenticationResolverStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				wasCalled = true
				assert.Equal(t, providedHashes, hashes)
				return errExpected
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paResolver, nil
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
		paResolver := &mock.PeerAuthenticationResolverStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				wasCalled = true
				assert.Equal(t, providedHashes, hashes)
				return nil
			},
		}
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (dataRetriever.Resolver, error) {
					assert.Equal(t, common.PeerAuthenticationTopic, baseTopic)
					return paResolver, nil
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

	t.Run("MetaChainResolver returns error", func(t *testing.T) {
		providedHash := []byte("provided hash")
		wasCalled := false
		res := &mock.ResolverStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
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
		res := &mock.ResolverStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				assert.True(t, bytes.Equal(providedHash, hash))
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
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

	t.Run("MetaChainResolver returns error", func(t *testing.T) {
		providedHash := []byte("provided hash")
		wasCalled := false
		res := &mock.ResolverStub{
			RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
					return res, errors.New("provided err")
				},
			},
			&mock.RequestedItemsHandlerStub{},
			&mock.WhiteListHandlerStub{},
			100,
			0,
			time.Second,
		)

		rrh.RequestValidatorsInfo([][]byte{providedHash}, 0)
		assert.False(t, wasCalled)
	})
	t.Run("cast fails", func(t *testing.T) {
		providedHash := []byte("provided hash")
		mbResolver := &mock.ResolverStub{} // uncastable to HashSliceResolver
		wasCalled := false
		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
					return mbResolver, nil
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

		rrh.RequestValidatorsInfo([][]byte{providedHash}, 0)
		assert.False(t, wasCalled)
	})
	t.Run("should work", func(t *testing.T) {
		providedHashes := [][]byte{[]byte("provided hash 1"), []byte("provided hash 2")}
		wasCalled := false
		res := &mock.HashSliceResolverStub{
			RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
				assert.Equal(t, providedHashes, hashes)
				wasCalled = true
				return nil
			},
		}

		rrh, _ := NewResolverRequestHandler(
			&mock.ResolversFinderStub{
				MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
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

		rrh.RequestValidatorsInfo(providedHashes, 0)
		assert.True(t, wasCalled)
	})
}
