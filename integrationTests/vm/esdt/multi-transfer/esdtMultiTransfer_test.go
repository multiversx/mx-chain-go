package multitransfer

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
)

type esdtTransfer struct {
	tokenIdentifier string
	nonce           int64
	amount          int64
}

func TestFungibleESDTMultiTransferToVault(t *testing.T) {
	logger.ToggleLoggerName(true)
	_ = logger.SetLogLevel("*:INFO,integrationtests:NONE,p2p/libp2p:NONE,process/block:NONE,process/smartcontract:TRACE,process/smartcontract/blockchainhook:NONE")

	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers := esdt.CreateNodesAndPrepareBalances(1)

	expectedIssuerBalance := make(map[string]map[int64]int64)
	expectedVaultBalance := make(map[string]map[int64]int64)

	defer func() {
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

	// deploy vault SC
	vaultScAddress := esdt.DeployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round,
		"../testdata/vault.wasm")

	// issue two fungible tokens
	fungibleTokenIdentifier1 := issueFungibleToken(t, nodes, idxProposers, &nonce, &round, 1000, "FUNG1")
	fungibleTokenIdentifier2 := issueFungibleToken(t, nodes, idxProposers, &nonce, &round, 1000, "FUNG2")

	expectedIssuerBalance[fungibleTokenIdentifier1] = make(map[int64]int64)
	expectedIssuerBalance[fungibleTokenIdentifier2] = make(map[int64]int64)
	expectedVaultBalance[fungibleTokenIdentifier1] = make(map[int64]int64)
	expectedVaultBalance[fungibleTokenIdentifier2] = make(map[int64]int64)

	expectedIssuerBalance[fungibleTokenIdentifier1][0] = 1000
	expectedIssuerBalance[fungibleTokenIdentifier2][0] = 1000

	// send a single ESDT with multi-transfer
	transfers := []esdtTransfer{
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          100,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)

	// send two identical transfers with multi-transfer

	transfers = []esdtTransfer{
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          50,
		},
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          50,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)

	// send two different transfers amounts, same token

	transfers = []esdtTransfer{
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          50,
		},
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          100,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)

	// send two different tokens, same amount

	transfers = []esdtTransfer{
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          100,
		},
		{
			tokenIdentifier: fungibleTokenIdentifier2,
			nonce:           0,
			amount:          100,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)
}

func issueFungibleToken(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	nonce *uint64, round *uint64, initialSupply int64, ticker string) string {

	tokenIssuer := nodes[0]
	nrRoundsToPropagateMultiShard := 15

	esdt.IssueTestToken(nodes, initialSupply, ticker)
	waitForOperationCompletion(t, nodes, idxProposers, nrRoundsToPropagateMultiShard, nonce, round)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	esdt.CheckAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes,
		tokenIdentifier, initialSupply)

	return tokenIdentifier
}

func buildEsdtMultiTransferTxData(receiverAddress []byte, transfers []esdtTransfer,
	endpointName string, arguments ...[]byte) string {

	nrTransfers := len(transfers)

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(receiverAddress)
	txData.Int(nrTransfers)

	for _, transfer := range transfers {
		txData.Str(transfer.tokenIdentifier)
		txData.Int64(transfer.nonce)
		txData.Int64(transfer.amount)
	}

	if len(endpointName) > 0 {
		txData.Str(endpointName)

		for _, arg := range arguments {
			txData.Bytes(arg)
		}
	}

	return txData.ToString()
}

func waitForOperationCompletion(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	roundsToWait int, nonce *uint64, round *uint64) {

	time.Sleep(time.Second)
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, roundsToWait, *nonce, *round, idxProposers)
	time.Sleep(time.Second)
}

func multiTransferToVault(t *testing.T,
	nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	vaultScAddress []byte, transfers []esdtTransfer,
	userBalances map[string]map[int64]int64, scBalances map[string]map[int64]int64,
	nonce *uint64, round *uint64) {

	acceptMultiTransferEndpointName := "accept_funds_multi_transfer"
	tokenIssuerAddress := nodes[0].OwnAccount.Address
	intraShardOpWaitTimeInRounds := 1

	txData := buildEsdtMultiTransferTxData(vaultScAddress,
		transfers,
		acceptMultiTransferEndpointName,
	)

	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		tokenIssuerAddress,
		txData,
		integrationTests.AdditionalGasLimit,
	)
	waitForOperationCompletion(t, nodes, idxProposers, intraShardOpWaitTimeInRounds, nonce, round)

	// update expected balances after transfers
	for _, transfer := range transfers {
		userBalances[transfer.tokenIdentifier][transfer.nonce] -= transfer.amount
		scBalances[transfer.tokenIdentifier][transfer.nonce] += transfer.amount
	}

	// check expected vs actual values
	for _, transfer := range transfers {
		expectedUserBalance := userBalances[transfer.tokenIdentifier][transfer.nonce]
		expectedScBalance := scBalances[transfer.tokenIdentifier][transfer.nonce]

		esdt.CheckAddressHasTokens(t, tokenIssuerAddress, nodes,
			transfer.tokenIdentifier, transfer.nonce, expectedUserBalance)
		esdt.CheckAddressHasTokens(t, vaultScAddress, nodes,
			transfer.tokenIdentifier, transfer.nonce, expectedScBalance)
	}
}
