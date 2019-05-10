package consensus

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func initNodesAndTest(numNodes, consensusSize, numInvalid uint32, roundTime uint64) ([]*testNode, p2p.Messenger) {
	fmt.Println("Step 1. Setup nodes...")

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodes(
		int(numNodes),
		int(consensusSize),
		roundTime,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	if numInvalid < numNodes {
		for i := uint32(0); i < numInvalid; i++ {
			nodes[i].blkProcessor.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				return process.ErrInvalidBlockHash
			}
		}
	}

	return nodes, advertiser
}

func calculatedNrSynced(nodes []*testNode, waitTime, roundTime uint64) uint32 {
	highestCommitCalled := uint32(0)
	for _, n := range nodes {
		if n.blkProcessor.NrCommitBlockCalled > highestCommitCalled {
			highestCommitCalled = n.blkProcessor.NrCommitBlockCalled
		}
	}

	committedLimit := (waitTime/roundTime)/2 - 2
	if committedLimit < 2 {
		committedLimit = 2
	}
	if committedLimit > uint64(highestCommitCalled) {
		return 0
	}

	nrSynced := uint32(0)
	for _, n := range nodes {
		if n.blkProcessor.NrCommitBlockCalled < 2 {
			continue
		}

		if highestCommitCalled-n.blkProcessor.NrCommitBlockCalled < 2 {
			nrSynced += 1
		}
	}

	return nrSynced
}

func TestConsensusFullTest(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(0)
	roundTime := uint64(2000)
	nodes, advertiser := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime)

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

	waitTime := time.Second * 20
	fmt.Println("Run for 20 seconds...")
	time.Sleep(waitTime)

	nrSynced := calculatedNrSynced(nodes, uint64(waitTime), uint64(roundTime*uint64(time.Millisecond)))

	passed := nrSynced > uint32((len(nodes)*2)/3)
	assert.Equal(t, true, passed)
}

func TestConsensusOnlyTestValidatorsAtLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(6)
	roundTime := uint64(2000)
	nodes, advertiser := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime)

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

	waitTime := time.Second * 20
	fmt.Println("Run for 20 seconds...")
	time.Sleep(waitTime)

	nrSynced := calculatedNrSynced(nodes, uint64(waitTime), uint64(roundTime*uint64(time.Millisecond)))

	passed := nrSynced > uint32((len(nodes)*2)/3)
	assert.Equal(t, true, passed)
}

func TestConsensusNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(7)
	roundTime := uint64(2000)
	nodes, advertiser := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime)

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

	waitTime := time.Second * 20
	fmt.Println("Run for 20 seconds...")
	time.Sleep(waitTime)

	nrSynced := calculatedNrSynced(nodes, uint64(waitTime), uint64(roundTime*uint64(time.Millisecond)))

	passed := nrSynced > uint32((len(nodes)*2)/3)
	assert.Equal(t, false, passed)
}
