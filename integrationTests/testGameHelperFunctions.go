package integrationTests

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

var gasPriceForGameSC = uint64(0)

// ScCallTxWithParams creates and sends a SC tx call or deploy with all major parameters provided
func ScCallTxWithParams(
	senderNode *TestProcessorNode,
	sk crypto.PrivateKey,
	nonce uint64,
	data string,
	value *big.Int,
	gasLimit uint64,
	gasPrice uint64,
) {

	fmt.Println("Deploying SC...")
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	txArgsObject := &txArgs{
		nonce:    nonce,
		value:    value,
		rcvAddr:  make([]byte, 32),
		sndAddr:  pkBuff,
		data:     data,
		gasLimit: gasLimit,
		gasPrice: gasPrice,
	}

	txDeploy := generateTx(
		sk,
		senderNode.OwnAccount.SingleSigner,
		txArgsObject,
	)

	_, _ = senderNode.SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(StepDelay)
}

// DeployScTx creates and sends a SC tx
func DeployScTx(nodes []*TestProcessorNode, senderIdx int, scCode string, vmType []byte, initArguments string) {
	fmt.Println("Deploying SC...")
	scCodeMetadataString := "0000"
	data := scCode + "@" + hex.EncodeToString(vmType) + "@" + scCodeMetadataString
	if len(initArguments) > 0 {
		data += "@" + initArguments
	}

	txDeploy := generateTx(
		nodes[senderIdx].OwnAccount.SkTxSign,
		nodes[senderIdx].OwnAccount.SingleSigner,
		&txArgs{
			nonce:    nodes[senderIdx].OwnAccount.Nonce,
			value:    big.NewInt(0),
			rcvAddr:  make([]byte, 32),
			sndAddr:  nodes[senderIdx].OwnAccount.PkTxSignBytes,
			data:     data,
			gasLimit: MaxGasLimitPerBlock - 1,
			gasPrice: MinTxGasPrice,
		})
	nodes[senderIdx].OwnAccount.Nonce++
	_, _ = nodes[senderIdx].SendTransaction(txDeploy)
	fmt.Println("Delaying for disseminating the deploy tx...")
	time.Sleep(StepDelay)

	fmt.Println(MakeDisplayTable(nodes))
}

// PlayerJoinsGame creates and sends a join game transaction to the SC
func PlayerJoinsGame(
	nodes []*TestProcessorNode,
	player *TestWalletAccount,
	joinGameVal *big.Int,
	round int32,
	scAddress []byte,
) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address)
	fmt.Println("Calling SC.joinGame...")
	txScCall := generateTx(
		player.SkTxSign,
		player.SingleSigner,
		&txArgs{
			nonce:    player.Nonce,
			value:    joinGameVal,
			rcvAddr:  scAddress,
			sndAddr:  player.Address,
			data:     fmt.Sprintf("joinGame@00%s", hex.EncodeToString(big.NewInt(0).SetInt64(int64(round)).Bytes())),
			gasLimit: 5000,
			gasPrice: gasPriceForGameSC,
		})
	player.Nonce++
	newBalance := big.NewInt(0)
	newBalance = newBalance.Sub(player.Balance, joinGameVal)
	player.Balance = player.Balance.Set(newBalance)

	fmt.Printf("Join %s\n", hex.EncodeToString(player.Address))
	_, _ = txDispatcherNode.SendTransaction(txScCall)
}

// PlayerSendsTransaction creates and sends a transaction to the SC
func PlayerSendsTransaction(
	nodes []*TestProcessorNode,
	player *TestWalletAccount,
	scAddress []byte,
	value *big.Int,
	txData string,
	gasLimit uint64,
) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address)
	txScCall := generateTx(
		player.SkTxSign,
		player.SingleSigner,
		&txArgs{
			nonce:    player.Nonce,
			value:    value,
			rcvAddr:  scAddress,
			sndAddr:  player.Address,
			data:     txData,
			gasLimit: gasLimit,
			gasPrice: MinTxGasPrice,
		})
	player.Nonce++
	newBalance := big.NewInt(0)
	newBalance = newBalance.Sub(player.Balance, value)
	player.Balance = player.Balance.Set(newBalance)

	_, _ = txDispatcherNode.SendTransaction(txScCall)
}

// NodeCallsRewardAndSend - smart contract owner sends reward transaction
func NodeCallsRewardAndSend(
	nodes []*TestProcessorNode,
	idxNodeOwner int,
	winnerPlayer *TestWalletAccount,
	prize *big.Int,
	round int32,
	scAddress []byte,
) {
	fmt.Println("Calling SC.rewardAndSendToWallet...")
	winnerAddress := winnerPlayer.Address
	txScCall := generateTx(
		nodes[idxNodeOwner].OwnAccount.SkTxSign,
		nodes[idxNodeOwner].OwnAccount.SingleSigner,
		&txArgs{
			nonce:    nodes[idxNodeOwner].OwnAccount.Nonce,
			value:    big.NewInt(0),
			rcvAddr:  scAddress,
			sndAddr:  nodes[idxNodeOwner].OwnAccount.PkTxSignBytes,
			data:     fmt.Sprintf("rewardAndSendToWallet@%X@%s@%s", hex.EncodeToString(big.NewInt(0).SetInt64(int64(round)).Bytes()), hex.EncodeToString(winnerAddress), hex.EncodeToString(prize.Bytes())),
			gasLimit: 30000,
			gasPrice: gasPriceForGameSC,
		})
	nodes[idxNodeOwner].OwnAccount.Nonce++

	newBalance := big.NewInt(0)
	newBalance = newBalance.Sub(nodes[idxNodeOwner].OwnAccount.Balance, prize)
	nodes[idxNodeOwner].OwnAccount.Balance = nodes[idxNodeOwner].OwnAccount.Balance.Set(newBalance)

	newBalance = big.NewInt(0)
	newBalance = newBalance.Add(winnerPlayer.Balance, prize)
	winnerPlayer.Balance = winnerPlayer.Balance.Set(newBalance)

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
			nonce:    nodes[idxNode].OwnAccount.Nonce,
			value:    big.NewInt(0),
			rcvAddr:  scAddress,
			sndAddr:  nodes[idxNode].OwnAccount.PkTxSignBytes,
			data:     fmt.Sprintf("withdraw@00%s", hex.EncodeToString(withdrawValue.Bytes())),
			gasLimit: 5000,
			gasPrice: gasPriceForGameSC,
		})
	nodes[idxNode].OwnAccount.Nonce++
	_, _ = nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(StepDelay)

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
			nonce:    nodes[idxNode].OwnAccount.Nonce,
			value:    topUpValue,
			rcvAddr:  scAddress,
			sndAddr:  nodes[idxNode].OwnAccount.PkTxSignBytes,
			data:     "topUp",
			gasLimit: 5000,
			gasPrice: gasPriceForGameSC,
		})
	nodes[idxNode].OwnAccount.Nonce++
	_, _ = nodes[idxNode].SendTransaction(txScCall)
	fmt.Println("Delaying for disseminating SC call tx...")
	time.Sleep(StepDelay)

	fmt.Println(MakeDisplayTable(nodes))
}
func getNodeWithinSameShardAsPlayer(
	nodes []*TestProcessorNode,
	player []byte,
) *TestProcessorNode {
	nodeWithCaller := nodes[0]
	playerShId := nodeWithCaller.ShardCoordinator.ComputeId(player)
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
	nodeWithCaller := getNodeWithinSameShardAsPlayer(nodes, player.Address)

	fmt.Println("Checking sender has initial-topUp val...")
	accnt, _ := nodeWithCaller.AccntState.GetExistingAccount(player.Address)
	assert.NotNil(t, accnt)
	ok := assert.Equal(t, player.Balance.Uint64(), accnt.(state.UserAccountHandler).GetBalance().Uint64())
	if !ok {
		fmt.Printf("Expected player balance %d Actual player balance %d\n", player.Balance.Uint64(), accnt.(state.UserAccountHandler).GetBalance().Uint64())
	}
}

// CheckBalanceIsDoneCorrectlySCSideAndReturnExpectedVal verifies is smart contract balance is correct according
// to top-up and withdraw and returns the expected value
func CheckBalanceIsDoneCorrectlySCSideAndReturnExpectedVal(
	t *testing.T,
	nodes []*TestProcessorNode,
	idxNodeScExists int,
	topUpVal *big.Int,
	withdraw *big.Int,
	scAddressBytes []byte,
) *big.Int {

	nodeWithSc := nodes[idxNodeScExists]

	fmt.Println("Checking SC account has topUp-withdraw val...")
	accnt, _ := nodeWithSc.AccntState.GetExistingAccount(scAddressBytes)
	assert.NotNil(t, accnt)
	expectedSC := big.NewInt(0).Set(topUpVal)
	expectedSC.Sub(expectedSC, withdraw)
	ok := assert.Equal(t, expectedSC.Uint64(), accnt.(state.UserAccountHandler).GetBalance().Uint64())
	if !ok {
		fmt.Printf("Expected smart contract val %d Actual smart contract val %d\n", expectedSC.Uint64(), accnt.(state.UserAccountHandler).GetBalance().Uint64())
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
	CheckBalanceIsDoneCorrectlySCSideAndReturnExpectedVal(
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
	CheckBalanceIsDoneCorrectlySCSideAndReturnExpectedVal(
		t,
		nodes,
		idxProposer,
		allTopUpValue,
		withdrawValue,
		hardCodedScResultingAddress,
	)
}

// CheckSenderBalanceOkAfterTopUpAndWithdraw checks if sender balance is ok after top-up and withdraw
func CheckSenderBalanceOkAfterTopUpAndWithdraw(
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
	accnt, _ := nodeWithCaller.AccntState.GetExistingAccount(nodeWithCaller.OwnAccount.PkTxSignBytes)
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedSender, accnt.(state.UserAccountHandler).GetBalance())
}

// CheckSenderBalanceOkAfterTopUp checks if sender balance is ok after top-up
func CheckSenderBalanceOkAfterTopUp(
	t *testing.T,
	nodeWithCaller *TestProcessorNode,
	initialVal *big.Int,
	topUpVal *big.Int,
) {
	fmt.Println("Checking sender has initial-topUp val...")
	expectedVal := big.NewInt(0).Set(initialVal)
	expectedVal.Sub(expectedVal, topUpVal)
	accnt, _ := nodeWithCaller.AccntState.GetExistingAccount(nodeWithCaller.OwnAccount.PkTxSignBytes)
	assert.NotNil(t, accnt)
	assert.Equal(t, expectedVal, accnt.(state.UserAccountHandler).GetBalance())
}

// CheckScTopUp checks if sc received the top-up value
func CheckScTopUp(
	t *testing.T,
	nodeWithSc *TestProcessorNode,
	topUpVal *big.Int,
	scAddressBytes []byte,
) {
	fmt.Println("Checking SC account received topUp val...")
	accnt, err := nodeWithSc.AccntState.GetExistingAccount(scAddressBytes)
	assert.Nil(t, err)
	assert.NotNil(t, accnt)
	assert.Equal(t, topUpVal, accnt.(state.UserAccountHandler).GetBalance())
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

	query := process.SCQuery{
		ScAddress: scAddressBytes,
		FuncName:  "balanceOf",
		Arguments: [][]byte{nodeWithCaller.OwnAccount.PkTxSignBytes},
	}

	vmOutput, _ := nodeWithSc.SCQueryService.ExecuteQuery(&query)

	retrievedValue := vmOutput.ReturnData[0]
	fmt.Printf("SC balanceOf returned %d\n", retrievedValue)
	assert.Equal(t, expectedSC, big.NewInt(0).SetBytes(retrievedValue))
}
