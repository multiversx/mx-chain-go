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
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

var agarioFile = "agarioV2.hex"
var stepDelay = time.Second * 2

func TestShouldProcessBlocksInMultiShardArchitectureWithScTxsTopUpAndWithdraw(t *testing.T) {
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
	hardCodedSk, _ := hex.DecodeString("cad6df09f1e958a38e7d41edcca4421072075c53011b48de58d5ffaf06ed0b06")
	hardCodedScResultingAddress, _ := hex.DecodeString("00000000000000000000dac025f619852c170e981e9902fb1352d31e724a54a0")
	nodeShard0.LoadTxSignSkBytes(hardCodedSk)
	nodeShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	nodeMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)

	nodes := []*integrationTests.TestProcessorNode{nodeShard0, nodeShard1, nodeMeta}
	idxNodeShard0 := 0
	idxNodeShard1 := 1

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

	round := uint32(1)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	stepMintAllNodes(nodes, initialVal)

	stepDeployScTx(nodes, idxNodeShard0, string(scCode))
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepNodeSh1DoesTopUp(nodes, idxNodeShard1, topUpValue, hardCodedScResultingAddress)
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++

	stepCheckTopUpIsDoneCorrectly(
		t,
		nodes,
		idxNodeShard0,
		idxNodeShard1,
		initialVal,
		topUpValue,
		hardCodedScResultingAddress,
	)

	stepNodeSh1DoesWithdraw(nodes, idxNodeShard1, withdrawValue, hardCodedScResultingAddress)
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++

	stepCheckWithdrawIsDoneCorrectly(
		t,
		nodes,
		idxNodeShard0,
		idxNodeShard1,
		initialVal,
		topUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)
}

func TestShouldProcessBlocksInMultiShardArchitectureWithScTxsJoinAndReward(t *testing.T) {
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
	hardCodedSk, _ := hex.DecodeString("cad6df09f1e958a38e7d41edcca4421072075c53011b48de58d5ffaf06ed0b06")
	hardCodedScResultingAddress, _ := hex.DecodeString("00000000000000000000dac025f619852c170e981e9902fb1352d31e724a54a0")
	nodeShard0.LoadTxSignSkBytes(hardCodedSk)
	nodeShard1 := integrationTests.NewTestProcessorNode(maxShards, 1, 1, advertiserAddr)
	nodeMeta := integrationTests.NewTestProcessorNode(maxShards, sharding.MetachainShardId, 0, advertiserAddr)

	nodes := []*integrationTests.TestProcessorNode{nodeShard0, nodeShard1, nodeMeta}
	idxNodeShard0 := 0
	idxNodeShard1 := 1

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

	round := uint32(1)

	initialVal := big.NewInt(10000000)
	topUpValue := big.NewInt(500)
	withdrawValue := big.NewInt(10)
	stepMintAllNodes(nodes, initialVal)

	stepDeployScTx(nodes, idxNodeShard0, string(scCode))
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepNodeSh1DoesJoinGame(nodes, idxNodeShard1, topUpValue, hardCodedScResultingAddress)
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++

	stepCheckJoinGameIsDoneCorrectly(
		t,
		nodes,
		idxNodeShard0,
		idxNodeShard1,
		initialVal,
		topUpValue,
		hardCodedScResultingAddress,
	)

	stepNodeSh0CallsRewardAndSend(nodes, idxNodeShard0, idxNodeShard1, withdrawValue, hardCodedScResultingAddress)
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++
	stepProposeBlock(nodes, round)
	round++

	stepCheckRewardIsDoneCorrectly(
		t,
		nodes,
		idxNodeShard0,
		idxNodeShard1,
		initialVal,
		topUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)
}

func stepMintAllNodes(nodes []*integrationTests.TestProcessorNode, value *big.Int) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == sharding.MetachainShardId {
			continue
		}

		integrationTests.MintAddress(n.AccntState, n.PkTxSignBytes, value)
	}
}

func stepDeployScTx(nodes []*integrationTests.TestProcessorNode, senderIdx int, scCode string) {
	fmt.Println("Deploying SC...")
	txDeploy := createTxDeploy(nodes[senderIdx], string(scCode))
	nodes[senderIdx].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func stepProposeBlock(nodes []*integrationTests.TestProcessorNode, round uint32) {
	fmt.Println("All shards propose blocks...")
	for _, n := range nodes {
		body, header := n.ProposeBlockOnlyWithSelf(round)
		n.BroadcastAndCommit(body, header)
	}

	fmt.Println("Delaying for disseminating headers and miniblocks...")
	time.Sleep(stepDelay)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func stepNodeSh1DoesTopUp(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	topUpValue *big.Int,
	scAddress []byte) {

	fmt.Println("Calling SC.topUp...")
	txDeploy := createTxTopUp(nodes[idxNode], topUpValue, scAddress)
	nodes[idxNode].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func stepNodeSh1DoesJoinGame(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	joinGameVal *big.Int,
	scAddress []byte) {

	fmt.Println("Calling SC.joinGame...")
	txDeploy := createTxJoinGame(nodes[idxNode], joinGameVal, scAddress)
	nodes[idxNode].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func stepCheckTopUpIsDoneCorrectly(
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
	assert.Equal(t, expectedVal, accnt.(*state.Account).Balance)
}

func stepCheckJoinGameIsDoneCorrectly(
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
	assert.Equal(t, topUpVal, accnt.(*state.Account).Balance)

	fmt.Println("Checking sender has initial-topUp val...")
	expectedVal := big.NewInt(0).Set(initialVal)
	expectedVal.Sub(expectedVal, topUpVal)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
	assert.Equal(t, expectedVal, accnt.(*state.Account).Balance)
}

func stepNodeSh1DoesWithdraw(
	nodes []*integrationTests.TestProcessorNode,
	idxNode int,
	withdrawValue *big.Int,
	scAddress []byte) {

	fmt.Println("Calling SC.withdraw...")
	txDeploy := createTxWithdraw(nodes[idxNode], withdrawValue, scAddress)
	nodes[idxNode].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(time.Second * 1)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func stepNodeSh0CallsRewardAndSend(
	nodes []*integrationTests.TestProcessorNode,
	idxNodeOwner int,
	idxNodeUser int,
	prize *big.Int,
	scAddress []byte) {

	fmt.Println("Calling SC.rewardAndSendToWallet...")
	txDeploy := createTxRewardAndSendToWallet(nodes[idxNodeOwner], nodes[idxNodeUser], prize, scAddress)
	nodes[idxNodeOwner].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(time.Second * 1)

	fmt.Println(integrationTests.MakeDisplayTable(nodes))
}

func stepCheckWithdrawIsDoneCorrectly(
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
	assert.Equal(t, expectedSender, accnt.(*state.Account).Balance)
}

func stepCheckRewardIsDoneCorrectly(
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
	expectedSC := big.NewInt(0).Set(topUpVal)
	expectedSC.Sub(expectedSC, withdraw)
	assert.Equal(t, expectedSC, accnt.(*state.Account).Balance)

	fmt.Println("Checking sender has initial-topUp+withdraw val...")
	expectedSender := big.NewInt(0).Set(initialVal)
	expectedSender.Sub(expectedSender, topUpVal)
	expectedSender.Add(expectedSender, withdraw)
	accnt, _ = nodeWithCaller.AccntState.GetExistingAccount(integrationTests.CreateAddresFromAddrBytes(nodeWithCaller.PkTxSignBytes))
	assert.Equal(t, expectedSender, accnt.(*state.Account).Balance)
}

func createTxDeploy(tn *integrationTests.TestProcessorNode, scCode string) *transaction.Transaction {
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

func createTxTopUp(tn *integrationTests.TestProcessorNode, topUpVal *big.Int, scAddress []byte) *transaction.Transaction {
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

func createTxJoinGame(tn *integrationTests.TestProcessorNode, joinGameVal *big.Int, scAddress []byte) *transaction.Transaction {
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

func createTxWithdraw(tn *integrationTests.TestProcessorNode, withdrawVal *big.Int, scAddress []byte) *transaction.Transaction {
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

func createTxRewardAndSendToWallet(tnOwner *integrationTests.TestProcessorNode, tnUser *integrationTests.TestProcessorNode, prizeVal *big.Int, scAddress []byte) *transaction.Transaction {
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
