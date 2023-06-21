package resolvers_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/batch"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fromConnectedPeer = core.PeerID("from connected peer")

func createMockArgTrieNodeResolver() resolvers.ArgTrieNodeResolver {
	return resolvers.ArgTrieNodeResolver{
		ArgBaseResolver: createMockArgBaseResolver(),
		TrieDataGetter:  &trieMock.TrieStub{},
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
	arg.Marshaller = nil
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

func TestTrieNodeResolver_ProcessReceivedAntiflooderCanProcessMessageErrShouldErr(t *testing.T) {
	t.Parallel()

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

	err := tnRes.ProcessReceivedMessage(&p2pmocks.P2PMessageMock{}, fromConnectedPeer)
	assert.True(t, errors.Is(err, expectedErr))
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestTrieNodeResolver_ProcessReceivedMessageNilMessageShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	err := tnRes.ProcessReceivedMessage(nil, fromConnectedPeer)
	assert.Equal(t, dataRetriever.ErrNilMessage, err)
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.False(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestTrieNodeResolver_ProcessReceivedMessageWrongTypeShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	arg := createMockArgTrieNodeResolver()
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.NonceType, Value: []byte("aaa")})
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, dataRetriever.ErrRequestTypeNotImplemented, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestTrieNodeResolver_ProcessReceivedMessageNilValueShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	arg := createMockArgTrieNodeResolver()
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: nil})
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, dataRetriever.ErrNilValue, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

//TODO in this PR: add more unit tests

func TestTrieNodeResolver_ProcessReceivedMessageShouldGetFromTrieAndSend(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	getSerializedNodesWasCalled := false
	sendWasCalled := false

	tr := &trieMock.TrieStub{
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
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			sendWasCalled = true
			return nil
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)

	assert.Nil(t, err)
	assert.True(t, getSerializedNodesWasCalled)
	assert.True(t, sendWasCalled)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
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
	arg.Marshaller = marshalizerStub
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := marshalizerMock.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, errExpected, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestTrieNodeResolver_ProcessReceivedMessageTrieErrorsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	arg.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodeCalled: func(_ []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := arg.Marshaller.Marshal(&dataRetriever.RequestData{Type: dataRetriever.HashType, Value: []byte("node1")})
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, expectedErr, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesUnmarshalFails(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	initialMarshaller := arg.Marshaller
	cnt := 0
	arg.Marshaller = &mock.MarshalizerStub{
		MarshalCalled: initialMarshaller.Marshal,
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			cnt++
			if cnt > 1 {
				return expectedErr
			}
			return initialMarshaller.Unmarshal(obj, buff)
		},
	}
	arg.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodeCalled: func(_ []byte) ([]byte, error) {
			assert.Fail(t, "should have not called send")
			return nil, nil
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("hash1")},
	}
	buffBatch, _ := arg.Marshaller.Marshal(b)

	data, _ := arg.Marshaller.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Equal(t, expectedErr, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesGetSerializedNodeErrorsShouldNotSend(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			assert.Fail(t, "should have not called send")
			return nil
		},
	}
	arg.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodeCalled: func(_ []byte) ([]byte, error) {
			return nil, expectedErr
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	b := &batch.Batch{
		Data: [][]byte{[]byte("hash1")},
	}
	buffBatch, _ := arg.Marshaller.Marshal(b)

	data, _ := arg.Marshaller.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesGetSerializedNodesErrorsShouldNotSendSubtrie(t *testing.T) {
	t.Parallel()

	nodes := [][]byte{[]byte("node1")}
	hashes := [][]byte{[]byte("hash1")}

	var receivedNodes [][]byte
	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			b := &batch.Batch{}
			err := arg.Marshaller.Unmarshal(b, buff)
			require.Nil(t, err)
			receivedNodes = b.Data

			return nil
		},
	}
	arg.TrieDataGetter = &trieMock.TrieStub{
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
	buffBatch, _ := arg.Marshaller.Marshal(b)

	data, _ := arg.Marshaller.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
	require.Equal(t, 1, len(receivedNodes))
	assert.Equal(t, nodes[0], receivedNodes[0])
}

func TestTrieNodeResolver_ProcessReceivedMessageMultipleHashesNotEnoughSpaceShouldNotReadSubtries(t *testing.T) {
	t.Parallel()

	nodes := [][]byte{bytes.Repeat([]byte{1}, core.MaxBufferSizeToSendTrieNodes)}
	hashes := [][]byte{[]byte("hash1")}

	var receivedNodes [][]byte
	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			b := &batch.Batch{}
			err := arg.Marshaller.Unmarshal(b, buff)
			require.Nil(t, err)
			receivedNodes = b.Data

			return nil
		},
	}
	arg.TrieDataGetter = &trieMock.TrieStub{
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
	buffBatch, _ := arg.Marshaller.Marshal(b)

	data, _ := arg.Marshaller.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
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
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			b := &batch.Batch{}
			err := arg.Marshaller.Unmarshal(b, buff)
			require.Nil(t, err)
			receivedNodes = b.Data

			return nil
		},
	}
	arg.TrieDataGetter = &trieMock.TrieStub{
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
	buffBatch, _ := arg.Marshaller.Marshal(b)

	data, _ := arg.Marshaller.Marshal(
		&dataRetriever.RequestData{
			Type:  dataRetriever.HashArrayType,
			Value: buffBatch,
		},
	)
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
	require.Equal(t, 4, len(receivedNodes))
	for _, n := range nodes {
		assert.True(t, buffInSlice(n, receivedNodes))
	}
	for _, n := range subtriesNodes {
		assert.True(t, buffInSlice(n, receivedNodes))
	}
}

func testTrieNodeResolverProcessReceivedMessageLargeTrieNode(
	t *testing.T,
	largeBuffer []byte,
	chunkIndex uint32,
	maxComputedChunks uint32,
	startIndexBuff int,
	endIndexBuff int,
) {
	nodes := [][]byte{largeBuffer, []byte("node2")}
	hashes := [][]byte{[]byte("hash1"), []byte("hash2")}

	sendWasCalled := false
	arg := createMockArgTrieNodeResolver()
	arg.SenderResolver = &mock.TopicResolverSenderStub{
		SendCalled: func(buff []byte, peer core.PeerID, network p2p.Network) error {
			b := &batch.Batch{}
			err := arg.Marshaller.Unmarshal(b, buff)
			require.Nil(t, err)
			sendWasCalled = true
			assert.Equal(t, maxComputedChunks, b.MaxChunks)
			assert.Equal(t, chunkIndex, b.ChunkIndex)
			require.Equal(t, 1, len(b.Data))
			chunk := b.Data[0]
			assert.Equal(t, largeBuffer[startIndexBuff:endIndexBuff], chunk)

			return nil
		},
	}
	arg.TrieDataGetter = &trieMock.TrieStub{
		GetSerializedNodeCalled: func(hash []byte) ([]byte, error) {
			for i := 0; i < len(hashes); i++ {
				if bytes.Equal(hash, hashes[i]) {
					return nodes[i], nil
				}
			}

			return nil, fmt.Errorf("not found")
		},
		GetSerializedNodesCalled: func(i []byte, u uint64) ([][]byte, uint64, error) {
			return make([][]byte, 0), 0, nil
		},
	}
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	data, _ := arg.Marshaller.Marshal(
		&dataRetriever.RequestData{
			Type:       dataRetriever.HashType,
			Value:      []byte("hash1"),
			ChunkIndex: chunkIndex,
		},
	)
	msg := &p2pmocks.P2PMessageMock{DataField: data}

	err := tnRes.ProcessReceivedMessage(msg, fromConnectedPeer)
	assert.Nil(t, err)
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).StartWasCalled())
	assert.True(t, arg.Throttler.(*mock.ThrottlerStub).EndWasCalled())
	require.True(t, sendWasCalled)
}

func TestTrieNodeResolver_ProcessReceivedMessageLargeTrieNodeShouldSendFirstChunk(t *testing.T) {
	t.Parallel()

	randBuff := make([]byte, 1<<20) //1MB
	_, _ = rand.Read(randBuff)
	testTrieNodeResolverProcessReceivedMessageLargeTrieNode(t, randBuff, 0, 4, 0, core.MaxBufferSizeToSendTrieNodes)
}

func TestTrieNodeResolver_ProcessReceivedMessageLargeTrieNodeShouldSendRequiredChunk(t *testing.T) {
	t.Parallel()

	randBuff := make([]byte, 1<<20) //1MB
	_, _ = rand.Read(randBuff)
	testTrieNodeResolverProcessReceivedMessageLargeTrieNode(
		t,
		randBuff,
		1,
		4,
		core.MaxBufferSizeToSendTrieNodes,
		2*core.MaxBufferSizeToSendTrieNodes,
	)
	testTrieNodeResolverProcessReceivedMessageLargeTrieNode(
		t,
		randBuff,
		2,
		4,
		2*core.MaxBufferSizeToSendTrieNodes,
		3*core.MaxBufferSizeToSendTrieNodes,
	)
	testTrieNodeResolverProcessReceivedMessageLargeTrieNode(
		t,
		randBuff,
		3,
		4,
		3*core.MaxBufferSizeToSendTrieNodes,
		4*core.MaxBufferSizeToSendTrieNodes,
	)

	randBuff = make([]byte, 1<<20+1) //1MB + 1 byte
	_, _ = rand.Read(randBuff)
	startIndex := len(randBuff) - 1
	endIndex := len(randBuff)
	testTrieNodeResolverProcessReceivedMessageLargeTrieNode(t, randBuff, 4, 5, startIndex, endIndex)
}

func buffInSlice(buff []byte, slice [][]byte) bool {
	for _, b := range slice {
		if bytes.Equal(b, buff) {
			return true
		}
	}

	return false
}

func TestTrieNodeResolver_Close(t *testing.T) {
	t.Parallel()

	arg := createMockArgTrieNodeResolver()
	tnRes, _ := resolvers.NewTrieNodeResolver(arg)

	assert.Nil(t, tnRes.Close())
}
