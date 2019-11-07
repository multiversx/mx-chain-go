package state

import (
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestNode_RequestInterceptTrieNodesWithMessenger(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var nrOfShards uint32 = 1
	var shardID uint32 = 0
	var txSignPrivKeyShardId uint32 = 0
	requesterNodeAddr := "0"
	resolverNodeAddr := "1"

	fmt.Println("Requester:	")
	nRequester := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, requesterNodeAddr)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nRequester.Node.Start()
	_ = nResolver.Node.Start()
	defer func() {
		_ = nRequester.Node.Stop()
		_ = nResolver.Node.Stop()
	}()

	time.Sleep(2 * time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	_ = nResolver.StateTrie.Update([]byte("doe"), []byte("reindeer"))
	_ = nResolver.StateTrie.Update([]byte("dog"), []byte("puppy"))
	_ = nResolver.StateTrie.Update([]byte("dogglesworth"), []byte("cat"))
	_ = nResolver.StateTrie.Commit()
	rootHash, _ := nResolver.StateTrie.Root()

	nilRootHash, _ := nRequester.StateTrie.Root()
	trieNodesResolver, _ := nRequester.ResolverFinder.IntraShardResolver(factory.TrieNodesTopic)

	trieSyncer, _ := trie.NewTrieSyncer(trieNodesResolver, nRequester.ShardDataPool.TrieNodes(), nRequester.StateTrie, time.Second)
	err = trieSyncer.StartSyncing(rootHash)
	assert.Nil(t, err)

	newRootHash, _ := nRequester.StateTrie.Root()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)
}
