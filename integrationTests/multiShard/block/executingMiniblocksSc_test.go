package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
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
	nodeShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodeShard1.LoadTxSignSkBytes(hardCodedSk)
	nodeMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)

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
	round = integrationTests.IncrementAndPrintRound(round)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	integrationTests.DeployScTx(nodes, idxNodeShard1, string(scCode))

	integrationTests.ProposeBlock(nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)

	integrationTests.NodeDoesTopUp(nodes, idxNodeShard0, topUpValue, hardCodedScResultingAddress)

	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
	}

	checkTopUpIsDoneCorrectly(
		t,
		nodes,
		idxNodeShard1,
		idxNodeShard0,
		initialVal,
		topUpValue,
		hardCodedScResultingAddress,
	)

	integrationTests.NodeDoesWithdraw(nodes, idxNodeShard0, withdrawValue, hardCodedScResultingAddress)

	roundsToWait = 12
	for i := 0; i < roundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
	}

	checkWithdrawIsDoneCorrectly(
		t,
		nodes,
		idxNodeShard1,
		idxNodeShard0,
		initialVal,
		topUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)
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

	nodeProposerShard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	nodeValidatorShard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)

	nodeProposerShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodeProposerShard1.LoadTxSignSkBytes(hardCodedSk)
	nodeValidatorShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)

	nodeProposerMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)
	nodeValidatorMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)

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
	round = integrationTests.IncrementAndPrintRound(round)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	integrationTests.DeployScTx(nodes, idxProposerShard1, string(scCode))

	integrationTests.ProposeBlock(nodes, idxProposers, round)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)

	integrationTests.PlayerJoinsGame(
		nodes,
		nodes[idxProposerShard0].OwnAccount,
		topUpValue,
		"aaaa",
		hardCodedScResultingAddress,
	)
	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		idxValidators, idxProposers = idxProposers, idxValidators
	}

	checkJoinGameIsDoneCorrectly(
		t,
		nodes,
		idxProposerShard1,
		idxProposerShard0,
		initialVal,
		topUpValue,
		hardCodedScResultingAddress,
	)

	integrationTests.NodeCallsRewardAndSend(
		nodes,
		idxProposerShard1,
		nodes[idxProposerShard0].OwnAccount.Address.Bytes(),
		withdrawValue,
		"aaaa",
		hardCodedScResultingAddress,
	)

	//TODO investigate why do we need 7 rounds here
	roundsToWait = 7
	for i := 0; i < roundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		idxValidators, idxProposers = idxProposers, idxValidators
	}

	checkRewardIsDoneCorrectly(
		t,
		nodes,
		idxProposerShard1,
		idxProposerShard0,
		initialVal,
		topUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)

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
	nodeValidatorShard0 := integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)

	nodeProposerShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodeProposerShard1.LoadTxSignSkBytes(hardCodedSk)
	nodeValidatorShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)

	nodeProposerMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)
	nodeValidatorMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)

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
	round = integrationTests.IncrementAndPrintRound(round)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	integrationTests.MintAllNodes(nodes, initialVal)

	integrationTests.DeployScTx(nodes, idxProposerShard1, string(scCode))
	integrationTests.ProposeBlock(nodes, idxProposers, round)
	integrationTests.SyncBlock(t, nodes, idxProposers, round)
	round = integrationTests.IncrementAndPrintRound(round)

	integrationTests.PlayerJoinsGame(
		nodes,
		nodes[idxProposerShard0].OwnAccount,
		topUpValue,
		"aaaa",
		hardCodedScResultingAddress,
	)
	maxRoundsToWait := 10
	for i := 0; i < maxRoundsToWait; i++ {
		integrationTests.ProposeBlock(nodes, idxProposersWithoutShard1, round)

		hdr, body, isBodyEmpty := integrationTests.ProposeBlockSignalsEmptyBlock(nodes[idxProposerShard1], round)
		if isBodyEmpty {
			nodes[idxProposerShard1].CommitBlock(body, hdr)
			integrationTests.SyncBlock(t, nodes, idxProposers, round)
			round = integrationTests.IncrementAndPrintRound(round)
			continue
		}

		//shard 1' proposer got to process at least 1 tx, should not commit but the shard 1's validator
		//should be able to process the block

		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		integrationTests.CheckRootHashes(t, nodes, idxProposers)
		break
	}

	checkJoinGameIsDoneCorrectly(
		t,
		nodes,
		idxProposerShard1,
		idxProposerShard0,
		initialVal,
		topUpValue,
		hardCodedScResultingAddress,
	)
}

func checkTopUpIsDoneCorrectly(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxNodeScExists int,
	idxNodeCallerExists int,
	initialVal *big.Int,
	topUpVal *big.Int,
	scAddressBytes []byte,
) {

	nodeWithSc := nodes[idxNodeScExists]
	nodeWithCaller := nodes[idxNodeCallerExists]

	fmt.Println("Checking SC account received topUp val...")
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, topUpVal, accnt.(*state.Account).Balance)

	fmt.Println("Checking SC.balanceOf...")
	bytesValue, _ := nodeWithSc.ScDataGetter.Get(
		scAddressBytes,
		"balanceOf",
		nodeWithCaller.OwnAccount.PkTxSignBytes,
	)
	retrievedValue := big.NewInt(0).SetBytes(bytesValue)
	fmt.Printf("SC balanceOf returned %d\n", retrievedValue)
	assert.Equal(t, topUpVal, retrievedValue)

	fmt.Println("Checking sender has initial-topUp val...")
	expectedVal := big.NewInt(0).Set(initialVal)
	expectedVal.Sub(expectedVal, topUpVal)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(nodeWithCaller.OwnAccount.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedVal, accnt.(*state.Account).Balance)
}

func checkJoinGameIsDoneCorrectly(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxNodeScExists int,
	idxNodeCallerExists int,
	initialVal *big.Int,
	topUpVal *big.Int,
	scAddressBytes []byte,
) {

	nodeWithSc := nodes[idxNodeScExists]
	nodeWithCaller := nodes[idxNodeCallerExists]

	fmt.Println("Checking SC account received topUp val...")
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, topUpVal, accnt.(*state.Account).Balance)

	fmt.Println("Checking sender has initial-topUp val...")
	expectedVal := big.NewInt(0).Set(initialVal)
	expectedVal.Sub(expectedVal, topUpVal)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(nodeWithCaller.OwnAccount.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedVal, accnt.(*state.Account).Balance)
}

func checkWithdrawIsDoneCorrectly(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxNodeScExists int,
	idxNodeCallerExists int,
	initialVal *big.Int,
	topUpVal *big.Int,
	withdraw *big.Int,
	scAddressBytes []byte,
) {

	nodeWithSc := nodes[idxNodeScExists]
	nodeWithCaller := nodes[idxNodeCallerExists]

	fmt.Println("Checking SC account has topUp-withdraw val...")
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	expectedSC := big.NewInt(0).Set(topUpVal)
	expectedSC.Sub(expectedSC, withdraw)
	assert.Equal(t, expectedSC, accnt.(*state.Account).Balance)

	fmt.Println("Checking SC.balanceOf...")
	bytesValue, _ := nodeWithSc.ScDataGetter.Get(
		scAddressBytes,
		"balanceOf",
		nodeWithCaller.OwnAccount.PkTxSignBytes,
	)
	retrievedValue := big.NewInt(0).SetBytes(bytesValue)
	fmt.Printf("SC balanceOf returned %d\n", retrievedValue)
	assert.Equal(t, expectedSC, retrievedValue)

	fmt.Println("Checking sender has initial-topUp+withdraw val...")
	expectedSender := big.NewInt(0).Set(initialVal)
	expectedSender.Sub(expectedSender, topUpVal)
	expectedSender.Add(expectedSender, withdraw)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(nodeWithCaller.OwnAccount.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedSender, accnt.(*state.Account).Balance)
}

func checkRewardIsDoneCorrectly(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxNodeScExists int,
	idxNodeCallerExists int,
	initialVal *big.Int,
	topUpVal *big.Int,
	withdraw *big.Int,
	scAddressBytes []byte,
) {

	nodeWithSc := nodes[idxNodeScExists]
	nodeWithCaller := nodes[idxNodeCallerExists]

	fmt.Println("Checking SC account has topUp-withdraw val...")
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	expectedSC := big.NewInt(0).Set(topUpVal)
	expectedSC.Sub(expectedSC, withdraw)
	assert.Equal(t, expectedSC, accnt.(*state.Account).Balance)

	fmt.Println("Checking sender has initial-topUp+withdraw val...")
	expectedSender := big.NewInt(0).Set(initialVal)
	expectedSender.Sub(expectedSender, topUpVal)
	expectedSender.Add(expectedSender, withdraw)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddressFromAddrBytes(nodeWithCaller.OwnAccount.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedSender, accnt.(*state.Account).Balance)
}
