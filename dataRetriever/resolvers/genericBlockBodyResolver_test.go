package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fromConnectedPeerId = p2p.PeerID("from connected peer Id")

func createMockArgGenericBlockBodyResolver() resolvers.ArgGenericBlockBodyResolver {
	return resolvers.ArgGenericBlockBodyResolver{
		SenderResolver:   &mock.TopicResolverSenderStub{},
		MiniBlockPool:    &mock.CacherStub{},
		MiniBlockStorage: &mock.StorerStub{},
		Marshalizer:      &mock.MarshalizerMock{},
		AntifloodHandler: &mock.P2PAntifloodHandlerStub{},
		Throttler:        &mock.ThrottlerStub{},
	}
}

//------- NewBlockBodyResolver

func TestNewGenericBlockBodyResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	arg.SenderResolver = nil
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.True(t, check.IfNil(gbbRes))
}

func TestNewGenericBlockBodyResolver_NilBlockBodyPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	arg.MiniBlockPool = nil
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilBlockBodyPool, err)
	assert.True(t, check.IfNil(gbbRes))
}

func TestNewGenericBlockBodyResolver_NilBlockBodyStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	arg.MiniBlockStorage = nil
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilBlockBodyStorage, err)
	assert.True(t, check.IfNil(gbbRes))
}

func TestNewGenericBlockBodyResolver_NilBlockMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	arg.Marshalizer = nil
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(gbbRes))
}

func TestNewGenericBlockBodyResolver_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	arg.AntifloodHandler = nil
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	assert.True(t, check.IfNil(gbbRes))
}

func TestNewGenericBlockBodyResolver_NilThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	arg.Throttler = nil
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilThrottler, err)
	assert.True(t, check.IfNil(gbbRes))
}

func TestNewGenericBlockBodyResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(gbbRes))
}

func TestRequestDataFromHashArray_MarshalErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	arg.Marshalizer.(*mock.MarshalizerMock).Fail = true
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)
	assert.Nil(t, err)

	err = gbbRes.RequestDataFromHashArray([][]byte{[]byte("hash")}, 0)
	require.NotNil(t, err)
}

func TestRequestDataFromHashArray(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArgGenericBlockBodyResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
			called = true
			return nil
		},
	}
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(arg)
	assert.Nil(t, err)

	err = gbbRes.RequestDataFromHashArray([][]byte{[]byte("hash")}, 0)
	require.Nil(t, err)
	require.True(t, called)
}

//------- ProcessReceivedMessage

func TestNewGenericBlockBodyResolver_ProcessReceivedAntifloodErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgGenericBlockBodyResolver()
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
			return expectedErr
		},
	}
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId)
	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestNewGenericBlockBodyResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgGenericBlockBodyResolver()
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, make([]byte, 0)), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrInvalidRequestType, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageFoundInPoolShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	mbHash := []byte("aaa")
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, _ := marshalizer.Marshal(miniBlockList)

	wasResolved := false
	wasSent := false

	cache := &mock.CacherStub{}
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, mbHash) {
			wasResolved = true
			return &block.MiniBlock{}, true
		}

		return nil, false
	}

	arg := createMockArgGenericBlockBodyResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer p2p.PeerID) error {
			wasSent = true
			return nil
		},
	}
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return make([]byte, 0), nil
		},
	}
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	err := gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageFoundInPoolMarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected error")
	goodMarshalizer := &mock.MarshalizerMock{}
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			return nil, errExpected
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {

			return goodMarshalizer.Unmarshal(obj, buff)
		},
	}
	mbHash := []byte("aaa")
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, _ := goodMarshalizer.Marshal(miniBlockList)

	cache := &mock.CacherStub{}
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, mbHash) {
			return &block.MiniBlock{}, true
		}

		return nil, false
	}

	arg := createMockArgGenericBlockBodyResolver()
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			body := block.MiniBlock{}
			buff, _ := goodMarshalizer.Marshal(&body)
			return buff, nil
		},
	}
	arg.Marshalizer = marshalizer
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	err := gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Equal(t, errExpected, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageNotFoundInPoolShouldRetFromStorageAndSend(t *testing.T) {
	t.Parallel()

	mbHash := []byte("aaa")
	marshalizer := &mock.MarshalizerMock{}
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, _ := marshalizer.Marshal(miniBlockList)

	wasResolved := false
	wasSend := false

	cache := &mock.CacherStub{}
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &mock.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		wasResolved = true
		mb, _ := marshalizer.Marshal(&block.MiniBlock{})
		return mb, nil
	}

	arg := createMockArgGenericBlockBodyResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer p2p.PeerID) error {
			wasSend = true
			return nil
		},
	}
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = store
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	err := gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageMissingDataShouldNotSend(t *testing.T) {
	t.Parallel()

	mbHash := []byte("aaa")
	marshalizer := &mock.MarshalizerMock{}
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, _ := marshalizer.Marshal(miniBlockList)

	wasSent := false

	cache := &mock.CacherStub{}
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &mock.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		return nil, errors.New("key not found")
	}

	arg := createMockArgGenericBlockBodyResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer p2p.PeerID) error {
			wasSent = true
			return nil
		},
	}
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = store
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	_ = gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.False(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

//------- Requests

func TestBlockBodyResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	buffRequested := []byte("aaaa")

	arg := createMockArgGenericBlockBodyResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
			wasCalled = true
			return nil
		},
	}
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(arg)

	assert.Nil(t, gbbRes.RequestDataFromHash(buffRequested, 0))
	assert.True(t, wasCalled)
}
