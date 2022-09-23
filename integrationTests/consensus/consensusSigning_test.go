package consensus

import (
	"fmt"
	"sync"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/assert"
)

func initNodesWithTestSigner(
	numNodes,
	consensusSize,
	numInvalid uint32,
	roundTime uint64,
	consensusType string,
) []*integrationTests.TestConsensusNode {

	fmt.Println("Step 1. Setup nodes...")

	nodes := integrationTests.CreateNodesWithTestConsensusNode(
		1,
		int(numNodes),
		int(consensusSize),
		roundTime,
		consensusType,
	)

	for _, nodesList := range nodes {
		displayAndStartNodes(nodesList)
	}

	time.Sleep(p2pBootstrapDelay)

	if numInvalid < numNodes {
		for i := uint32(0); i < numInvalid; i++ {
			iCopy := i
			nodes[0][i].MultiSigner = &cryptoMocks.MultiSignerStub{
				CreateSignatureShareCalled: func(privateKeyBytes, message []byte) ([]byte, error) {
					fmt.Println("invalid sig share from ",
						getPkEncoded(nodes[0][iCopy].NodeKeys.Pk),
					)
					return []byte("invalid sig share"), nil
				},
			}
		}
	}

	return nodes[0]
}

func TestConsensusWithInvalidSigners(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:INFO,*:DEBUG")

	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(1)
	roundTime := uint64(1000)
	numCommBlock := uint64(8)

	nodes := initNodesWithTestSigner(numNodes, consensusSize, numInvalid, roundTime, blsConsensusType)

	mutex := &sync.Mutex{}
	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second)

	nonceForRoundMap := make(map[uint64]uint64)
	totalCalled := 0
	err := startNodesWithCommitBlock(nodes, mutex, nonceForRoundMap, &totalCalled)
	assert.Nil(t, err)

	chDone := make(chan bool)
	go checkBlockProposedEveryRound(numCommBlock, nonceForRoundMap, mutex, chDone, t)

	extraTime := uint64(2)
	endTime := time.Duration(roundTime) * time.Duration(numCommBlock+extraTime) * time.Millisecond
	select {
	case <-chDone:
	case <-time.After(endTime):
		mutex.Lock()
		fmt.Println("currently saved nonces for rounds: \n", nonceForRoundMap)
		assert.Fail(t, "consensus too slow, not working.")
		mutex.Unlock()
		return
	}
}
