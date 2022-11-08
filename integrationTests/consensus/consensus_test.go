package consensus

import (
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/pubkeyConverter"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/config"
	consensusComp "github.com/ElrondNetwork/elrond-go/factory/consensus"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	consensusMocks "github.com/ElrondNetwork/elrond-go/testscommon/consensus"
	"github.com/stretchr/testify/assert"
)

const (
	consensusTimeBetweenRounds = time.Second
	blsConsensusType           = "bls"
)

var (
	p2pBootstrapDelay      = time.Second * 5
	testPubkeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(32)
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
			nodes[0][i].BlockProcessor.ProcessBlockCalled = func(
				header data.HeaderHandler,
				body data.BodyHandler,
				haveTime func() time.Duration,
			) error {
				fmt.Println(
					"process block invalid ",
					header.GetRound(),
					header.GetNonce(),
					getPkEncoded(nodes[0][iCopy].NodeKeys.Pk),
				)
				return process.ErrBlockHashDoesNotMatch
			}
			nodes[0][i].BlockProcessor.CreateBlockCalled = func(
				header data.HeaderHandler,
				haveTime func() bool,
			) (data.HeaderHandler, data.BodyHandler, error) {
				return nil, nil, process.ErrWrongTypeAssertion
			}
		}
	}

	return nodes[0]
}

func startNodesWithCommitBlock(nodes []*integrationTests.TestConsensusNode, mutex *sync.Mutex, nonceForRoundMap map[uint64]uint64, totalCalled *int) error {
	for _, n := range nodes {
		nCopy := n
		n.BlockProcessor.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
			nCopy.BlockProcessor.NrCommitBlockCalled++
			_ = nCopy.ChainHandler.SetCurrentBlockHeaderAndRootHash(header, header.GetRootHash())

			mutex.Lock()
			nonceForRoundMap[header.GetRound()] = header.GetNonce()
			*totalCalled += 1
			mutex.Unlock()

			return nil
		}

		statusComponents := integrationTests.GetDefaultStatusComponents()

		consensusArgs := consensusComp.ConsensusComponentsFactoryArgs{
			Config: config.Config{
				Consensus: config.ConsensusConfig{
					Type: blsConsensusType,
				},
				ValidatorPubkeyConverter: config.PubkeyConfig{
					Length:          96,
					Type:            "bls",
					SignatureLength: 48,
				},
				TrieSync: config.TrieSyncConfig{
					NumConcurrentTrieSyncers:  5,
					MaxHardCapForMissingNodes: 5,
					TrieSyncerVersion:         2,
					CheckNodesOnDisk:          false,
				},
				GeneralSettings: config.GeneralSettingsConfig{
					SyncProcessTimeInMillis: 6000,
				},
			},
			BootstrapRoundIndex:  0,
			CoreComponents:       n.Node.GetCoreComponents(),
			NetworkComponents:    n.Node.GetNetworkComponents(),
			CryptoComponents:     n.Node.GetCryptoComponents(),
			DataComponents:       n.Node.GetDataComponents(),
			ProcessComponents:    n.Node.GetProcessComponents(),
			StateComponents:      n.Node.GetStateComponents(),
			StatusComponents:     statusComponents,
			StatusCoreComponents: n.Node.GetStatusCoreComponents(),
			ScheduledProcessor:   &consensusMocks.ScheduledProcessorStub{},
			IsInImportMode:       n.Node.IsInImportMode(),
		}

		consensusFactory, err := consensusComp.NewConsensusComponentsFactory(consensusArgs)
		if err != nil {
			return fmt.Errorf("NewConsensusComponentsFactory failed: %w", err)
		}

		managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusFactory)
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

		time.Sleep(consensusTimeBetweenRounds)
	}
}

func runFullConsensusTest(t *testing.T, consensusType string) {
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(0)
	roundTime := uint64(1000)
	numCommBlock := uint64(8)

	nodes := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, consensusType)

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
	roundTime := uint64(1000)
	nodes := initNodesAndTest(numNodes, consensusSize, numInvalid, roundTime, consensusType)

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

	waitTime := time.Second * 30
	fmt.Println("Run for 30 seconds...")
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

func displayAndStartNodes(nodes []*integrationTests.TestConsensusNode) {
	for _, n := range nodes {
		skBuff, _ := n.NodeKeys.Sk.ToByteArray()
		pkBuff, _ := n.NodeKeys.Pk.ToByteArray()

		fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
			n.ShardCoordinator.SelfId(),
			hex.EncodeToString(skBuff),
			testPubkeyConverter.Encode(pkBuff),
		)
	}
}
