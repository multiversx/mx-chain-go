//go:build !race

package localFuncs

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/integrationTests"
	esdtCommon "github.com/multiversx/mx-chain-go/integrationTests/vm/esdt"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/stretchr/testify/assert"
)

func TestESDTLocalMintAndBurnFromSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := esdtCommon.CreateNodesAndPrepareBalances(1)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddress := esdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/local-esdt-and-nft.wasm")

	ESDTLocalMintAndBurnFromSC_RunTestsAndAsserts(t, nodes, nodes[0].OwnAccount, scAddress, idxProposers, nonce, round)
}

func ESDTLocalMintAndBurnFromSC_RunTestsAndAsserts(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	scAddress []byte,
	idxProposers []int,
	nonce uint64,
	round uint64,
) {
	tokenIdentifier := esdtCommon.PrepareFungibleTokensWithLocalBurnAndMintWithIssuerAccount(t, nodes, ownerWallet, scAddress, idxProposers, &nonce, &round)

	txData := []byte("localMint" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(100).Bytes()))
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 2
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 200)

	txData = []byte("localBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(50).Bytes()))
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransactionWithSenderAccount(
		nodes[0],
		nodes,
		big.NewInt(0),
		ownerWallet,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 100)
}

func TestESDTSetRolesAndLocalMintAndBurnFromSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := esdtCommon.CreateNodesAndPrepareBalances(1)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddress := esdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/local-esdt-and-nft.wasm")

	issuePrice := big.NewInt(1000)
	txData := []byte("issueFungibleToken" + "@" + hex.EncodeToString([]byte("TOKEN")) +
		"@" + hex.EncodeToString([]byte("TKR")) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issuePrice,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("TKR")))
	txData = []byte("setLocalRoles" + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + "01" + "@" + "02")
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	txData = []byte("localMint" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(100).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 201)

	txData = []byte("localBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(50).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtCommon.CheckAddressHasTokens(t, scAddress, nodes, []byte(tokenIdentifier), 0, 101)
}

func TestESDTSetTransferRoles(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := esdtCommon.CreateNodesAndPrepareBalances(2)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddress := esdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/use-module.wasm")
	nrRoundsToPropagateMultiShard := 12
	tokenIdentifier := esdtCommon.PrepareFungibleTokensWithLocalBurnAndMint(t, nodes, scAddress, idxProposers, &nonce, &round)

	esdtCommon.SetRoles(nodes, scAddress, []byte(tokenIdentifier), [][]byte{[]byte(core.ESDTRoleTransfer)})

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	destAddress := nodes[1].OwnAccount.Address

	amount := int64(100)
	txData := txDataBuilder.NewBuilder()
	txData.Clear().TransferESDT(tokenIdentifier, amount).Str("forwardPayments").Bytes(destAddress).Str("fund")

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 10, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtCommon.CheckAddressHasTokens(t, destAddress, nodes, []byte(tokenIdentifier), 0, amount)
}

func TestESDTSetTransferRolesForwardAsyncCallFailsIntra(t *testing.T) {
	testESDTWithTransferRoleAndForwarder(t, 1)
}

func TestESDTSetTransferRolesForwardAsyncCallFailsCross(t *testing.T) {
	testESDTWithTransferRoleAndForwarder(t, 2)
}

func testESDTWithTransferRoleAndForwarder(t *testing.T, numShards int) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := esdtCommon.CreateNodesAndPrepareBalances(numShards)

	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddressA := esdtCommon.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/use-module.wasm")
	scAddressB := esdtCommon.DeployNonPayableSmartContractFromNode(t, nodes, 1, idxProposers, &nonce, &round, "../testdata/use-module.wasm")
	nrRoundsToPropagateMultiShard := 12
	tokenIdentifier := esdtCommon.PrepareFungibleTokensWithLocalBurnAndMint(t, nodes, scAddressA, idxProposers, &nonce, &round)

	esdtCommon.SetRoles(nodes, scAddressA, []byte(tokenIdentifier), [][]byte{[]byte(core.ESDTRoleTransfer)})

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	amount := int64(100)
	txData := txDataBuilder.NewBuilder()
	txData.Clear().TransferESDT(tokenIdentifier, amount).Str("forwardPayments").Bytes(scAddressB).Str("depositTokensForAction").Str("fund")

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddressA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 15, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdtCommon.CheckAddressHasTokens(t, scAddressB, nodes, []byte(tokenIdentifier), 0, 0)
	esdtCommon.CheckAddressHasTokens(t, scAddressA, nodes, []byte(tokenIdentifier), 0, 0)
	esdtCommon.CheckAddressHasTokens(t, nodes[0].OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, amount)
}

func TestAsyncCallsAndCallBacksArgumentsIntra(t *testing.T) {
	testAsyncCallAndCallBacksArguments(t, 1)
}

func TestAsyncCallsAndCallBacksArgumentsCross(t *testing.T) {
	testAsyncCallAndCallBacksArguments(t, 2)
}

func testAsyncCallAndCallBacksArguments(t *testing.T, numShards int) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers := esdtCommon.CreateNodesAndPrepareBalances(numShards)
	defer func() {
		for _, n := range nodes {
			n.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	scAddressA := esdtCommon.DeployNonPayableSmartContractFromNode(t, nodes, 0, idxProposers, &nonce, &round, "forwarder.wasm")
	scAddressB := esdtCommon.DeployNonPayableSmartContractFromNode(t, nodes, 1, idxProposers, &nonce, &round, "vault.wasm")

	txData := txDataBuilder.NewBuilder()
	txData.Clear().Func("echo_args_async").Bytes(scAddressB).Str("AA").Str("BB")

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddressA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 15, nonce, round, idxProposers)
	time.Sleep(time.Second)

	callbackArgs := append([]byte("success"), []byte{0}...)
	callbackArgs = append(callbackArgs, []byte("AABB")...)
	checkDataFromAccountAndKey(t, nodes, scAddressA, []byte("callbackStorage"), callbackArgs)

	txData.Clear().Func("echo_args_async").Bytes(scAddressB)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddressA,
		txData.ToString(),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 15, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkDataFromAccountAndKey(t, nodes, scAddressA, []byte("callbackStorage"), append([]byte("success"), []byte{0}...))
}

func checkDataFromAccountAndKey(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	address []byte,
	key []byte,
	expectedData []byte,
) {
	userAcc := esdtCommon.GetUserAccountWithAddress(t, address, nodes)
	val, _, _ := userAcc.RetrieveValue(key)
	assert.Equal(t, expectedData, val)
}
