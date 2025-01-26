package consensus

import (
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
)

const (
	consensusTimeBetweenRounds = time.Second
	blsConsensusType           = "bls"
)

var (
	p2pBootstrapDelay      = time.Second * 5
	testPubkeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(32)
	log                    = logger.GetOrCreate("integrationtests/consensus")
)

func TestConsensusBLSFullTestSingleKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runFullConsensusTest(t, blsConsensusType, 1, false)
}

func TestConsensusBLSFullTestMultiKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runFullConsensusTest(t, blsConsensusType, 5, false)
}

func TestConsensusBLSFullTestSingleKeys_WithEquivalentProofs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	logger.ToggleLoggerName(true)
	logger.SetLogLevel("*:DEBUG,consensus:TRACE")

	runFullConsensusTest(t, blsConsensusType, 1, true)
}

func TestConsensusBLSNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runConsensusWithNotEnoughValidators(t, blsConsensusType)
}

func TestConsensusBLSWithProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:DEBUG,process:TRACE,consensus:TRACE")
	logger.ToggleLoggerName(true)

	numKeysOnEachNode := 1
	numMetaNodes := uint32(2)
	numNodes := uint32(2)
	consensusSize := uint32(2 * numKeysOnEachNode)
	roundTime := uint64(1000)

	log.Info("runFullConsensusTest",
		"numNodes", numNodes,
		"numKeysOnEachNode", numKeysOnEachNode,
		"consensusSize", consensusSize,
	)

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()

	equivalentProodsActivationEpoch := uint32(0)

	enableEpochsConfig.EquivalentMessagesEnableEpoch = equivalentProodsActivationEpoch
	enableEpochsConfig.FixedOrderInConsensusEnableEpoch = equivalentProodsActivationEpoch

	fmt.Println("Step 1. Setup nodes...")

	nodes := integrationTests.CreateNodesWithTestConsensusNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		blsConsensusType,
		numKeysOnEachNode,
		enableEpochsConfig,
	)

	// leaders := []*integrationTests.TestConsensusNode{}
	for shardID, nodesList := range nodes {
		// leaders = append(leaders, nodesList[0])

		displayAndStartNodes(shardID, nodesList)
	}

	time.Sleep(p2pBootstrapDelay)

	// round := uint64(0)
	// nonce := uint64(0)
	// round = integrationTests.IncrementAndPrintRound(round)
	// integrationTests.UpdateRound(nodes, round)
	// nonce++

	// numRoundsToTest := 5
	// for i := 0; i < numRoundsToTest; i++ {
	// 	integrationTests.ProposeBlock(nodes, leaders, round, nonce)

	// 	time.Sleep(integrationTests.SyncDelay)

	// 	round = integrationTests.IncrementAndPrintRound(round)
	// 	integrationTests.UpdateRound(nodes, round)
	// 	nonce++
	// }

	for _, nodesList := range nodes {
		for _, n := range nodesList {
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
			require.Nil(t, err)

			managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusFactory)
			require.Nil(t, err)

			err = managedConsensusComponents.Create()
			require.Nil(t, err)
		}
	}

	time.Sleep(100 * time.Second)

	fmt.Println("Checking shards...")

	for _, nodesList := range nodes {
		// expectedNonce := nodesList[0].Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetNonce()
		expectedNonce := 1
		for _, n := range nodesList {
			for i := 1; i < len(nodes); i++ {
				if check.IfNil(n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader()) {
					assert.Fail(t, fmt.Sprintf("Node with idx %d does not have a current block", i))
				} else {
					assert.Equal(t, expectedNonce, n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetNonce())
				}
			}
		}
	}
}

func TestConsensusBLSWithFullProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	_ = logger.SetLogLevel("*:DEBUG,process:TRACE,consensus:TRACE")
	logger.ToggleLoggerName(true)

	numKeysOnEachNode := 1
	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(4 * numKeysOnEachNode)
	roundTime := uint64(1000)

	// maxShards := uint32(1)
	// shardId := uint32(0)
	// numNodesPerShard := 3

	log.Info("runFullNodesTest",
		"numNodes", numNodes,
		"numKeysOnEachNode", numKeysOnEachNode,
		"consensusSize", consensusSize,
	)

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()

	equivalentProodsActivationEpoch := uint32(10)

	enableEpochsConfig.EquivalentMessagesEnableEpoch = equivalentProodsActivationEpoch
	enableEpochsConfig.FixedOrderInConsensusEnableEpoch = equivalentProodsActivationEpoch

	fmt.Println("Step 1. Setup nodes...")

	nodes := integrationTests.CreateNodesWithTestFullNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		blsConsensusType,
		numKeysOnEachNode,
		enableEpochsConfig,
	)

	for shardID, nodesList := range nodes {
		for _, n := range nodesList {
			skBuff, _ := n.NodeKeys.MainKey.Sk.ToByteArray()
			pkBuff, _ := n.NodeKeys.MainKey.Pk.ToByteArray()

			encodedNodePkBuff := testPubkeyConverter.SilentEncode(pkBuff, log)

			fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
				shardID,
				hex.EncodeToString(skBuff),
				encodedNodePkBuff,
			)
		}
	}

	time.Sleep(p2pBootstrapDelay)

	defer func() {
		for _, nodesList := range nodes {
			for _, n := range nodesList {
				n.Close()
			}
		}
	}()

	for _, nodesList := range nodes {
		for _, n := range nodesList {
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
			require.Nil(t, err)

			managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusFactory)
			require.Nil(t, err)

			err = managedConsensusComponents.Create()
			require.Nil(t, err)
		}
	}

	time.Sleep(10 * time.Second)

	fmt.Println("Checking shards...")

	for _, nodesList := range nodes {
		expectedNonce := uint64(0)
		if !check.IfNil(nodesList[0].Node.GetDataComponents().Blockchain().GetCurrentBlockHeader()) {
			expectedNonce = nodesList[0].Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetNonce()
		}
		for _, n := range nodesList {
			for i := 1; i < len(nodes); i++ {
				if check.IfNil(n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader()) {
					// assert.Fail(t, fmt.Sprintf("Node with idx %d does not have a current block", i))
				} else {
					fmt.Println("FOUND")
					assert.Equal(t, expectedNonce, n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetNonce())
				}
			}
		}
	}
}

func initNodesAndTest(
	numMetaNodes,
	numNodes,
	consensusSize,
	numInvalid uint32,
	roundTime uint64,
	consensusType string,
	numKeysOnEachNode int,
	enableEpochsConfig config.EnableEpochs,
) map[uint32][]*integrationTests.TestConsensusNode {

	fmt.Println("Step 1. Setup nodes...")

	nodes := integrationTests.CreateNodesWithTestConsensusNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		consensusType,
		numKeysOnEachNode,
		enableEpochsConfig,
	)

	for shardID, nodesList := range nodes {
		displayAndStartNodes(shardID, nodesList)
	}

	time.Sleep(p2pBootstrapDelay)

	for shardID := range nodes {
		if numInvalid < numNodes {
			for i := uint32(0); i < numInvalid; i++ {
				iCopy := i
				nodes[shardID][i].BlockProcessor.ProcessBlockCalled = func(
					header data.HeaderHandler,
					body data.BodyHandler,
					haveTime func() time.Duration,
				) error {
					fmt.Println(
						"process block invalid ",
						header.GetRound(),
						header.GetNonce(),
						getPkEncoded(nodes[shardID][iCopy].NodeKeys.Pk),
					)
					return process.ErrBlockHashDoesNotMatch
				}
				nodes[shardID][i].BlockProcessor.CreateBlockCalled = func(
					header data.HeaderHandler,
					haveTime func() bool,
				) (data.HeaderHandler, data.BodyHandler, error) {
					return nil, nil, process.ErrWrongTypeAssertion
				}
			}
		}
	}

	return nodes
}

func startNodesWithCommitBlock(nodes []*integrationTests.TestConsensusNode, mutex *sync.Mutex, nonceForRoundMap map[uint64]uint64, totalCalled *int) error {
	for _, n := range nodes {
		nCopy := n
		n.BlockProcessor.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
			nCopy.BlockProcessor.NumCommitBlockCalled++
			headerHash, _ := core.CalculateHash(
				n.Node.GetCoreComponents().InternalMarshalizer(),
				n.Node.GetCoreComponents().Hasher(),
				header,
			)
			nCopy.ChainHandler.SetCurrentBlockHeaderHash(headerHash)
			_ = nCopy.ChainHandler.SetCurrentBlockHeaderAndRootHash(header, header.GetRootHash())

			log.Info("BlockProcessor.CommitBlockCalled", "shard", header.GetShardID(), "nonce", header.GetNonce(), "round", header.GetRound())

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
						log.Error("currently saved nonces for rounds", "nonceForRoundMap", nonceForRoundMap)
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

func runFullConsensusTest(
	t *testing.T,
	consensusType string,
	numKeysOnEachNode int,
	withEquivalentProofs bool,
) {
	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(3 * numKeysOnEachNode)
	numInvalid := uint32(0)
	roundTime := uint64(1000)
	numCommBlock := uint64(8)

	log.Info("runFullConsensusTest",
		"numNodes", numNodes,
		"numKeysOnEachNode", numKeysOnEachNode,
		"consensusSize", consensusSize,
	)

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()

	equivalentProodsActivationEpoch := integrationTests.UnreachableEpoch
	if withEquivalentProofs {
		equivalentProodsActivationEpoch = 0
	}

	enableEpochsConfig.EquivalentMessagesEnableEpoch = equivalentProodsActivationEpoch
	enableEpochsConfig.FixedOrderInConsensusEnableEpoch = equivalentProodsActivationEpoch

	nodes := initNodesAndTest(
		numMetaNodes,
		numNodes,
		consensusSize,
		numInvalid,
		roundTime,
		consensusType,
		numKeysOnEachNode,
		enableEpochsConfig,
	)

	defer func() {
		for shardID := range nodes {
			for _, n := range nodes[shardID] {
				_ = n.MainMessenger.Close()
				_ = n.FullArchiveMessenger.Close()
			}
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second * 2)

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
			log.Info("consensus done", "shard", shardID)
		case <-time.After(endTime):
			mutex.Lock()
			fmt.Println("currently saved nonces for rounds: \n", nonceForRoundMap)
			assert.Fail(t, "consensus too slow, not working.")
			mutex.Unlock()
			return
		}
	}
}

func runConsensusWithNotEnoughValidators(t *testing.T, consensusType string) {
	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(4)
	numInvalid := uint32(2)
	roundTime := uint64(1000)
	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.EquivalentMessagesEnableEpoch = integrationTests.UnreachableEpoch
	enableEpochsConfig.FixedOrderInConsensusEnableEpoch = integrationTests.UnreachableEpoch
	nodes := initNodesAndTest(numMetaNodes, numNodes, consensusSize, numInvalid, roundTime, consensusType, 1, enableEpochsConfig)

	defer func() {
		for shardID := range nodes {
			for _, n := range nodes[shardID] {
				_ = n.MainMessenger.Close()
				_ = n.FullArchiveMessenger.Close()
			}
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Start consensus...")
	time.Sleep(time.Second)

	for shardID := range nodes {
		totalCalled := 0
		mutex := &sync.Mutex{}
		nonceForRoundMap := make(map[uint64]uint64)

		err := startNodesWithCommitBlock(nodes[shardID], mutex, nonceForRoundMap, &totalCalled)
		assert.Nil(t, err)

		waitTime := time.Second * 30
		fmt.Println("Run for 30 seconds...")
		time.Sleep(waitTime)

		mutex.Lock()
		assert.Equal(t, 0, totalCalled)
		mutex.Unlock()
	}
}

func displayAndStartNodes(shardID uint32, nodes []*integrationTests.TestConsensusNode) {
	for _, n := range nodes {
		skBuff, _ := n.NodeKeys.Sk.ToByteArray()
		pkBuff, _ := n.NodeKeys.Pk.ToByteArray()

		encodedNodePkBuff := testPubkeyConverter.SilentEncode(pkBuff, log)

		fmt.Printf("Shard ID: %v, sk: %s, pk: %s\n",
			shardID,
			hex.EncodeToString(skBuff),
			encodedNodePkBuff,
		)
	}
}

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
