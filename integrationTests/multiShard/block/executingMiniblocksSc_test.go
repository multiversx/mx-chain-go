package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "agarioV2.hex"
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
	round = incrementAndPrintRound(round)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	deployScTx(nodes, idxNodeShard1, string(scCode))

	proposeBlockWithScTxs(nodes, round, idxProposers)
	round = incrementAndPrintRound(round)

	nodeDoesTopUp(nodes, idxNodeShard0, topUpValue, hardCodedScResultingAddress)

	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		proposeBlockWithScTxs(nodes, round, idxProposers)
		round = incrementAndPrintRound(round)
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

	nodeDoesWithdraw(nodes, idxNodeShard0, withdrawValue, hardCodedScResultingAddress)

	roundsToWait = 12
	for i := 0; i < roundsToWait; i++ {
		proposeBlockWithScTxs(nodes, round, idxProposers)
		round = incrementAndPrintRound(round)
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
	round = incrementAndPrintRound(round)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	integrationTests.MintAllNodes(nodes, initialVal)

	deployScTx(nodes, idxProposerShard1, string(scCode))

	proposeBlockWithScTxs(nodes, round, idxProposers)
	syncBlock(t, nodes, idxProposers, round)
	round = incrementAndPrintRound(round)

	nodeJoinsGame(nodes, idxProposerShard0, topUpValue, hardCodedScResultingAddress)

	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		proposeBlockWithScTxs(nodes, round, idxProposers)
		syncBlock(t, nodes, idxProposers, round)
		round = incrementAndPrintRound(round)
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

	nodeCallsRewardAndSend(nodes, idxProposerShard1, idxProposerShard0, withdrawValue, hardCodedScResultingAddress)

	//TODO investigate why do we need 7 rounds here
	roundsToWait = 7
	for i := 0; i < roundsToWait; i++ {
		proposeBlockWithScTxs(nodes, round, idxProposers)
		syncBlock(t, nodes, idxProposers, round)
		round = incrementAndPrintRound(round)
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

	checkRootHashes(t, nodes, idxProposers)
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
	round = incrementAndPrintRound(round)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	integrationTests.MintAllNodes(nodes, initialVal)

	deployScTx(nodes, idxProposerShard1, string(scCode))
	proposeBlockWithScTxs(nodes, round, idxProposers)
	syncBlock(t, nodes, idxProposers, round)
	round = incrementAndPrintRound(round)

	nodeJoinsGame(nodes, idxProposerShard0, topUpValue, hardCodedScResultingAddress)
	maxRoundsToWait := 10
	for i := 0; i < maxRoundsToWait; i++ {
		proposeBlockWithScTxs(nodes, round, idxProposersWithoutShard1)

		hdr, body, isBodyEmpty := proposeBlockSignalsEmptyBlock(nodes[idxProposerShard1], round)
		if isBodyEmpty {
			nodes[idxProposerShard1].CommitBlock(body, hdr)
			syncBlock(t, nodes, idxProposers, round)
			round = incrementAndPrintRound(round)
			continue
		}

		//shard 1' proposer got to process at least 1 tx, should not commit but the shard 1's validator
		//should be able to process the block

		syncBlock(t, nodes, idxProposers, round)
		round = incrementAndPrintRound(round)
		checkRootHashes(t, nodes, idxProposers)
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

func incrementAndPrintRound(round uint64) uint64 {
	round++
	fmt.Printf("#################################### ROUND %d BEGINS ####################################\n\n", round)

	time.Sleep(stepDelay)
	return round
}

func deployScTx(
	nodes []*integrationTests.TestProcessorNode,
	senderIdx int,
	scCode string,
) {

	fmt.Println("Deploying SC...")
	txDeploy := createTxDeploy(nodes[senderIdx], scCode)
	_, _ = nodes[senderIdx].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func proposeBlockWithScTxs(
	nodes []*integrationTests.TestProcessorNode,
	round uint64,
	idxProposers []int,
) {

	fmt.Println("All shards propose blocks...")
	for idx, n := range nodes {
		if !isIntInSlice(idx, idxProposers) {
			continue
		}

		body, header, _ := n.ProposeBlock(round)
		n.BroadcastBlock(body, header)
		n.CommitBlock(body, header)
	}

	fmt.Println("Delaying for disseminating headers and miniblocks...")
	time.Sleep(stepDelay)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func syncBlock(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	round uint64,
) {

	fmt.Println("All other shard nodes sync the proposed block...")
	for idx, n := range nodes {
		if isIntInSlice(idx, idxProposers) {
			continue
		}

		err := n.SyncNode(round)
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}
	}

	time.Sleep(stepDelay)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func isIntInSlice(idx int, slice []int) bool {
	for _, value := range slice {
		if value == idx {
			return true
		}
	}

	return false
}

func nodeDoesTopUp(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	topUpValue *big.Int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.topUp...")
	txScCall := createTxTopUp(nodes[idxNode], topUpValue, scAddress)
	_, _ = nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func nodeJoinsGame(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	joinGameVal *big.Int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.joinGame...")
	txScCall := createTxJoinGame(nodes[idxNode], joinGameVal, scAddress)
	_, _ = nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func proposeBlockSignalsEmptyBlock(
	node *integrationTests.TestProcessorNode,
	round uint64,
) (data.HeaderHandler, data.BodyHandler, bool) {

	fmt.Println("Proposing block without commit...")

	body, header, txHashes := node.ProposeBlock(round)
	node.BroadcastBlock(body, header)
	isEmptyBlock := len(txHashes) == 0

	fmt.Println("Delaying for disseminating headers and miniblocks...")
	time.Sleep(stepDelay)

	return header, body, isEmptyBlock
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
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, topUpVal, accnt.(*state.Account).Balance)

	fmt.Println("Checking SC.balanceOf...")
	bytesValue, _ := nodeWithSc.ScDataGetter.Get(
		scAddressBytes,
		"balanceOf",
		nodeWithCaller.PkTxSignBytes,
	)
	retrievedValue := big.NewInt(0).SetBytes(bytesValue)
	fmt.Printf("SC balanceOf returned %d\n", retrievedValue)
	assert.Equal(t, topUpVal, retrievedValue)

	fmt.Println("Checking sender has initial-topUp val...")
	expectedVal := big.NewInt(0).Set(initialVal)
	expectedVal.Sub(expectedVal, topUpVal)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
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
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, topUpVal, accnt.(*state.Account).Balance)

	fmt.Println("Checking sender has initial-topUp val...")
	expectedVal := big.NewInt(0).Set(initialVal)
	expectedVal.Sub(expectedVal, topUpVal)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedVal, accnt.(*state.Account).Balance)
}

func nodeDoesWithdraw(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	withdrawValue *big.Int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.withdraw...")
	txScCall := createTxWithdraw(nodes[idxNode], withdrawValue, scAddress)
	_, _ = nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(time.Second * 1)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func nodeCallsRewardAndSend(
	nodes []*integrationTests.TestProcessorNode,
	idxNodeOwner int,
	idxNodeUser int,
	prize *big.Int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.rewardAndSendToWallet...")
	txScCall := createTxRewardAndSendToWallet(nodes[idxNodeOwner], nodes[idxNodeUser], prize, scAddress)
	_, _ = nodes[idxNodeOwner].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(time.Second * 1)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
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
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	expectedSC := big.NewInt(0).Set(topUpVal)
	expectedSC.Sub(expectedSC, withdraw)
	assert.Equal(t, expectedSC, accnt.(*state.Account).Balance)

	fmt.Println("Checking SC.balanceOf...")
	bytesValue, _ := nodeWithSc.ScDataGetter.Get(
		scAddressBytes,
		"balanceOf",
		nodeWithCaller.PkTxSignBytes,
	)
	retrievedValue := big.NewInt(0).SetBytes(bytesValue)
	fmt.Printf("SC balanceOf returned %d\n", retrievedValue)
	assert.Equal(t, expectedSC, retrievedValue)

	fmt.Println("Checking sender has initial-topUp+withdraw val...")
	expectedSender := big.NewInt(0).Set(initialVal)
	expectedSender.Sub(expectedSender, topUpVal)
	expectedSender.Add(expectedSender, withdraw)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
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
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	expectedSC := big.NewInt(0).Set(topUpVal)
	expectedSC.Sub(expectedSC, withdraw)
	assert.Equal(t, expectedSC, accnt.(*state.Account).Balance)

	fmt.Println("Checking sender has initial-topUp+withdraw val...")
	expectedSender := big.NewInt(0).Set(initialVal)
	expectedSender.Sub(expectedSender, topUpVal)
	expectedSender.Add(expectedSender, withdraw)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedSender, accnt.(*state.Account).Balance)
}

func checkRootHashes(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
) {

	for _, idx := range idxProposers {
		checkRootHashInShard(t, nodes, idx)
	}
}

func checkRootHashInShard(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposer int,
) {

	proposerNode := nodes[idxProposer]
	proposerRootHash, _ := proposerNode.AccntState.RootHash()

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]

		if node.ShardCoordinator.SelfId() != proposerNode.ShardCoordinator.SelfId() {
			continue
		}

		fmt.Printf("Testing roothash for node index %d, shard ID %d...\n", i, node.ShardCoordinator.SelfId())
		nodeRootHash, _ := node.AccntState.RootHash()
		assert.Equal(t, proposerRootHash, nodeRootHash)
	}
}

func createTxDeploy(
	tn *integrationTests.TestProcessorNode,
	scCode string,
) *transaction.Transaction {

	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  make([]byte, 32),
		SndAddr:  tn.PkTxSignBytes,
		Data:     scCode,
		GasPrice: 0,
		GasLimit: 100000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tn.SingleSigner.Sign(tn.SkTxSign, txBuff)

	return tx
}

func createTxTopUp(
	tn *integrationTests.TestProcessorNode,
	topUpVal *big.Int,
	scAddress []byte,
) *transaction.Transaction {

	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    topUpVal,
		RcvAddr:  scAddress,
		SndAddr:  tn.PkTxSignBytes,
		Data:     fmt.Sprintf("topUp"),
		GasPrice: 0,
		GasLimit: 100000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tn.SingleSigner.Sign(tn.SkTxSign, txBuff)

	return tx
}

func createTxJoinGame(
	tn *integrationTests.TestProcessorNode,
	joinGameVal *big.Int,
	scAddress []byte,
) *transaction.Transaction {

	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    joinGameVal,
		RcvAddr:  scAddress,
		SndAddr:  tn.PkTxSignBytes,
		Data:     fmt.Sprintf("joinGame@aaaa"),
		GasPrice: 0,
		GasLimit: 100000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tn.SingleSigner.Sign(tn.SkTxSign, txBuff)

	return tx
}

func createTxWithdraw(
	tn *integrationTests.TestProcessorNode,
	withdrawVal *big.Int,
	scAddress []byte,
) *transaction.Transaction {

	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  tn.PkTxSignBytes,
		Data:     fmt.Sprintf("withdraw@%X", withdrawVal),
		GasPrice: 0,
		GasLimit: 100000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tn.SingleSigner.Sign(tn.SkTxSign, txBuff)

	return tx
}

func createTxRewardAndSendToWallet(
	tnOwner *integrationTests.TestProcessorNode,
	tnUser *integrationTests.TestProcessorNode,
	prizeVal *big.Int,
	scAddress []byte,
) *transaction.Transaction {

	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  tnOwner.PkTxSignBytes,
		Data:     fmt.Sprintf("rewardAndSendToWallet@aaaa@%s@%X", hex.EncodeToString(tnUser.PkTxSignBytes), prizeVal),
		GasPrice: 0,
		GasLimit: 100000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tnOwner.SingleSigner.Sign(tnOwner.SkTxSign, txBuff)

	return tx
}
