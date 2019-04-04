package resolvers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/block/resolvers"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewHeaderResolverBase

func TestNewHeaderResolverBase_NilSenderResolverShouldErr(t *testing.T) {
	t.Parallel()

	hdrResBase, err := resolvers.NewHeaderResolverBase(
		nil,
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilResolverSender, err)
	assert.Nil(t, hdrResBase)
}

func TestNewHeaderResolverBase_NilHeadersShouldErr(t *testing.T) {
	t.Parallel()

	hdrResBase, err := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilHeadersDataPool, err)
	assert.Nil(t, hdrResBase)
}

func TestNewHeaderResolverBase_NilHeadersStorageShouldErr(t *testing.T) {
	t.Parallel()

	hdrResBase, err := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilHeadersStorage, err)
	assert.Nil(t, hdrResBase)
}

func TestNewHeaderResolverBase_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hdrResBase, err := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, hdrResBase)
}

func TestNewHeaderResolverBase_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdrResBase, err := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.NotNil(t, hdrResBase)
	assert.Nil(t, err)
}

//------- ParseReceivedMessage

func TestHeaderResolverBase_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	hdrResBase, _ := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)
	rd, err := hdrResBase.ParseReceivedMessage(nil)

	assert.Nil(t, rd)
	assert.Equal(t, process.ErrNilMessage, err)
}

func TestHeaderResolverBase_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	hdrResBase, _ := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)
	rd, err := hdrResBase.ParseReceivedMessage(createRequestMsg(process.NonceType, nil))

	assert.Nil(t, rd)
	assert.Equal(t, process.ErrNilValue, err)
}

func TestHeaderResolverBase_ProcessReceivedMessageOkValsShouldWork(t *testing.T) {
	t.Parallel()

	hdrResBase, _ := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		&mock.CacherStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)
	expectedValue := []byte("expected value")
	rd, err := hdrResBase.ParseReceivedMessage(createRequestMsg(process.HashType, expectedValue))
	expectedRequestData := &process.RequestData{
		Type:  process.HashType,
		Value: expectedValue,
	}

	assert.Equal(t, expectedRequestData, rd)
	assert.Nil(t, err)
}

func TestHeaderResolverBase_ValidateRequestHashTypeFoundInHdrPoolShouldSearchAndSend(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	resolvedData := []byte("resolved data")
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				return resolvedData, true
			}
			return nil, false
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	hdrResBase, _ := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		headers,
		&mock.StorerStub{},
		marshalizer,
	)

	buffExpected, _ := marshalizer.Marshal(resolvedData)
	buff, err := hdrResBase.ResolveHeaderFromHash(requestedData)
	assert.Nil(t, err)
	assert.Equal(t, buffExpected, buff)
}

func TestHeaderResolverBase_ProcessReceivedMessageRequestHashTypeFoundInHdrPoolMarshalizerFailsShouldErr(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	resolvedData := []byte("resolved data")
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			if bytes.Equal(requestedData, key) {
				return resolvedData, true
			}
			return nil, false
		},
	}
	errExpected := errors.New("MarshalizerMock generic error")
	marshalizerStub := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			return nil, errExpected
		},
	}
	hdrResBase, _ := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		headers,
		&mock.StorerStub{},
		marshalizerStub,
	)

	buff, err := hdrResBase.ResolveHeaderFromHash(requestedData)
	assert.Equal(t, errExpected, err)
	assert.Nil(t, buff)
}

func TestHeaderResolverBase_ValidateRequestHashTypeNotFoundInHdrPoolShouldReturnFromStorage(t *testing.T) {
	t.Parallel()

	requestedData := []byte("aaaa")
	resolvedData := []byte("resolved data")
	headers := &mock.CacherStub{
		PeekCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
	}
	marshalizer := &mock.MarshalizerMock{}
	hdrResBase, _ := resolvers.NewHeaderResolverBase(
		&mock.TopicResolverSenderStub{},
		headers,
		&mock.StorerStub{
			GetCalled: func(key []byte) (i []byte, e error) {
				if bytes.Equal(requestedData, key) {
					return resolvedData, nil
				}
				return nil, nil
			},
		},
		marshalizer,
	)

	buff, err := hdrResBase.ResolveHeaderFromHash(requestedData)
	assert.Nil(t, err)
	assert.Equal(t, resolvedData, buff)
}

//------- Requests

func TestHeaderResolverBase_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	buffRequested := []byte("aaaa")
	wasRequested := false
	hdrResBase, _ := resolvers.NewHeaderResolverBase(
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

	assert.Nil(t, hdrResBase.RequestDataFromHash(buffRequested))
	assert.True(t, wasRequested)
}
