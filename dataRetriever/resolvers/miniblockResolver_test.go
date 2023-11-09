package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

var fromConnectedPeerId = core.PeerID("from connected peer Id")

func createMockArgMiniblockResolver() resolvers.ArgMiniblockResolver {
	return resolvers.ArgMiniblockResolver{
		ArgBaseResolver:  createMockArgBaseResolver(),
		MiniBlockPool:    testscommon.NewCacherStub(),
		MiniBlockStorage: &storageStubs.StorerStub{},
		DataPacker:       &mock.DataPackerStub{},
	}
}

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
	arg.Marshaller = nil
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

func TestNewMiniblockResolver_NilDataPackerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.DataPacker = nil
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilDataPacker, err)
	assert.True(t, check.IfNil(mbRes))
}

func TestNewMiniblockResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	mbRes, err := resolvers.NewMiniblockResolver(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(mbRes))
}

func TestMiniblockResolver_ProcessReceivedAntifloodErrorsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return expectedErr
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId, &p2pmocks.MessengerStub{})
	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestMiniblockResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, nil), fromConnectedPeerId, &p2pmocks.MessengerStub{})
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestMiniblockResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, make([]byte, 0)), fromConnectedPeerId, &p2pmocks.MessengerStub{})

	assert.True(t, errors.Is(err, dataRetriever.ErrRequestTypeNotImplemented))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
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
		SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
			wasSent = true
			return nil
		},
	}
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			return make([]byte, 0), nil
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
		&p2pmocks.MessengerStub{},
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
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
	arg.MiniBlockStorage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			body := block.MiniBlock{}
			buff, _ := goodMarshalizer.Marshal(&body)
			return buff, nil
		},
	}
	arg.Marshaller = marshalizer
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
		&p2pmocks.MessengerStub{},
	)

	assert.True(t, errors.Is(err, errExpected))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestMiniblockResolver_ProcessReceivedMessageUnmarshalFails(t *testing.T) {
	t.Parallel()

	goodMarshalizer := &mock.MarshalizerMock{}
	cnt := 0
	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: goodMarshalizer.Marshal,
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			cnt++
			if cnt > 1 {
				return expectedErr
			}
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
		return nil, false
	}

	arg := createMockArgMiniblockResolver()
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			body := block.MiniBlock{}
			buff, _ := goodMarshalizer.Marshal(&body)
			return buff, nil
		},
	}
	arg.Marshaller = marshalizer
	arg.DataPacker = &mock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			assert.Fail(t, "should not have been called")
			return nil, nil
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
		&p2pmocks.MessengerStub{},
	)

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestMiniblockResolver_ProcessReceivedMessagePackDataInChunksFails(t *testing.T) {
	t.Parallel()

	goodMarshalizer := &mock.MarshalizerMock{}
	mbHash := []byte("aaa")
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, merr := goodMarshalizer.Marshal(&batch.Batch{Data: miniBlockList})

	assert.Nil(t, merr)

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	arg := createMockArgMiniblockResolver()
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			body := block.MiniBlock{}
			buff, _ := goodMarshalizer.Marshal(&body)
			return buff, nil
		},
	}
	arg.Marshaller = goodMarshalizer
	arg.DataPacker = &mock.DataPackerStub{
		PackDataInChunksCalled: func(data [][]byte, limit int) ([][]byte, error) {
			return nil, expectedErr
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
		&p2pmocks.MessengerStub{},
	)

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestMiniblockResolver_ProcessReceivedMessageSendFails(t *testing.T) {
	t.Parallel()

	goodMarshalizer := &mock.MarshalizerMock{}
	mbHash := []byte("aaa")
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, merr := goodMarshalizer.Marshal(&batch.Batch{Data: miniBlockList})

	assert.Nil(t, merr)

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	arg := createMockArgMiniblockResolver()
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = &storageStubs.StorerStub{
		GetCalled: func(key []byte) (i []byte, e error) {
			body := block.MiniBlock{}
			buff, _ := goodMarshalizer.Marshal(&body)
			return buff, nil
		},
	}
	arg.Marshaller = goodMarshalizer
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
			return expectedErr
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashArrayType, requestedBuff),
		fromConnectedPeerId,
		&p2pmocks.MessengerStub{},
	)

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
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

	store := &storageStubs.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		wasResolved = true
		mb, _ := marshalizer.Marshal(&block.MiniBlock{})
		return mb, nil
	}

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
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
		&p2pmocks.MessengerStub{},
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestMiniblockResolver_ProcessReceivedMessageMarshalFails(t *testing.T) {
	t.Parallel()

	mbHash := []byte("aaa")
	marshalizer := &mock.MarshalizerMock{}
	miniBlockList := make([][]byte, 0)
	miniBlockList = append(miniBlockList, mbHash)
	requestedBuff, _ := marshalizer.Marshal(&batch.Batch{Data: miniBlockList})

	wasResolved := false

	cache := testscommon.NewCacherStub()
	cache.PeekCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &storageStubs.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		wasResolved = true
		mb, _ := marshalizer.Marshal(&block.MiniBlock{})
		return mb, nil
	}

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
			assert.Fail(t, "should have not been called")
			return nil
		},
	}
	arg.MiniBlockPool = cache
	arg.MiniBlockStorage = store
	arg.Marshaller = &mock.MarshalizerStub{
		UnmarshalCalled: marshalizer.Unmarshal,
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedErr
		},
	}
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	err := mbRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.HashType, requestedBuff),
		fromConnectedPeerId,
		&p2pmocks.MessengerStub{},
	)

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, wasResolved)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
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

	store := &storageStubs.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		return nil, errors.New("key not found")
	}

	arg := createMockArgMiniblockResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, source p2p.MessageHandler) error {
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
		&p2pmocks.MessengerStub{},
	)

	assert.False(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestMiniblockResolver_Close(t *testing.T) {
	t.Parallel()

	arg := createMockArgMiniblockResolver()
	mbRes, _ := resolvers.NewMiniblockResolver(arg)

	assert.Nil(t, mbRes.Close())
}
