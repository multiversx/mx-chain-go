package consensus

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/config"
	consensusComp "github.com/multiversx/mx-chain-go/factory/consensus"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	vmFactory "github.com/multiversx/mx-chain-go/process/factory"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
)

const (
	consensusTimeBetweenRounds = time.Second
	blsConsensusType           = "bls"
	roundsPerEpoch             = 10
	roundTime                  = uint64(1000) // 1 second
)

var (
	p2pBootstrapDelay      = time.Second * 5
	testPubkeyConverter, _ = pubkeyConverter.NewHexPubkeyConverter(32)
	log                    = logger.GetOrCreate("integrationtests/consensus")
)

type generatedTxsParams struct {
	numScTxs          int
	numMoveBalanceTxs int
}

func TestConsensusBLSFullTestSingleKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runFullConsensusTest(t, blsConsensusType, 1)
}

func TestConsensusBLSFullTestMultiKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runFullConsensusTest(t, blsConsensusType, 5)
}

func TestConsensusBLSNotEnoughValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	runConsensusWithNotEnoughValidators(t, blsConsensusType)
}

func TestConsensusBLSWithFullProcessing_BeforeEquivalentProofs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = integrationTests.UnreachableEpoch
	enableEpochsConfig.SupernovaEnableEpoch = integrationTests.UnreachableEpoch
	numKeysOnEachNode := 1
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          100,
		numMoveBalanceTxs: 5000,
	}
	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func TestConsensusBLSWithFullProcessing_WithEquivalentProofs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = uint32(0)
	enableEpochsConfig.SupernovaEnableEpoch = integrationTests.UnreachableEpoch
	numKeysOnEachNode := 1
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          100,
		numMoveBalanceTxs: 5000,
	}

	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func TestConsensusBLSWithFullProcessing_TransitionWithEquivalentProofs(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = uint32(1)
	enableEpochsConfig.SupernovaEnableEpoch = integrationTests.UnreachableEpoch
	numKeysOnEachNode := 1
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          100,
		numMoveBalanceTxs: 5000,
	}

	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func TestConsensusBLSWithFullProcessing_WithEquivalentProofs_MultiKeys(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = uint32(0)
	enableEpochsConfig.SupernovaEnableEpoch = integrationTests.UnreachableEpoch
	numKeysOnEachNode := 3
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          100,
		numMoveBalanceTxs: 5000,
	}

	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func TestConsensusBLSWithFullProcessing_TransitionToSupernova(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = uint32(0)
	enableEpochsConfig.SupernovaEnableEpoch = uint32(1)
	numKeysOnEachNode := 3
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          0,
		numMoveBalanceTxs: 0,
	}

	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func TestConsensusBLSWithFullProcessing_AfterSupernova(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = uint32(0)
	enableEpochsConfig.SupernovaEnableEpoch = uint32(0)
	numKeysOnEachNode := 3
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          0,
		numMoveBalanceTxs: 0,
	}

	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func TestConsensusBLSWithFullProcessing_TransitionToSupernova_HighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = uint32(0)
	enableEpochsConfig.SupernovaEnableEpoch = uint32(1)
	numKeysOnEachNode := 3
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          500,
		numMoveBalanceTxs: 10000,
	}

	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func TestConsensusBLSWithFullProcessing_AfterSupernova_HighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = uint32(0)
	enableEpochsConfig.SupernovaEnableEpoch = uint32(0)
	numKeysOnEachNode := 3
	targetEpoch := uint32(2)
	txs := &generatedTxsParams{
		numScTxs:          500,
		numMoveBalanceTxs: 10000,
	}

	testConsensusBLSWithFullProcessing(t, enableEpochsConfig, numKeysOnEachNode, roundsPerEpoch, roundTime, targetEpoch, txs)
}

func testConsensusBLSWithFullProcessing(
	t *testing.T,
	enableEpochsConfig config.EnableEpochs,
	numKeysOnEachNode int,
	roundsPerEpoch int64,
	roundTime uint64,
	targetEpoch uint32,
	txs *generatedTxsParams,
) {
	numMetaNodes := uint32(2)
	numNodes := uint32(2)
	consensusSize := uint32(2 * numKeysOnEachNode)

	log.Info("runFullNodesTest",
		"numNodes", numNodes,
		"numKeysOnEachNode", numKeysOnEachNode,
		"consensusSize", consensusSize,
	)

	fmt.Println("Step 1. Setup nodes...")

	nodes := integrationTests.CreateNodesWithTestFullNode(
		int(numMetaNodes),
		int(numNodes),
		int(consensusSize),
		roundTime,
		blsConsensusType,
		numKeysOnEachNode,
		enableEpochsConfig,
		integrationTests.GetSupernovaRoundsConfig(),
		true,
		roundsPerEpoch,
	)
	shard0Node := nodes[0][0]

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
			err := startFullConsensusNode(n)
			require.Nil(t, err)
		}
	}

	fmt.Println("Wait for several rounds...")

	nodesList := make([]*integrationTests.TestProcessorNode, 0)
	shard0Nodes := nodes[0]
	for _, n := range shard0Nodes {
		nodesList = append(nodesList, n.TestProcessorNode)
	}
	integrationTests.MintAllNodes(nodesList, big.NewInt(100000000000))

	waitForEpoch(shard0Node, enableEpochsConfig.SCDeployEnableEpoch)

	scTxs(t, shard0Node, txs.numScTxs, nodesList)

	encodedReceiverAddr, err := integrationTests.TestAddressPubkeyConverter.Encode(integrationTests.CreateRandomBytes(32))
	assert.Nil(t, err)
	moveBalanceTxs(t, shard0Node.TestProcessorNode, encodedReceiverAddr, txs.numMoveBalanceTxs)

	waitForEpoch(shard0Node, targetEpoch)

	fmt.Println("Checking shards...")

	receiverBytes, err := integrationTests.TestAddressPubkeyConverter.Decode(encodedReceiverAddr)
	require.Nil(t, err)
	acc, err := shard0Node.AccntState.LoadAccount(receiverBytes)
	require.Nil(t, err)

	if txs.numMoveBalanceTxs > 0 {
		require.NotEqual(t, uint64(0), acc.(data.UserAccountHandler).GetBalance().Uint64())
	}

	expectedNonce := uint64(10)
	for _, nodesList := range nodes {
		for _, n := range nodesList {
			for i := 1; i < len(nodes); i++ {
				if check.IfNil(n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader()) {
					assert.Fail(t, fmt.Sprintf("Node with idx %d does not have a current block", i))
				} else {
					assert.GreaterOrEqual(t, n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetNonce(), expectedNonce)
					assert.Equal(t, targetEpoch, n.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader().GetEpoch())
				}
			}
		}
	}
}

func startFullConsensusNode(
	n *integrationTests.TestFullNode,
) error {
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
		return err
	}

	managedConsensusComponents, err := consensusComp.NewManagedConsensusComponents(consensusFactory)
	if err != nil {
		return err
	}

	return managedConsensusComponents.Create()
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
) {
	numMetaNodes := uint32(4)
	numNodes := uint32(4)
	consensusSize := uint32(3 * numKeysOnEachNode)
	numInvalid := uint32(0)
	numCommBlock := uint64(8)

	log.Info("runFullConsensusTest",
		"numNodes", numNodes,
		"numKeysOnEachNode", numKeysOnEachNode,
		"consensusSize", consensusSize,
	)

	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()

	equivalentProofsActivationEpoch := integrationTests.UnreachableEpoch
	enableEpochsConfig.AndromedaEnableEpoch = equivalentProofsActivationEpoch

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
	enableEpochsConfig := integrationTests.CreateEnableEpochsConfig()
	enableEpochsConfig.AndromedaEnableEpoch = integrationTests.UnreachableEpoch
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

func waitForEpoch(node *integrationTests.TestFullNode, targetEpoch uint32) {
	epochReached := false
	for !epochReached {
		blockHeader := node.Node.GetDataComponents().Blockchain().GetCurrentBlockHeader()
		if check.IfNil(blockHeader) {
			time.Sleep(time.Second)
			continue
		}
		epochReached = blockHeader.GetEpoch() == targetEpoch
	}

	time.Sleep(time.Second * 3) // wait for all nodes to change epoch
	fmt.Println("Wait for all nodes to change epoch...")
}

func scTxs(t *testing.T, senderNode *integrationTests.TestFullNode, numTxs int, nodesList []*integrationTests.TestProcessorNode) {
	if numTxs <= 0 {
		return
	}

	numPlayers := 10
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		players[i] = integrationTests.CreateTestWalletAccount(senderNode.ShardCoordinator, 0)
	}
	initialVal := big.NewInt(100000000000)
	integrationTests.MintAllPlayers(nodesList, players, initialVal)

	scCode, err := os.ReadFile("../vm/wasm/testdata/erc20-c-03/wrc20_wasm.wasm")
	require.Nil(t, err)

	scAddress, _ := senderNode.TestProcessorNode.BlockchainHook.NewAddress(senderNode.OwnAccount.Address, senderNode.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)
	initialSupply := hex.EncodeToString(big.NewInt(100000000000).Bytes())
	integrationTests.DeployScTx(nodesList, 0, hex.EncodeToString(scCode), vmFactory.WasmVirtualMachine, initialSupply)
	time.Sleep(time.Second)

	for i := 0; i < numTxs; i++ {
		playersDoTransfer(senderNode.TestProcessorNode, players, scAddress, big.NewInt(100))
	}
}

func moveBalanceTxs(t *testing.T, senderNode *integrationTests.TestProcessorNode, receiverAddr string, numTxs int) {
	if numTxs <= 0 {
		return
	}

	err := senderNode.Node.GenerateAndSendBulkTransactions(
		receiverAddr,
		big.NewInt(1),
		uint64(numTxs),
		senderNode.OwnAccount.SkTxSign,
		nil,
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)
	assert.Nil(t, err)
}

func playersDoTransfer(
	node *integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount,
	scAddress []byte,
	txValue *big.Int,
) {
	for _, playerToTransfer := range players {
		createAndSendTx(node, node.OwnAccount, big.NewInt(0), 200000, scAddress,
			[]byte("transferToken@"+hex.EncodeToString(playerToTransfer.Address)+"@"+hex.EncodeToString(txValue.Bytes())))
	}
}

func createAndSendTx(
	node *integrationTests.TestProcessorNode,
	player *integrationTests.TestWalletAccount,
	txValue *big.Int,
	gasLimit uint64,
	rcvAddress []byte,
	txData []byte,
) {
	tx := &transaction.Transaction{
		Nonce:    player.Nonce,
		Value:    txValue,
		SndAddr:  player.Address,
		RcvAddr:  rcvAddress,
		Data:     txData,
		GasPrice: node.EconomicsData.GetMinGasPrice(),
		GasLimit: gasLimit,
		Version:  integrationTests.MinTransactionVersion,
		ChainID:  integrationTests.ChainID,
	}

	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
	tx.Signature, _ = player.SingleSigner.Sign(player.SkTxSign, txBuff)

	_, _ = node.SendTransaction(tx)
	player.Nonce++
}
