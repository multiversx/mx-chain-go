package transaction

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/mock"
	"github.com/stretchr/testify/assert"
)

//------- NewTxResolver

func TestNewTxResolver_NilResolverShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		nil,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilResolverSender, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilTxDataPool, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilTxStorageShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, process.ErrNilTxStorage, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		nil,
	)

	assert.Equal(t, process.ErrNilMarshalizer, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	res := &mock.TopicResolverSenderStub{}

	txRes, err := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, err)
	assert.NotNil(t, txRes)
}

//------- ProcessReceivedMessage

func TestTxResolver_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	err := txRes.ProcessReceivedMessage(nil)

	assert.Equal(t, process.ErrNilMessage, err)
}

func TestTxResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&process.RequestData{Type: process.NonceType, Value: []byte("aaa")})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrResolveNotHashType, err)
}

func TestTxResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&process.RequestData{Type: process.HashType, Value: nil})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Equal(t, process.ErrNilValue, err)
}

func TestTxResolver_ProcessReceivedMessageFoundInTxPoolShouldSearchAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	searchWasCalled := false
	sendWasCalled := false

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal([]byte("aaa"), key) {
			searchWasCalled = true
			return make([]byte, 0), true
		}

		return nil, false
	}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		txPool,
		&mock.StorerStub{},
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&process.RequestData{Type: process.HashType, Value: []byte("aaa")})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
}

func TestTxResolver_ProcessReceivedMessageFoundInTxPoolMarshalizerFailShouldRetNilAndErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("MarshalizerMock generic error")

	marshalizerMock := &mock.MarshalizerMock{}
	marshalizerStub := &mock.MarshalizerStub{
		MarshalCalled: func(obj interface{}) (i []byte, e error) {
			return nil, errExpected
		},
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return marshalizerMock.Unmarshal(obj, buff)
		},
	}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal([]byte("aaa"), key) {
			return "value", true
		}

		return nil, false
	}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		txPool,
		&mock.StorerStub{},
		marshalizerStub,
	)

	data, _ := marshalizerMock.Marshal(&process.RequestData{Type: process.HashType, Value: []byte("aaa")})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Equal(t, errExpected, err)
}

func TestTxResolver_ProcessReceivedMessageFoundInTxStorageShouldRetValAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		//not found in txPool
		return nil, false
	}

	searchWasCalled := false
	sendWasCalled := false

	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal([]byte("aaa"), key) {
			searchWasCalled = true
			return make([]byte, 0), nil
		}

		return nil, nil
	}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		txPool,
		txStorage,
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&process.RequestData{Type: process.HashType, Value: []byte("aaa")})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	assert.True(t, searchWasCalled)
	assert.True(t, sendWasCalled)
}

func TestTxResolver_ProcessReceivedMessageFoundInTxStorageCheckRetError(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		//not found in txPool
		return nil, false
	}

	errExpected := errors.New("expected error")

	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal([]byte("aaa"), key) {
			return nil, errExpected
		}

		return nil, nil
	}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		txPool,
		txStorage,
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&process.RequestData{Type: process.HashType, Value: []byte("aaa")})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Equal(t, errExpected, err)

}

//------- RequestTransactionFromHash

func TestTxResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	requested := &process.RequestData{}

	res := &mock.TopicResolverSenderStub{}
	res.SendOnRequestTopicCalled = func(rd *process.RequestData) error {
		requested = rd
		return nil
	}

	buffRequested := []byte("aaaa")

	txRes, _ := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, txRes.RequestDataFromHash(buffRequested))
	assert.Equal(t, &process.RequestData{
		Type:  process.HashType,
		Value: buffRequested,
	}, requested)

}
