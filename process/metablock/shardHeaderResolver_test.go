package metablock_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/metablock"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

func createRequestMsg(dataType process.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&process.RequestData{Type: dataType, Value: val})
	return &mock.P2PMessageMock{DataField: buff}
}

//------- NewShardHeaderResolver

func TestNewShardHeaderResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, err := metablock.NewShardHeaderResolver(
		nil,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilResolverSender, err)
	assert.Nil(t, shardHdrRes)
}

func TestNewShardHeaderResolver_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, err := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, shardHdrRes)
}

func TestNewShardHeaderResolver_NilHeadersStorageShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, err := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilHeadersStorage, err)
	assert.Nil(t, shardHdrRes)
}

func TestNewShardHeaderResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, err := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, shardHdrRes)
}

func TestNewShardHeaderResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	shardHdrRes, err := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.NotNil(t, shardHdrRes)
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestShardHeaderResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(process.NonceType, nil))
	assert.Equal(t, process.ErrNilValue, err)
}

func TestShardHeaderResolver_ProcessReceivedMessageRequestUnknownTypeShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(254, make([]byte, 0)))
	assert.Equal(t, process.ErrResolveTypeUnknown, err)

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
	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		headers,
		&mock.StorerStub{},
		marshalizer,
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
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
	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				return nil
			},
		},
		headers,
		&mock.StorerStub{},
		marshalizerStub,
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Equal(t, errExpected, err)
}

func TestShardHeaderResolver_ProcessReceivedMessageRequestRetFromStorageShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}
	wasGotFromStorage := false
	wasSent := false
	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal(key, requestedData) {
			wasGotFromStorage = true
			return make([]byte, 0), nil
		}

		return nil, nil
	}
	marshalizer := &mock.MarshalizerMock{}
	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		headers,
		store,
		marshalizer,
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Nil(t, err)
	assert.True(t, wasGotFromStorage)
	assert.True(t, wasSent)
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
	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		headers,
		store,
		marshalizer,
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(process.HashType, requestedData))
	assert.Equal(t, errExpected, err)
}

func TestShardHeaderResolver_ProcessReceivedMessageRequestNonceTypeShouldErr(t *testing.T) {
	t.Parallel()

	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	err := shardHdrRes.ProcessReceivedMessage(createRequestMsg(process.NonceType, []byte("aaa")))
	assert.Equal(t, process.ErrResolveTypeUnknown, err)
}

//------- Requests

func TestShardHeaderResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	buffRequested := []byte("aaaa")
	wasRequested := false
	shardHdrRes, _ := metablock.NewShardHeaderResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *process.RequestData) error {
				if bytes.Equal(rd.Value, buffRequested) {
					wasRequested = true
				}

				return nil
			},
		},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, shardHdrRes.RequestDataFromHash(buffRequested))
	assert.True(t, wasRequested)
}
