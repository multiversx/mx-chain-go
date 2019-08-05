package integrationTests

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/assert"
)

// DeployScTx creates and sends a SC tx
func DeployScTx(nodes []*TestProcessorNode, senderIdx int, scCode string) {
	fmt.Println("Deploying SC...")
	txDeploy := createTxDeploy(nodes[senderIdx].OwnAccount, scCode)
	_, _ = nodes[senderIdx].SendTransaction(txDeploy)

	fmt.Println(MakeDisplayTable(nodes))
}

func createTxDeploy(twa *TestWalletAccount, scCode string) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  make([]byte, 32),
		SndAddr:  twa.PkTxSignBytes,
		Data:     scCode,
		GasPrice: 0,
		GasLimit: 1000000000000,
	}
	txBuff, _ := TestMarshalizer.Marshal(tx)
	tx.Signature, _ = twa.SingleSigner.Sign(twa.SkTxSign, txBuff)

	return tx
}

func createTxEndGame(twa *TestWalletAccount, round int, scAddress []byte) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		RcvAddr:  scAddress,
		SndAddr:  twa.PkTxSignBytes,
		Data:     fmt.Sprintf("endGame@%d", round),
		GasPrice: 0,
		GasLimit: 1000000000000,
	}
	txBuff, _ := TestMarshalizer.Marshal(tx)
	tx.Signature, _ = twa.SingleSigner.Sign(twa.SkTxSign, txBuff)

	fmt.Printf("End %s\n", hex.EncodeToString(twa.PkTxSignBytes))

	return tx
}

func createTxJoinGame(twa *TestWalletAccount, joinGameVal *big.Int, round int, scAddress []byte) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    joinGameVal,
		RcvAddr:  scAddress,
		SndAddr:  twa.Address.Bytes(),
		Data:     fmt.Sprintf("joinGame@%d", round),
		GasPrice: 0,
		GasLimit: 1000000000000,
	}
	txBuff, _ := TestMarshalizer.Marshal(tx)
	tx.Signature, _ = twa.SingleSigner.Sign(twa.SkTxSign, txBuff)

	fmt.Printf("Join %s\n", hex.EncodeToString(twa.Address.Bytes()))

	return tx
}

// PlayerJoinsGame creates and sends a join game transaction to the SC
func PlayerJoinsGame(
	nodes []*TestProcessorNode,
	player *TestWalletAccount,
	joinGameVal *big.Int,
	round int,
	scAddress []byte,
) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address.Bytes())
	fmt.Println("Calling SC.joinGame...")
	txScCall := createTxJoinGame(player, joinGameVal, round, scAddress)
	_, _ = txDispatcherNode.SendTransaction(txScCall)
}

// NodeEndGame creates and sends an end game transaction to the SC
func NodeEndGame(
	nodes []*TestProcessorNode,
	idxNode int,
	round int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.endGame...")
	txScCall := createTxEndGame(nodes[idxNode].OwnAccount, round, scAddress)
	_, _ = nodes[idxNode].SendTransaction(txScCall)
}

func createTxRewardAndSendToWallet(
	tnOwner *TestWalletAccount,
	winnerAddress []byte,
	prizeVal *big.Int,
	round int,
	scAddress []byte,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  tnOwner.PkTxSignBytes,
		Data:     fmt.Sprintf("rewardAndSendToWallet@%d@%s@%X", round, hex.EncodeToString(winnerAddress), prizeVal),
		GasPrice: 0,
		GasLimit: 100000,
	}
	txBuff, _ := TestMarshalizer.Marshal(tx)
	tx.Signature, _ = tnOwner.SingleSigner.Sign(tnOwner.SkTxSign, txBuff)

	fmt.Printf("Reward %s\n", hex.EncodeToString(winnerAddress))

	return tx
}

// NodeCallsRewardAndSend - smart contract owner sends reward transaction
func NodeCallsRewardAndSend(
	nodes []*TestProcessorNode,
	idxNodeOwner int,
	winnerAddress []byte,
	prize *big.Int,
	round int,
	scAddress []byte,
) {

	fmt.Println("Calling SC.rewardAndSendToWallet...")
	txScCall := createTxRewardAndSendToWallet(nodes[idxNodeOwner].OwnAccount, winnerAddress, prize, round, scAddress)
	_, _ = nodes[idxNodeOwner].SendTransaction(txScCall)
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

// CheckPlayerBalanceTheSameWithBlockchain verifies if player blaance is the same as in the blockchain
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
) {

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
