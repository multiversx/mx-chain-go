//go:build !race

package roles

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/esdt"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
)

// Test scenario
// 1 - issue an ESDT token
// 2 - set special roles for the owner of the token (local burn and local mint)
// 3 - do a local burn - should work
// 4 - do a local mint - should work
func TestESDTRolesIssueAndTransactionsOnMultiShardEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// ------- send token issue

	initialSupply := big.NewInt(10000000000)
	esdt.IssueTestToken(nodes, initialSupply.Int64(), "FTT")
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 6
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("FTT")))

	// ----- set special role
	setRole(nodes, nodes[0].OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleLocalMint))
	setRole(nodes, nodes[0].OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleLocalBurn))

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply.Int64())

	// mint local new tokens
	txData := []byte(core.BuiltInFunctionESDTLocalMint + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(500).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// check balance ofter local mint
	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, big.NewInt(10000000500).Int64())

	// burn local  tokens
	txData = []byte(core.BuiltInFunctionESDTLocalBurn + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(200).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// check balance ofter local mint
	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, big.NewInt(10000000300).Int64())
}

// Test scenario
// 1 - issue an ESDT token
// 2 - set special role for the owner of the token (local mint)
// 3 - unset special role (local mint)
// 3 - do a local mint - ESDT balance should not change
// 4 - set special role (local burn)
// 5 - do a local burn - should work
func TestESDTRolesSetRolesAndUnsetRolesIssueAndTransactionsOnMultiShardEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	// ------- send token issue

	initialSupply := big.NewInt(10000000000)
	esdt.IssueTestToken(nodes, initialSupply.Int64(), "FTT")
	tokenIssuer := nodes[0]

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("FTT")))

	// ----- set special role
	setRole(nodes, nodes[0].OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleLocalMint))

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// unset special role
	unsetRole(nodes, nodes[0].OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleLocalMint))

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply.Int64())
	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply.Int64())

	// mint local new tokens
	txData := []byte(core.BuiltInFunctionESDTLocalMint + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(500).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 7
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// check balance ofter local mint
	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, big.NewInt(10000000000).Int64())

	setRole(nodes, nodes[0].OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleLocalBurn))
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// burn local  tokens
	txData = []byte(core.BuiltInFunctionESDTLocalBurn + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(200).Bytes()))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// check balance ofter local mint
	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, big.NewInt(9999999800).Int64())
}

func setRole(nodes []*integrationTests.TestProcessorNode, addrForRole []byte, tokenIdentifier []byte, roles []byte) {
	tokenIssuer := nodes[0]

	txData := "setSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(addrForRole) +
		"@" + hex.EncodeToString(roles)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)
}

func unsetRole(nodes []*integrationTests.TestProcessorNode, addrForRole []byte, tokenIdentifier []byte, roles []byte) {
	tokenIssuer := nodes[0]

	txData := "unSetSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(addrForRole) +
		"@" + hex.EncodeToString(roles)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)
}

func TestESDTMintTransferAndExecute(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	scAddress := esdt.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "../testdata/egld-esdt-swap.wasm")

	// issue ESDT by calling exec on dest context on child contract
	ticker := "DSN"
	name := "DisplayName"
	issueCost := big.NewInt(1000)
	txIssueData := txDataBuilder.NewBuilder()
	txIssueData.Func("issueWrappedEgld").
		Str(name).
		Str(ticker)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issueCost,
		scAddress,
		txIssueData.ToString(),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 15
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := integrationTests.GetTokenIdentifier(nodes, []byte(ticker))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		"setLocalRoles",
		integrationTests.AdditionalGasLimit,
	)
	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	valueToWrap := big.NewInt(1000)
	for _, n := range nodes {
		txData := []byte("wrapEgld")
		integrationTests.CreateAndSendTransaction(
			n,
			nodes,
			valueToWrap,
			scAddress,
			string(txData),
			integrationTests.AdditionalGasLimit,
		)
	}

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	for i, n := range nodes {
		if i == 0 {
			continue
		}
		esdt.CheckAddressHasTokens(t, n.OwnAccount.Address, nodes, tokenIdentifier, 0, valueToWrap.Int64())
	}

	for _, n := range nodes {
		txUnWrap := txDataBuilder.NewBuilder()
		txUnWrap.Func(core.BuiltInFunctionESDTTransfer).Str(string(tokenIdentifier)).BigInt(valueToWrap).Str("unwrapEgld")
		integrationTests.CreateAndSendTransaction(
			n,
			nodes,
			big.NewInt(0),
			scAddress,
			txUnWrap.ToString(),
			integrationTests.AdditionalGasLimit,
		)
	}
	time.Sleep(time.Second)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	userAccount := esdt.GetUserAccountWithAddress(t, scAddress, nodes)
	require.Equal(t, userAccount.GetBalance(), big.NewInt(0))
}

func TestESDTLocalBurnFromAnyoneOfThisToken(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	enableEpochs := config.EnableEpochs{
		ScheduledMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
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
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

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
		esdt.CheckAddressHasTokens(t, node.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, valueToSend)
		finalSupply = finalSupply - valueToSend
		txData.Clear().LocalBurnESDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, finalSupply)
	txData.Clear().LocalBurnESDT(tokenIdentifier, finalSupply)
	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), tokenIssuer.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	for _, node := range nodes {
		esdt.CheckAddressHasTokens(t, node.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, 0)
	}
}

func TestESDTWithTransferRoleCrossShardShouldWork(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	enableEpochs := config.EnableEpochs{
		ScheduledMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
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
	nrRoundsToPropagateMultiShard := 12
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))
	setRole(nodes, tokenIssuer.OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleTransfer))
	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 2, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// send tx to other nodes
	valueToSend := int64(100)
	for _, node := range nodes[1:] {
		txData.Clear().TransferESDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), node.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// send value back to the initial node
	for _, node := range nodes[1:] {
		esdt.CheckAddressHasTokens(t, node.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, valueToSend)
		txData.Clear().TransferESDT(tokenIdentifier, valueToSend)
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), tokenIssuer.OwnAccount.Address, txData.ToString(), integrationTests.AdditionalGasLimit)
	}

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	for _, node := range nodes[1:] {
		esdt.CheckAddressHasTokens(t, node.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, 0)
	}
	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes, []byte(tokenIdentifier), 0, initialSupply)
}
