package multitransfer

import (
	"math/big"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
)

func TestESDTMultiTransferToVault(t *testing.T) {
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

	// issue two NFT, with multiple NFTCreate
	nonFungibleTokenIdentifier1 := issueNft(t, nodes, idxProposers, &nonce, &round, "NFT1", false)
	nonFungibleTokenIdentifier2 := issueNft(t, nodes, idxProposers, &nonce, &round, "NFT2", false)

	expectedIssuerBalance[nonFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedIssuerBalance[nonFungibleTokenIdentifier2] = make(map[int64]int64)

	expectedVaultBalance[nonFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedVaultBalance[nonFungibleTokenIdentifier2] = make(map[int64]int64)

	for i := int64(1); i <= 10; i++ {
		createNFT(t, nodes, idxProposers, nonFungibleTokenIdentifier1, i, &nonce, &round)
		createNFT(t, nodes, idxProposers, nonFungibleTokenIdentifier2, i, &nonce, &round)

		expectedIssuerBalance[nonFungibleTokenIdentifier1][i] = 1
		expectedIssuerBalance[nonFungibleTokenIdentifier2][i] = 1
	}

	// issue two SFTs, with two NFTCreate for each
	semiFungibleTokenIdentifier1 := issueNft(t, nodes, idxProposers, &nonce, &round, "SFT1", true)
	semiFungibleTokenIdentifier2 := issueNft(t, nodes, idxProposers, &nonce, &round, "SFT2", true)

	expectedIssuerBalance[semiFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedIssuerBalance[semiFungibleTokenIdentifier2] = make(map[int64]int64)

	expectedVaultBalance[semiFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedVaultBalance[semiFungibleTokenIdentifier2] = make(map[int64]int64)

	for i := int64(1); i <= 2; i++ {
		createSFT(t, nodes, idxProposers, semiFungibleTokenIdentifier1, i, 1000, &nonce, &round)
		createSFT(t, nodes, idxProposers, semiFungibleTokenIdentifier2, i, 1000, &nonce, &round)

		expectedIssuerBalance[semiFungibleTokenIdentifier1][i] = 1000
		expectedIssuerBalance[semiFungibleTokenIdentifier2][i] = 1000
	}

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

	// send single NFT
	transfers = []esdtTransfer{
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           1,
			amount:          1,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)

	// send two NFTs, same token ID
	transfers = []esdtTransfer{
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           2,
			amount:          1,
		},
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           3,
			amount:          1,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)

	// send two NFTs, different token ID
	transfers = []esdtTransfer{
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           4,
			amount:          1,
		},
		{
			tokenIdentifier: nonFungibleTokenIdentifier2,
			nonce:           1,
			amount:          1,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)

	// send fours NFTs, two of each different token ID
	transfers = []esdtTransfer{
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           5,
			amount:          1,
		},
		{
			tokenIdentifier: nonFungibleTokenIdentifier2,
			nonce:           2,
			amount:          1,
		},
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           6,
			amount:          1,
		},
		{
			tokenIdentifier: nonFungibleTokenIdentifier2,
			nonce:           3,
			amount:          1,
		}}
	multiTransferToVault(t, nodes, idxProposers,
		vaultScAddress, transfers,
		expectedIssuerBalance, expectedVaultBalance,
		&nonce, &round,
	)
}
