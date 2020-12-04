package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fromConnectedPeerId = core.PeerID("from connected peer Id")

func createMockArgMiniblockResolver() resolvers.ArgMiniblockResolver {
	return resolvers.ArgMiniblockResolver{
		SenderResolver:   &mock.TopicResolverSenderStub{},
		MiniBlockPool:    testscommon.NewCacherStub(),
		MiniBlockStorage: &mock.StorerStub{},
		Marshalizer:      &mock.MarshalizerMock{},
		AntifloodHandler: &mock.P2PAntifloodHandlerStub{},
		Throttler:        &mock.ThrottlerStub{},
		DataPacker:       &mock.DataPackerStub{},
	}
}

//------- NewMiniblockResolver

func TestNewMiniblockResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = nil
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.True(t, check.IfNil(mbRes))
}

func TestNewMiniblockResolver_NilBlockBodyPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.MiniBlockPool = nil
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilMiniblocksPool, err)
	assert.True(t, check.IfNil(mbRes))
}

func TestNewMiniblockResolver_NilBlockBodyStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.MiniBlockStorage = nil
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilMiniblocksStorage, err)
	assert.True(t, check.IfNil(mbRes))
}

func TestNewMiniblockResolver_NilBlockMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.Marshalizer = nil
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(mbRes))
}

func TestNewMiniblockResolver_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.AntifloodHandler = nil
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	assert.True(t, check.IfNil(mbRes))
}

func TestNewMiniblockResolver_NilThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.Throttler = nil
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilThrottler, err)
	assert.True(t, check.IfNil(mbRes))
}

func TestNewMiniblockResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(mbRes))
}

func TestMiniblockResolver_RequestDataFromHashArrayMarshalErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.Marshalizer.(*mock.MarshalizerMock).Fail = true
	mbRes, err := resolvers.NewMiniblockResolver(arg)
	assert.Nil(t, err)

	err = mbRes.RequestDataFromHashArray([][]byte{[]byte("hash")}, 0)
	require.NotNil(t, err)
}

func TestMiniblockResolver_RequestDataFromHashArray(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, hashes [][]byte) error {
			called = true
			return nil
		},
	}
	mbRes, err := resolvers.NewMiniblockResolver(arg)
	assert.Nil(t, err)

	err = mbRes.RequestDataFromHashArray([][]byte{[]byte("hash")}, 0)
	require.Nil(t, err)
	require.True(t, called)
}

//------- ProcessReceivedMessage

func TestMiniblockResolver_ProcessReceivedAntifloodErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgMiniblockResolver()
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return expectedErr
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId)
	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestMiniblockResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestMiniblockResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, make([]byte, 0)), fromConnectedPeerId)

	assert.True(t, errors.Is(err, dataRetriever.ErrRequestTypeNotImplemented))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestMiniblockResolver_ProcessReceivedMessageFoundInPoolShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	mbHash := []byte("aaa")
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, merr := marshalizer.Marshal(&batch.Batch{Data: miniBlockList})

	assert.Nil(t, merr)

	wasResolved := false
	wasSent := false

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, mbHash) {
			wasResolved = true
			return &block.MiniBlock{}, true
		}

		return nil, false
	}

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
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
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestMiniblockResolver_ProcessReceivedMessageFoundInPoolMarshalizerFailShouldErr(t *testing.T) {
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
	requestedBuff, merr := goodMarshalizer.Marshal(&batch.Batch{Data: miniBlockList})

	assert.Nil(t, merr)

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, mbHash) {
			return &block.MiniBlock{}, true
		}

		return nil, false
	}

	arg := createMockArgMiniblockResolver()
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			body := block.MiniBlock{}
			buff, _ := goodMarshalizer.Marshal(&body)
			return buff, nil
		},
	}
	arg.Marshalizer = marshalizer
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.True(t, errors.Is(err, errExpected))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestMiniblockResolver_ProcessReceivedMessageNotFoundInPoolShouldRetFromStorageAndSend(t *testing.T) {
	t.Parallel()

	mbHash := []byte("aaa")
	marshalizer := &mock.MarshalizerMock{}
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, _ := marshalizer.Marshal(&batch.Batch{Data: miniBlockList})

	wasResolved := false
	wasSend := false

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &mock.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		wasResolved = true
		mb, _ := marshalizer.Marshal(&block.MiniBlock{})
		return mb, nil
	}

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			wasSend = true
			return nil
		},
	}
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = store
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestMiniblockResolver_ProcessReceivedMessageMissingDataShouldNotSend(t *testing.T) {
	t.Parallel()

	mbHash := []byte("aaa")
	marshalizer := &mock.MarshalizerMock{}
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, _ := marshalizer.Marshal(&batch.Batch{Data: miniBlockList})

	wasSent := false

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &mock.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		return nil, errors.New("key not found")
	}

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			wasSent = true
			return nil
		},
	}
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = store
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	_ = mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashType, requestedBuff),
		fromConnectedPeerId,
	)

	assert.False(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

//------- Requests

func TestMiniblockResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	buffRequested := []byte("aaaa")

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, hashes [][]byte) error {
			wasCalled = true
			return nil
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	assert.Nil(t, mbRes.RequestDataFromHash(buffRequested, 0))
	assert.True(t, wasCalled)
}

//------ NumPeersToQuery setter and getter

func TestMiniblockResolver_SetAndGetNumPeersToQuery(t *testing.T) {
	t.Parallel()

	expectedIntra := 5
	expectedCross := 7

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		GetNumPeersToQueryCalled: func() (int, int) {
			return expectedIntra, expectedCross
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	mbRes.SetNumPeersToQuery(expectedIntra, expectedCross)
	actualIntra, actualCross := mbRes.NumPeersToQuery()
	assert.Equal(t, expectedIntra, actualIntra)
	assert.Equal(t, expectedCross, actualCross)
}
