package txScenarios

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

func createGeneralSetupForTxTest(initialBalance *big.Int) (
	[]*integrationTests.TestProcessorNode,
	[]int,
	[]*integrationTests.TestWalletAccount,
	p2p.Messenger,
) {
	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

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

	integrationTests.MintAllNodes(nodes, initialBalance)

	numPlayers := 10
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		shardId := uint32(i) % nodes[0].ShardCoordinator.NumberOfShards()
		players[i] = integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, shardId)
	}

	integrationTests.MintAllPlayers(nodes, players, initialBalance)

	return nodes, idxProposers, players, advertiser
}

func createAndSendTransaction(
	senderNode *integrationTests.TestProcessorNode,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	txData []byte,
	gasPrice uint64,
	gasLimit uint64,
) *transaction.Transaction {
	userTx := createUserTx(player, rcvAddr, value, txData, gasPrice, gasLimit)

	_, err := senderNode.SendTransaction(userTx)
	if err != nil {
		fmt.Println(err.Error())
	}

	return userTx
}

func createUserTx(
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	txData []byte,
	gasPrice uint64,
	gasLimit uint64,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    player.Nonce,
		Value:    big.NewInt(0).Set(value),
		RcvAddr:  rcvAddr,
		SndAddr:  player.Address,
		GasPrice: gasPrice,
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

func getUserAccount(
	nodes []*integrationTests.TestProcessorNode,
	address []byte,
) state.UserAccountHandler {
	shardID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == shardID {
			acc, _ := node.AccntState.GetExistingAccount(address)
			userAcc := acc.(state.UserAccountHandler)
			return userAcc
		}
	}
	return nil
}
