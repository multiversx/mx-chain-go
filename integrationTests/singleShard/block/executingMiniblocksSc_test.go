package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/pkg/profile"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "agarioV2.hex"
var stepDelay = time.Second

func TestShouldProcessWithScTxsJoinAndRewardTheOwner(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	log := logger.DefaultLogger()
	log.SetLevel(logger.LogDebug)

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(1)
	numOfNodes := 4
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	}

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

	idxProposer := 0
	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodes[idxProposer].LoadTxSignSkBytes(hardCodedSk)

	deployScTx(nodes, idxProposer, string(scCode))
	proposeBlock(nodes, idxProposer, round)
	syncBlock(t, nodes, idxProposer, round)
	round = incrementAndPrintRound(round)

	nodeJoinsGame(nodes, idxProposer, topUpValue, 0, hardCodedScResultingAddress)
	proposeBlock(nodes, idxProposer, round)
	syncBlock(t, nodes, idxProposer, round)
	round = incrementAndPrintRound(round)

	checkJoinGameIsDoneCorrectly(
		t,
		nodes,
		idxProposer,
		idxProposer,
		initialVal,
		topUpValue,
		hardCodedScResultingAddress,
	)

	nodeCallsRewardAndSend(nodes, idxProposer, idxProposer, withdrawValue, 0, hardCodedScResultingAddress)
	proposeBlock(nodes, idxProposer, round)
	syncBlock(t, nodes, idxProposer, round)
	round = incrementAndPrintRound(round)

	checkRewardIsDoneCorrectly(
		t,
		nodes,
		idxProposer,
		idxProposer,
		initialVal,
		topUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)

	checkRootHashes(t, nodes, []int{0})

	time.Sleep(1 * time.Second)
}

func TestProcessesJoinGameOf100PlayersRewardAndEndgame(t *testing.T) {
	t.Skip("this is a stress test for VM and AGAR.IO")

	stepDelay = time.Nanosecond

	p := profile.Start(profile.MemProfile, profile.ProfilePath("."), profile.NoShutdownHook)
	defer p.Stop()

	log := logger.DefaultLogger()
	log.SetLevel(logger.LogDebug)

	scCode, err := ioutil.ReadFile(agarioFile)
	assert.Nil(t, err)

	maxShards := uint32(1)
	numOfNodes := 1
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)

	nodes := make([]*integrationTests.TestProcessorNode, numOfNodes)
	for i := 0; i < numOfNodes; i++ {
		nodes[i] = integrationTests.NewTestProcessorNode(maxShards, 0, 0, advertiserAddr)
	}

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

	idxProposer := 0
	hardCodedSk, _ := hex.DecodeString("5561d28b0d89fa425bbbf9e49a018b5d1e4a462c03d2efce60faf9ddece2af06")
	hardCodedScResultingAddress, _ := hex.DecodeString("000000000000000000005fed9c659422cd8429ce92f8973bba2a9fb51e0eb3a1")
	nodes[idxProposer].LoadTxSignSkBytes(hardCodedSk)

	deployScTx(nodes, idxProposer, string(scCode))
	proposeBlock(nodes, idxProposer, round)
	syncBlock(t, nodes, idxProposer, round)
	round = incrementAndPrintRound(round)

	rMonitor := &statistics.ResourceMonitor{}
	fmt.Println(rMonitor.GenerateStatistics())

	for rr := 0; rr < 50; rr++ {
		for i := 0; i < 100; i++ {
			nodeJoinsGame(nodes, idxProposer, topUpValue, rr, hardCodedScResultingAddress)
		}
		time.Sleep(time.Second)

		startTime := time.Now()
		proposeBlock(nodes, idxProposer, round)
		elapsedTime := time.Since(startTime)
		fmt.Printf("Block Created in %s\n", elapsedTime)
		round = incrementAndPrintRound(round)

		nodeCallsRewardAndSend(nodes, idxProposer, idxProposer, withdrawValue, rr, hardCodedScResultingAddress)
		nodeEndGame(nodes, idxProposer, rr, hardCodedScResultingAddress)
		time.Sleep(time.Second)

		startTime = time.Now()
		proposeBlock(nodes, idxProposer, round)
		elapsedTime = time.Since(startTime)
		fmt.Printf("Block Created in %s\n", elapsedTime)
		round = incrementAndPrintRound(round)

		fmt.Println(rMonitor.GenerateStatistics())
	}

	checkRootHashes(t, nodes, []int{0})

	time.Sleep(1 * time.Second)
}

func incrementAndPrintRound(round uint64) uint64 {
	round++
	fmt.Printf("#################################### ROUND %d BEGINS ####################################\n\n", round)

	return round
}

func deployScTx(nodes []*integrationTests.TestProcessorNode, senderIdx int, scCode string) {
	fmt.Println("Deploying SC...")
	txDeploy := createTxDeploy(nodes[senderIdx], scCode)
	nodes[senderIdx].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func proposeBlock(nodes []*integrationTests.TestProcessorNode, idxProposer int, round uint64) {
	fmt.Println("Proposing block...")
	for idx, n := range nodes {
		if idx != idxProposer {
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

func syncBlock(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposer int, round uint64) {
	fmt.Println("All other shard nodes sync the proposed block...")
	for idx, n := range nodes {
		if idx == idxProposer {
			continue
		}

		err := n.SyncNode(uint64(round))
		if err != nil {
			assert.Fail(t, err.Error())
			return
		}
	}

	time.Sleep(stepDelay)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func nodeJoinsGame(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	joinGameVal *big.Int,
	round int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.joinGame...")
	txScCall := createTxJoinGame(nodes[idxNode], joinGameVal, round, scAddress)
	nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)
}

func nodeEndGame(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	round int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.endGame...")
	txScCall := createTxEndGame(nodes[idxNode], round, scAddress)
	nodes[idxNode].SendTransaction(txScCall)
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func nodeCallsRewardAndSend(
	nodes []*integrationTests.TestProcessorNode,
	idxNodeOwner int,
	idxNodeUser int,
	prize *big.Int,
	round int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.rewardAndSendToWallet...")
	txScCall := createTxRewardAndSendToWallet(nodes[idxNodeOwner], nodes[idxNodeUser], prize, round, scAddress)
	nodes[idxNodeOwner].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)
}

func createTxDeploy(tn *integrationTests.TestProcessorNode, scCode string) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  make([]byte, 32),
		SndAddr:  tn.PkTxSignBytes,
		Data:     scCode,
		GasPrice: 0,
		GasLimit: 1000000000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tn.SingleSigner.Sign(tn.SkTxSign, txBuff)

	return tx
}

func createTxEndGame(tn *integrationTests.TestProcessorNode, round int, scAddress []byte) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(100),
		RcvAddr:  scAddress,
		SndAddr:  tn.PkTxSignBytes,
		Data:     fmt.Sprintf("endGame@%d", round),
		GasPrice: 0,
		GasLimit: 10000000000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tn.SingleSigner.Sign(tn.SkTxSign, txBuff)

	fmt.Printf("End %s\n", hex.EncodeToString(tn.PkTxSignBytes))

	return tx
}

func createTxJoinGame(tn *integrationTests.TestProcessorNode, joinGameVal *big.Int, round int, scAddress []byte) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    joinGameVal,
		RcvAddr:  scAddress,
		SndAddr:  tn.PkTxSignBytes,
		Data:     fmt.Sprintf("joinGame@%d", round),
		GasPrice: 0,
		GasLimit: 10000000000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tn.SingleSigner.Sign(tn.SkTxSign, txBuff)

	fmt.Printf("Join %s\n", hex.EncodeToString(tn.PkTxSignBytes))

	return tx
}

func createTxRewardAndSendToWallet(tnOwner *integrationTests.TestProcessorNode, tnUser *integrationTests.TestProcessorNode, prizeVal *big.Int, round int, scAddress []byte) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  tnOwner.PkTxSignBytes,
		Data:     fmt.Sprintf("rewardAndSendToWallet@%d@%s@%X", round, hex.EncodeToString(tnUser.PkTxSignBytes), prizeVal),
		GasPrice: 0,
		GasLimit: 10000000000,
	}
	txBuff, _ := integrationTests.TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tnOwner.SingleSigner.Sign(tnOwner.SkTxSign, txBuff)

	fmt.Printf("Reward %s\n", hex.EncodeToString(tnUser.PkTxSignBytes))

	return tx
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
	fmt.Printf("Checking %s\n", hex.EncodeToString(nodeWithCaller.PkTxSignBytes))
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedVal, accnt.(*state.Account).Balance)
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
	fmt.Printf("Checking %s\n", hex.EncodeToString(nodeWithCaller.PkTxSignBytes))
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedSender, accnt.(*state.Account).Balance)
}

func checkRootHashes(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int) {
	for _, idx := range idxProposers {
		checkRootHashInShard(t, nodes, idx)
	}
}

func checkRootHashInShard(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposer int) {
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
