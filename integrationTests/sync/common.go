package sync

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var stepDelay = time.Second
var delayP2pBootstrap = time.Second * 2
var stepSync = time.Second * 2

func setupSyncNodesOneShardAndMeta(
	numNodesPerShard int,
	numNodesMeta int,
) ([]*integrationTests.TestProcessorNode, p2p.Messenger, []int) {

	maxShards := uint32(1)
	shardId := uint32(0)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, 0)
	for i := 0; i < numNodesPerShard; i++ {
		shardNode := integrationTests.NewTestSyncNode(
			maxShards,
			shardId,
			shardId,
			advertiserAddr,
		)
		nodes = append(nodes, shardNode)
	}
	idxProposerShard0 := 0

	for i := 0; i < numNodesMeta; i++ {
		metaNode := integrationTests.NewTestSyncNode(
			maxShards,
			sharding.MetachainShardId,
			shardId,
			advertiserAddr,
		)
		nodes = append(nodes, metaNode)
	}
	idxProposerMeta := len(nodes) - 1

	idxProposers := []int{idxProposerShard0, idxProposerMeta}

	return nodes, advertiser, idxProposers
}

func startSyncingBlocks(nodes []*integrationTests.TestProcessorNode) {
	for _, n := range nodes {
		_ = n.StartSync()
	}

	fmt.Println("Delaying for nodes to start syncing blocks...")
	time.Sleep(stepDelay)
}

func updateRound(nodes []*integrationTests.TestProcessorNode, round uint64) {
	for _, n := range nodes {
		n.Rounder.IndexField = int64(round)
	}
}

func proposeAndSyncBlocks(
	nodes []*integrationTests.TestProcessorNode,
	round *uint64,
	idxProposers []int,
	nonces []*uint64,
	numOfRounds int,
) {

	for i := 0; i < numOfRounds; i++ {
		crtRound := atomic.LoadUint64(round)
		proposeBlocks(nodes, idxProposers, nonces, crtRound)

		time.Sleep(stepSync)

		crtRound = integrationTests.IncrementAndPrintRound(crtRound)
		atomic.StoreUint64(round, crtRound)
		updateRound(nodes, crtRound)
		incrementNonces(nonces)
	}
	time.Sleep(stepSync)
}

func incrementNonces(nonces []*uint64) {
	for i := 0; i < len(nonces); i++ {
		atomic.AddUint64(nonces[i], 1)
	}
}

func proposeBlocks(
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	nonces []*uint64,
	crtRound uint64,
) {
	for idx, proposer := range idxProposers {
		crtNonce := atomic.LoadUint64(nonces[idx])
		integrationTests.ProposeBlock(nodes, []int{proposer}, crtRound, crtNonce)
	}
}

func forkChoiceOneBlock(nodes []*integrationTests.TestProcessorNode, shardId uint32) {
	for idx, n := range nodes {
		if n.ShardCoordinator.SelfId() != shardId {
			continue
		}
		err := n.Bootstrapper.ForkChoice(false)
		if err != nil {
			fmt.Println(err)
		}

		newNonce := n.BlockChain.GetCurrentBlockHeader().GetNonce()
		fmt.Printf("Node's id %d is at block height %d\n", idx, newNonce)
	}
}

func emptyDataPools(nodes []*integrationTests.TestProcessorNode, shardId uint32) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != shardId {
			continue
		}

		emptyNodeDataPool(n)
	}
}

func emptyNodeDataPool(node *integrationTests.TestProcessorNode) {
	if node.ShardDataPool != nil {
		emptyShardDataPool(node.ShardDataPool)
	}
	if node.MetaDataPool != nil {
		emptyMetaDataPool(node.MetaDataPool)
	}
}

func emptyShardDataPool(sdp dataRetriever.PoolsHolder) {
	sdp.HeadersNonces().Clear()
	sdp.Headers().Clear()
	sdp.UnsignedTransactions().Clear()
	sdp.Transactions().Clear()
	sdp.MetaBlocks().Clear()
	sdp.MiniBlocks().Clear()
	sdp.PeerChangesBlocks().Clear()
}

func emptyMetaDataPool(holder dataRetriever.MetaPoolsHolder) {
	holder.HeadersNonces().Clear()
	holder.MetaBlocks().Clear()
	holder.MiniBlockHashes().Clear()
	holder.ShardHeaders().Clear()
}

func resetHighestProbableNonce(nodes []*integrationTests.TestProcessorNode, shardId uint32, targetNonce uint64) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() != shardId {
			continue
		}
		if n.BlockChain.GetCurrentBlockHeader().GetNonce() != targetNonce {
			continue
		}

		n.Bootstrapper.SetProbableHighestNonce(targetNonce)
	}
}
