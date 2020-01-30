package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fromConnectedPeerId = p2p.PeerID("from connected peer Id")

func createMockP2PAntifloodHandler() *mock.P2PAntifloodHandlerStub {
	return &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
			return nil
		},
	}
}

//------- NewBlockBodyResolver

func TestNewGenericBlockBodyResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		nil,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockBodyPoolShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilBlockBodyPool, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockBodyStorageShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		nil,
		&mock.MarshalizerMock{},
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilBlockBodyStorage, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockMArshalizerShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		nil,
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		createMockP2PAntifloodHandler(),
	)

	assert.Nil(t, err)
	assert.NotNil(t, gbbRes)
}

func TestRequestDataFromHashArray_MarshalErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{
			Fail: true,
		},
	)
	assert.Nil(t, err)

	err = gbbRes.RequestDataFromHashArray([][]byte{[]byte("hash")}, 0)
	require.NotNil(t, err)
}

func TestRequestDataFromHashArray(t *testing.T) {
	t.Parallel()

	called := false
	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
				called = true
				return nil
			},
		},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, err)

	err = gbbRes.RequestDataFromHashArray([][]byte{[]byte("hash")}, 0)
	require.Nil(t, err)
	require.True(t, called)
}

//------- ProcessReceivedMessage

func TestNewGenericBlockBodyResolver_ProcessReceivedAntifloodErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		&mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
				return expectedErr
			},
		},
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId)
	assert.Equal(t, expectedErr, err)
}

func TestNewGenericBlockBodyResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		createMockP2PAntifloodHandler(),
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		createMockP2PAntifloodHandler(),
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, make([]byte, 0)), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrInvalidRequestType, err)
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

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		cache,
		&mock.StorerStub{
			GetCalled: func(key []byte) (i []byte, e error) {
				return make([]byte, 0), nil
			},
		},
		marshalizer,
		createMockP2PAntifloodHandler(),
	)

	err := gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
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

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		cache,
		&mock.StorerStub{
			GetCalled: func(key []byte) (i []byte, e error) {
				body := block.MiniBlock{}
				buff, _ := goodMarshalizer.Marshal(&body)
				return buff, nil
			},
		},
		marshalizer,
		createMockP2PAntifloodHandler(),
	)

	err := gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Equal(t, errExpected, err)

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

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSend = true
				return nil
			},
		},
		cache,
		store,
		marshalizer,
		createMockP2PAntifloodHandler(),
	)

	err := gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
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

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		cache,
		store,
		marshalizer,
		createMockP2PAntifloodHandler(),
	)

	_ = gbbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.False(t, wasSent)
}

//------- Requests

func TestBlockBodyResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	buffRequested := []byte("aaaa")

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
				wasCalled = true
				return nil
			},
		},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		createMockP2PAntifloodHandler(),
	)

	assert.Nil(t, gbbRes.RequestDataFromHash(buffRequested, 0))
	assert.True(t, wasCalled)
}
