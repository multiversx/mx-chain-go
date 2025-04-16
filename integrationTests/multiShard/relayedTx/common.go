package relayedTx

import (
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("relayedtests")

// CreateGeneralSetupForRelayTxTest will create the general setup for relayed transactions
func CreateGeneralSetupForRelayTxTest(baseCostFixEnabled bool) ([]*integrationTests.TestProcessorNode, []*integrationTests.TestProcessorNode, []*integrationTests.TestWalletAccount, *integrationTests.TestWalletAccount) {
	initialVal := big.NewInt(10000000000)
	epochsConfig := integrationTests.GetDefaultEnableEpochsConfig()
	if !baseCostFixEnabled {
		epochsConfig.FixRelayedBaseCostEnableEpoch = integrationTests.UnreachableEpoch
		epochsConfig.FixRelayedMoveBalanceToNonPayableSCEnableEpoch = integrationTests.UnreachableEpoch
	}
	nodes, leaders := createAndMintNodes(initialVal, epochsConfig)

	players, relayerAccount := createAndMintPlayers(baseCostFixEnabled, nodes, initialVal)

	return nodes, leaders, players, relayerAccount
}

func createAndMintNodes(initialVal *big.Int, enableEpochsConfig *config.EnableEpochs) ([]*integrationTests.TestProcessorNode, []*integrationTests.TestProcessorNode) {
	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodesWithEnableEpochsConfig(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochsConfig,
	)

	leaders := make([]*integrationTests.TestProcessorNode, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		leaders[i] = nodes[i*nodesPerShard]
	}
	leaders[numOfShards] = nodes[numOfShards*nodesPerShard]

	integrationTests.DisplayAndStartNodes(nodes)

	integrationTests.MintAllNodes(nodes, initialVal)

	return nodes, leaders
}

func createAndMintPlayers(
	intraShard bool,
	nodes []*integrationTests.TestProcessorNode,
	initialVal *big.Int,
) ([]*integrationTests.TestWalletAccount, *integrationTests.TestWalletAccount) {
	relayerShard := uint32(0)
	numPlayers := 5
	numShards := nodes[0].ShardCoordinator.NumberOfShards()
	players := make([]*integrationTests.TestWalletAccount, numPlayers)
	for i := 0; i < numPlayers; i++ {
		shardId := uint32(i) % numShards
		if intraShard {
			shardId = relayerShard
		}
		players[i] = integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, shardId)
	}

	relayerAccount := integrationTests.CreateTestWalletAccount(nodes[0].ShardCoordinator, relayerShard)
	integrationTests.MintAllPlayers(nodes, []*integrationTests.TestWalletAccount{relayerAccount}, initialVal)

	return players, relayerAccount
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
) (*transaction.Transaction, *transaction.Transaction) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, relayer.Address)

	userTx := createUserTx(player, rcvAddr, value, gasLimit, txData)
	relayedTx := createRelayedTx(txDispatcherNode.EconomicsData, relayer, userTx)

	_, err := txDispatcherNode.SendTransaction(relayedTx)
	if err != nil {
		log.Error("CreateAndSendRelayedAndUserTx.SendTransaction", "error", err)
	}

	return relayedTx, userTx
}

// CreateAndSendRelayedAndUserTxV2 will create and send a relayed user transaction for relayed v2
func CreateAndSendRelayedAndUserTxV2(
	nodes []*integrationTests.TestProcessorNode,
	relayer *integrationTests.TestWalletAccount,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) (*transaction.Transaction, *transaction.Transaction) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, relayer.Address)

	userTx := createUserTx(player, rcvAddr, value, 0, txData)
	relayedTx := createRelayedTxV2(txDispatcherNode.EconomicsData, relayer, userTx, gasLimit)

	_, err := txDispatcherNode.SendTransaction(relayedTx)
	if err != nil {
		log.Error("CreateAndSendRelayedAndUserTxV2.SendTransaction", "error", err)
	}

	return relayedTx, userTx
}

// CreateAndSendRelayedAndUserTxV3 will create and send a relayed user transaction v3
func CreateAndSendRelayedAndUserTxV3(
	nodes []*integrationTests.TestProcessorNode,
	relayer *integrationTests.TestWalletAccount,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) (*transaction.Transaction, *transaction.Transaction) {
	txDispatcherNode := getNodeWithinSameShardAsPlayer(nodes, relayer.Address)

	relayedTx := createRelayedTxV3(txDispatcherNode.EconomicsData, relayer, player, rcvAddr, value, gasLimit, txData)

	_, err := txDispatcherNode.SendTransaction(relayedTx)
	if err != nil {
		log.Error("CreateAndSendRelayedAndUserTxV3.SendTransaction", "error", err)
	}

	return relayedTx, relayedTx
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
	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
	tx.Signature, _ = player.SingleSigner.Sign(player.SkTxSign, txBuff)
	player.Nonce++
	player.Balance.Sub(player.Balance, value)
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
		Value:    big.NewInt(0),
		RcvAddr:  userTx.SndAddr,
		SndAddr:  relayer.Address,
		GasPrice: integrationTests.MinTxGasPrice,
		Data:     []byte(txData),
		ChainID:  userTx.ChainID,
		Version:  userTx.Version,
	}
	gasLimit := economicsFee.ComputeGasLimit(tx)
	tx.GasLimit = userTx.GasLimit + gasLimit

	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
	tx.Signature, _ = relayer.SingleSigner.Sign(relayer.SkTxSign, txBuff)
	relayer.Nonce++

	relayer.Balance.Sub(relayer.Balance, tx.Value)

	txFee := economicsFee.ComputeTxFee(tx)
	relayer.Balance.Sub(relayer.Balance, txFee)

	return tx
}

func createRelayedTxV2(
	economicsFee process.FeeHandler,
	relayer *integrationTests.TestWalletAccount,
	userTx *transaction.Transaction,
	gasLimitForUserTx uint64,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:    relayer.Nonce,
		Value:    big.NewInt(0).Set(userTx.Value),
		RcvAddr:  userTx.SndAddr,
		SndAddr:  relayer.Address,
		GasPrice: integrationTests.MinTxGasPrice,
		Data:     integrationTests.PrepareRelayedTxDataV2(userTx),
		ChainID:  userTx.ChainID,
		Version:  userTx.Version,
	}
	gasLimit := economicsFee.ComputeGasLimit(tx)
	tx.GasLimit = gasLimitForUserTx + gasLimit

	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
	tx.Signature, _ = relayer.SingleSigner.Sign(relayer.SkTxSign, txBuff)
	relayer.Nonce++

	relayer.Balance.Sub(relayer.Balance, tx.Value)

	txFee := economicsFee.ComputeTxFee(tx)
	relayer.Balance.Sub(relayer.Balance, txFee)

	return tx
}

func createRelayedTxV3(
	economicsFee process.FeeHandler,
	relayer *integrationTests.TestWalletAccount,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) *transaction.Transaction {
	tx := &transaction.Transaction{
		Nonce:       player.Nonce,
		Value:       big.NewInt(0).Set(value),
		RcvAddr:     rcvAddr,
		SndAddr:     player.Address,
		GasPrice:    integrationTests.MinTxGasPrice,
		GasLimit:    gasLimit + integrationTests.MinTxGasLimit,
		Data:        txData,
		ChainID:     integrationTests.ChainID,
		Version:     integrationTests.MinTransactionVersion,
		RelayerAddr: relayer.Address,
	}
	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer, integrationTests.TestTxSignHasher)
	tx.Signature, _ = player.SingleSigner.Sign(player.SkTxSign, txBuff)
	tx.RelayerSignature, _ = relayer.SingleSigner.Sign(relayer.SkTxSign, txBuff)

	player.Nonce++
	player.Balance.Sub(player.Balance, value)

	txFee := economicsFee.ComputeTxFee(tx)
	relayer.Balance.Sub(relayer.Balance, txFee)

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
		log.Error("createAndSendSimpleTransaction.SendTransaction", "error", err)
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
