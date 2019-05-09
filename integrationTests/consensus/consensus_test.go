package consensus

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConsensusOnlyTest(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Step 1. Setup nodes...")
	startingPort := 20000
	nodesPerShard := 21

	advertiser := createMessengerWithKadDht(context.Background(), startingPort, "")
	advertiser.Bootstrap()
	startingPort++

	nodes := createNodes(
		startingPort,
		nodesPerShard,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	for _, n := range nodes {
		_ = n.node.StartConsensus()
	}

	fmt.Println("Wait 10 seconds...")
	time.Sleep(time.Second * 30)

	for _, n := range nodes {
		isHigher := (n.blkProcessor.NrCommitBlockCalled > 2)
		assert.True(t, isHigher)
	}
}

func TestConsensusWithMetaBlockProcessor(t *testing.T) {

}
