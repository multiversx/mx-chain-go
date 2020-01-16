package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

//------- NewHeaderResolver

func TestNewHeaderResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		nil,
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersPoolShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilHeadersDataPool, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersStorageShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilHeadersStorage, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilHeadersNoncesStorageShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		nil,
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilHeadersNoncesStorage, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		nil,
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilNonceConverterShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nil,
		createMockP2PAntifloodHandler(),
	)

	assert.Equal(t, dataRetriever.ErrNilUint64ByteSliceConverter, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_NilAntifloodHandlerShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	assert.Nil(t, hdrRes)
}

func TestNewHeaderResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdrRes, err := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	assert.NotNil(t, hdrRes)
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestHeaderResolver_ProcessReceivedAntifloodErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		&mock.P2PAntifloodHandlerStub{
			CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer p2p.PeerID) error {
				return expectedErr
			},
		},
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, nil), fromConnectedPeerId)
	assert.Equal(t, expectedErr, err)
}

func TestHeaderResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, nil), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
}

func TestHeaderResolver_ProcessReceivedMessageRequestUnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(254, make([]byte, 0)), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrResolveTypeUnknown, err)

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

	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		headers,
		&mock.StorerStub{},
		&mock.StorerStub{},
		marshalizer,
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
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

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				return nil
			},
		},
		headers,
		&mock.StorerStub{},
		&mock.StorerStub{},
		marshalizerStub,
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Equal(t, errExpected, err)
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
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			wasGotFromStorage = true
			return make([]byte, 0), nil
		}

		return nil, errors.New("should have not reach this point")
	}

	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		headers,
		store,
		&mock.StorerStub{},
		marshalizer,
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData), fromConnectedPeerId)
	assert.Nil(t, err)
	assert.True(t, wasGotFromStorage)
	assert.True(t, wasSent)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeInvalidSliceShouldErr(t *testing.T) {
	t.Parallel()

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, errors.New("key not found")
			},
		},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("aaa")), fromConnectedPeerId)
	assert.Equal(t, dataRetriever.ErrInvalidNonceByteSlice, err)
}

func TestHeaderResolver_ProcessReceivedMessageRequestNonceTypeNotFoundInHdrNoncePoolAndStorageShouldRetNilAndNotSend(t *testing.T) {
	t.Parallel()

	requestedNonce := uint64(67)
	nonceConverter := mock.NewNonceHashConverterMock()

	expectedErr := errors.New("err")
	wasSent := false

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
			TargetShardIDCalled: func() uint32 {
				return 1
			},
		},
		&mock.HeadersCacherStub{
			GetHeaderByNonceAndShardIdCalled: func(hdrNonce uint64, shardId uint32) (handlers []data.HeaderHandler, i [][]byte, e error) {
				return nil, nil, expectedErr
			},
		},
		&mock.StorerStub{},
		&mock.StorerStub{
			GetCalled: func(key []byte) (i []byte, e error) {
				return nil, errors.New("key not found")
			},
		},
		&mock.MarshalizerMock{},
		nonceConverter,
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, nonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)
	assert.Equal(t, expectedErr, err)
	assert.False(t, wasSent)
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

	nonceConverter := mock.NewNonceHashConverterMock()
	marshalizer := &mock.MarshalizerMock{}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
			TargetShardIDCalled: func() uint32 {
				return targetShardId
			},
		},
		headers,
		&mock.StorerStub{},
		&mock.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, errors.New("key not found")
			},
		},
		marshalizer,
		nonceConverter,
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, nonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
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

	nonceConverter := mock.NewNonceHashConverterMock()
	marshalizer := &mock.MarshalizerMock{}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, hash) {
			wasResolved = true
			return make([]byte, 0), nil
		}

		return nil, errors.New("should have not reach this point")
	}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSend = true
				return nil
			},
			TargetShardIDCalled: func() uint32 {
				return targetShardId
			},
		},
		headers,
		store,
		&mock.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, errors.New("key not found")
			},
		},
		marshalizer,
		nonceConverter,
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, nonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
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

	nonceConverter := mock.NewNonceHashConverterMock()
	marshalizer := &mock.MarshalizerMock{}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, []byte("aaaa")) {
			return nil, errExpected
		}

		return nil, errors.New("should have not reach this point")
	}

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				return nil
			},
			TargetShardIDCalled: func() uint32 {
				return targetShardId
			},
		},
		headers,
		store,
		&mock.StorerStub{
			GetCalled: func(key []byte) ([]byte, error) {
				return nil, errors.New("key not found")
			},
		},
		marshalizer,
		nonceConverter,
		createMockP2PAntifloodHandler(),
	)

	err := hdrRes.ProcessReceivedMessage(
		createRequestMsg(dataRetriever.NonceType, nonceConverter.ToByteSlice(requestedNonce)),
		fromConnectedPeerId,
	)

	assert.Equal(t, errExpected, err)
}

//------- Requests

func TestHeaderResolver_RequestDataFromNonceShouldWork(t *testing.T) {
	t.Parallel()

	nonceRequested := uint64(67)
	wasRequested := false

	nonceConverter := mock.NewNonceHashConverterMock()

	buffToExpect := nonceConverter.ToByteSlice(nonceRequested)

	hdrRes, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
				if bytes.Equal(rd.Value, buffToExpect) {
					wasRequested = true
				}
				return nil
			},
		},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
		createMockP2PAntifloodHandler(),
	)

	assert.Nil(t, hdrRes.RequestDataFromNonce(nonceRequested))
	assert.True(t, wasRequested)
}

func TestHeaderResolverBase_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	buffRequested := []byte("aaaa")
	wasRequested := false
	nonceConverter := mock.NewNonceHashConverterMock()
	hdrResBase, _ := resolvers.NewHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
				if bytes.Equal(rd.Value, buffRequested) {
					wasRequested = true
				}

				return nil
			},
		},
		&mock.HeadersCacherStub{},
		&mock.StorerStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nonceConverter,
		createMockP2PAntifloodHandler(),
	)

	assert.Nil(t, hdrResBase.RequestDataFromHash(buffRequested))
	assert.True(t, wasRequested)
}
