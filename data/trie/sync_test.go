package trie

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie/statistics"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrieSync_InterceptedNodeShouldNotBeAddedToNodesForTrieIfNodeReceived(t *testing.T) {
	t.Parallel()

	marshalizer, hasher := getTestMarshalizerAndHasher()
	arg := ArgTrieSyncer{
		RequestHandler:                 &mock.RequestHandlerStub{},
		InterceptedNodes:               testscommon.NewCacherMock(),
		Trie:                           &patriciaMerkleTrie{},
		ShardId:                        0,
		Topic:                          "trieNodes",
		TrieSyncStatistics:             statistics.NewTrieSyncStatistics(),
		TimeoutBetweenTrieNodesCommits: time.Second * 10,
	}
	ts, err := NewTrieSyncer(arg)
	require.Nil(t, err)

	bn, collapsedBn := getBnAndCollapsedBn(marshalizer, hasher)
	encodedNode, err := collapsedBn.getEncodedNode()
	assert.Nil(t, err)

	interceptedNode, err := NewInterceptedTrieNode(encodedNode, marshalizer, hasher)
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
	arg := ArgTrieSyncer{
		RequestHandler:   &mock.RequestHandlerStub{},
		InterceptedNodes: testscommon.NewCacherMock(),
		Trie: &patriciaMerkleTrie{
			trieStorage: &mock.StorageManagerStub{
				DatabaseCalled: func() data.DBWriteCacher {
					return mock.NewMemDbMock()
				},
			},
		},
		ShardId:                        0,
		Topic:                          "trieNodes",
		TrieSyncStatistics:             statistics.NewTrieSyncStatistics(),
		TimeoutBetweenTrieNodesCommits: timeout,
	}
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
	marshalizer, hasher := getTestMarshalizerAndHasher()
	bn, _ := getBnAndCollapsedBn(marshalizer, hasher)
	err := bn.setHash()
	require.Nil(t, err)
	rootHash := bn.getHash()
	db := mock.NewMemDbMock()

	err = bn.commit(true, 2, 2, db, db)
	require.Nil(t, err)

	arg := ArgTrieSyncer{
		RequestHandler: &mock.RequestHandlerStub{
			RequestTrieNodesCalled: func(destShardID uint32, hashes [][]byte, topic string) {
				assert.Fail(t, "should have not requested trie nodes")
			},
		},
		InterceptedNodes: testscommon.NewCacherMock(),
		Trie: &patriciaMerkleTrie{
			trieStorage: &mock.StorageManagerStub{
				DatabaseCalled: func() data.DBWriteCacher {
					return db
				},
			},
			marshalizer: marshalizer,
			hasher:      hasher,
		},
		ShardId:                        0,
		Topic:                          "trieNodes",
		TrieSyncStatistics:             statistics.NewTrieSyncStatistics(),
		TimeoutBetweenTrieNodesCommits: timeout,
	}
	ts, err := NewTrieSyncer(arg)
	require.Nil(t, err)

	err = ts.StartSyncing(rootHash, context.Background())
	assert.Nil(t, err)
}
