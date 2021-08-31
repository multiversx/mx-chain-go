package multitransfer

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
)

const NR_ROUNDS_CROSS_SHARD = 15
const NR_ROUNDS_SAME_SHARD = 1

type esdtTransfer struct {
	tokenIdentifier string
	nonce           int64
	amount          int64
}

func issueFungibleToken(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	nonce *uint64, round *uint64, initialSupply int64, ticker string) string {

	tokenIssuer := nodes[0]

	esdt.IssueTestToken(nodes, initialSupply, ticker)
	waitForOperationCompletion(t, nodes, idxProposers, NR_ROUNDS_CROSS_SHARD, nonce, round)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	esdt.CheckAddressHasTokens(t, tokenIssuer.OwnAccount.Address, nodes,
		tokenIdentifier, 0, initialSupply)

	return tokenIdentifier
}

func issueNft(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	nonce *uint64, round *uint64, ticker string, semiFungible bool) string {

	tokenType := core.NonFungibleESDT
	if semiFungible {
		tokenType = core.SemiFungibleESDT
	}

	esdt.IssueNFT(nodes, tokenType, ticker)
	waitForOperationCompletion(t, nodes, idxProposers, NR_ROUNDS_CROSS_SHARD, nonce, round)

	issuerAddress := nodes[0].OwnAccount.Address
	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte(ticker)))

	esdt.SetRoles(nodes, issuerAddress, []byte(tokenIdentifier), [][]byte{
		[]byte("ESDTRoleNFTCreate"),
	})
	waitForOperationCompletion(t, nodes, idxProposers, NR_ROUNDS_CROSS_SHARD, nonce, round)

	return tokenIdentifier
}

func createSFT(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	tokenIdentifier string, createdTokenNonce int64, initialSupply int64,
	nonce *uint64, round *uint64) {

	issuerAddress := nodes[0].OwnAccount.Address

	tokenName := "token"
	royalties := big.NewInt(0)
	hash := "someHash"
	attributes := "cool nft"
	uri := "www.my-cool-nfts.com"

	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionESDTNFTCreate)
	txData.Str(tokenIdentifier)
	txData.Int64(initialSupply)
	txData.Str(tokenName)
	txData.BigInt(royalties)
	txData.Str(hash)
	txData.Str(attributes)
	txData.Str(uri)

	integrationTests.CreateAndSendTransaction(nodes[0],
		nodes,
		big.NewInt(0),
		issuerAddress,
		txData.ToString(),
		integrationTests.AdditionalGasLimit)
	waitForOperationCompletion(t, nodes, idxProposers, NR_ROUNDS_SAME_SHARD, nonce, round)

	esdt.CheckAddressHasTokens(t, issuerAddress, nodes,
		tokenIdentifier, createdTokenNonce, initialSupply)
}

func createNFT(t *testing.T, nodes []*integrationTests.TestProcessorNode, idxProposers []int,
	tokenIdentifier string, createdTokenNonce int64,
	nonce *uint64, round *uint64) {

	createSFT(t, nodes, idxProposers, tokenIdentifier, createdTokenNonce, 1, nonce, round)
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
	waitForOperationCompletion(t, nodes, idxProposers, NR_ROUNDS_SAME_SHARD, nonce, round)

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
