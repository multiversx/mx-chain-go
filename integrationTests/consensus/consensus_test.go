package consensus

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
)

func encodeAddress(address []byte) string {
	return hex.EncodeToString(address)
}

func getPkEncoded(pubKey crypto.PublicKey) string {
	pk, err := pubKey.ToByteArray()
	if err != nil {
		return err.Error()
	}

	return encodeAddress(pk)
}

func initNodesAndTest(numNodes, consensusSize, numInvalid uint32, roundTime uint64, consensusType string) ([]*testNode, p2p.Messenger, *sync.Map) {
	fmt.Println("Step 1. Setup nodes...")

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	concMap := &sync.Map{}

	nodes := createNodes(
		int(numNodes),
		int(consensusSize),
		roundTime,
		getConnectableAddress(advertiser),
		consensusType,
	)
	displayAndStartNodes(nodes)

	if numInvalid < numNodes {
		for i := uint32(0); i < numInvalid; i++ {
			nodes[i].blkProcessor.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
				fmt.Println("process block invalid ", header.GetRound(), header.GetNonce(), getPkEncoded(nodes[i].pk))
				return process.ErrInvalidBlockHash
			}
			nodes[i].blkProcessor.CreateBlockHeaderCalled = func(body data.BodyHandler, round int32, haveTime func() bool) (handler data.HeaderHandler, e error) {
				return nil, process.ErrAccountStateDirty
			}
			nodes[i].blkProcessor.CreateBlockCalled = func(round int32, haveTime func() bool) (handler data.BodyHandler, e error) {
				return nil, process.ErrWrongTypeAssertion
			}
		}
	}

	return nodes, advertiser, concMap
}

func checkBlockProposedEveryRound(numCommBlock uint32, combinedMap map[uint32]uint64, mutex *sync.Mutex, chDone chan bool, t *testing.T) {
	for {
		mutex.Lock()

		minRound := ^uint32(0)
		maxRound := uint32(0)
		if uint32(len(combinedMap)) >= numCommBlock {
			for k := range combinedMap {
				if k > maxRound {
					maxRound = k
				}
				if k < minRound {
					minRound = k
				}
			}

			if maxRound-minRound >= numCommBlock {
				for i := minRound; i <= maxRound; i++ {
					if _, ok := combinedMap[i]; !ok {
						assert.Fail(t, "consensus not reached in each round")
						fmt.Println("combined map: \n", combinedMap)
						mutex.Unlock()
						return
					}
				}
				chDone <- true
				mutex.Unlock()
				return
			}
		}

		mutex.Unlock()

		time.Sleep(time.Second * 2)
	}
}

func TestConsensusBNFullTest(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(0)
	roundTime := uint64(4000)
	numCommBlock := uint32(10)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, bnConsensusType)

	mutex := &sync.Mutex{}
	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
		mutex.Lock()
		mutex.Unlock()
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	combinedMap := make(map[uint32]uint64)
	totalCalled := 0

	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			combinedMap[header.GetRound()] = header.GetNonce()
			totalCalled += 1
			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	chDone := make(chan bool, 0)
	go checkBlockProposedEveryRound(numCommBlock, combinedMap, mutex, chDone, t)

	extraTime := uint32(2)
	endTime := time.Duration(roundTime) * time.Duration(numCommBlock+extraTime) * time.Millisecond
	select {
	case <-chDone:
	case <-time.After(endTime):
		mutex.Lock()
		fmt.Println("combined map: \n", combinedMap)
		assert.Fail(t, "consensus too slow, not working.")
		mutex.Unlock()
		return
	}
}

func TestConsensusBNNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(2)
	roundTime := uint64(4000)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, bnConsensusType)

	mutex := &sync.Mutex{}
	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
		mutex.Lock()
		mutex.Unlock()
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	maxNonce := uint64(0)
	minNonce := ^uint64(0)
	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			if maxNonce < header.GetNonce() {
				maxNonce = header.GetNonce()
			}

			if minNonce < header.GetNonce() {
				minNonce = header.GetNonce()
			}
			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	waitTime := time.Second * 60
	fmt.Println("Run for 60 seconds...")
	time.Sleep(waitTime)

	mutex.Lock()
	assert.Equal(t, uint64(0), maxNonce)
	mutex.Unlock()
}

func TestConsensusBLSFullTest(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(0)
	roundTime := uint64(4000)
	numCommBlock := uint32(10)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, blsConsensusType)

	mutex := &sync.Mutex{}
	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
		mutex.Lock()
		mutex.Unlock()
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	combinedMap := make(map[uint32]uint64)
	totalCalled := 0

	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			combinedMap[header.GetRound()] = header.GetNonce()
			totalCalled += 1
			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	chDone := make(chan bool, 0)
	go checkBlockProposedEveryRound(numCommBlock, combinedMap, mutex, chDone, t)

	extraTime := uint32(2)
	endTime := time.Duration(roundTime) * time.Duration(numCommBlock+extraTime) * time.Millisecond
	select {
	case <-chDone:
	case <-time.After(endTime):
		mutex.Lock()
		fmt.Println("combined map: \n", combinedMap)
		assert.Fail(t, "consensus too slow, not working")
		mutex.Unlock()
		return
	}
}

func TestConsensusBLSNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(2)
	roundTime := uint64(4000)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, blsConsensusType)

	mutex := &sync.Mutex{}
	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
		mutex.Lock()
		mutex.Unlock()
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 1)

	maxNonce := uint64(0)
	minNonce := ^uint64(0)
	for _, n := range nodes {
		n.blkProcessor.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
			n.blkProcessor.NrCommitBlockCalled++
			_ = blockChain.SetCurrentBlockHeader(header)
			_ = blockChain.SetCurrentBlockBody(body)

			mutex.Lock()
			if maxNonce < header.GetNonce() {
				maxNonce = header.GetNonce()
			}

			if minNonce < header.GetNonce() {
				minNonce = header.GetNonce()
			}
			mutex.Unlock()

			return nil
		}
		_ = n.node.StartConsensus()
	}

	waitTime := time.Second * 60
	fmt.Println("Run for 60 seconds...")
	time.Sleep(waitTime)

	mutex.Lock()
	assert.Equal(t, uint64(0), maxNonce)
	mutex.Unlock()
}
