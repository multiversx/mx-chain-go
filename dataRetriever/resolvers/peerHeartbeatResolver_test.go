package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/heartbeat"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgPeerHeartbeatResolver() resolvers.ArgPeerHeartbeatResolver {
	return resolvers.ArgPeerHeartbeatResolver{
		SenderResolver:     &mock.TopicResolverSenderStub{},
		PeerHeartbeatsPool: testscommon.NewCacherStub(),
		Marshalizer:        &mock.MarshalizerMock{},
		AntifloodHandler:   &mock.P2PAntifloodHandlerStub{},
		Throttler:          &mock.ThrottlerStub{},
		DataPacker:         &mock.DataPackerStub{},
	}
}

func TestNewPeerHeartbeatResolverResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	arg.SenderResolver = nil
	res, err := resolvers.NewPeerHeartbeatResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.True(t, check.IfNil(res))
}

func TestNewPeerHeartbeatResolverResolver_NilPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	arg.PeerHeartbeatsPool = nil
	res, err := resolvers.NewPeerHeartbeatResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilPeerHeartbeatPool, err)
	assert.True(t, check.IfNil(res))
}

func TestNewPeerHeartbeatResolverResolver_NilBlockMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	arg.Marshalizer = nil
	res, err := resolvers.NewPeerHeartbeatResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(res))
}

func TestNewPeerHeartbeatResolverResolver_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	arg.AntifloodHandler = nil
	res, err := resolvers.NewPeerHeartbeatResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	assert.True(t, check.IfNil(res))
}

func TestNewPeerHeartbeatResolverResolver_NilThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	arg.Throttler = nil
	res, err := resolvers.NewPeerHeartbeatResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilThrottler, err)
	assert.True(t, check.IfNil(res))
}

func TestNewPeerHeartbeatResolverResolver_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	arg.DataPacker = nil
	res, err := resolvers.NewPeerHeartbeatResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
	assert.True(t, check.IfNil(res))
}

func TestNewPeerHeartbeatResolverResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	res, err := resolvers.NewPeerHeartbeatResolver(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(res))
}

func TestPeerHeartbeatResolverResolver_RequestDataFromPubKeyArrayMarshalErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	arg.Marshalizer.(*mock.MarshalizerMock).Fail = true
	mbRes, err := resolvers.NewPeerHeartbeatResolver(arg)
	assert.Nil(t, err)

	err = mbRes.RequestDataFromPublicKeyArray([][]byte{[]byte("pk1")}, 0)
	require.NotNil(t, err)
}

func TestPeerHeartbeatResolverResolver_RequestDataFromPubKeyArray(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArgPeerHeartbeatResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, hashes [][]byte) error {
			called = true
			return nil
		},
	}
	mbRes, err := resolvers.NewPeerHeartbeatResolver(arg)
	assert.Nil(t, err)

	err = mbRes.RequestDataFromPublicKeyArray([][]byte{[]byte("pk1")}, 0)
	require.Nil(t, err)
	require.True(t, called)
}

func TestPeerHeartbeatResolverResolver_ProcessReceivedAntifloodErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgPeerHeartbeatResolver()
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return expectedErr
		},
	}
	mbRes, _ := resolvers.NewPeerHeartbeatResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.PubkeyArrayType, nil), fromConnectedPeerId)
	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestPeerHeartbeatResolverResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	mbRes, _ := resolvers.NewPeerHeartbeatResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.PubkeyArrayType, nil), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestPeerHeartbeatResolverResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgPeerHeartbeatResolver()
	mbRes, _ := resolvers.NewPeerHeartbeatResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, make([]byte, 0)), fromConnectedPeerId)

	assert.True(t, errors.Is(err, dataRetriever.ErrRequestTypeNotImplemented))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestPeerHeartbeatResolverResolver_ProcessReceivedMessageFoundInPoolShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	pk := []byte("pk1")
	pkList := [][]byte{pk}
	requestedBuff, merr := marshalizer.Marshal(&batch.Batch{Data: pkList})
	require.Nil(t, merr)

	wasResolved := false
	wasSent := false

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, pk) {
			wasResolved = true
			return &heartbeat.PeerHeartbeat{}, true
		}

		return nil, false
	}

	arg := createMockArgPeerHeartbeatResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			wasSent = true
			return nil
		},
	}
	arg.PeerHeartbeatsPool = cache
	mbRes, _ := resolvers.NewPeerHeartbeatResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.PubkeyArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestPeerHeartbeatResolverResolver_ProcessReceivedMessageFoundInPoolMarshalizerFailShouldErr(t *testing.T) {
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
	pk := []byte("pk1")
	pkList := [][]byte{pk}
	requestedBuff, merr := goodMarshalizer.Marshal(&batch.Batch{Data: pkList})
	require.Nil(t, merr)

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, pk) {
			return &heartbeat.PeerHeartbeat{}, true
		}

		return nil, false
	}

	arg := createMockArgPeerHeartbeatResolver()
	arg.PeerHeartbeatsPool = cache
	arg.Marshalizer = marshalizer
	mbRes, _ := resolvers.NewPeerHeartbeatResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.PubkeyArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.True(t, errors.Is(err, errExpected))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestPeerHeartbeatResolverResolver_ProcessReceivedMessageMissingDataShouldNotSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	pk := []byte("pk1")
	pkList := [][]byte{pk}
	requestedBuff, merr := marshalizer.Marshal(&batch.Batch{Data: pkList})
	require.Nil(t, merr)

	wasSent := false

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	arg := createMockArgPeerHeartbeatResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			wasSent = true
			return nil
		},
	}
	arg.PeerHeartbeatsPool = cache
	mbRes, _ := resolvers.NewPeerHeartbeatResolver(arg)

	_ = mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.PubkeyArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.False(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}
