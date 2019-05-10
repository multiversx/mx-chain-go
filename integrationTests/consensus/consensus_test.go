package consensus

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func initNodesAndTest(numNodes, consensusSize, numInvalid uint32, roundTime uint64, numCommBlock uint32) ([]*testNode, p2p.Messenger, *sync.Map) {
	fmt.Println("Step 1. Setup nodes...")

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	concMap := &sync.Map{}

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

	return nodes, advertiser, concMap
}

func calculatedNrSynced(nodes []*testNode, waitTime, roundTime uint64) uint32 {
	highestCommitCalled := uint32(0)
	for _, n := range nodes {
		if n.blkProcessor.NrCommitBlockCalled > highestCommitCalled {
			highestCommitCalled = n.blkProcessor.NrCommitBlockCalled
		}
	}

	committedLimit := uint64(3)
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
	numCommBlock := uint32(10)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, numCommBlock)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	combinedMap := make(map[uint32]uint64)
	totalCalled := 0
	mutex := sync.Mutex{}

	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			combinedMap[header.GetRound()] = 1
			totalCalled += 1
			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	time.Sleep(time.Second * 20)
	chDone := make(chan bool, 0)
	minRound := ^uint32(0)
	maxRound := uint32(0)
	go func() {
		mutex.Lock()

		if uint32(len(combinedMap)) >= numCommBlock {
			for k, _ := range combinedMap {
				if k > maxRound {
					maxRound = k
				}
				if k < minRound {
					minRound = k
				}
			}

			if maxRound-minRound >= numCommBlock {
				for i := minRound; i <= maxRound; i++ {
					if _, ok := combinedMap[i]; ok {
						assert.Fail(t, "consensus not reached in each round")
					}
				}
				chDone <- true
				return
			}
		}

		mutex.Unlock()

		time.Sleep(time.Second)
	}()

	select {
	case <-chDone:
	case <-time.After(40 * time.Second):
		mutex.Lock()
		assert.Equal(t, 0, minRound)
		assert.Equal(t, 10000, maxRound)
		assert.Equal(t, 10, len(combinedMap))
		assert.Equal(t, 10, totalCalled)
		assert.Fail(t, "consensus too slow, not working %d %d")
		mutex.Unlock()
		return
	}
}

func TestConsensusOnlyTestValidatorsAtLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(21)
	consensusSize := uint32(21)
	numInvalid := uint32(0)
	roundTime := uint64(2000)
	numCommBlock := uint32(10)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, numCommBlock)

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

	waitTime := time.Second * 60
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
	numInvalid := uint32(0)
	roundTime := uint64(2000)
	numCommBlock := uint32(10)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, numCommBlock)

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
