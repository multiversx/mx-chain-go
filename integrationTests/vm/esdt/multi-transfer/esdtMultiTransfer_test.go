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

func TestESDTMultiTransferToVault(t *testing.T) {
	logger.ToggleLoggerName(true)
	logger.SetLogLevel("*:INFO,integrationtests:NONE,p2p/libp2p:NONE,process/block:NONE,process/smartcontract:TRACE,process/smartcontract/blockchainhook:NONE")

	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers := esdt.CreateNodesAndPrepareBalances(1)
	tokenIssuerAddress := nodes[0].OwnAccount.Address
	intraShardOpWaitTimeInRounds := 1

	expectedIssuerBalance := make(map[string]int64)
	expectedVaultBalance := make(map[string]int64)

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
	acceptMultiTransferEndpointName := "accept_funds_multi_transfer"

	// issue two fungible tokens
	fungibleTokenIdentifier1 := issueFungibleToken(t, nodes, idxProposers, &nonce, &round, 1000, "FUNG1")
	fungibleTokenIdentifier2 := issueFungibleToken(t, nodes, idxProposers, &nonce, &round, 1000, "FUNG2")

	expectedIssuerBalance[fungibleTokenIdentifier1] = 1000
	expectedIssuerBalance[fungibleTokenIdentifier2] = 1000

	// send a single ESDT with multi-transfer
	txData := buildEsdtMultiTransferTxData(vaultScAddress, []esdtTransfer{
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          100,
		}},
		acceptMultiTransferEndpointName,
	)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		tokenIssuerAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)
	waitForOperationCompletion(t, nodes, idxProposers, intraShardOpWaitTimeInRounds, &nonce, &round)

	expectedIssuerBalance[fungibleTokenIdentifier1] -= 100
	expectedVaultBalance[fungibleTokenIdentifier1] += 100

	esdt.CheckAddressHasESDTTokens(t, tokenIssuerAddress, nodes,
		fungibleTokenIdentifier1, expectedIssuerBalance[fungibleTokenIdentifier1])
	esdt.CheckAddressHasESDTTokens(t, vaultScAddress, nodes,
		fungibleTokenIdentifier1, expectedVaultBalance[fungibleTokenIdentifier1])
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
	endpointName string, arguments ...[]byte) []byte {

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

	return txData.ToBytes()
}

func waitForOperationCompletion(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	roundsToWait int, nonce *uint64, round *uint64) {

	time.Sleep(time.Second)
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, roundsToWait, *nonce, *round, idxProposers)
	time.Sleep(time.Second)
}
