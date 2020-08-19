package consensus

import (
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
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

func initNodesAndTest(
	numNodes,
	consensusSize,
	numInvalid uint32,
	roundTime uint64,
	consensusType string,
) ([]*testNode, p2p.Messenger, *sync.Map) {

	fmt.Println("Step 1. Setup nodes...")

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	concMap := &sync.Map{}

	nodes := createNodes(
		int(numNodes),
		int(consensusSize),
		roundTime,
		getConnectableAddress(advertiser),
		consensusType,
	)

	for _, nodesList := range nodes {
		displayAndStartNodes(nodesList)
	}

	time.Sleep(p2pBootstrapDelay)

	if numInvalid < numNodes {
		for i := uint32(0); i < numInvalid; i++ {
			iCopy := i
			nodes[0][i].blkProcessor.ProcessBlockCalled = func(
				header data.HeaderHandler,
				body data.BodyHandler,
				haveTime func() time.Duration,
			) error {
				fmt.Println(
					"process block invalid ",
					header.GetRound(),
					header.GetNonce(),
					getPkEncoded(nodes[0][iCopy].pk),
				)
				return process.ErrBlockHashDoesNotMatch
			}
			nodes[0][i].blkProcessor.CreateBlockCalled = func(
				header data.HeaderHandler,
				haveTime func() bool,
			) (data.HeaderHandler, data.BodyHandler, error) {
				return nil, nil, process.ErrWrongTypeAssertion
			}
		}
	}

	return nodes[0], advertiser, concMap
}

func startNodesWithCommitBlock(nodes []*testNode, mutex *sync.Mutex, nonceForRoundMap map[uint64]uint64, totalCalled *int) error {
	for _, n := range nodes {
		nCopy := n
		n.blkProcessor.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
			nCopy.blkProcessor.NrCommitBlockCalled++
			_ = nCopy.blkc.SetCurrentBlockHeader(header)

			mutex.Lock()
			nonceForRoundMap[header.GetRound()] = header.GetNonce()
			*totalCalled += 1
			mutex.Unlock()

			return nil
		}

		statusComponents := integrationTests.GetDefaultStatusComponents()

		consensusArgs := factory.ConsensusComponentsFactoryArgs{
			Config: config.Config{
				Consensus: config.ConsensusConfig{
					Type: blsConsensusType,
				},
				ValidatorPubkeyConverter: config.PubkeyConfig{
					Length:          96,
					Type:            "bls",
					SignatureLength: 48,
				},
			},

			ConsensusGroupSize:  n.node.GetConsensusGroupSize(),
			BootstrapRoundIndex: 0,
			HardforkTrigger:     n.node.GetHardforkTrigger(),
			CoreComponents:      n.node.GetCoreComponents(),
			NetworkComponents:   n.node.GetNetworkComponents(),
			CryptoComponents:    n.node.GetCryptoComponents(),
			DataComponents:      n.node.GetDataComponents(),
			ProcessComponents:   n.node.GetProcessComponents(),
			StateComponents:     n.node.GetStateComponents(),
			StatusComponents:    statusComponents,
		}

		managedConsensusComponents, err := factory.NewManagedConsensusComponents(consensusArgs)
		if err != nil {
			return err
		}

		err = managedConsensusComponents.Create()
		if err != nil {
			return err
		}
	}
	return nil
}

func checkBlockProposedEveryRound(numCommBlock uint64, nonceForRoundMap map[uint64]uint64, mutex *sync.Mutex, chDone chan bool, t *testing.T) {
	for {
		mutex.Lock()

		minRound := ^uint64(0)
		maxRound := uint64(0)
		if uint64(len(nonceForRoundMap)) >= numCommBlock {
			for k := range nonceForRoundMap {
				if k > maxRound {
					maxRound = k
				}
				if k < minRound {
					minRound = k
				}
			}

			if maxRound-minRound >= numCommBlock {
				for i := minRound; i <= maxRound; i++ {
					if _, ok := nonceForRoundMap[i]; !ok {
						assert.Fail(t, "consensus not reached in each round")
						fmt.Println("currently saved nonces for rounds: \n", nonceForRoundMap)
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

func runFullConsensusTest(t *testing.T, consensusType string) {
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(0)
	roundTime := uint64(5000)
	numCommBlock := uint64(8)

	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, consensusType)

	mutex := &sync.Mutex{}
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.mesenger.Close()
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

func TestConsensusBLSFullTest(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runFullConsensusTest(t, blsConsensusType)
}

func runConsensusWithNotEnoughValidators(t *testing.T, consensusType string) {
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(2)
	roundTime := uint64(4000)
	nodes, advertiser, _ := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, consensusType)

	mutex := &sync.Mutex{}
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.mesenger.Close()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second)

	nonceForRoundMap := make(map[uint64]uint64)
	totalCalled := 0
	err := startNodesWithCommitBlock(nodes, mutex, nonceForRoundMap, &totalCalled)
	assert.Nil(t, err)

	waitTime := time.Second * 60
	fmt.Println("Run for 60 seconds...")
	time.Sleep(waitTime)

	mutex.Lock()
	assert.Equal(t, 0, totalCalled)
	mutex.Unlock()
}

func TestConsensusBLSNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runConsensusWithNotEnoughValidators(t, blsConsensusType)
}
