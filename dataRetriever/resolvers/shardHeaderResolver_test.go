package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/stretchr/testify/assert"
)

//------- NewShardHeaderResolver

func TestNewShardHeaderResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, err := resolvers.NewShardHeaderResolver(
		nil,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.Nil(t, shardHdrRes)
}

func TestNewShardHeaderResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardHdrRes, err := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.NotNil(t, shardHdrRes)
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestShardHeaderResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, nil))
	assert.Equal(t, dataRetriever.ErrNilValue, err)
}

func TestShardHeaderResolver_ProcessReceivedMessageRequestUnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, make([]byte, 0)))
	assert.Equal(t, dataRetriever.ErrResolveTypeUnknown, err)
}

func TestShardHeaderResolver_ValidateRequestHashTypeFoundInHdrPoolShouldSearchAndSend(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	searchWasCalled := false
	sendWasCalled := false
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				searchWasCalled = true
				return make([]byte, 0), true
			}
			return nil, false
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		headers,
		&mock.StorerStub{},
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData))
	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
}

func TestShardHeaderResolver_NotFoundShouldReturnNilAndNotSend(t *testing.T) {
	t.Parallel()

	sendWasCalled := false
	requestedData := []byte("aaaa")
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		headers,
		&mock.StorerStub{
			GetCalled: func(key []byte) (i []byte, e error) {
				return nil, nil
			},
		},
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData))
	assert.Nil(t, err)
	assert.False(t, sendWasCalled)
}

func TestShardHeaderResolver_ProcessReceivedMessageRequestHashTypeFoundInHdrPoolMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	resolvedData := []byte("bbbb")
	errExpected := errors.New("MarshalizerMock generic error")
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				return resolvedData, true
			}
			return nil, false
		},
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
	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				return nil
			},
		},
		headers,
		&mock.StorerStub{},
		marshalizerStub,
		mock.NewNonceHashConverterMock(),
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData))
	assert.Equal(t, errExpected, err)
}

func TestShardHeaderResolver_ProcessReceivedMessageRequestRetFromStorageCheckRetError(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}
	errExpected := errors.New("expected error")
	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			return nil, errExpected
		}

		return nil, nil
	}
	marshalizer := &mock.MarshalizerMock{}
	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		headers,
		store,
		marshalizer,
		mock.NewNonceHashConverterMock(),
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.HashType, requestedData))
	assert.Equal(t, errExpected, err)
}

func TestShardHeaderResolver_ProcessReceivedMessageRequestNonceTypeShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(dataRetriever.NonceType, []byte("aaa")))
	assert.Equal(t, dataRetriever.ErrResolveTypeUnknown, err)
}

//------- Requests

func TestShardHeaderResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	buffRequested := []byte("aaaa")
	wasRequested := false
	shardHdrRes, _ := resolvers.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *dataRetriever.RequestData) error {
				if bytes.Equal(rd.Value, buffRequested) {
					wasRequested = true
				}

				return nil
			},
		},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		mock.NewNonceHashConverterMock(),
	)

	assert.Nil(t, shardHdrRes.RequestDataFromHash(buffRequested))
	assert.True(t, wasRequested)
}
