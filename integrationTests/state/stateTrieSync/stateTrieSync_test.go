package stateTrieSync

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/trie"
	factory2 "github.com/ElrondNetwork/elrond-go/data/trie/factory"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/requestHandlers"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/interceptors"
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
	_ = nRequester.Messenger.CreateTopic(core.ConsensusTopic+nRequester.ShardCoordinator.CommunicationIdentifier(nRequester.ShardCoordinator.SelfId()), true)

	fmt.Println("Resolver:")
	nResolver := integrationTests.NewTestProcessorNode(nrOfShards, shardID, txSignPrivKeyShardId, resolverNodeAddr)
	_ = nResolver.Messenger.CreateTopic(core.ConsensusTopic+nResolver.ShardCoordinator.CommunicationIdentifier(nResolver.ShardCoordinator.SelfId()), true)
	defer func() {
		_ = nRequester.Messenger.Close()
		_ = nResolver.Messenger.Close()
	}()

	time.Sleep(time.Second)
	err := nRequester.Messenger.ConnectToPeer(integrationTests.GetConnectableAddress(nResolver.Messenger))
	assert.Nil(t, err)

	time.Sleep(integrationTests.SyncDelay)

	resolverTrie := nResolver.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	//we have tested even with the 50000 value and found out that it worked in a reasonable amount of time ~21 seconds
	for i := 0; i < 10000; i++ {
		_ = resolverTrie.Update([]byte(strconv.Itoa(i)), []byte(strconv.Itoa(i)))
	}

	_ = resolverTrie.Commit()
	rootHash, _ := resolverTrie.Root()

	_, err = resolverTrie.GetAllLeaves()
	assert.Nil(t, err)

	requesterTrie := nRequester.TrieContainer.Get([]byte(factory2.UserAccountTrie))
	nilRootHash, _ := requesterTrie.Root()
	whiteListHandler, _ := interceptors.NewWhiteListDataVerifier(
		&mock.CacherStub{
			PutCalled: func(_ []byte, _ interface{}, _ int) (evicted bool) {
				return false
			},
		},
	)
	requestHandler, _ := requestHandlers.NewResolverRequestHandler(
		nRequester.ResolverFinder,
		&mock.RequestedItemsHandlerStub{},
		whiteListHandler,
		10000,
		nRequester.ShardCoordinator.SelfId(),
		time.Second,
	)

	waitTime := 100 * time.Second
	trieSyncer, _ := trie.NewTrieSyncer(requestHandler, nRequester.DataPool.TrieNodes(), requesterTrie, shardID, factory.AccountTrieNodesTopic)
	ctx, cancel := context.WithTimeout(context.Background(), waitTime)
	defer cancel()

	err = trieSyncer.StartSyncing(rootHash, ctx)
	assert.Nil(t, err)

	newRootHash, _ := requesterTrie.Root()
	assert.NotEqual(t, nilRootHash, newRootHash)
	assert.Equal(t, rootHash, newRootHash)

	_, err = requesterTrie.GetAllLeaves()
	assert.Nil(t, err)
}
