package block

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

// TestShardShouldNotProposeAndExecuteTwoBlocksInSameRound tests that a shard can not continue building on a
// chain with 2 blocks in the same round
func TestShardShouldNotProposeAndExecuteTwoBlocksInSameRound(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	log := logger.DefaultLogger()
	log.SetLevel(logger.LogDebug)

	maxShards := uint32(1)
	numOfNodes := 4
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	}

	idxProposer := 0

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(1)
	round = integrationTests.IncrementAndPrintRound(round)

	err := proposeAndCommitBlock(nodes[idxProposer], round, nonce)
	assert.Nil(t, err)

	integrationTests.SyncBlock(t, nodes, []int{idxProposer}, nonce)

	time.Sleep(20 * stepDelay)

	checkCurrentBlockHeight(t, nodes, nonce)

	//only nonce increases, round stays the same
	nonce++

	err = proposeAndCommitBlock(nodes[idxProposer], round, nonce)
	assert.Equal(t, process.ErrLowerRoundInBlock, err)

	//mockTestingT is used as in normal case SyncBlock would fail as it doesn't find the header with nonce 2
	mockTestingT := &testing.T{}
	integrationTests.SyncBlock(mockTestingT, nodes, []int{idxProposer}, nonce)

	time.Sleep(stepDelay)

	checkCurrentBlockHeight(t, nodes, nonce-1)
}

func proposeAndCommitBlock(node *integrationTests.TestProcessorNode, round uint64, nonce uint64) error {
	body, hdr, _ := node.ProposeBlock(round, nonce)
	err := node.BlockProcessor.CommitBlock(node.BlockChain, hdr, body)
	if err != nil {
		return err
	}

	node.BroadcastBlock(body, hdr)
	time.Sleep(stepDelay)
	return nil
}

func checkCurrentBlockHeight(t *testing.T, nodes []*integrationTests.TestProcessorNode, nonce uint64) {
	for _, n := range nodes {
		assert.Equal(t, nonce, n.BlockChain.GetCurrentBlockHeader().GetNonce())
	}
}
