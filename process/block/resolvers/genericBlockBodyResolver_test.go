package resolvers_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

//------- NewBlockBodyResolver

func TestNewGenericBlockBodyResolver_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		nil,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilResolverSender, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockBodyPoolShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilBlockBodyPool, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockBodyStorageShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilBlockBodyStorage, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_NilBlockMArshalizerShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, gbbRes)
}

func TestNewGenericBlockBodyResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	gbbRes, err := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, gbbRes)
}

//------- ProcessReceivedMessage

func TestNewGenericBlockBodyResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(process.HashType, nil))
	assert.Equal(t, process.ErrNilValue, err)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(process.NonceType, make([]byte, 0)))
	assert.Equal(t, process.ErrResolveNotHashType, err)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageFoundInPoolShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	requestedBuff := []byte("aaa")

	wasResolved := false
	wasSent := false

	cache := &mock.CacherStub{}
	cache.GetCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, requestedBuff) {
			wasResolved = true
			return make([]byte, 0), true
		}

		return nil, false
	}

	marshalizer := &mock.MarshalizerMock{}

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				wasSent = true
				return nil
			},
		},
		cache,
		&mock.StorerStub{},
		marshalizer,
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(
		process.HashType,
		requestedBuff))

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSent)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageFoundInPoolMarshalizerFailShouldErr(t *testing.T) {
	t.Parallel()

	requestedBuff := []byte("aaa")

	errExpected := errors.New("expected error")

	cache := &mock.CacherStub{}
	cache.GetCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(key, requestedBuff) {
			return make([]byte, 0), true
		}

		return nil, false
	}

	marshalizer := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			return nil, errExpected
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			m := &mock.MarshalizerMock{}

			return m.Unmarshal(obj, buff)
		},
	}

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{},
		cache,
		&mock.StorerStub{},
		marshalizer,
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(
		process.HashType,
		requestedBuff))

	assert.Equal(t, errExpected, err)

}

func TestGenericBlockBodyResolver_ProcessReceivedMessageNotFoundInPoolShouldRetFromStorageAndSend(t *testing.T) {
	t.Parallel()

	requestedBuff := []byte("aaa")

	wasResolved := false
	wasSend := false

	cache := &mock.CacherStub{}
	cache.GetCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		wasResolved = true
		return make([]byte, 0), nil
	}

	marshalizer := &mock.MarshalizerMock{}

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
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(
		process.HashType,
		requestedBuff))

	assert.Nil(t, err)
	assert.True(t, wasResolved)
	assert.True(t, wasSend)
}

func TestGenericBlockBodyResolver_ProcessReceivedMessageMissingDataShouldNotSend(t *testing.T) {
	t.Parallel()

	requestedBuff := []byte("aaa")

	wasSend := false

	cache := &mock.CacherStub{}
	cache.GetCalled = func(key []byte) (value interface{}, ok bool) {
		return nil, false
	}

	store := &mock.StorerStub{}
	store.GetCalled = func(key []byte) (i []byte, e error) {
		return nil, nil
	}

	marshalizer := &mock.MarshalizerMock{}

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
	)

	err := gbbRes.ProcessReceivedMessage(createRequestMsg(
		process.HashType,
		requestedBuff))

	assert.Nil(t, err)
	assert.False(t, wasSend)
}

//------- Requests

func TestBlockBodyResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	wasCalled := false

	buffRequested := []byte("aaaa")

	gbbRes, _ := resolvers.NewGenericBlockBodyResolver(
		&mock.TopicResolverSenderStub{
			SendOnRequestTopicCalled: func(rd *process.RequestData) error {
				wasCalled = true
				return nil
			},
		},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, gbbRes.RequestDataFromHash(buffRequested))
	assert.True(t, wasCalled)
}
