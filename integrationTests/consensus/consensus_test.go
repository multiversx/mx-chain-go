package consensus

import (
	"context"
	"fmt"
	"gotest.tools/assert"
	"testing"
	"time"
)

func TestConsensusOnlyTest(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Step 1. Setup nodes...")
	nodesPerShard := 21

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodes(
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

	fmt.Println("Run for 20 seconds...")
	time.Sleep(time.Second * 20)

	// test is good if 2/3+1 of validators are somewhat synchronized - blocks committed.
	highestCommitCalled := uint32(0)
	for _, n := range nodes {
		if n.blkProcessor.NrCommitBlockCalled > highestCommitCalled {
			highestCommitCalled += n.blkProcessor.NrCommitBlockCalled
		}
	}

	nrSynced := 0
	for _, n := range nodes {
		if highestCommitCalled-n.blkProcessor.NrCommitBlockCalled < 2 {
			nrSynced += 1
		}
	}

	passed := nrSynced > (len(nodes)*2)/3
	assert.Equal(t, true, passed)
}

func TestConsensusWithMetaBlockProcessor(t *testing.T) {

}
