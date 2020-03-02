package trie_test

import (
	"io/ioutil"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
	"github.com/ElrondNetwork/elrond-go/storage/storageUnit"
	"github.com/stretchr/testify/assert"
)

func getInterceptedNodes(tr data.Trie, marshalizer marshal.Marshalizer, hasher hashing.Hasher) []*trie.InterceptedTrieNode {
	nodes, _ := getEncodedTrieNodesAndHashes(tr)

	interceptedNodes := make([]*trie.InterceptedTrieNode, 0)
	for i := range nodes {
		node, _ := trie.NewInterceptedTrieNode(nodes[i], marshalizer, hasher)
		interceptedNodes = append(interceptedNodes, node)
	}

	return interceptedNodes
}

func TestTrieSyncer_StartSyncing(t *testing.T) {
	t.Parallel()

	db := mock.NewMemDbMock()
	marshalizer := &mock.ProtobufMarshalizerMock{}
	hasher := &mock.KeccakMock{}
	tempDir, _ := ioutil.TempDir("", strconv.Itoa(rand.Intn(100000)))
	cfg := config.DBConfig{
		FilePath:          tempDir,
		Type:              string(storageUnit.LvlDBSerial),
		BatchDelaySeconds: 1,
		MaxBatchSize:      1,
		MaxOpenFiles:      10,
	}

	evictionWaitingList, _ := mock.NewEvictionWaitingList(100, memorydb.New(), marshalizer)
	trieStorage, _ := trie.NewTrieStorageManager(db, cfg, evictionWaitingList)
	tr, _ := trie.NewTrie(trieStorage, marshalizer, hasher)

	syncTrie := initTrie()
	interceptedNodesCacher, _ := lrucache.NewCache(100)
	interceptedNodes := getInterceptedNodes(syncTrie, marshalizer, hasher)
	nrNodesToSend := 2
	nodesIndex := 0
	nrRequests := 0
	expectedRequests := 3

	resolver := &mock.TrieNodesResolverStub{
		RequestDataFromHashCalled: func(hash []byte) error {
			requestedNode := interceptedNodes[nodesIndex]
			for i := nodesIndex; i < nodesIndex+nrNodesToSend; i++ {
				interceptedNodesCacher.Put(interceptedNodes[i].Hash(), interceptedNodes[i])
			}

			requestedNode.CreateEndOfProcessingTriggerNode()
			interceptedNodesCacher.Put(requestedNode.Hash(), requestedNode)
			nodesIndex += nrNodesToSend
			nrRequests++

			return nil
		},
	}

	rootHash, _ := syncTrie.Root()
	sync, _ := trie.NewTrieSyncer(resolver, interceptedNodesCacher, tr, 10*time.Second)

	_ = sync.StartSyncing(rootHash)
	newTrieRootHash, _ := tr.Root()
	assert.Equal(t, rootHash, newTrieRootHash)
	assert.Equal(t, expectedRequests, nrRequests)
}
