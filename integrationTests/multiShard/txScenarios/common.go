package txScenarios

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/state"
)

func createGeneralTestnetForTxTest(
	t *testing.T,
	initialBalance uint64,
) *integrationTests.TestNetwork {
	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()

	net.MintNodeAccountsUint64(initialBalance)

	numPlayers := 10
	net.CreateWallets(numPlayers)
	net.MintWalletsUint64(initialBalance)

	return net
}

func createGeneralSetupForTxTest(initialBalance *big.Int) (
	[]*integrationTests.TestProcessorNode,
	[]int,
	[]*integrationTests.TestWalletAccount,
) {
	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	enableEpochs := config.EnableEpochs{
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:              integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:        integrationTests.UnreachableEpoch,
		ConsensusPropagationChangesEnableEpoch:      integrationTests.UnreachableEpoch,
	}

	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
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

	return nodes, idxProposers, players
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
	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
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
