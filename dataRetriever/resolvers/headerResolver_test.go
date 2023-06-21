package resolvers_test

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func createMockArgBaseResolver() resolvers.ArgBaseResolver {
	return resolvers.ArgBaseResolver{
		SenderResolver:   &mock.TopicResolverSenderStub{},
		Marshaller:       &mock.MarshalizerMock{},
		AntifloodHandler: &mock.P2PAntifloodHandlerStub{},
		Throttler:        &mock.ThrottlerStub{},
	}
}

func createMockArgHeaderResolver() resolvers.ArgHeaderResolver {
	return resolvers.ArgHeaderResolver{
		ArgBaseResolver:      createMockArgBaseResolver(),
		Headers:              &mock.HeadersCacherStub{},
		HdrStorage:           &storageStubs.StorerStub{},
		HeadersNoncesStorage: &storageStubs.StorerStub{},
		NonceConverter:       mock.NewNonceHashConverterMock(),
		ShardCoordinator:     mock.NewOneShardCoordinatorMock(),
	}
}

var errKeyNotFound = errors.New("key not found")

func TestNewHeaderResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.Headers = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilHeadersDataPool, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.HdrStorage = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersNoncesStorageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.HeadersNoncesStorage = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilHeadersNoncesStorage, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.Marshaller = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilNonceConverterShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.NonceConverter = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.ShardCoordinator = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilShardCoordinator, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.AntifloodHandler = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.Throttler = nil
	hdrRes, err := resolvers.NewHeaderResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilThrottler, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, err := resolvers.NewHeaderResolver(arg)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(hdrRes))
}

func TestHeaderResolver_ProcessReceivedCanProcessMessageErrorsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return expectedErr
		},
		CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
			return nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, nil), fromConnectedPeerId)
	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, nil), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessage_WrongIdentifierStartBlock(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	requestedData := []byte("request")
	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.EpochType, requestedData), "")
	assert.Equal(t, core.ErrInvalidIdentifierForEpochStartBlockRequest, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageEpochTypeUnknownEpochShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return []byte("hash"), nil
		},
	}
	wasSent := false
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			wasSent = true
			return nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	requestedData := []byte(fmt.Sprintf("epoch_%d", math.MaxUint32))
	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.EpochType, requestedData), "")
	assert.NoError(t, err)
	assert.True(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessage_Ok(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	requestedData := []byte("request_1")
	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.EpochType, requestedData), "")
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestUnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(254, make([]byte, 0)), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrResolveTypeUnknown, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ValidateRequestHashTypeFoundInHdrPoolShouldSearchAndSend(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")

	searchWasCalled := false
	sendWasCalled := false

	headers := &mock.HeadersCacherStub{}

	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		if bytes.Equal(requestedData, hash) {
			searchWasCalled = true
			return &block.Header{}, nil
		}
		return nil, errors.New("0")
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			sendWasCalled = true
			return nil
		},
	}
	arg.Headers = headers
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ValidateRequestHashTypeFoundInHdrPoolShouldSearchAndSendFullHistory(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")

	searchWasCalled := false
	sendWasCalled := false

	headers := &mock.HeadersCacherStub{}

	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		if bytes.Equal(requestedData, hash) {
			searchWasCalled = true
			return &block.Header{}, nil
		}
		return nil, errors.New("0")
	}

	arg := createMockArgHeaderResolver()
	arg.IsFullHistoryNode = true
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			sendWasCalled = true
			return nil
		},
	}
	arg.Headers = headers
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestHashTypeFoundInHdrPoolMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")

	errExpected := errors.New("MarshalizerMock generic error")

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		if bytes.Equal(requestedData, hash) {
			return &block.Header{}, nil
		}
		return nil, errors.New("err")
	}

	marshalizerMock := &mock.MarshalizerMock{}
	marshalizerStub := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			return nil, errExpected
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return marshalizerMock.Unmarshal(obj, buff)
		},
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			return nil
		},
	}
	arg.Marshaller = marshalizerStub
	arg.Headers = headers
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Equal(t, errExpected, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestRetFromStorageShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		return nil, errors.New("err")
	}

	wasGotFromStorage := false
	wasSent := false

	store := &storageStubs.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			wasGotFromStorage = true
			return make([]byte, 0), nil
		}

		return nil, errors.New("should have not reach this point")
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			wasSent = true
			return nil
		},
	}
	arg.Headers = headers
	arg.HdrStorage = store
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Nil(t, err)
	assert.True(t, wasGotFromStorage)
	assert.True(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeInvalidSliceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.HdrStorage = &storageStubs.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			return nil, errKeyNotFound
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("aaa")), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrInvalidNonceByteSlice, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceShouldCallWithTheCorrectEpoch(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(7)
	arg := createMockArgHeaderResolver()
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			assert.Equal(t, expectedEpoch, epoch)
			return nil, nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	buff, _ := arg.Marshaller.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.NonceType,
			Value: []byte("aaa"),
			Epoch: expectedEpoch,
		},
	)
	msg := &p2pmocks.P2PMessageMock{DataField: buff}
	_ = hdrRes.ProcessReceivedMessage(msg, "")
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeNotFoundInHdrNoncePoolAndStorageShouldRetNilAndNotSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)

	wasSent := false

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			wasSent = true
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return 1
		},
	}
	arg.Headers = &mock.HeadersCacherStub{
		GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
			return nil, nil, expectedErr
		},
	}
	arg.HdrStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return nil, errKeyNotFound
		},
	}
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) (i []byte, e error) {
			return nil, errKeyNotFound
		},
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return nil, errKeyNotFound
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, arg.NonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)
	assert.Equal(t, expectedErr, err)
	assert.False(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolShouldRetFromPoolAndSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	targetShardId := uint32(9)
	wasResolved := false
	wasSent := false

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
		wasResolved = true
		return []data.HeaderHandler{&block.Header{}, &block.Header{}}, [][]byte{[]byte("1"), []byte("2")}, nil
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			wasSent = true
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return targetShardId
		},
	}
	arg.Headers = headers
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			return nil, errKeyNotFound
		},
		SearchFirstCalled: func(key []byte) ([]byte, error) {
			return nil, errKeyNotFound
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, arg.NonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolShouldRetFromStorageAndSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	targetShardId := uint32(9)
	wasResolved := false
	wasSend := false
	hash := []byte("aaaa")

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		return nil, errors.New("err")
	}
	headers.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
		wasResolved = true
		return []data.HeaderHandler{&block.Header{}, &block.Header{}}, [][]byte{[]byte("1"), []byte("2")}, nil
	}

	store := &storageStubs.StorerStub{}
	store.GetFromEpochCalled = func(key []byte, epoch uint32) (i []byte, e error) {
		if bytes.Equal(key, hash) {
			wasResolved = true
			return make([]byte, 0), nil
		}

		return nil, errors.New("should have not reach this point")
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			wasSend = true
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return targetShardId
		},
	}
	arg.Headers = headers
	arg.HdrStorage = store
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			return nil, errKeyNotFound
		},
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return nil, errKeyNotFound
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, arg.NonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolButMarshalFailsShouldError(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	targetShardId := uint32(9)
	wasResolved := false

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		return nil, errors.New("err")
	}
	headers.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
		wasResolved = true
		return []data.HeaderHandler{&block.Header{}, &block.Header{}}, [][]byte{[]byte("1"), []byte("2")}, nil
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			assert.Fail(t, "should not have been called")
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return targetShardId
		},
	}
	arg.Headers = headers
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			return nil, errKeyNotFound
		},
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return nil, errKeyNotFound
		},
	}
	initialMarshaller := arg.Marshaller
	arg.Marshaller = &mock.MarshalizerStub{
		UnmarshalCalled: initialMarshaller.Unmarshal,
		MarshalCalled: func(obj interface{}) ([]byte, error) {
			return nil, expectedErr
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, arg.NonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.True(t, errors.Is(err, expectedErr))
	assert.True(t, wasResolved)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeNotFoundInHdrNoncePoolShouldRetFromPoolAndSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	wasSend := false
	hash := []byte("aaaa")

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		return &block.Header{}, nil
	}
	headers.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
		assert.Fail(t, "should not have been called")
		return nil, nil, nil
	}
	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			wasSend = true
			return nil
		},
	}
	arg.Headers = headers
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return hash, nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, arg.NonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasSend)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeFoundInHdrNoncePoolCheckRetErr(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	targetShardId := uint32(9)
	errExpected := errors.New("expected error")

	headers := &mock.HeadersCacherStub{}
	headers.GetHeaderByHashCalled = func(hash []byte) (handler data.HeaderHandler, e error) {
		return nil, errors.New("err")
	}
	headers.GetHeaderByNonceAndShardIdCalled = func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
		return nil, nil, errExpected
	}

	store := &storageStubs.StorerStub{}
	store.GetFromEpochCalled = func(key []byte, epoch uint32) (i []byte, e error) {
		if bytes.Equal(key, []byte("aaaa")) {
			return nil, errExpected
		}

		return nil, errors.New("should have not reach this point")
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return targetShardId
		},
	}
	arg.Headers = headers
	arg.HdrStorage = store
	arg.HeadersNoncesStorage = &storageStubs.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			return nil, errKeyNotFound
		},
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return nil, errKeyNotFound
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, arg.NonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.Equal(t, errExpected, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestHeaderResolver_SetEpochHandlerNilShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.SetEpochHandler(nil)

	assert.Equal(t, dataRetriever.ErrNilEpochHandler, err)
}

func TestHeaderResolver_SetEpochHandlerShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	eh := &mock.EpochHandlerStub{}
	err := hdrRes.SetEpochHandler(eh)

	assert.Nil(t, err)
	assert.True(t, eh == hdrRes.EpochHandler())
}

func TestHeaderResolver_Close(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	assert.Nil(t, hdrRes.Close())
}

func TestHeaderResolver_SetEpochHandlerConcurrency(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	eh := &mock.EpochHandlerStub{}
	var wg sync.WaitGroup
	numCalls := 1000
	wg.Add(numCalls)
	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			defer wg.Done()

			if idx == 0 {
				err := hdrRes.SetEpochHandler(eh)
				assert.Nil(t, err)
				return
			}
			err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.EpochType, []byte("request_1")), fromConnectedPeerId)
			assert.Nil(t, err)
		}(i)
	}
	wg.Wait()
}
