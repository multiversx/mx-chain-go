package block

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "../../agarioV3.hex"
var stepDelay = time.Second

// TestShouldProcessBlocksInMultiShardArchitectureWithScTxsTopUpAndWithdrawOnlyProposers tests the following scenario:
// There are 2 shards and 1 meta, each with only one node (proposer).
// Shard 1's proposer deploys a SC. There is 1 round for proposing block that will create the SC account.
// Shard 0's proposer sends a topUp SC call tx and then there are another 6 blocks added to all blockchains.
// After that there is a first check that the topUp was made. Shard 0's proposer sends a withdraw SC call tx and after
// 12 more blocks the results are checked again
func TestProcessWithScTxsTopUpAndWithdrawOnlyProposers(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(2)
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodeShard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	nodeShard0.EconomicsData.SetMinGasPrice(0)

	nodeShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	nodeShard1.EconomicsData.SetMinGasPrice(0)

	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	nodeShard1.LoadTxSignSkBytes(hardCodedSk)
	nodeMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)
	nodeMeta.EconomicsData.SetMinGasPrice(0)

	nodes := []*integrationTests.TestProcessorNode{nodeShard0, nodeShard1, nodeMeta}

	idxNodeShard0 := 0
	idxNodeShard1 := 1
	idxNodeMeta := 2
	idxProposers := []int{idxNodeShard0, idxNodeShard1, idxNodeMeta}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	hardCodedScResultingAddress, _ := nodeShard1.BlockchainHook.NewAddress(
		nodes[idxNodeShard1].OwnAccount.Address.Bytes(),
		nodes[idxNodeShard1].OwnAccount.Nonce,
		factory.IELEVirtualMachine,
	)
	integrationTests.DeployScTx(nodes, idxNodeShard1, string(scCode))

	integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	integrationTests.NodeDoesTopUp(nodes, idxNodeShard0, topUpValue, hardCodedScResultingAddress)

	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	nodeWithSc := nodes[idxNodeShard1]
	nodeWithCaller := nodes[idxNodeShard0]

	integrationTests.CheckScTopUp(t, nodeWithSc, topUpValue, hardCodedScResultingAddress)
	integrationTests.CheckScBalanceOf(t, nodeWithSc, nodeWithCaller, topUpValue, hardCodedScResultingAddress)
	integrationTests.CheckSenderBalanceOkAfterTopUp(t, nodeWithCaller, initialVal, topUpValue)

	integrationTests.NodeDoesWithdraw(nodes, idxNodeShard0, withdrawValue, hardCodedScResultingAddress)

	roundsToWait = 12
	for i := 0; i < roundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	expectedSC := integrationTests.CheckBalanceIsDoneCorrectlySCSideAndReturnExpectedVal(t, nodes, idxNodeShard1, topUpValue, withdrawValue, hardCodedScResultingAddress)
	integrationTests.CheckScBalanceOf(t, nodeWithSc, nodeWithCaller, expectedSC, hardCodedScResultingAddress)
	integrationTests.CheckSenderBalanceOkAfterTopUpAndWithdraw(t, nodeWithCaller, initialVal, topUpValue, withdrawValue)
}

// TestShouldProcessBlocksInMultiShardArchitectureWithScTxsJoinAndRewardProposersAndValidators tests the following scenario:
// There are 2 shards and 1 meta, each with one proposer and one validator.
// Shard 1's proposer deploys a SC. There is 1 round for proposing block that will create the SC account.
// Shard 0's proposer sends a joinGame SC call tx and then there are another 6 blocks added to all blockchains.
// After that there is a first check that the joinGame was made. Shard 1's proposer sends a rewardAndSendFunds SC call
// tx and after 6 more blocks the results are checked again
func TestProcessWithScTxsJoinAndRewardTwoNodesInShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(2)
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)
	nodeProposerShard0 := integrationTests.NewTestProcessorNode(
		maxShards,
		0,
		0,
		advertiserAddr,
	)
	nodeProposerShard0.EconomicsData.SetMinGasPrice(0)

	nodeValidatorShard0 := integrationTests.NewTestProcessorNode(
		maxShards,
		0,
		0,
		advertiserAddr,
	)
	nodeValidatorShard0.EconomicsData.SetMinGasPrice(0)

	nodeProposerShard1 := integrationTests.NewTestProcessorNode(
		maxShards,
		1,
		1,
		advertiserAddr,
	)
	nodeProposerShard1.EconomicsData.SetMinGasPrice(0)

	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000001006c560111a94e434413c1cdaafbc3e1348947d1d5b3a1")
	nodeProposerShard1.LoadTxSignSkBytes(hardCodedSk)
	nodeValidatorShard1 := integrationTests.NewTestProcessorNode(
		maxShards,
		1,
		1,
		advertiserAddr,
	)
	nodeValidatorShard1.EconomicsData.SetMinGasPrice(0)

	nodeProposerMeta := integrationTests.NewTestProcessorNode(
		maxShards,
		sharding.MetachainShardId,
		0,
		advertiserAddr,
	)
	nodeProposerMeta.EconomicsData.SetMinGasPrice(0)

	nodeValidatorMeta := integrationTests.NewTestProcessorNode(
		maxShards,
		sharding.MetachainShardId,
		0,
		advertiserAddr,
	)
	nodeValidatorMeta.EconomicsData.SetMinGasPrice(0)

	nodes := []*integrationTests.TestProcessorNode{
		nodeProposerShard0,
		nodeProposerShard1,
		nodeProposerMeta,
		nodeValidatorShard0,
		nodeValidatorShard1,
		nodeValidatorMeta,
	}

	idxProposerShard0 := 0
	idxProposerShard1 := 1
	idxProposerMeta := 2
	idxProposers := []int{idxProposerShard0, idxProposerShard1, idxProposerMeta}
	idxValidators := []int{3, 4, 5}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	integrationTests.DeployScTx(nodes, idxProposerShard1, string(scCode))

	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	integrationTests.PlayerJoinsGame(
		nodes,
		nodes[idxProposerShard0].OwnAccount,
		topUpValue,
		"aaaa",
		hardCodedScResultingAddress,
	)

	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		idxValidators, idxProposers = idxProposers, idxValidators
	}

	nodeWithSc := nodes[idxProposerShard1]
	nodeWithCaller := nodes[idxProposerShard0]

	integrationTests.CheckScTopUp(t, nodeWithSc, topUpValue, hardCodedScResultingAddress)
	integrationTests.CheckSenderBalanceOkAfterTopUp(t, nodeWithCaller, initialVal, topUpValue)

	integrationTests.NodeCallsRewardAndSend(
		nodes,
		idxProposerShard1,
		nodes[idxProposerShard0].OwnAccount,
		withdrawValue,
		"aaaa",
		hardCodedScResultingAddress,
	)

	//TODO investigate why do we need 7 rounds here
	roundsToWait = 7
	for i := 0; i < roundsToWait; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		idxValidators, idxProposers = idxProposers, idxValidators
	}

	_ = integrationTests.CheckBalanceIsDoneCorrectlySCSideAndReturnExpectedVal(t, nodes, idxProposerShard1, topUpValue, withdrawValue, hardCodedScResultingAddress)
	integrationTests.CheckSenderBalanceOkAfterTopUpAndWithdraw(t, nodeWithCaller, initialVal, topUpValue, withdrawValue)
	integrationTests.CheckRootHashes(t, nodes, idxProposers)
}

// TestShouldProcessWithScTxsJoinNoCommitShouldProcessedByValidators tests the following scenario:
// There are 2 shards and 1 meta, each with one proposer and one validator.
// Shard 1's proposer deploys a SC. There is 1 round for proposing block that will create the SC account.
// Shard 0's proposer sends a joinGame SC call tx, proposes a block (not committing it) and the validator
// should be able to sync it.
// Test will fail with any variant before commit d79898991f83188118a1c60003f5277bc71209e6
func TestShouldProcessWithScTxsJoinNoCommitShouldProcessedByValidators(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(2)
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodeProposerShard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	nodeProposerShard0.EconomicsData.SetMinGasPrice(0)
	nodeValidatorShard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	nodeValidatorShard0.EconomicsData.SetMinGasPrice(0)

	nodeProposerShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	nodeProposerShard1.EconomicsData.SetMinGasPrice(0)

	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000001006c560111a94e434413c1cdaafbc3e1348947d1d5b3a1")
	nodeProposerShard1.LoadTxSignSkBytes(hardCodedSk)
	nodeValidatorShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	nodeValidatorShard1.EconomicsData.SetMinGasPrice(0)

	nodeProposerMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)
	nodeProposerMeta.EconomicsData.SetMinGasPrice(0)
	nodeValidatorMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)
	nodeValidatorMeta.EconomicsData.SetMinGasPrice(0)

	nodes := []*integrationTests.TestProcessorNode{
		nodeProposerShard0,
		nodeProposerShard1,
		nodeProposerMeta,
		nodeValidatorShard0,
		nodeValidatorShard1,
		nodeValidatorMeta,
	}

	idxProposerShard0 := 0
	idxProposerShard1 := 1
	idxProposerMeta := 2
	idxProposers := []int{idxProposerShard0, idxProposerShard1, idxProposerMeta}
	idxProposersWithoutShard1 := []int{idxProposerShard0, idxProposerMeta}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	integrationTests.MintAllNodes(nodes, initialVal)

	integrationTests.DeployScTx(nodes, idxProposerShard1, string(scCode))
	round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)

	integrationTests.PlayerJoinsGame(
		nodes,
		nodes[idxProposerShard0].OwnAccount,
		topUpValue,
		"aaaa",
		hardCodedScResultingAddress,
	)

	maxRoundsToWait := 10
	for i := 0; i < maxRoundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposersWithoutShard1, round, nonce)

		hdr, body, isBodyEmpty := integrationTests.ProposeBlockSignalsEmptyBlock(nodes[idxProposerShard1], round, nonce)
		if isBodyEmpty {
			nodes[idxProposerShard1].CommitBlock(body, hdr)
			integrationTests.SyncBlock(t, nodes, idxProposers, round)
			round = integrationTests.IncrementAndPrintRound(round)
			nonce++
			continue
		}

		//shard 1' proposer got to process at least 1 tx, should not commit but the shard 1's validator
		//should be able to process the block

		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
		integrationTests.CheckRootHashes(t, nodes, idxProposers)
		break
	}

	nodeWithSc := nodes[idxProposerShard1]
	nodeWithCaller := nodes[idxProposerShard0]

	integrationTests.CheckScTopUp(t, nodeWithSc, topUpValue, hardCodedScResultingAddress)
	integrationTests.CheckSenderBalanceOkAfterTopUp(t, nodeWithCaller, initialVal, topUpValue)
}

// TestShouldSubtractTheCorrectTxFee uses the mock VM as it's gas model is predictable
// The test checks the tx fee subtraction from the sender account when deploying a SC
// It also checks the fee obtained by the leader is correct
// Test parameters: 2 shards + meta, each with 2 nodes
func TestShouldSubtractTheCorrectTxFee(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	logger.DefaultLogger().SetLevel("DEBUG")

	maxShards := 2
	consensusGroupSize := 2
	nodesPerShard := 2
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	// create map of shard - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinator(
		nodesPerShard,
		nodesPerShard,
		maxShards,
		consensusGroupSize,
		consensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
		integrationTests.SetEconomicsParameters(nodes, integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit)
		//set rewards = 0 so we can easily tests the balances taking into account only the tx fee
		for _, n := range nodes {
			n.EconomicsData.SetRewards(big.NewInt(0))
		}
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	initialVal := big.NewInt(10000000)
	senders := integrationTests.CreateSendersWithInitialBalances(nodesMap, initialVal)

	deployValue := big.NewInt(0)
	gasLimit := integrationTests.OpGasValueForMockVm + integrationTests.MinTxGasLimit
	gasPrice := integrationTests.MinTxGasPrice
	txNonce := uint64(0)
	owner := senders[0][0]
	nodeShard0 := nodesMap[0][0]
	ownerPk, _ := owner.GeneratePublic().ToByteArray()
	ownerAddr, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(ownerPk)
	integrationTests.ScCallTxWithParams(
		nodeShard0,
		owner,
		txNonce,
		"DEADBEEF@"+hex.EncodeToString(factory.InternalTestingVM)+"@00",
		deployValue,
		gasLimit,
		gasPrice,
	)
	txNonce++

	randomness := generateInitialRandomness(uint32(maxShards))
	_, _, _, randomness = integrationTests.AllShardsProposeBlock(round, nonce, randomness, nodesMap)
	leaderPkBytes := nodeShard0.SpecialAddressHandler.LeaderAddress()
	leaderAddress, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(leaderPkBytes)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// test sender account decreased its balance with gasPrice * gasLimit
	accnt, err := nodeShard0.AccntState.GetExistingAccount(ownerAddr)
	assert.Nil(t, err)
	ownerAccnt := accnt.(*state.Account)
	expectedBalance := big.NewInt(0).Set(initialVal)
	txCost := big.NewInt(0).SetUint64(gasPrice * (integrationTests.MinTxGasLimit + integrationTests.OpGasValueForMockVm))
	expectedBalance.Sub(expectedBalance, txCost)
	assert.Equal(t, expectedBalance, ownerAccnt.Balance)

	printContainingTxs(nodeShard0, nodeShard0.BlockChain.GetCurrentBlockHeader().(*block.Header))

	accnt, err = nodeShard0.AccntState.GetExistingAccount(leaderAddress)
	assert.Nil(t, err)
	leaderAccnt := accnt.(*state.Account)
	expectedBalance = big.NewInt(0).Set(txCost)
	expectedBalance.Div(expectedBalance, big.NewInt(2))
	assert.Equal(t, expectedBalance, leaderAccnt.Balance)
}

func printContainingTxs(tpn *integrationTests.TestProcessorNode, hdr *block.Header) {
	for _, miniblockHdr := range hdr.MiniBlockHeaders {
		miniblockBytes, err := tpn.Storage.Get(dataRetriever.MiniBlockUnit, miniblockHdr.Hash)
		if err != nil {
			fmt.Println("miniblock " + base64.StdEncoding.EncodeToString(miniblockHdr.Hash) + "not found")
			continue
		}

		miniblock := &block.MiniBlock{}
		err = integrationTests.TestMarshalizer.Unmarshal(miniblock, miniblockBytes)
		if err != nil {
			fmt.Println("can not unmarshal miniblock " + base64.StdEncoding.EncodeToString(miniblockHdr.Hash))
			continue
		}

		for _, txHash := range miniblock.TxHashes {
			txBytes := []byte("not found")

			mbType := miniblockHdr.Type
			switch mbType {
			case block.TxBlock:
				txBytes, err = tpn.Storage.Get(dataRetriever.TransactionUnit, txHash)
				if err != nil {
					fmt.Println("tx hash " + base64.StdEncoding.EncodeToString(txHash) + " not found")
					continue
				}
			case block.SmartContractResultBlock:
				txBytes, err = tpn.Storage.Get(dataRetriever.UnsignedTransactionUnit, txHash)
				if err != nil {
					fmt.Println("scr hash " + base64.StdEncoding.EncodeToString(txHash) + " not found")
					continue
				}
			case block.RewardsBlock:
				txBytes, err = tpn.Storage.Get(dataRetriever.RewardTransactionUnit, txHash)
				if err != nil {
					fmt.Println("reward hash " + base64.StdEncoding.EncodeToString(txHash) + " not found")
					continue
				}
			}

			fmt.Println(string(txBytes))
		}
	}
}
