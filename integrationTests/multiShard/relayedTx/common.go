package relayedTx

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateGeneralSetupForRelayTxTest -
func CreateGeneralSetupForRelayTxTest() ([]*integrationTests.TestProcessorNode, []int, []*integrationTests.TestWalletAccount, *integrationTests.TestWalletAccount, p2p.Messenger) {
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

	maxGasPerBlock := uint64(1500000000)
	for _, node := range nodes {
		node.EconomicsData.SetMaxGasLimitPerBlock(maxGasPerBlock)
	}

	return nodes, idxProposers, players, relayerAccount, advertiser
}

// CreateAndSendRelayedAndUserTx -
func CreateAndSendRelayedAndUserTx(
	nodes []*integrationTests.TestProcessorNode,
	relayer *integrationTests.TestWalletAccount,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) *transaction.Transaction {
	txDispatcherNode := GetNodeWithinSameShardAsPlayer(nodes, player.Address)

	userTx := CreateUserTx(player, rcvAddr, value, gasLimit, txData)
	relayedTx := CreateRelayedTx(txDispatcherNode.EconomicsData, relayer, userTx)

	_, err := txDispatcherNode.SendTransaction(relayedTx)
	if err != nil {
		fmt.Println(err.Error())
	}

	return relayedTx
}

// CreateUserTx -
func CreateUserTx(
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

// CreateRelayedTx -
func CreateRelayedTx(
	feeHandler process.FeeHandler,
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
	gasLimit := feeHandler.ComputeGasLimit(tx)
	tx.GasLimit = userTx.GasLimit + gasLimit

	txBuff, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	tx.Signature, _ = relayer.SingleSigner.Sign(relayer.SkTxSign, txBuff)
	relayer.Nonce++
	txFee := big.NewInt(0).Mul(big.NewInt(0).SetUint64(tx.GasLimit), big.NewInt(0).SetUint64(tx.GasPrice))
	relayer.Balance.Sub(relayer.Balance, txFee)
	relayer.Balance.Sub(relayer.Balance, tx.Value)

	return tx
}

// CreateAndSendSimpleTransaction -
func CreateAndSendSimpleTransaction(
	nodes []*integrationTests.TestProcessorNode,
	player *integrationTests.TestWalletAccount,
	rcvAddr []byte,
	value *big.Int,
	gasLimit uint64,
	txData []byte,
) {
	txDispatcherNode := GetNodeWithinSameShardAsPlayer(nodes, player.Address)

	userTx := CreateUserTx(player, rcvAddr, value, gasLimit, txData)
	_, err := txDispatcherNode.SendTransaction(userTx)
	if err != nil {
		fmt.Println(err.Error())
	}

}

// GetNodeWithinSameShardAsPlayer -
func GetNodeWithinSameShardAsPlayer(
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

// CheckAttestedPublicKeys -
func CheckAttestedPublicKeys(
	t *testing.T,
	node *integrationTests.TestProcessorNode,
	scAddress []byte,
	obfuscatedData []byte,
	userAddress []byte,
) {
	scQuery := node.SCQueryService
	vmOutput, err := scQuery.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "getPublicKey",
		Arguments: [][]byte{obfuscatedData},
	})
	require.Nil(t, err)
	require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
	require.Equal(t, vmOutput.ReturnData[0], userAddress)
}

// CheckSCBalance -
func CheckSCBalance(t *testing.T, node *integrationTests.TestProcessorNode, scAddress []byte, userAddress []byte, balance *big.Int) {
	scQuery := node.SCQueryService
	vmOutput, err := scQuery.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "balanceOf",
		Arguments: [][]byte{userAddress},
	})
	assert.Nil(t, err)
	actualBalance := big.NewInt(0).SetBytes(vmOutput.ReturnData[0])
	assert.Equal(t, actualBalance.Cmp(balance), 0)
}

// CheckPlayerBalances -
func CheckPlayerBalances(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	players []*integrationTests.TestWalletAccount) {
	for _, player := range players {
		userAcc := GetUserAccount(nodes, player.Address)
		assert.Equal(t, player.Balance.String(), userAcc.GetBalance().String())
		assert.Equal(t, player.Nonce, userAcc.GetNonce())
	}
}

// CheckAddressHasESDTTokens -
func CheckAddressHasESDTTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value *big.Int,
) {
	userAcc := GetUserAccount(nodes, address)

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenName)
	esdtData, err := GetESDTDataFromKey(userAcc, tokenKey)
	assert.Nil(t, err)

	assert.Equal(t, esdtData.Value.Cmp(value), 0)
}

// GetESDTDataFromKey -
func GetESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*builtInFunctions.ESDigitalToken, error) {
	esdtData := &builtInFunctions.ESDigitalToken{Value: big.NewInt(0)}
	marshaledData, err := userAcnt.DataTrieTracker().RetrieveValue(key)
	if err != nil {
		return esdtData, nil
	}

	err = integrationTests.TestMarshalizer.Unmarshal(esdtData, marshaledData)
	if err != nil {
		return nil, err
	}

	return esdtData, nil
}
