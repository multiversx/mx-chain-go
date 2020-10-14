package esdt

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestESDTIssueAndTransactionsOnMultiShardEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

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

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send token issue
	tokenName := "robertWhyNot"
	issuePrice := big.NewInt(1000)
	initalSupply := big.NewInt(10000000000)
	tokenIssuer := nodes[0]
	hexEncodedTrue := hex.EncodeToString([]byte("true"))
	txData := "issue" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(initalSupply.Bytes())
	properties := "@" + hex.EncodeToString([]byte("canFreeze")) + "@" + hexEncodedTrue +
		"@" + hex.EncodeToString([]byte("canWipe")) + "@" + hexEncodedTrue +
		"@" + hex.EncodeToString([]byte("canPause")) + "@" + hexEncodedTrue +
		"@" + hex.EncodeToString([]byte("canMint")) + "@" + hexEncodedTrue +
		"@" + hex.EncodeToString([]byte("canBurn")) + "@" + hexEncodedTrue
	txData += properties
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, initalSupply)

	/////////------ send tx to other nodes
	valueToSend := big.NewInt(100)
	for _, node := range nodes[1:] {
		txData = core.BuiltInFunctionESDTTransfer + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(valueToSend.Bytes())
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData, integrationTests.AdditionalGasLimit)
	}

	mintValue := big.NewInt(10000)
	txData = "mint" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(mintValue.Bytes())
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	finalSupply := big.NewInt(0).Add(initalSupply, mintValue)
	for _, node := range nodes[1:] {
		checkAddressHasESDTTokens(t, node.OwnAccount.Address, nodes, tokenName, valueToSend)
		finalSupply.Sub(finalSupply, valueToSend)
	}

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, finalSupply)

	txData = core.BuiltInFunctionESDTBurn + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(mintValue.Bytes())
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	txData = "freeze" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(nodes[1].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	txData = "wipe" + "@" + hex.EncodeToString([]byte(tokenName)) + "@" + hex.EncodeToString(nodes[2].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	txData = "pause" + "@" + hex.EncodeToString([]byte(tokenName))
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtFrozenData := getESDTTokenData(t, nodes[1].OwnAccount.Address, nodes, tokenName)
	esdtUserMetaData := builtInFunctions.ESDTUserMetadataFromBytes(esdtFrozenData.Properties)
	assert.True(t, esdtUserMetaData.Frozen)

	wipedAcc := getUserAccountWithAddress(t, nodes[2].OwnAccount.Address, nodes)
	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenName)
	retrievedData, _ := wipedAcc.DataTrieTracker().RetrieveValue(tokenKey)
	assert.Equal(t, 0, len(retrievedData))

	systemSCAcc := getUserAccountWithAddress(t, core.SystemAccountAddress, nodes)
	retrievedData, _ = systemSCAcc.DataTrieTracker().RetrieveValue(tokenKey)
	esdtGlobalMetaData := builtInFunctions.ESDTGlobalMetadataFromBytes(retrievedData)
	assert.True(t, esdtGlobalMetaData.Paused)

	finalSupply.Sub(finalSupply, mintValue)
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenName, finalSupply)

	esdtSCAcc := getUserAccountWithAddress(t, vm.ESDTSCAddress, nodes)
	retrievedData, _ = esdtSCAcc.DataTrieTracker().RetrieveValue([]byte(tokenName))
	tokenInSystemSC := &systemSmartContracts.ESDTData{}
	_ = integrationTests.TestMarshalizer.Unmarshal(tokenInSystemSC, retrievedData)
	assert.True(t, tokenInSystemSC.MintedValue.Cmp(big.NewInt(0).Add(initalSupply, mintValue)) == 0)
	assert.True(t, tokenInSystemSC.BurntValue.Cmp(mintValue) == 0)
	assert.True(t, tokenInSystemSC.IsPaused)
}

func getESDTTokenData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
) *builtInFunctions.ESDigitalToken {
	userAcc := getUserAccountWithAddress(t, address, nodes)
	require.False(t, check.IfNil(userAcc))

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenName)
	esdtData, err := getESDTDataFromKey(userAcc, tokenKey)
	assert.Nil(t, err)

	return esdtData
}

func checkAddressHasESDTTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value *big.Int,
) {
	esdtData := getESDTTokenData(t, address, nodes, tokenName)
	assert.Equal(t, esdtData.Value.Cmp(value), 0)
}

func getUserAccountWithAddress(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
) state.UserAccountHandler {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				acc, err := helperNode.AccntState.LoadAccount(address)
				require.Nil(t, err)
				return acc.(state.UserAccountHandler)
			}
		}
	}

	return nil
}

func getESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*builtInFunctions.ESDigitalToken, error) {
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
