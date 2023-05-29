package trie

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie/statistics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgument(timeout time.Duration) ArgTrieSyncer {
	_, trieStorage := newEmptyTrie()
	memDb := testscommon.NewMemDbMock()
	trieStorage.mainStorer = &trieMock.SnapshotPruningStorerStub{
		MemDbMock: memDb,
		PutInEpochCalled: func(key []byte, data []byte, epoch uint32) error {
			return memDb.Put(key, data)
		},
	}

	return ArgTrieSyncer{
		RequestHandler:            &testscommon.RequestHandlerStub{},
		InterceptedNodes:          testscommon.NewCacherMock(),
		DB:                        trieStorage,
		Hasher:                    &hashingMocks.HasherMock{},
		Marshalizer:               &testscommon.MarshalizerMock{},
		ShardId:                   0,
		Topic:                     "topic",
		TrieSyncStatistics:        statistics.NewTrieSyncStatistics(),
		TimeoutHandler:            testscommon.NewTimeoutHandlerMock(timeout),
		MaxHardCapForMissingNodes: 500,
		LeavesChan:                make(chan core.KeyValueHolder, 100),
	}
}

func TestNewTrieSyncer(t *testing.T) {
	t.Parallel()

	t.Run("nil request handler", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.RequestHandler = nil

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.Equal(t, err, ErrNilRequestHandler)
	})

	t.Run("nil intercepted nodes", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.InterceptedNodes = nil

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.Equal(t, err, data.ErrNilCacher)
	})

	t.Run("empty topic should fail", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.Topic = ""

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.Equal(t, err, ErrInvalidTrieTopic)
	})

	t.Run("nil trie statistics", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.TrieSyncStatistics = nil

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.Equal(t, err, ErrNilTrieSyncStatistics)
	})

	t.Run("nil database", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.DB = nil

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.True(t, errors.Is(err, ErrNilDatabase))
	})

	t.Run("nil marshalizer", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.Marshalizer = nil

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.True(t, errors.Is(err, ErrNilMarshalizer))
	})

	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.Hasher = nil

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.True(t, errors.Is(err, ErrNilHasher))
	})

	t.Run("nil timeout handler", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.TimeoutHandler = nil

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.True(t, errors.Is(err, ErrNilTimeoutHandler))
	})

	t.Run("invalid max hard capacity for missing nodes", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)
		arg.MaxHardCapForMissingNodes = 0

		ts, err := NewTrieSyncer(arg)
		assert.True(t, check.IfNil(ts))
		assert.True(t, errors.Is(err, ErrInvalidMaxHardCapForMissingNodes))
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArgument(time.Minute)

		ts, err := NewTrieSyncer(arg)
		assert.False(t, check.IfNil(ts))
		assert.Nil(t, err)
	})
}

func TestTrieSync_InterceptedNodeShouldNotBeAddedToNodesForTrieIfNodeReceived(t *testing.T) {
	t.Parallel()

	testMarshalizer, testHasher := getTestMarshalizerAndHasher()
	arg := createMockArgument(time.Second * 10)
	arg.MaxHardCapForMissingNodes = 500

	ts, err := NewTrieSyncer(arg)
	require.Nil(t, err)

	bn, collapsedBn := getBnAndCollapsedBn(testMarshalizer, testHasher)
	encodedNode, err := collapsedBn.getEncodedNode()
	assert.Nil(t, err)

	interceptedNode, err := NewInterceptedTrieNode(encodedNode, testHasher)
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
	arg := createMockArgument(timeout)
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

	_, trieStorage := newEmptyTrie()
	db := testscommon.NewMemDbMock()
	trieStorage.mainStorer = &trieMock.SnapshotPruningStorerStub{
		MemDbMock: db,
		PutInEpochWithoutCacheCalled: func(key []byte, data []byte, epoch uint32) error {
			return db.Put(key, data)
		},
	}

	err = bn.commitSnapshot(db, nil, nil, context.Background(), statistics.NewTrieStatistics(), &testscommon.ProcessStatusHandlerStub{}, 0)
	require.Nil(t, err)

	leaves, err := bn.getChildren(db)
	require.Nil(t, err)
	numLeaves := len(leaves)

	arg := createMockArgument(timeout)
	arg.RequestHandler = &testscommon.RequestHandlerStub{
		RequestTrieNodesCalled: func(destShardID uint32, hashes [][]byte, topic string) {
			assert.Fail(t, "should have not requested trie nodes")
		},
	}
	arg.DB = trieStorage
	arg.Marshalizer = testMarshalizer
	arg.Hasher = testHasher

	ts, err := NewTrieSyncer(arg)
	require.Nil(t, err)

	err = ts.StartSyncing(rootHash, context.Background())
	assert.Nil(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(numLeaves)

	numLeavesOnChan := 0
	go func() {
		for range arg.LeavesChan {
			numLeavesOnChan++
			wg.Done()
		}
	}()

	wg.Wait()

	assert.Equal(t, numLeaves, numLeavesOnChan)
}

func TestTrieSync_StartSyncing(t *testing.T) {
	t.Parallel()

	t.Run("nil hash should return nil", func(t *testing.T) {
		t.Parallel()

		timeout := time.Second * 2
		arg := createMockArgument(timeout)
		ts, _ := NewTrieSyncer(arg)

		err := ts.StartSyncing(nil, context.Background())
		assert.NoError(t, err)
	})
	t.Run("empty trie hash should return nil", func(t *testing.T) {
		t.Parallel()

		timeout := time.Second * 2
		arg := createMockArgument(timeout)
		ts, _ := NewTrieSyncer(arg)

		err := ts.StartSyncing(common.EmptyTrieHash, context.Background())
		assert.NoError(t, err)
	})
	t.Run("nil context should error", func(t *testing.T) {
		t.Parallel()

		timeout := time.Second * 2
		arg := createMockArgument(timeout)
		ts, _ := NewTrieSyncer(arg)

		err := ts.StartSyncing([]byte("roothash"), nil)
		assert.Equal(t, ErrNilContext, err)
	})
	t.Run("closed context should error", func(t *testing.T) {
		t.Parallel()

		timeout := time.Second * 2
		arg := createMockArgument(timeout)
		ts, _ := NewTrieSyncer(arg)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := ts.StartSyncing([]byte("roothash"), ctx)
		assert.Equal(t, core.ErrContextClosing, err)
	})
}
