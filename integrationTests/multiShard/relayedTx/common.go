package relayedTx

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// CreateGeneralSetupForRelayTxTest will create the general setup for relayed transactions
func CreateGeneralSetupForRelayTxTest() ([]*integrationTests.TestProcessorNode, []int, []*integrationTests.TestWalletAccount, *integrationTests.TestWalletAccount, p2p.Messenger) {
	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	initialVal := big.NewInt(1000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	numPlayers := 5
	numShards := nodes[0].ShardCoordinator.NumberOfShards()
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		shardId := uint32(i) % numShards
		players[i] = integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, shardId)
	}

	relayerAccount := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, 0)
	integrationTests.MintAllPlayers(nodes, []*integrationTests.TestWalletAccount{relayerAccount}, initialVal)

	return nodes, idxProposers, players, relayerAccount, advertiser
}

// CreateAndSendRelayedAndUserTx will create and send a relayed user transaction
func CreateAndSendRelayedAndUserTx(
	nodes []*integrationTests.TestProcessorNode,
	relayer *integrationTests.TestWalletAccount,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) *transaction.Transaction {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, relayer.Address)

	userTx := createUserTx(player, rcvAddr, value, gasLimit, txData)
	relayedTx := createRelayedTx(txDispatcherNode.EconomicsData, relayer, userTx)

	_, err := txDispatcherNode.SendTransaction(relayedTx)
	if err != nil {
		fmt.Println(err.Error())
	}

	return relayedTx
}

func createUserTx(
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    player.Nonce,
		Value:    big.NewInt(0).Set(value),
		RcvAddr:  rcvAddr,
		SndAddr:  player.Address,
		GasPrice: integrationTests.MinTxGasPrice,
		GasLimit: gasLimit,
		Data:     txData,
		ChainID:  integrationTests.ChainID,
		Version:  integrationTests.MinTransactionVersion,
	}
	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	tx.Signature, _ = player.SingleSigner.Sign(player.SkTxSign, txBuff)
	player.Nonce++
	return tx
}

func createRelayedTx(
	economicsFee process.FeeHandler,
	relayer *integrationTests.TestWalletAccount,
	userTx *transaction.Transaction,
) *transaction.Transaction {

	userTxMarshaled, _ := integrationTests.TestTxSignMarshalizer.Marshal(userTx)
	txData := core.RelayedTransaction + "@" + hex.EncodeToString(userTxMarshaled)
	tx := &transaction.Transaction{
		Nonce:    relayer.Nonce,
		Value:    big.NewInt(0).Set(userTx.Value),
		RcvAddr:  userTx.SndAddr,
		SndAddr:  relayer.Address,
		GasPrice: integrationTests.MinTxGasPrice,
		Data:     []byte(txData),
		ChainID:  userTx.ChainID,
		Version:  userTx.Version,
	}
	gasLimit := economicsFee.ComputeGasLimit(tx)
	tx.GasLimit = userTx.GasLimit + gasLimit

	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	tx.Signature, _ = relayer.SingleSigner.Sign(relayer.SkTxSign, txBuff)
	relayer.Nonce++
	txFee := economicsFee.ComputeTxFee(tx)
	relayer.Balance.Sub(relayer.Balance, txFee)
	relayer.Balance.Sub(relayer.Balance, tx.Value)

	return tx
}

func createAndSendSimpleTransaction(
	nodes []*integrationTests.TestProcessorNode,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, player.Address)

	userTx := createUserTx(player, rcvAddr, value, gasLimit, txData)
	_, err := txDispatcherNode.SendTransaction(userTx)
	if err != nil {
		fmt.Println(err.Error())
	}

}

func getNodeWithinSameShardAsPlayer(
	nodes []*integrationTests.TestProcessorNode,
	player []byte,
) *integrationTests.TestProcessorNode {
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

// GetUserAccount -
func GetUserAccount(
	nodes []*integrationTests.TestProcessorNode,
	address []byte,
) state.UserAccountHandler {
	shardID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == shardID {
			acc, _ := node.AccntState.GetExistingAccount(address)
			if check.IfNil(acc) {
				return nil
			}
			userAcc := acc.(state.UserAccountHandler)
			return userAcc
		}
	}
	return nil
}
