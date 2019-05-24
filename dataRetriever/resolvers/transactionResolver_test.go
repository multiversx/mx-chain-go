package resolvers

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
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
		&mock.SliceSplitterStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilTxPoolShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		&mock.SliceSplitterStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTxDataPool, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilTxStorageShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		nil,
		&mock.MarshalizerMock{},
		&mock.SliceSplitterStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilTxStorage, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	txRes, err := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		nil,
		&mock.SliceSplitterStub{},
	)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.Nil(t, txRes)
}

func TestNewTxResolver_NilSliceSplitterShouldErr(t *testing.T) {
	t.Parallel()

	res := &mock.TopicResolverSenderStub{}

	txRes, err := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilSliceSplitter, err)
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
		&mock.SliceSplitterStub{},
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
		&mock.SliceSplitterStub{},
	)

	err := txRes.ProcessReceivedMessage(nil)

	assert.Equal(t, dataRetriever.ErrNilMessage, err)
}

func TestTxResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		marshalizer,
		&mock.SliceSplitterStub{},
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.NonceType, Value: []byte("aaa")})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Equal(t, dataRetriever.ErrRequestTypeNotImplemented, err)
}

func TestTxResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		marshalizer,
		&mock.SliceSplitterStub{},
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: nil})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Equal(t, dataRetriever.ErrNilValue, err)
}

func TestTxResolver_ProcessReceivedMessageFoundInTxPoolShouldSearchAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	searchWasCalled := false
	sendWasCalled := false
	txReturned := &transaction.Transaction{
		Nonce: 10,
	}
	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal([]byte("aaa"), key) {
			searchWasCalled = true
			return txReturned, true
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
		&mock.SliceSplitterStub{},
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("aaa")})

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
	txReturned := &transaction.Transaction{
		Nonce: 10,
	}
	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal([]byte("aaa"), key) {
			return txReturned, true
		}

		return nil, false
	}

	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{},
		txPool,
		&mock.StorerStub{},
		marshalizerStub,
		&mock.SliceSplitterStub{},
	)

	data, _ := marshalizerMock.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("aaa")})

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
	txReturned := &transaction.Transaction{
		Nonce: 10,
	}
	txReturnedAsBuffer, _ := marshalizer.Marshal(txReturned)
	txStorage := &mock.StorerStub{}
	txStorage.GetCalled = func(key []byte) (i []byte, e error) {
		if bytes.Equal([]byte("aaa"), key) {
			searchWasCalled = true
			return txReturnedAsBuffer, nil
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
		&mock.SliceSplitterStub{},
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("aaa")})

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
		&mock.SliceSplitterStub{},
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("aaa")})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Equal(t, errExpected, err)

}

func TestTxResolver_ProcessReceivedMessageRequestedTwoSmallTransactionsShouldCallSliceSplitter(t *testing.T) {
	t.Parallel()

	txHash1 := []byte("txHash1")
	txHash2 := []byte("txHash2")

	tx1 := &transaction.Transaction{
		Nonce: 10,
	}
	tx2 := &transaction.Transaction{
		Nonce: 20,
	}

	marshalizer := &mock.MarshalizerMock{}
	txPool := &mock.ShardedDataStub{}
	txPool.SearchFirstDataCalled = func(key []byte) (value interface{}, ok bool) {
		if bytes.Equal(txHash1, key) {
			return tx1, true
		}
		if bytes.Equal(txHash2, key) {
			return tx2, true
		}

		return nil, false
	}

	sendSliceWasCalled := false
	txRes, _ := NewTxResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				return nil
			},
		},
		txPool,
		&mock.StorerStub{},
		marshalizer,
		&mock.SliceSplitterStub{
			SendDataInChunksCalled: func(data [][]byte, sendHandler func(buff []byte) error, maxPacketSize int) error {
				if len(data) != 2 {
					return errors.New("should have been 2 data pieces")
				}

				sendSliceWasCalled = true
				return nil
			},
		},
	)

	buff, _ := marshalizer.Marshal([][]byte{txHash1, txHash2})
	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashArrayType, Value: buff})

	msg := &mock.P2PMessageMock{DataField: data}

	err := txRes.ProcessReceivedMessage(msg)

	assert.Nil(t, err)
	assert.True(t, sendSliceWasCalled)
}

//------- RequestTransactionFromHash

func TestTxResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	requested := &dataRetriever.RequestData{}

	res := &mock.TopicResolverSenderStub{}
	res.SendOnRequestTopicCalled = func(rd *dataRetriever.RequestData) error {
		requested = rd
		return nil
	}

	buffRequested := []byte("aaaa")

	txRes, _ := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
		&mock.SliceSplitterStub{},
	)

	assert.Nil(t, txRes.RequestDataFromHash(buffRequested))
	assert.Equal(t, &dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: buffRequested,
	}, requested)

}

//------- RequestDataFromHashArray

func TestTxResolver_RequestDataFromHashArrayShouldWork(t *testing.T) {
	t.Parallel()

	requested := &dataRetriever.RequestData{}

	res := &mock.TopicResolverSenderStub{}
	res.SendOnRequestTopicCalled = func(rd *dataRetriever.RequestData) error {
		requested = rd
		return nil
	}

	buffRequested := [][]byte{[]byte("aaaa"), []byte("bbbb")}

	marshalizer := &mock.MarshalizerMock{}
	txRes, _ := NewTxResolver(
		res,
		&mock.ShardedDataStub{},
		&mock.StorerStub{},
		marshalizer,
		&mock.SliceSplitterStub{},
	)

	buff, _ := marshalizer.Marshal(buffRequested)

	assert.Nil(t, txRes.RequestDataFromHashArray(buffRequested))
	assert.Equal(t, &dataRetriever.RequestData{
		Type:  dataRetriever.HashArrayType,
		Value: buff,
	}, requested)

}
