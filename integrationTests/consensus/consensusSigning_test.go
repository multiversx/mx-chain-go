package consensus

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/stretchr/testify/assert"
)

func initNodesWithTestSigner(
	numMetaNodes,
	numNodes,
	consensusSize,
	numInvalid uint32,
	roundTime uint64,
	consensusType string,
) map[uint32][]*integrationTests.TestConsensusNode {

	fmt.Println("Step 1. Setup nodes...")

	nodes := integrationTests.CreateNodesWithTestConsensusNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		consensusType,
		1,
	)

	for shardID, nodesList := range nodes {
		displayAndStartNodes(shardID, nodesList)
	}

	time.Sleep(p2pBootstrapDelay)

	for shardID := range nodes {
		if numInvalid < numNodes {
			for i := uint32(0); i < numInvalid; i++ {
				ii := numNodes - i - 1
				nodes[shardID][ii].MultiSigner.CreateSignatureShareCalled = func(privateKeyBytes, message []byte) ([]byte, error) {
					var invalidSigShare []byte
					if i%2 == 0 {
						// invalid sig share but with valid format
						invalidSigShare, _ = hex.DecodeString("2ee350b9a821e20df97ba487a80b0d0ffffca7da663185cf6a562edc7c2c71e3ca46ed71b31bccaf53c626b87f2b6e08")
					} else {
						// sig share with invalid size
						invalidSigShare = bytes.Repeat([]byte("a"), 3)
					}
					log.Warn("invalid sig share from ", "pk", getPkEncoded(nodes[shardID][ii].NodeKeys.Pk), "sig", invalidSigShare)

					return invalidSigShare, nil
				}
			}
		}
	}

	return nodes
}

func TestConsensusWithInvalidSigners(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(1)
	roundTime := uint64(1000)
	numCommBlock := uint64(8)

	nodes := initNodesWithTestSigner(numMetaNodes, numNodes, consensusSize, numInvalid, roundTime, blsConsensusType)

	defer func() {
		for shardID := range nodes {
			for _, n := range nodes[shardID] {
				_ = n.Messenger.Close()
			}
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second)

	for shardID := range nodes {
		mutex := &sync.Mutex{}
		nonceForRoundMap := make(map[uint64]uint64)
		totalCalled := 0

		err := startNodesWithCommitBlock(nodes[shardID], mutex, nonceForRoundMap, &totalCalled)
		assert.Nil(t, err)

		chDone := make(chan bool)
		go checkBlockProposedEveryRound(numCommBlock, nonceForRoundMap, mutex, chDone, t)

		extraTime := uint64(2)
		endTime := time.Duration(roundTime)*time.Duration(numCommBlock+extraTime)*time.Millisecond + time.Minute
		select {
		case <-chDone:
		case <-time.After(endTime):
			mutex.Lock()
			log.Error("currently saved nonces for rounds", "nonceForRoundMap", nonceForRoundMap)
			assert.Fail(t, "consensus too slow, not working.")
			mutex.Unlock()
			return
		}
	}
}
