package multitransfer

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
)

type EsdtTransfer struct {
	tokenIdentifier string
	nonce           uint64
	amount          int64
}

func TestESDTMultiTransferToVault(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	nodes, idxProposers := esdt.CreateNodesAndPrepareBalances(1)
	tokenIssuerAddress := nodes[0].OwnAccount.Address

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
	txData := buildEsdtMultiTransferTxData(vaultScAddress, []EsdtTransfer{
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

	esdt.IssueTestToken(nodes, initialSupply, ticker)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 15
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	esdt.CheckAddressHasESDTTokens(t, tokenIssuer.OwnAccount.Address, nodes,
		tokenIdentifier, initialSupply)

	return tokenIdentifier
}

func buildEsdtMultiTransferTxData(receiverAddress []byte, transfers []EsdtTransfer,
	endpointName string, arguments ...[]byte) []byte {

	nrTransfers := len(transfers)
	txData := []byte(core.BuiltInFunctionMultiESDTNFTTransfer +
		"@" + hex.EncodeToString(receiverAddress) +
		"@" + hex.EncodeToString(big.NewInt(int64(nrTransfers)).Bytes()))

	for _, transfer := range transfers {
		txData = append(txData, []byte("@"+hex.EncodeToString([]byte(transfer.tokenIdentifier))+
			"@"+hex.EncodeToString(big.NewInt(int64(transfer.nonce)).Bytes())+
			"@"+hex.EncodeToString(big.NewInt(transfer.amount).Bytes()))...)
	}

	if len(endpointName) > 0 {
		txData = append(txData, []byte(endpointName)...)

		for _, arg := range arguments {
			txData = append(txData, arg...)
		}
	}

	return txData
}
