package integrationTests

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/stretchr/testify/assert"
)

// DeployScTx creates and sends a SC tx
func DeployScTx(nodes []*TestProcessorNode, senderIdx int, scCode string) {
	fmt.Println("Deploying SC...")
	txDeploy := generateTx(
		nodes[senderIdx].OwnAccount.SkTxSign,
		nodes[senderIdx].OwnAccount.SingleSigner,
		&txArgs{
			value:    big.NewInt(0),
			rcvAddr:  make([]byte, 32),
			sndAddr:  nodes[senderIdx].OwnAccount.PkTxSignBytes,
			data:     scCode,
			gasLimit: 1000000000000,
		})
	_, _ = nodes[senderIdx].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(stepDelay)

	fmt.Println(MakeDisplayTable(nodes))
}

// PlayerJoinsGame creates and sends a join game transaction to the SC
func PlayerJoinsGame(
	nodes []*TestProcessorNode,
	player *TestWalletAccount,
	joinGameVal *big.Int,
	round string,
	scAddress []byte,
) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address.Bytes())
	fmt.Println("Calling SC.joinGame...")
	txScCall := generateTx(
		player.SkTxSign,
		player.SingleSigner,
		&txArgs{
			value:    joinGameVal,
			rcvAddr:  scAddress,
			sndAddr:  player.Address.Bytes(),
			data:     fmt.Sprintf("joinGame@%s", round),
			gasLimit: 1000000000000,
		})
	fmt.Printf("Join %s\n", hex.EncodeToString(player.Address.Bytes()))
	_, _ = txDispatcherNode.SendTransaction(txScCall)
}

// NodeCallsRewardAndSend - smart contract owner sends reward transaction
func NodeCallsRewardAndSend(
	nodes []*TestProcessorNode,
	idxNodeOwner int,
	winnerAddress []byte,
	prize *big.Int,
	round string,
	scAddress []byte,
) {

	fmt.Println("Calling SC.rewardAndSendToWallet...")
	txScCall := generateTx(
		nodes[idxNodeOwner].OwnAccount.SkTxSign,
		nodes[idxNodeOwner].OwnAccount.SingleSigner,
		&txArgs{
			value:    big.NewInt(0),
			rcvAddr:  scAddress,
			sndAddr:  nodes[idxNodeOwner].OwnAccount.PkTxSignBytes,
			data:     fmt.Sprintf("rewardAndSendToWallet@%s@%s@%X", round, hex.EncodeToString(winnerAddress), prize),
			gasLimit: 100000,
		})
	fmt.Printf("Reward %s\n", hex.EncodeToString(winnerAddress))
	_, _ = nodes[idxNodeOwner].SendTransaction(txScCall)

	fmt.Println(MakeDisplayTable(nodes))
}

// NodeDoesWithdraw creates and sends a withdraw tx to the SC
func NodeDoesWithdraw(
	nodes []*TestProcessorNode,
	idxNode int,
	withdrawValue *big.Int,
	scAddress []byte,
) {
	fmt.Println("Calling SC.withdraw...")
	txScCall := generateTx(
		nodes[idxNode].OwnAccount.SkTxSign,
		nodes[idxNode].OwnAccount.SingleSigner,
		&txArgs{
			value:    big.NewInt(0),
			rcvAddr:  scAddress,
			sndAddr:  nodes[idxNode].OwnAccount.PkTxSignBytes,
			data:     fmt.Sprintf("withdraw@%X", withdrawValue),
			gasLimit: 100000,
		})
	_, _ = nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)

	fmt.Println(MakeDisplayTable(nodes))
}

// NodeDoesTopUp creates and sends a
func NodeDoesTopUp(
	nodes []*TestProcessorNode,
	idxNode int,
	topUpValue *big.Int,
	scAddress []byte,
) {
	fmt.Println("Calling SC.topUp...")
	txScCall := generateTx(
		nodes[idxNode].OwnAccount.SkTxSign,
		nodes[idxNode].OwnAccount.SingleSigner,
		&txArgs{
			value:    topUpValue,
			rcvAddr:  scAddress,
			sndAddr:  nodes[idxNode].OwnAccount.PkTxSignBytes,
			data:     fmt.Sprintf("topUp"),
			gasLimit: 100000,
		})
	_, _ = nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(stepDelay)

	fmt.Println(MakeDisplayTable(nodes))
}
func getNodeWithinSameShardAsPlayer(
	nodes []*TestProcessorNode,
	player []byte,
) *TestProcessorNode {
	nodeWithCaller := nodes[0]
	playerShId := nodeWithCaller.ShardCoordinator.ComputeId(CreateAddressFromAddrBytes(player))
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == playerShId {
			nodeWithCaller = node
			break
		}
	}

	return nodeWithCaller
}

// CheckPlayerBalanceTheSameWithBlockchain verifies if player balance is the same as in the blockchain
func CheckPlayerBalanceTheSameWithBlockchain(
	t *testing.T,
	nodes []*TestProcessorNode,
	player *TestWalletAccount,
) {
	nodeWithCaller := getNodeWithinSameShardAsPlayer(nodes, player.Address.Bytes())

	fmt.Println("Checking sender has initial-topUp val...")
	accnt, _ := nodeWithCaller.AccntState.GetExistingAccount(CreateAddressFromAddrBytes(player.Address.Bytes()))
	assert.NotNil(t, accnt)
	ok := assert.Equal(t, player.Balance.Uint64(), accnt.(*state.Account).Balance.Uint64())
	if !ok {
		fmt.Printf("Expected player balance %d Actual player balance %d\n", player.Balance.Uint64(), accnt.(*state.Account).Balance.Uint64())
	}
}

// CheckBalanceIsDoneCorrectlySCSide verifies is smart contract balance is correct according to topup and withdraw
func CheckBalanceIsDoneCorrectlySCSide(
	t *testing.T,
	nodes []*TestProcessorNode,
	idxNodeScExists int,
	topUpVal *big.Int,
	withdraw *big.Int,
	scAddressBytes []byte,
) *big.Int {

	nodeWithSc := nodes[idxNodeScExists]

	fmt.Println("Checking SC account has topUp-withdraw val...")
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(CreateAddressFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	expectedSC := big.NewInt(0).Set(topUpVal)
	expectedSC.Sub(expectedSC, withdraw)
	ok := assert.Equal(t, expectedSC.Uint64(), accnt.(*state.Account).Balance.Uint64())
	if !ok {
		fmt.Printf("Expected smart contract val %d Actual smart contract val %d\n", expectedSC.Uint64(), accnt.(*state.Account).Balance.Uint64())
	}

	return expectedSC
}

// CheckJoinGame verifies if joinGame was done correctly by players
func CheckJoinGame(
	t *testing.T,
	nodes []*TestProcessorNode,
	players []*TestWalletAccount,
	topUpValue *big.Int,
	idxProposer int,
	hardCodedScResultingAddress []byte,
) {
	for _, player := range players {
		CheckPlayerBalanceTheSameWithBlockchain(
			t,
			nodes,
			player,
		)
	}

	numPlayers := len(players)
	allTopUpValue := big.NewInt(0).SetUint64(topUpValue.Uint64() * uint64(numPlayers))
	CheckBalanceIsDoneCorrectlySCSide(
		t,
		nodes,
		idxProposer,
		allTopUpValue,
		big.NewInt(0),
		hardCodedScResultingAddress,
	)
}

// CheckRewardsDistribution checks that players and smart contract balances are correct
func CheckRewardsDistribution(
	t *testing.T,
	nodes []*TestProcessorNode,
	players []*TestWalletAccount,
	topUpValue *big.Int,
	withdrawValue *big.Int,
	hardCodedScResultingAddress []byte,
	idxProposer int,
) {
	for _, player := range players {
		CheckPlayerBalanceTheSameWithBlockchain(
			t,
			nodes,
			player,
		)
	}

	numPlayers := len(players)
	allTopUpValue := big.NewInt(0).SetUint64(topUpValue.Uint64() * uint64(numPlayers))
	CheckBalanceIsDoneCorrectlySCSide(
		t,
		nodes,
		idxProposer,
		allTopUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)
}

// CheckSenderOkBalanceAfterTopUpAndWithdraw checks if sender balance is ok after top-up and withdraw
func CheckSenderOkBalanceAfterTopUpAndWithdraw(
	t *testing.T,
	nodeWithCaller *TestProcessorNode,
	initialVal *big.Int,
	topUpVal *big.Int,
	withdraw *big.Int,
) {
	fmt.Println("Checking sender has initial-topUp+withdraw val...")
	expectedSender := big.NewInt(0).Set(initialVal)
	expectedSender.Sub(expectedSender, topUpVal)
	expectedSender.Add(expectedSender, withdraw)
	accnt, _ := nodeWithCaller.AccntState.GetExistingAccount(CreateAddressFromAddrBytes(nodeWithCaller.OwnAccount.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedSender, accnt.(*state.Account).Balance)
}

// CheckSenderOkBalanceAfterTopUp checks if sender balance is ok after top-up
func CheckSenderOkBalanceAfterTopUp(
	t *testing.T,
	nodeWithCaller *TestProcessorNode,
	initialVal *big.Int,
	topUpVal *big.Int,
) {
	fmt.Println("Checking sender has initial-topUp val...")
	expectedVal := big.NewInt(0).Set(initialVal)
	expectedVal.Sub(expectedVal, topUpVal)
	accnt, _ := nodeWithCaller.AccntState.GetExistingAccount(CreateAddressFromAddrBytes(nodeWithCaller.OwnAccount.PkTxSignBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedVal, accnt.(*state.Account).Balance)
}

// CheckScTopUp checks if sc received the top-up value
func CheckScTopUp(
	t *testing.T,
	nodeWithSc *TestProcessorNode,
	topUpVal *big.Int,
	scAddressBytes []byte,
) {
	fmt.Println("Checking SC account received topUp val...")
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(CreateAddressFromAddrBytes(scAddressBytes))
	assert.NotNil(t, accnt)
	assert.Equal(t, topUpVal, accnt.(*state.Account).Balance)
}

// CheckScBalanceOf checks the balance of a SC
func CheckScBalanceOf(
	t *testing.T,
	nodeWithSc *TestProcessorNode,
	nodeWithCaller *TestProcessorNode,
	expectedSC *big.Int,
	scAddressBytes []byte,
) {
	fmt.Println("Checking SC.balanceOf...")
	bytesValue, _ := nodeWithSc.ScDataGetter.Get(
		scAddressBytes,
		"balanceOf",
		nodeWithCaller.OwnAccount.PkTxSignBytes,
	)
	retrievedValue := big.NewInt(0).SetBytes(bytesValue)
	fmt.Printf("SC balanceOf returned %d\n", retrievedValue)
	assert.Equal(t, expectedSC, retrievedValue)
}
