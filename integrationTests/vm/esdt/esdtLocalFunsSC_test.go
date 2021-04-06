package esdt

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	testVm "github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/require"
)

func TestESDTLocalMintAndBurnFromSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	advertiser, nodes, idxProposers := createNodesAndPrepareBalances(1)

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

	scAddress := deployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "./testdata/local-esdt-and-nft.wasm")
	tokenIdentifier := prepareFungibleTokensWithLocalBurnAndMint(t, nodes, scAddress, idxProposers, &nonce, &round)

	txData := []byte("localMint" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
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
	nrRoundsToPropagateMultiShard := 2
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkAddressHasESDTTokens(t, scAddress, nodes, tokenIdentifier, 200)

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

	checkAddressHasESDTTokens(t, scAddress, nodes, tokenIdentifier, 100)
}

func TestESDTSetRolesAndLocalMintAndBurnFromSC(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	advertiser, nodes, idxProposers := createNodesAndPrepareBalances(1)

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

	scAddress := deployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round, "./testdata/local-esdt-and-nft.wasm")

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

	checkAddressHasESDTTokens(t, scAddress, nodes, tokenIdentifier, 201)

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

	checkAddressHasESDTTokens(t, scAddress, nodes, tokenIdentifier, 101)
}

func deployNonPayableSmartContract(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	nonce *uint64,
	round *uint64,
	fileName string,
) []byte {
	// deploy Smart Contract which can do local mint and local burn
	scCode := arwen.GetSCCode(fileName)
	scAddress, _ := nodes[0].BlockchainHook.NewAddress(nodes[0].OwnAccount.Address, nodes[0].OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		arwen.CreateDeployTxDataNonPayable(scCode),
		integrationTests.AdditionalGasLimit,
	)

	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, 4, *nonce, *round, idxProposers)
	_, err := nodes[0].AccntState.GetExistingAccount(scAddress)
	require.Nil(t, err)

	return scAddress
}

func prepareFungibleTokensWithLocalBurnAndMint(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	addressWithRoles []byte,
	idxProposers []int,
	round *uint64,
	nonce *uint64,
) string {
	issueTestToken(nodes, 100, "TKN")

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("TKN")))

	setRoles(nodes, addressWithRoles, []byte(tokenIdentifier), [][]byte{[]byte(core.ESDTRoleLocalMint), []byte(core.ESDTRoleLocalBurn)})

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	return tokenIdentifier
}
