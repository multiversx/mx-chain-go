package requestHandlers

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var timeoutSendRequests = time.Second * 2

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

//------- NewMetaResolver

func TestNewMetaResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		nil,
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		100,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewMetaResolverRequestHandlerNilRequestedItemsHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{},
		nil,
		&mock.EpochHandlerStub{},
		100,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilRequestedItemsHandler, err)
}

func TestNewMetaResolverRequestHandlerNilEpochHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		nil,
		100,
	)
	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilEpochHandler, err)
}

func TestNewMetaResolverRequestHandlerMaxTxRequestTooSmall(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrInvalidMaxTxRequest, err)
}

func TestNewMetaResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		100,
	)
	assert.Nil(t, err)
	assert.NotNil(t, rrh)
}

//------- NewShardResolver

func TestNewShardResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		nil,
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewShardResolverRequestHandlerNilRequestedItemsHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{},
		nil,
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilRequestedItemsHandler, err)
}

func TestNewShardResolverRequestHandlerNilEpochHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		nil,
		1,
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilEpochHandler, err)
}

func TestNewShardResolverRequestHandlerMaxTxRequestTooSmall(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		0,
		0,
	)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrInvalidMaxTxRequest, err)
}

func TestNewShardResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	assert.Nil(t, err)
	assert.NotNil(t, rrh)
}

//------- RequestTransaction

func TestResolverRequestHandler_RequestTransactionErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return wrongTxResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

func TestResolverRequestHandler_RequestTransactionErrorsOnRequestShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestTransaction(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

//------- RequestMiniBlock

func TestResolverRequestHandler_RequestMiniBlockErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	errExpected := errors.New("expected error")
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestMiniBlock(0, []byte("mbHash"))

	assert.True(t, wasCalled)
}

func TestResolverRequestHandler_RequestMiniBlockShouldCallWithTheCorrectEpoch(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(7)
	epochHandler := &mock.EpochHandlerStub{
		EpochCalled: func() uint32 {
			return expectedEpoch
		},
	}
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			assert.Equal(t, expectedEpoch, epoch)
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		epochHandler,
		1,
		0,
	)

	rrh.RequestMiniBlock(0, []byte("mbHash"))
}

//------- RequestShardHeader

func TestResolverRequestHandler_RequestShardHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestShardHeader(0, make([]byte, 0))
}

func TestResolverRequestHandler_RequestShardHeaderHashBadRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestShardHeader(0, []byte("hdrHash"))

	assert.True(t, wasCalled)
}

//------- RequestMetaHeader

func TestResolverRequestHandler_RequestMetadHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestMetaHeader([]byte("hdrHash"))

	assert.True(t, wasCalled)
}

//------- RequestShardHeaderByNonce

func TestResolverRequestHandler_RequestShardHeaderByNonceAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(0, 0)
}

func TestResolverRequestHandler_RequestShardHeaderByNonceBadRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(1, 0)
}

func TestResolverRequestHandler_RequestShardHeaderByNonceFinderReturnsErrorShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	errExpected := errors.New("expected error")
	hdrResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	errExpected := errors.New("expected error")
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte, epoch uint32) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, shardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestShardHeaderByNonce(0, 0)

	assert.True(t, wasCalled)
}

//------- RequestMetaHeaderByNonce

func TestResolverRequestHandler_RequestMetaHeaderHashAlreadyRequestedShouldNotRequest(t *testing.T) {
	t.Parallel()

	rrh, _ := NewShardResolverRequestHandler(
		createResolversFinderStubThatShouldNotBeCalled(t),
		&mock.RequestedItemsHandlerStub{
			HasCalled: func(key string) bool {
				return true
			},
		},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		100,
		0,
	)

	rrh.RequestMetaHeaderByNonce(0)

	assert.True(t, wasCalled)
}

//------- RequestSmartContractResult

func TestResolverRequestHandler_RequestScrErrorWhenGettingCrossShardResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return wrongTxResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
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

	errExpected := errors.New("expected error")
	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestUnsignedTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}

//------- RequestRewardTransaction

func TestResolverRequestHandler_RequestRewardShouldRequestReward(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte, epoch uint32) error {
			chTxRequested <- struct{}{}
			return nil
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return txResolver, nil
			},
		},
		&mock.RequestedItemsHandlerStub{},
		&mock.EpochHandlerStub{},
		1,
		0,
	)

	rrh.RequestRewardTransactions(0, [][]byte{[]byte("txHash")})

	select {
	case <-chTxRequested:
	case <-time.After(timeoutSendRequests):
		assert.Fail(t, "timeout while waiting to call RequestDataFromHashArray")
	}

	time.Sleep(time.Second)
}
