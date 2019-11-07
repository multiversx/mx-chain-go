package resolvers

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/p2p"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewTrieNodeResolver_NilResolverShouldErr(t *testing.T) {
	t.Parallel()

	tnRes, err := NewTrieNodeResolver(
		nil,
		&mock.TrieStub{},
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_NilTrieShouldErr(t *testing.T) {
	t.Parallel()

	tnRes, err := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{},
		nil,
		&mock.MarshalizerMock{},
	)

	assert.Equal(t, dataRetriever.ErrNilTrie, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	tnRes, err := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{},
		&mock.TrieStub{},
		nil,
	)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	tnRes, err := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{},
		&mock.TrieStub{},
		&mock.MarshalizerMock{},
	)

	assert.NotNil(t, tnRes)
	assert.Nil(t, err)
}

//------- ProcessReceivedMessage

func TestTrieNodeResolver_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	tnRes, _ := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{},
		&mock.TrieStub{},
		&mock.MarshalizerMock{},
	)

	err := tnRes.ProcessReceivedMessage(nil, nil)
	assert.Equal(t, dataRetriever.ErrNilMessage, err)
}

func TestTrieNodeResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	tnRes, _ := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{},
		&mock.TrieStub{},
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.NonceType, Value: []byte("aaa")})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, nil)
	assert.Equal(t, dataRetriever.ErrRequestTypeNotImplemented, err)
}

func TestTrieNodeResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	tnRes, _ := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{},
		&mock.TrieStub{},
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: nil})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, nil)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
}

func TestTrieNodeResolver_ProcessReceivedMessageShouldGetFromTrieAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	getSerializedNodesWasCalled := false
	sendWasCalled := false
	returnedEncNodes := [][]byte{[]byte("node1"), []byte("node2")}

	tr := &mock.TrieStub{
		GetSerializedNodesCalled: func(hash []byte, maxSize uint64) ([][]byte, error) {
			if bytes.Equal([]byte("node1"), hash) {
				getSerializedNodesWasCalled = true
				return returnedEncNodes, nil
			}

			return nil, errors.New("wrong hash")
		},
	}

	tnRes, _ := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{
			SendCalled: func(buff []byte, peer p2p.PeerID) error {
				sendWasCalled = true
				return nil
			},
		},
		tr,
		marshalizer,
	)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, nil)

	assert.Nil(t, err)
	assert.True(t, getSerializedNodesWasCalled)
	assert.True(t, sendWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageShouldGetFromTrieAndMarshalizerFailShouldRetNilAndErr(t *testing.T) {
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

	tnRes, _ := NewTrieNodeResolver(
		&mock.TopicResolverSenderStub{},
		&mock.TrieStub{},
		marshalizerStub,
	)

	data, _ := marshalizerMock.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, nil)
	assert.Equal(t, errExpected, err)
}

//------- RequestTransactionFromHash

func TestTrieNodeResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	requested := &dataRetriever.RequestData{}

	res := &mock.TopicResolverSenderStub{}
	res.SendOnRequestTopicCalled = func(rd *dataRetriever.RequestData) error {
		requested = rd
		return nil
	}

	buffRequested := []byte("node1")

	tnRes, _ := NewTrieNodeResolver(
		res,
		&mock.TrieStub{},
		&mock.MarshalizerMock{},
	)

	assert.Nil(t, tnRes.RequestDataFromHash(buffRequested))
	assert.Equal(t, &dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: buffRequested,
	}, requested)

}
