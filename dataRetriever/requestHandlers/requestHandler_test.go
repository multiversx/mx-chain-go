package requestHandlers

import (
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var timeoutSendRequests = time.Second * 2

func TestNewMetaResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(nil, "topic")

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewMetaResolverRequestHandlerEmptyTopic(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(&mock.ResolversFinderStub{}, "")

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyHeaderRequestTopic, err)
}

func TestNewMetaResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(&mock.ResolversFinderStub{}, "topic")

	assert.Nil(t, err)
	assert.NotNil(t, rrh)
}

func TestNewShardResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(nil, "topic", "topic", "topic", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewShardResolverRequestHandlerTxTopicEmpty(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "", "topic", "topic", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyTxRequestTopic, err)
}

func TestNewShardResolverRequestHandlerMBTopicEmpty(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "", "topic", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyMiniBlockRequestTopic, err)
}

func TestNewShardResolverRequestHandlerHdrTopicEmpty(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "topic", "", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyHeaderRequestTopic, err)
}

func TestNewShardResolverRequestHandlerMaxTxRequestTooSmall(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "topic", "topic", 0)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrInvalidMaxTxRequest, err)
}

func TestNewShardResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "topic", "topic", 1)

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
		"txTopic",
		"topic",
		"topic",
		1,
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
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestTransaction(0, make([][]byte, 0))
}

func TestResolverRequestHandler_RequestTransactionShouldRequestTransactions(t *testing.T) {
	t.Parallel()

	chTxRequested := make(chan struct{})
	txResolver := &mock.HashSliceResolverStub{
		RequestDataFromHashArrayCalled: func(hashes [][]byte) error {
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
		"txTopic",
		"topic",
		"topic",
		1,
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
		RequestDataFromHashArrayCalled: func(hashes [][]byte) error {
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
		"txTopic",
		"topic",
		"topic",
		1,
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
		"txTopic",
		"topic",
		"topic",
		1,
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
		RequestDataFromHashCalled: func(hash []byte) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, crossShard uint32) (resolver dataRetriever.Resolver, e error) {
				return mbResolver, nil
			},
		},
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestMiniBlock(0, []byte("mbHash"))
}

func TestResolverRequestHandler_RequestMiniBlockShouldCallRequestOnResolver(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte) error {
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
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestMiniBlock(0, []byte("mbHash"))

	assert.True(t, wasCalled)
}

//------- RequestHeader

func TestResolverRequestHandler_RequestHeaderShouldCallRequestOnResolver(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mbResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte) error {
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
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestHeader(0, []byte("hdrHash"))

	assert.True(t, wasCalled)
}

//------- RequestHeaderByNonce

func TestResolverRequestHandler_RequestHeaderByNonceShardFinderReturnsErrorShouldNotPanic(t *testing.T) {
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
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return nil, errExpected
			},
		},
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestHeaderByNonce(0, 0)
}

func TestResolverRequestHandler_RequestHeaderByNonceShardFinderReturnsAWrongResolverShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	hdrResolver := &mock.ResolverStub{
		RequestDataFromHashCalled: func(hash []byte) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestHeaderByNonce(0, 0)
}

func TestResolverRequestHandler_RequestHeaderByNonceShardResolverFailsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	errExpected := errors.New("expected error")
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromHashCalled: func(hash []byte) error {
			return errExpected
		},
	}

	rrh, _ := NewShardResolverRequestHandler(
		&mock.ResolversFinderStub{
			MetaChainResolverCalled: func(baseTopic string) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestHeaderByNonce(0, 0)
}

func TestResolverRequestHandler_RequestHeaderByNonceShardShouldRequest(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromNonceCalled: func(nonce uint64) error {
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
		"txTopic",
		"topic",
		"topic",
		1,
	)

	rrh.RequestHeaderByNonce(0, 0)

	assert.True(t, wasCalled)
}

func TestResolverRequestHandler_RequestHeaderByNonceMetaShouldRequest(t *testing.T) {
	t.Parallel()

	wasCalled := false
	hdrResolver := &mock.HeaderResolverStub{
		RequestDataFromNonceCalled: func(nonce uint64) error {
			wasCalled = true
			return nil
		},
	}

	rrh, _ := NewMetaResolverRequestHandler(
		&mock.ResolversFinderStub{
			CrossShardResolverCalled: func(baseTopic string, destShardID uint32) (resolver dataRetriever.Resolver, e error) {
				return hdrResolver, nil
			},
		},
		"topic",
	)

	rrh.RequestHeaderByNonce(0, 0)

	assert.True(t, wasCalled)
}
