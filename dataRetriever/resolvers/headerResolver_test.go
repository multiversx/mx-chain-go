package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func createMockArgHeaderResolver() resolvers.ArgHeaderResolver {
	return resolvers.ArgHeaderResolver{
		SenderResolver:       &mock.TopicResolverSenderStub{},
		Headers:              &mock.HeadersCacherStub{},
		HdrStorage:           &mock.StorerStub{},
		HeadersNoncesStorage: &mock.StorerStub{},
		Marshalizer:          &mock.MarshalizerMock{},
		NonceConverter:       mock.NewNonceHashConverterMock(),
		ShardCoordinator:     mock.NewOneShardCoordinatorMock(),
		AntifloodHandler:     &mock.P2PAntifloodHandlerStub{},
		Throttler:            &mock.ThrottlerStub{},
	}
}

var errKeyNotFound = errors.New("key not found")

//------- NewHeaderResolver

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
	arg.Marshalizer = nil
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

//------- ProcessReceivedMessage

func TestHeaderResolver_ProcessReceivedCanProcessMessageErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
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
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestHeaderResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, nil), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestHeaderResolver_ProcessReceivedMessage_WrongIdentifierStartBlock(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	requestedData := []byte("request")
	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.EpochType, requestedData), "")
	assert.Equal(t, core.ErrInvalidIdentifierForEpochStartBlockRequest, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestHeaderResolver_ProcessReceivedMessage_Ok(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.HdrStorage = &mock.StorerStub{
		SearchFirstCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	requestedData := []byte("request_1")
	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.EpochType, requestedData), "")
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestHeaderResolver_RequestDataFromEpoch(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, hashes [][]byte) error {
			called = true
			return nil
		},
	}
	arg.HdrStorage = &mock.StorerStub{
		GetCalled: func(key []byte) (i []byte, err error) {
			return nil, nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	requestedData := []byte("request_1")
	err := hdrRes.RequestDataFromEpoch(requestedData)
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestHeaderResolver_ProcessReceivedMessageRequestUnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(254, make([]byte, 0)), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrResolveTypeUnknown, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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
		SendCalled: func(buff []byte, peer core.PeerID) error {
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
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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
		SendCalled: func(buff []byte, peer core.PeerID) error {
			return nil
		},
	}
	arg.Marshalizer = marshalizerStub
	arg.Headers = headers
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Equal(t, errExpected, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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

	store := &mock.StorerStub{}
	store.SearchFirstCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			wasGotFromStorage = true
			return make([]byte, 0), nil
		}

		return nil, errors.New("should have not reach this point")
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
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
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeInvalidSliceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	arg.HdrStorage = &mock.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			return nil, errKeyNotFound
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("aaa")), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrInvalidNonceByteSlice, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceShouldCallWithTheCorrectEpoch(t *testing.T) {
	t.Parallel()

	expectedEpoch := uint32(7)
	arg := createMockArgHeaderResolver()
	arg.HeadersNoncesStorage = &mock.StorerStub{
		GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
			assert.Equal(t, expectedEpoch, epoch)
			return nil, nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	buff, _ := arg.Marshalizer.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.NonceType,
			Value: []byte("aaa"),
			Epoch: expectedEpoch,
		},
	)
	msg := &mock.P2PMessageMock{DataField: buff}
	_ = hdrRes.ProcessReceivedMessage(msg, "")
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeNotFoundInHdrNoncePoolAndStorageShouldRetNilAndNotSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)

	expectedErr := errors.New("err")
	wasSent := false

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
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
	arg.HdrStorage = &mock.StorerStub{
		SearchFirstCalled: func(key []byte) (i []byte, e error) {
			return nil, errKeyNotFound
		},
	}
	arg.HeadersNoncesStorage = &mock.StorerStub{
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
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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
		SendCalled: func(buff []byte, peer core.PeerID) error {
			wasSent = true
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return targetShardId
		},
	}
	arg.Headers = headers
	arg.HeadersNoncesStorage = &mock.StorerStub{
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
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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

	store := &mock.StorerStub{}
	store.GetFromEpochCalled = func(key []byte, epoch uint32) (i []byte, e error) {
		if bytes.Equal(key, hash) {
			wasResolved = true
			return make([]byte, 0), nil
		}

		return nil, errors.New("should have not reach this point")
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			wasSend = true
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return targetShardId
		},
	}
	arg.Headers = headers
	arg.HdrStorage = store
	arg.HeadersNoncesStorage = &mock.StorerStub{
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
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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

	store := &mock.StorerStub{}
	store.GetFromEpochCalled = func(key []byte, epoch uint32) (i []byte, e error) {
		if bytes.Equal(key, []byte("aaaa")) {
			return nil, errExpected
		}

		return nil, errors.New("should have not reach this point")
	}

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			return nil
		},
		TargetShardIDCalled: func() uint32 {
			return targetShardId
		},
	}
	arg.Headers = headers
	arg.HdrStorage = store
	arg.HeadersNoncesStorage = &mock.StorerStub{
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
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

//------- Requests

func TestHeaderResolver_RequestDataFromNonceShouldWork(t *testing.T) {
	t.Parallel()

	nonceRequested := uint64(67)
	wasRequested := false

	nonceConverter := mock.NewNonceHashConverterMock()
	buffToExpect := nonceConverter.ToByteSlice(nonceRequested)

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, hashes [][]byte) error {
			if bytes.Equal(rd.Value, buffToExpect) {
				wasRequested = true
			}
			return nil
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	assert.Nil(t, hdrRes.RequestDataFromNonce(nonceRequested, 0))
	assert.True(t, wasRequested)
}

func TestHeaderResolverBase_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	buffRequested := []byte("aaaa")
	wasRequested := false
	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData, hashes [][]byte) error {
			if bytes.Equal(rd.Value, buffRequested) {
				wasRequested = true
			}

			return nil
		},
	}
	hdrResBase, _ := resolvers.NewHeaderResolver(arg)

	assert.Nil(t, hdrResBase.RequestDataFromHash(buffRequested, 0))
	assert.True(t, wasRequested)
}

//------- SetEpochHandler

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

//------ NumPeersToQuery setter and getter

func TestHeaderResolver_SetAndGetNumPeersToQuery(t *testing.T) {
	t.Parallel()

	expectedIntra := 5
	expectedCross := 7

	arg := createMockArgHeaderResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		GetNumPeersToQueryCalled: func() (int, int) {
			return expectedIntra, expectedCross
		},
	}
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	hdrRes.SetNumPeersToQuery(expectedIntra, expectedCross)
	actualIntra, actualCross := hdrRes.NumPeersToQuery()
	assert.Equal(t, expectedIntra, actualIntra)
	assert.Equal(t, expectedCross, actualCross)
}

func TestHeaderResolver_Close(t *testing.T) {
	t.Parallel()

	arg := createMockArgHeaderResolver()
	hdrRes, _ := resolvers.NewHeaderResolver(arg)

	assert.Nil(t, hdrRes.Close())
}
