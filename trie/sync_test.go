package trie

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/mock"
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgument() ArgTrieSyncer {
	return ArgTrieSyncer{
		RequestHandler:            &mock.RequestHandlerStub{},
		InterceptedNodes:          mock.NewCacherMock(),
		DB:                        mock.NewMemDbMock(),
		Hasher:                    mock.HasherMock{},
		Marshalizer:               &mock.MarshalizerMock{},
		ShardId:                   0,
		Topic:                     "topic",
		TrieSyncStatistics:        statistics.NewTrieSyncStatistics(),
		ReceivedNodesTimeout:      time.Minute,
		MaxHardCapForMissingNodes: 500,
	}
}

func TestNewTrieSyncer_NilRequestHandlerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.RequestHandler = nil

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.Equal(t, err, ErrNilRequestHandler)
}

func TestNewTrieSyncer_NilInterceptedNodesShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.InterceptedNodes = nil

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.Equal(t, err, data.ErrNilCacher)
}

func TestNewTrieSyncer_EmptyTopicShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Topic = ""

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.Equal(t, err, ErrInvalidTrieTopic)
}

func TestNewTrieSyncer_NilTrieStatisticsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.TrieSyncStatistics = nil

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.Equal(t, err, ErrNilTrieSyncStatistics)
}

func TestNewTrieSyncer_NilDatabaseShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.DB = nil

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.True(t, errors.Is(err, ErrNilDatabase))
}

func TestNewTrieSyncer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Marshalizer = nil

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.True(t, errors.Is(err, ErrNilMarshalizer))
}

func TestNewTrieSyncer_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.Hasher = nil

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.True(t, errors.Is(err, ErrNilHasher))
}

func TestNewTrieSyncer_InvalidTimeoutBetweenTrieNodesCommitsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.ReceivedNodesTimeout = time.Duration(0)

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.True(t, errors.Is(err, ErrInvalidTimeout))
}

func TestNewTrieSyncer_InvalidMaxHardCapForMissingNodesShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()
	arg.MaxHardCapForMissingNodes = 0

	ts, err := NewTrieSyncer(arg)
	assert.True(t, check.IfNil(ts))
	assert.True(t, errors.Is(err, ErrInvalidMaxHardCapForMissingNodes))
}

func TestNewTrieSyncer_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgument()

	ts, err := NewTrieSyncer(arg)
	assert.False(t, check.IfNil(ts))
	assert.Nil(t, err)
}

func TestTrieSync_InterceptedNodeShouldNotBeAddedToNodesForTrieIfNodeReceived(t *testing.T) {
	t.Parallel()

	testMarshalizer, testHasher := getTestMarshalizerAndHasher()
	arg := createMockArgument()
	arg.ReceivedNodesTimeout = time.Second * 10
	arg.MaxHardCapForMissingNodes = 500

	ts, err := NewTrieSyncer(arg)
	require.Nil(t, err)

	bn, collapsedBn := getBnAndCollapsedBn(testMarshalizer, testHasher)
	encodedNode, err := collapsedBn.getEncodedNode()
	assert.Nil(t, err)

	interceptedNode, err := NewInterceptedTrieNode(encodedNode, testMarshalizer, testHasher)
	assert.Nil(t, err)

	hash := "nodeHash"
	ts.nodesForTrie[hash] = trieNodeInfo{
		trieNode: bn,
		received: true,
	}

	ts.trieNodeIntercepted([]byte(hash), interceptedNode)

	nodeInfo, ok := ts.nodesForTrie[hash]
	assert.True(t, ok)
	assert.Equal(t, bn, nodeInfo.trieNode)
}

func TestTrieSync_InterceptedNodeTimedOut(t *testing.T) {
	t.Parallel()

	timeout := time.Second * 2
	arg := createMockArgument()
	arg.ReceivedNodesTimeout = timeout
	ts, err := NewTrieSyncer(arg)
	require.Nil(t, err)

	start := time.Now()
	err = ts.StartSyncing([]byte("roothash"), context.Background())
	end := time.Now()

	assert.True(t, errors.Is(err, ErrTimeIsOut))
	assert.True(t, timeout <= end.Sub(start))
}

func TestTrieSync_FoundInStorageShouldNotRequest(t *testing.T) {
	t.Parallel()

	timeout := time.Second * 200
	testMarshalizer, testHasher := getTestMarshalizerAndHasher()
	bn, _ := getBnAndCollapsedBn(testMarshalizer, testHasher)
	err := bn.setHash()
	require.Nil(t, err)
	rootHash := bn.getHash()
	db := mock.NewMemDbMock()

	err = bn.commitSnapshot(db, db, nil, context.Background())
	require.Nil(t, err)

	arg := createMockArgument()
	arg.RequestHandler = &mock.RequestHandlerStub{
		RequestTrieNodesCalled: func(destShardID uint32, hashes [][]byte, topic string) {
			assert.Fail(t, "should have not requested trie nodes")
		},
	}
	arg.DB = db
	arg.Marshalizer = testMarshalizer
	arg.Hasher = testHasher
	arg.ReceivedNodesTimeout = timeout

	ts, err := NewTrieSyncer(arg)
	require.Nil(t, err)

	err = ts.StartSyncing(rootHash, context.Background())
	assert.Nil(t, err)
}
