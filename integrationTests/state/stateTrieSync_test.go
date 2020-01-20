package state

import (
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/trie"
	factory2 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
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

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	resolverTrie := nResolver.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	_ = resolverTrie.Update([]byte("doe"), []byte("reindeer"))
	_ = resolverTrie.Update([]byte("dog"), []byte("puppy"))
	_ = resolverTrie.Update([]byte("dogglesworth"), []byte("cat"))
	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.Root()

	requesterTrie := nRequester.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	nilRootHash, _ := requesterTrie.Root()
	trieNodesResolver, _ := nRequester.ResolverFinder.CrossShardResolver(factory.AccountTrieNodesTopic, core.MetachainShardId)

	waitTime := 5 * time.Second
	trieSyncer, _ := trie.NewTrieSyncer(trieNodesResolver, nRequester.DataPool.TrieNodes(), requesterTrie, waitTime)
	err = trieSyncer.StartSyncing(rootHash)
	assert.Nil(t, err)

	newRootHash, _ := requesterTrie.Root()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)
}
