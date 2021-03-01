package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	testVm "github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract/builtInFunctions"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts"
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

	initialVal := int64(10000000000)
	integrationTests.MintAllNodes(nodes, big.NewInt(initialVal))

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// send token issue

	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	txData := txDataBuilder.NewBuilder()

	// send tx to other nodes
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData := txData.Clear().TransferESDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	mintValue := int64(10000)
	txData = txData.Clear().Func("mint").Str(tokenIdentifier).Int64(mintValue)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("freeze").Str(tokenIdentifier).Bytes(nodes[2].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	finalSupply := initialSupply + mintValue
	for _, node := range nodes[1:] {
		checkAddressHasESDTTokens(t, node.OwnAccount.Address, nodes, tokenIdentifier, valueToSend)
		finalSupply = finalSupply - valueToSend
	}

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, finalSupply)

	txData.Clear().BurnESDT(tokenIdentifier, mintValue)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("freeze").Str(tokenIdentifier).Bytes(nodes[1].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("wipe").Str(tokenIdentifier).Bytes(nodes[2].OwnAccount.Address)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	txData.Clear().Func("pause").Str(tokenIdentifier)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtFrozenData := getESDTTokenData(t, nodes[1].OwnAccount.Address, nodes, tokenIdentifier)
	esdtUserMetaData := builtInFunctions.ESDTUserMetadataFromBytes(esdtFrozenData.Properties)
	require.True(t, esdtUserMetaData.Frozen)

	wipedAcc := getUserAccountWithAddress(t, nodes[2].OwnAccount.Address, nodes)
	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenIdentifier)
	retrievedData, _ := wipedAcc.DataTrieTracker().RetrieveValue(tokenKey)
	require.Equal(t, 0, len(retrievedData))

	systemSCAcc := getUserAccountWithAddress(t, core.SystemAccountAddress, nodes)
	retrievedData, _ = systemSCAcc.DataTrieTracker().RetrieveValue(tokenKey)
	esdtGlobalMetaData := builtInFunctions.ESDTGlobalMetadataFromBytes(retrievedData)
	require.True(t, esdtGlobalMetaData.Paused)

	finalSupply = finalSupply - mintValue
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, finalSupply)

	esdtSCAcc := getUserAccountWithAddress(t, vm.ESDTSCAddress, nodes)
	retrievedData, _ = esdtSCAcc.DataTrieTracker().RetrieveValue([]byte(tokenIdentifier))
	tokenInSystemSC := &systemSmartContracts.ESDTData{}
	_ = integrationTests.TestMarshalizer.Unmarshal(tokenInSystemSC, retrievedData)
	require.Zero(t, tokenInSystemSC.MintedValue.Cmp(big.NewInt(initialSupply+mintValue)))
	require.Zero(t, tokenInSystemSC.BurntValue.Cmp(big.NewInt(mintValue)))
	require.True(t, tokenInSystemSC.IsPaused)
}

func TestESDTCallBurnOnANonBurnableToken(t *testing.T) {
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

	// send token issue
	ticker := "ALC"
	issuePrice := big.NewInt(1000)
	initialSupply := int64(10000000000)
	tokenIssuer := nodes[0]
	txData := txDataBuilder.NewBuilder()

	txData.Clear().IssueESDT("aliceToken", ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(false)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// send tx to other nodes
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData.Clear().TransferESDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	finalSupply := initialSupply
	for _, node := range nodes[1:] {
		checkAddressHasESDTTokens(t, node.OwnAccount.Address, nodes, tokenIdentifier, valueToSend)
		finalSupply = finalSupply - valueToSend
	}

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, finalSupply)

	burnValue := int64(77)
	txData.Clear().BurnESDT(tokenIdentifier, burnValue)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)

	time.Sleep(time.Second)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtSCAcc := getUserAccountWithAddress(t, vm.ESDTSCAddress, nodes)
	retrievedData, _ := esdtSCAcc.DataTrieTracker().RetrieveValue([]byte(tokenIdentifier))
	tokenInSystemSC := &systemSmartContracts.ESDTData{}
	_ = integrationTests.TestMarshalizer.Unmarshal(tokenInSystemSC, retrievedData)
	require.Equal(t, initialSupply, tokenInSystemSC.MintedValue.Int64())
	require.Zero(t, tokenInSystemSC.BurntValue.Int64())

	// if everything is ok, the caller should have received the amount of burnt tokens back because canBurn = false
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, finalSupply)
}

func TestESDTIssueFromASmartContractSimulated(t *testing.T) {
	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()
	metaNode := integrationTests.NewTestProcessorNode(1, core.MetachainShardId, 0, integrationTests.GetConnectableAddress(advertiser))
	defer func() {
		_ = advertiser.Close()
		_ = metaNode.Messenger.Close()
	}()

	txData := txDataBuilder.NewBuilder()

	ticker := "RBT"
	issuePrice := big.NewInt(1000)
	initialSupply := big.NewInt(10000000000)
	numDecimals := byte(6)

	txData.Clear().IssueESDT("robertWhyNot", ticker, initialSupply.Int64(), numDecimals)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true).Int(1000)
	scr := &smartContractResult.SmartContractResult{
		Nonce:          0,
		Value:          issuePrice,
		RcvAddr:        vm.ESDTSCAddress,
		SndAddr:        metaNode.OwnAccount.Address,
		Data:           []byte(txData.ToBytes()),
		PrevTxHash:     []byte("hash"),
		OriginalTxHash: []byte("hash"),
		GasLimit:       10000000,
		GasPrice:       1,
		CallType:       vmcommon.AsynchronousCall,
		OriginalSender: metaNode.OwnAccount.Address,
	}

	returnCode, err := metaNode.ScProcessor.ProcessSmartContractResult(scr)
	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, returnCode)

	interimProc, _ := metaNode.InterimProcContainer.Get(block.SmartContractResultBlock)
	mapCreatedSCRs := interimProc.GetAllCurrentFinishedTxs()

	require.Equal(t, len(mapCreatedSCRs), 1)
	for _, addedSCR := range mapCreatedSCRs {
		strings.Contains(string(addedSCR.GetData()), core.BuiltInFunctionESDTTransfer)
	}
}

func TestScSendsEsdtToUserWithMessage(t *testing.T) {
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

	// send token issue
	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// deploy the smart contract

	vaultScCode := arwen.GetSCCode("./testdata/vault.wasm")
	vaultScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(vaultScCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(vaultScAddress)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()

	// feed funds to the vault
	valueToSendToSc := int64(1000)
	txData.Clear().TransferESDT(tokenIdentifier, valueToSendToSc)
	txData.Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vaultScAddress, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply-valueToSendToSc)
	checkAddressHasESDTTokens(t, vaultScAddress, nodes, tokenIdentifier, valueToSendToSc)

	// take them back, with a message
	valueToRequest := valueToSendToSc / 4
	txData.Clear().Func("retrieve_funds").Str(tokenIdentifier).Int64(valueToRequest).Str("ESDT transfer message")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vaultScAddress, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply-valueToSendToSc+valueToRequest)
	checkAddressHasESDTTokens(t, vaultScAddress, nodes, tokenIdentifier, valueToSendToSc-valueToRequest)
}

func TestESDTcallsSC(t *testing.T) {
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

	// send token issue

	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// send tx to other nodes
	txData := txDataBuilder.NewBuilder()
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData.Clear().TransferESDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	numNodesWithoutIssuer := int64(len(nodes) - 1)
	issuerBalance := initialSupply - valueToSend*numNodesWithoutIssuer
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, issuerBalance)
	for i := 1; i < len(nodes); i++ {
		checkAddressHasESDTTokens(t, nodes[i].OwnAccount.Address, nodes, tokenIdentifier, valueToSend)
	}

	// deploy the smart contract
	scCode := arwen.GetSCCode("./testdata/crowdfunding-esdt.wasm")
	scAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(scCode)+"@"+
			hex.EncodeToString(big.NewInt(1000).Bytes())+"@"+
			hex.EncodeToString(big.NewInt(1000).Bytes())+"@"+
			hex.EncodeToString([]byte(tokenIdentifier)),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(scAddress)
	require.Nil(t, err)

	// call sc with esdt
	valueToSendToSc := int64(10)
	for _, node := range nodes {
		txData.Clear().TransferESDT(tokenIdentifier, valueToSendToSc).Str("fund")
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), scAddress, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	scQuery1 := &process.SCQuery{
		ScAddress: scAddress,
		FuncName:  "currentFunds",
		Arguments: [][]byte{},
	}
	vmOutput1, _ := nodes[0].SCQueryService.ExecuteQuery(scQuery1)
	require.Equal(t, big.NewInt(60).Bytes(), vmOutput1.ReturnData[0])

	nodesBalance := valueToSend - valueToSendToSc
	issuerBalance = issuerBalance - valueToSendToSc
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, issuerBalance)
	for i := 1; i < len(nodes); i++ {
		checkAddressHasESDTTokens(t, nodes[i].OwnAccount.Address, nodes, tokenIdentifier, nodesBalance)
	}
}

func TestScCallsScWithEsdtIntraShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
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

	// send token issue
	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// deploy the smart contracts

	vaultCode := arwen.GetSCCode("./testdata/vault.wasm")
	vault, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(vaultCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(vault)
	require.Nil(t, err)

	forwarderCode := arwen.GetSCCode("./testdata/forwarder-raw.wasm")
	forwarder, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(forwarderCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(forwarder)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()

	// call forwarder with esdt, and forwarder automatically calls second sc
	valueToSendToSc := int64(1000)
	txData.TransferESDT(tokenIdentifier, valueToSendToSc)
	txData.Str("forward_async_call_half_payment").Bytes(vault).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIssuerBalance := initialSupply - valueToSendToSc
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, tokenIssuerBalance)
	checkAddressHasESDTTokens(t, forwarder, nodes, tokenIdentifier, valueToSendToSc/2)
	checkAddressHasESDTTokens(t, vault, nodes, tokenIdentifier, valueToSendToSc/2)

	checkNumCallBacks(t, forwarder, nodes, 1)
	checkSavedCallBackData(t, forwarder, nodes, 1, "EGLD", big.NewInt(0), vmcommon.Ok, [][]byte{})

	// call forwarder to ask the second one to send it back some esdt
	valueToRequest := valueToSendToSc / 4
	txData.Clear().Func("forward_async_call").Bytes(vault).Str("retrieve_funds").Str(tokenIdentifier).Int64(valueToRequest)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, tokenIssuerBalance)
	checkAddressHasESDTTokens(t, forwarder, nodes, tokenIdentifier, valueToSendToSc*3/4)
	checkAddressHasESDTTokens(t, vault, nodes, tokenIdentifier, valueToSendToSc/4)

	checkNumCallBacks(t, forwarder, nodes, 2)
	checkSavedCallBackData(t, forwarder, nodes, 2, tokenIdentifier, big.NewInt(valueToRequest), vmcommon.Ok, [][]byte{})

	// call forwarder to ask the second one to execute a method
	valueToTransferWithExecSc := valueToSendToSc / 4
	txData.Clear().TransferESDT(tokenIdentifier, valueToTransferWithExecSc)
	txData.Str("forward_transf_exec").Bytes(vault).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(5 * time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(5 * time.Second)

	tokenIssuerBalance -= valueToTransferWithExecSc
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, tokenIssuerBalance)
	checkAddressHasESDTTokens(t, forwarder, nodes, tokenIdentifier, valueToSendToSc*3/4)
	checkAddressHasESDTTokens(t, vault, nodes, tokenIdentifier, valueToSendToSc/2)

	// call forwarder to ask the second one to execute a method that transfers ESDT twice, with execution
	valueToTransferWithExecSc = valueToSendToSc / 10
	txData.Clear().TransferESDT(tokenIdentifier, valueToTransferWithExecSc)
	txData.Str("forward_transf_exec_twice").Bytes(vault).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(5 * time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(5 * time.Second)

	tokenIssuerBalance -= valueToTransferWithExecSc
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, tokenIssuerBalance)
	checkAddressHasESDTTokens(t, forwarder, nodes, tokenIdentifier, valueToSendToSc*3/4)
	checkAddressHasESDTTokens(t, vault, nodes, tokenIdentifier, valueToSendToSc/2+valueToTransferWithExecSc)
}

func TestCallbackPaymentEgld(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
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

	// send token issue
	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// deploy the smart contracts

	vaultCode := arwen.GetSCCode("./testdata/vault.wasm")
	secondScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(vaultCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(secondScAddress)
	require.Nil(t, err)

	forwarderCode := arwen.GetSCCode("./testdata/forwarder-raw.wasm")
	forwarder, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(forwarderCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(forwarder)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()
	// call first sc with esdt, and first sc automatically calls second sc
	valueToSendToSc := int64(1000)
	txData.Clear().Func("forward_async_call_half_payment").Bytes(secondScAddress).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(valueToSendToSc), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNumCallBacks(t, forwarder, nodes, 1)
	checkSavedCallBackData(t, forwarder, nodes, 1, "EGLD", big.NewInt(0), vmcommon.Ok, [][]byte{})

	// call first sc to ask the second one to send it back some esdt
	valueToRequest := valueToSendToSc / 4
	txData.Clear().Func("forward_async_call").Bytes(secondScAddress).Str("retrieve_funds").Str("EGLD").Int64(valueToRequest)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNumCallBacks(t, forwarder, nodes, 2)
	checkSavedCallBackData(t, forwarder, nodes, 2, "EGLD", big.NewInt(valueToRequest), vmcommon.Ok, [][]byte{})
}

func TestScCallsScWithEsdtCrossShard(t *testing.T) {
	t.Skip("test is not ready yet")

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

	// send token issue

	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// deploy the smart contracts

	vaultCode := arwen.GetSCCode("./testdata/vault.wasm")
	secondScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(vaultCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(secondScAddress)
	require.Nil(t, err)

	forwarderCode := arwen.GetSCCode("./testdata/forwarder-raw.wasm")
	forwarder, _ := nodes[2].BlockchainHook.NewAddress(nodes[2].OwnAccount.Address, nodes[2].OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)
	integrationTests.CreateAndSendTransaction(
		nodes[2],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(forwarderCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[2].AccntState.GetExistingAccount(forwarder)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()

	// call forwarder with esdt, and the forwarder automatically calls second sc
	valueToSendToSc := int64(1000)
	txData.Clear().TransferESDT(tokenIdentifier, valueToSendToSc)
	txData.Str("forward_async_call_half_payment").Bytes(secondScAddress).Str("accept_funds")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply-valueToSendToSc)
	checkAddressHasESDTTokens(t, forwarder, nodes, tokenIdentifier, valueToSendToSc/2)
	checkAddressHasESDTTokens(t, secondScAddress, nodes, tokenIdentifier, valueToSendToSc/2)

	checkNumCallBacks(t, forwarder, nodes, 1)
	checkSavedCallBackData(t, forwarder, nodes, 1, "EGLD", big.NewInt(0), vmcommon.Ok, [][]byte{})

	// call forwarder to ask the second one to send it back some esdt
	valueToRequest := valueToSendToSc / 4
	txData.Clear().Func("forward_async_call").Bytes(secondScAddress)
	txData.Str("retrieve_funds").Str(tokenIdentifier).Int64(valueToRequest)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), forwarder, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, forwarder, nodes, tokenIdentifier, valueToSendToSc*3/4)
	checkAddressHasESDTTokens(t, secondScAddress, nodes, tokenIdentifier, valueToSendToSc/4)

	checkNumCallBacks(t, forwarder, nodes, 2)
	checkSavedCallBackData(t, forwarder, nodes, 1, tokenIdentifier, big.NewInt(valueToSendToSc), vmcommon.Ok, [][]byte{})
}

func TestScCallsScWithEsdtIntraShard_SecondScRefusesPayment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
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

	// send token issue
	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// deploy the smart contracts

	secondScCode := arwen.GetSCCode("./testdata/second-contract.wasm")
	secondScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(secondScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier)),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(secondScAddress)
	require.Nil(t, err)

	firstScCode := arwen.GetSCCode("./testdata/first-contract.wasm")
	firstScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(firstScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier))+"@"+
			hex.EncodeToString(secondScAddress),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	_, err = nodes[0].AccntState.GetExistingAccount(firstScAddress)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()

	// call first sc with esdt, and first sc automatically calls second sc which returns error
	valueToSendToSc := int64(1000)
	txData.Clear().TransferESDT(tokenIdentifier, valueToSendToSc)
	txData.Str("transfer_to_second_contract_rejected")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), firstScAddress, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply-valueToSendToSc)

	esdtData := getESDTTokenData(t, firstScAddress, nodes, tokenIdentifier)
	require.Equal(t, &esdt.ESDigitalToken{Value: big.NewInt(valueToSendToSc)}, esdtData)

	esdtData = getESDTTokenData(t, secondScAddress, nodes, tokenIdentifier)
	require.Equal(t, &esdt.ESDigitalToken{Value: big.NewInt(0)}, esdtData)
}

func TestScCallsScWithEsdtCrossShard_SecondScRefusesPayment(t *testing.T) {
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

	// send token issue

	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	// deploy the smart contracts

	secondScCode := arwen.GetSCCode("./testdata/second-contract.wasm")
	secondScAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(secondScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier)),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(secondScAddress)
	require.Nil(t, err)

	firstScCode := arwen.GetSCCode("./testdata/first-contract.wasm")
	firstScAddress, _ := nodes[2].BlockchainHook.NewAddress(nodes[2].OwnAccount.Address, nodes[2].OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)
	integrationTests.CreateAndSendTransaction(
		nodes[2],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(firstScCode)+"@"+
			hex.EncodeToString([]byte(tokenIdentifier))+"@"+
			hex.EncodeToString(secondScAddress),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err = nodes[2].AccntState.GetExistingAccount(firstScAddress)
	require.Nil(t, err)

	txData := txDataBuilder.NewBuilder()

	// call first sc with esdt, and first sc automatically calls second sc which returns error
	valueToSendToSc := int64(1000)
	txData.Clear().TransferESDT(tokenIdentifier, valueToSendToSc)
	txData.Str("transfer_to_second_contract_rejected")
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), firstScAddress, txData.ToString(), integrationTests.AdditionalGasLimit)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard*2, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply-valueToSendToSc)

	esdtData := getESDTTokenData(t, firstScAddress, nodes, tokenIdentifier)
	require.Equal(t, &esdt.ESDigitalToken{Value: big.NewInt(valueToSendToSc)}, esdtData)

	esdtData = getESDTTokenData(t, secondScAddress, nodes, tokenIdentifier)
	require.Equal(t, &esdt.ESDigitalToken{}, esdtData)
}

func TestESDTMultiTransferFromSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
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

	// send token issue

	initialSupply := int64(10000000000)
	issueTestToken(nodes, initialSupply)
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, initialSupply)

	txData := txDataBuilder.NewBuilder()

	// send tx to other nodes
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData.Clear().TransferESDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	numNodesWithoutIssuer := int64(len(nodes) - 1)
	issuerBalance := initialSupply - valueToSend*numNodesWithoutIssuer
	checkAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes, tokenIdentifier, issuerBalance)
	for i := 1; i < len(nodes); i++ {
		checkAddressHasESDTTokens(t, nodes[i].OwnAccount.Address, nodes, tokenIdentifier, valueToSend)
	}

	// deploy the smart contract
	scCode := arwen.GetSCCode("./testdata/transfer-esdt-hook.wasm")
	scAddress, _ := tokenIssuer.BlockchainHook.NewAddress(tokenIssuer.OwnAccount.Address, tokenIssuer.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxData(scCode),
		integrationTests.AdditionalGasLimit,
	)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(scAddress)
	require.Nil(t, err)

	// call sc with esdt
	valueToSendToSc := int64(100)
	for _, node := range nodes {
		txData.Clear().TransferESDT(tokenIdentifier, valueToSendToSc).Str("acceptEsdt")
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), scAddress, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// transferOne
	for _, node := range nodes {
		txData.Clear().Func("transferEsdtOnce").Bytes(node.OwnAccount.Address).Str(tokenIdentifier).Int(1)
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), scAddress, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// transferMultiple
	for _, node := range nodes {
		txData.Clear().Func("transferEsdtMultiple").Bytes(node.OwnAccount.Address).Str(tokenIdentifier).Int(1).Int(3)
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), scAddress, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 4, nonce, round, idxProposers)
	time.Sleep(time.Second)
}

func issueTestToken(nodes []*integrationTests.TestProcessorNode, initialSupply int64) {
	ticker := "TKN"
	tokenName := "token"
	issuePrice := big.NewInt(1000)

	tokenIssuer := nodes[0]

	txData := txDataBuilder.NewBuilder()

	txData.Clear().IssueESDT(tokenName, ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)
}

func getTokenIdentifier(nodes []*integrationTests.TestProcessorNode) []byte {
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != core.MetachainShardId {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  vm.ESDTSCAddress,
			FuncName:   "getAllESDTTokens",
			CallerAddr: vm.ESDTSCAddress,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		if err != nil || vmOutput == nil || vmOutput.ReturnCode != vmcommon.Ok {
			return nil
		}
		if len(vmOutput.ReturnData) == 0 {
			return nil
		}

		return vmOutput.ReturnData[0]
	}

	return nil
}

func getESDTTokenData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
) *esdt.ESDigitalToken {
	userAcc := getUserAccountWithAddress(t, address, nodes)
	require.False(t, check.IfNil(userAcc))

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + "esdt" + tokenName)
	esdtData, err := getESDTDataFromKey(userAcc, tokenKey)
	require.Nil(t, err)

	return esdtData
}

func checkAddressHasESDTTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	value int64,
) {
	esdtData := getESDTTokenData(t, address, nodes, tokenName)
	bigValue := big.NewInt(value)
	if esdtData.Value.Cmp(bigValue) != 0 {
		require.Fail(t, fmt.Sprintf("esdt balance difference. expected %s, but got %s", bigValue.String(), esdtData.Value.String()))
	}
}

func checkNumCallBacks(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	expectedNumCallbacks int) {

	contractID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != contractID {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  address,
			FuncName:   "callback_data",
			CallerAddr: address,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		require.Nil(t, err)
		require.NotNil(t, vmOutput)
		require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
		require.Equal(t, expectedNumCallbacks, len(vmOutput.ReturnData))
	}
}

func checkSavedCallBackData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	callbackIndex int,
	expectedTokenId string,
	expectedPayment *big.Int,
	expectedResultCode vmcommon.ReturnCode,
	expectedArguments [][]byte) {

	contractID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != contractID {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  address,
			FuncName:   "callback_data_at_index",
			CallerAddr: address,
			CallValue:  big.NewInt(0),
			Arguments: [][]byte{
				{byte(callbackIndex)},
			},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		require.Nil(t, err)
		require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
		require.GreaterOrEqual(t, 3, len(vmOutput.ReturnData))
		require.Equal(t, []byte(expectedTokenId), vmOutput.ReturnData[0])
		require.Equal(t, expectedPayment.Bytes(), vmOutput.ReturnData[1])
		if expectedResultCode == vmcommon.Ok {
			require.Equal(t, []byte{}, vmOutput.ReturnData[2])
		} else {
			require.Equal(t, []byte{byte(expectedResultCode)}, vmOutput.ReturnData[2])
		}
		require.Equal(t, expectedArguments, vmOutput.ReturnData[3:])
	}
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

func getESDTDataFromKey(userAcnt state.UserAccountHandler, key []byte) (*esdt.ESDigitalToken, error) {
	esdtData := &esdt.ESDigitalToken{Value: big.NewInt(0)}
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
