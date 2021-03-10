package resolvers_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fromConnectedPeer = core.PeerID("from connected peer")

func createMockArgTrieNodeResolver() resolvers.ArgTrieNodeResolver {
	return resolvers.ArgTrieNodeResolver{
		SenderResolver:   &mock.TopicResolverSenderStub{},
		TrieDataGetter:   &mock.TrieStub{},
		Marshalizer:      &mock.MarshalizerMock{},
		AntifloodHandler: &mock.P2PAntifloodHandlerStub{},
		Throttler:        &mock.ThrottlerStub{},
	}
}

func TestNewTrieNodeResolver_NilResolverShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = nil
	tnRes, err := resolvers.NewTrieNodeResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilResolverSender, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_NilTrieShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	arg.TrieDataGetter = nil
	tnRes, err := resolvers.NewTrieNodeResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilTrieDataGetter, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	arg.Marshalizer = nil
	tnRes, err := resolvers.NewTrieNodeResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_NilAntiflooderShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	arg.AntifloodHandler = nil
	tnRes, err := resolvers.NewTrieNodeResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilAntifloodHandler, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_NilThrottlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	arg.Throttler = nil
	tnRes, err := resolvers.NewTrieNodeResolver(arg)

	assert.Equal(t, dataRetriever.ErrNilThrottler, err)
	assert.Nil(t, tnRes)
}

func TestNewTrieNodeResolver_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	tnRes, err := resolvers.NewTrieNodeResolver(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(tnRes))
}

//------- ProcessReceivedMessage

func TestTrieNodeResolver_ProcessReceivedAntiflooderCanProcessMessageErrShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	arg := createMockArgTrieNodeResolver()
	arg.AntifloodHandler = &mock.P2PAntifloodHandlerStub{
		CanProcessMessageCalled: func(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
			return expectedErr
		},
		CanProcessMessagesOnTopicCalled: func(peer core.PeerID, topic string, numMessages uint32, totalSize uint64, sequence []byte) error {
			return nil
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	err := tnRes.ProcessReceivedMessage(&mock.P2PMessageMock{}, fromConnectedPeer)
	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	err := tnRes.ProcessReceivedMessage(nil, fromConnectedPeer)
	assert.Equal(t, dataRetriever.ErrNilMessage, err)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	arg := createMockArgTrieNodeResolver()
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.NonceType, Value: []byte("aaa")})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, dataRetriever.ErrRequestTypeNotImplemented, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	arg := createMockArgTrieNodeResolver()
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: nil})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

//TODO in this PR: add more unit tests

func TestTrieNodeResolver_ProcessReceivedMessageShouldGetFromTrieAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	getSerializedNodesWasCalled := false
	sendWasCalled := false

	tr := &mock.TrieStub{
		GetSerializedNodeCalled: func(hash []byte) ([]byte, error) {
			if bytes.Equal([]byte("node1"), hash) {
				getSerializedNodesWasCalled = true
				return []byte("node1"), nil
			}

			return nil, errors.New("wrong hash")
		},
	}

	arg := createMockArgTrieNodeResolver()
	arg.TrieDataGetter = tr
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			sendWasCalled = true
			return nil
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)

	assert.Nil(t, err)
	assert.True(t, getSerializedNodesWasCalled)
	assert.True(t, sendWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
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

	arg := createMockArgTrieNodeResolver()
	arg.Marshalizer = marshalizerStub
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizerMock.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, errExpected, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageTrieErrorsShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected err")
	arg := createMockArgTrieNodeResolver()
	arg.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodeCalled: func(_ []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := arg.Marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, expectedErr, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesGetSerializedNodeErrorsShouldNotSend(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected err")
	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			assert.Fail(t, "should have not called send")
			return nil
		},
	}
	arg.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodeCalled: func(_ []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("hash1")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	data, _ := arg.Marshalizer.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesGetSerializedNodesErrorsShouldNotSendSubtrie(t *testing.T) {
	t.Parallel()

	nodes := [][]byte{[]byte("node1")}
	hashes := [][]byte{[]byte("hash1")}

	var receivedNodes [][]byte
	arg := createMockArgTrieNodeResolver()
	expectedErr := errors.New("expected err")
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			b := &batch.Batch{}
			err := arg.Marshalizer.Unmarshal(b, buff)
			require.Nil(t, err)
			receivedNodes = b.Data

			return nil
		},
	}
	arg.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodeCalled: func(hash []byte) ([]byte, error) {
			for i := 0; i < len(hashes); i++ {
				if bytes.Equal(hash, hashes[i]) {
					return nodes[i], nil
				}
			}

			return nil, fmt.Errorf("not found")
		},
		GetSerializedNodesCalled: func(i []byte, u uint64) ([][]byte, uint64, error) {
			return nil, 0, expectedErr
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("hash1")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	data, _ := arg.Marshalizer.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
	require.Equal(t, 1, len(receivedNodes))
	assert.Equal(t, nodes[0], receivedNodes[0])
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesNotEnoughSpaceShouldNotReadSubtries(t *testing.T) {
	t.Parallel()

	nodes := [][]byte{bytes.Repeat([]byte{1}, resolvers.MaxBuffToSendTrieNodes)}
	hashes := [][]byte{[]byte("hash1")}

	var receivedNodes [][]byte
	arg := createMockArgTrieNodeResolver()
	expectedErr := errors.New("expected err")
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			b := &batch.Batch{}
			err := arg.Marshalizer.Unmarshal(b, buff)
			require.Nil(t, err)
			receivedNodes = b.Data

			return nil
		},
	}
	arg.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodeCalled: func(hash []byte) ([]byte, error) {
			for i := 0; i < len(hashes); i++ {
				if bytes.Equal(hash, hashes[i]) {
					return nodes[i], nil
				}
			}

			return nil, fmt.Errorf("not found")
		},
		GetSerializedNodesCalled: func(i []byte, u uint64) ([][]byte, uint64, error) {
			assert.Fail(t, "should have not called GetSerializedNodesCalled")
			return nil, 0, expectedErr
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("hash1")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	data, _ := arg.Marshalizer.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
	require.Equal(t, 1, len(receivedNodes))
	assert.Equal(t, nodes[0], receivedNodes[0])
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesShouldWorkWithSubtries(t *testing.T) {
	t.Parallel()

	nodes := [][]byte{[]byte("node1"), []byte("node2")}
	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}
	subtriesNodes := [][]byte{[]byte("subtrie nodes 0"), []byte("subtrie nodes 1")}

	var receivedNodes [][]byte
	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID) error {
			b := &batch.Batch{}
			err := arg.Marshalizer.Unmarshal(b, buff)
			require.Nil(t, err)
			receivedNodes = b.Data

			return nil
		},
	}
	arg.TrieDataGetter = &mock.TrieStub{
		GetSerializedNodeCalled: func(hash []byte) ([]byte, error) {
			for i := 0; i < len(hashes); i++ {
				if bytes.Equal(hash, hashes[i]) {
					return nodes[i], nil
				}
			}

			return nil, fmt.Errorf("not found")
		},
		GetSerializedNodesCalled: func(i []byte, u uint64) ([][]byte, uint64, error) {
			used := 0
			for _, buff := range subtriesNodes {
				used += len(buff)
			}

			return subtriesNodes, u - uint64(used), nil
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("hash1"), []byte("hash2")},
	}
	buffBatch, _ := arg.Marshalizer.Marshal(b)

	data, _ := arg.Marshalizer.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &mock.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled)
	require.Equal(t, 4, len(receivedNodes))
	for _, n := range nodes {
		assert.True(t, buffInSlice(n, receivedNodes))
	}
	for _, n := range subtriesNodes {
		assert.True(t, buffInSlice(n, receivedNodes))
	}
}

func buffInSlice(buff []byte, slice [][]byte) bool {
	for _, b := range slice {
		if bytes.Equal(b, buff) {
			return true
		}
	}

	return false
}

//------- RequestTransactionFromHash

func TestTrieNodeResolver_RequestDataFromHashShouldWork(t *testing.T) {
	t.Parallel()

	requested := &dataRetriever.RequestData{}

	res := &mock.TopicResolverSenderStub{}
	res.SendOnRequestTopicCalled = func(rd *dataRetriever.RequestData, hashes [][]byte) error {
		requested = rd
		return nil
	}

	buffRequested := []byte("node1")

	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = res
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	assert.Nil(t, tnRes.RequestDataFromHash(buffRequested, 0))
	assert.Equal(t, &dataRetriever.RequestData{
		Type:  dataRetriever.HashType,
		Value: buffRequested,
	}, requested)
}

//------ NumPeersToQuery setter and getter

func TestTrieNodeResolver_SetAndGetNumPeersToQuery(t *testing.T) {
	t.Parallel()

	expectedIntra := 5
	expectedCross := 7

	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		GetNumPeersToQueryCalled: func() (int, int) {
			return expectedIntra, expectedCross
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	tnRes.SetNumPeersToQuery(expectedIntra, expectedCross)
	actualIntra, actualCross := tnRes.NumPeersToQuery()
	assert.Equal(t, expectedIntra, actualIntra)
	assert.Equal(t, expectedCross, actualCross)
}
