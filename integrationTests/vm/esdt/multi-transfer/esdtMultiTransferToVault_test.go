package multitransfer

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
)

func TestESDTMultiTransferToVaultSameShard(t *testing.T) {
	esdtMultiTransferToVault(t, false)
}

func TestESDTMultiTransferToVaultCrossShard(t *testing.T) {
	esdtMultiTransferToVault(t, true)
}

func esdtMultiTransferToVault(t *testing.T, crossShard bool) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	// For cross shard, we use 2 nodes, with node[1] being the SC deployer, and node[0] being the caller
	numShards := 1
	nrRoundsToWait := numRoundsSameShard

	if crossShard {
		numShards = 2
		nrRoundsToWait = numRoundsCrossShard
	}

	net := integrationTests.NewTestNetworkSized(t, numShards, 1, 1)
	net.Start()
	defer net.Close()

	net.MintNodeAccountsUint64(10000000000)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	if crossShard {
		senderNode = net.NodesSharded[1][0]
	}

	expectedIssuerBalance := make(map[string]map[int64]int64)
	expectedVaultBalance := make(map[string]map[int64]int64)

	// deploy vault SC
	vaultScAddress := deployNonPayableSmartContract(t, net, net.NodesSharded[0][0], "../testdata/vaultV2.wasm")

	// issue two fungible tokens
	fungibleTokenIdentifier1 := issueFungibleToken(t, net, senderNode, "FUNG1", 1000)
	fungibleTokenIdentifier2 := issueFungibleToken(t, net, senderNode, "FUNG2", 1000)

	expectedIssuerBalance[fungibleTokenIdentifier1] = make(map[int64]int64)
	expectedIssuerBalance[fungibleTokenIdentifier2] = make(map[int64]int64)
	expectedVaultBalance[fungibleTokenIdentifier1] = make(map[int64]int64)
	expectedVaultBalance[fungibleTokenIdentifier2] = make(map[int64]int64)

	expectedIssuerBalance[fungibleTokenIdentifier1][0] = 1000
	expectedIssuerBalance[fungibleTokenIdentifier2][0] = 1000

	// issue two NFT, with multiple NFTCreate
	nonFungibleTokenIdentifier1 := issueNft(net, senderNode, "NFT1", false)
	nonFungibleTokenIdentifier2 := issueNft(net, senderNode, "NFT2", false)

	expectedIssuerBalance[nonFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedIssuerBalance[nonFungibleTokenIdentifier2] = make(map[int64]int64)

	expectedVaultBalance[nonFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedVaultBalance[nonFungibleTokenIdentifier2] = make(map[int64]int64)

	for i := int64(1); i <= 10; i++ {
		createNFT(t, net, senderNode, nonFungibleTokenIdentifier1, i)
		createNFT(t, net, senderNode, nonFungibleTokenIdentifier2, i)

		expectedIssuerBalance[nonFungibleTokenIdentifier1][i] = 1
		expectedIssuerBalance[nonFungibleTokenIdentifier2][i] = 1
	}

	// issue two SFTs, with two NFTCreate for each
	semiFungibleTokenIdentifier1 := issueNft(net, senderNode, "SFT1", true)
	semiFungibleTokenIdentifier2 := issueNft(net, senderNode, "SFT2", true)

	expectedIssuerBalance[semiFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedIssuerBalance[semiFungibleTokenIdentifier2] = make(map[int64]int64)

	expectedVaultBalance[semiFungibleTokenIdentifier1] = make(map[int64]int64)
	expectedVaultBalance[semiFungibleTokenIdentifier2] = make(map[int64]int64)

	for i := int64(1); i <= 2; i++ {
		createSFT(t, net, senderNode, semiFungibleTokenIdentifier1, i, 1000)
		createSFT(t, net, senderNode, semiFungibleTokenIdentifier2, i, 1000)

		expectedIssuerBalance[semiFungibleTokenIdentifier1][i] = 1000
		expectedIssuerBalance[semiFungibleTokenIdentifier2][i] = 1000
	}

	// send a single ESDT with multi-transfer
	transfers := []*esdtTransfer{
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          100,
		}}
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send two identical transfers with multi-transfer
	transfers = []*esdtTransfer{
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
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send two different transfers amounts, same token
	transfers = []*esdtTransfer{
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
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send two different tokens, same amount
	transfers = []*esdtTransfer{
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
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send single NFT
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           1,
			amount:          1,
		}}
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send two NFTs, same token ID
	transfers = []*esdtTransfer{
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
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send two NFTs, different token ID
	transfers = []*esdtTransfer{
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
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send fours NFTs, two of each different token ID
	transfers = []*esdtTransfer{
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
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send single SFT
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: semiFungibleTokenIdentifier1,
			nonce:           1,
			amount:          100,
		}}
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send two SFTs, same token ID
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: semiFungibleTokenIdentifier1,
			nonce:           1,
			amount:          100,
		},
		{
			tokenIdentifier: semiFungibleTokenIdentifier1,
			nonce:           2,
			amount:          100,
		}}
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send two SFTs, different token ID
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: semiFungibleTokenIdentifier1,
			nonce:           1,
			amount:          100,
		},
		{
			tokenIdentifier: semiFungibleTokenIdentifier2,
			nonce:           1,
			amount:          100,
		}}
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// send fours SFTs, two of each different token ID
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: semiFungibleTokenIdentifier1,
			nonce:           1,
			amount:          100,
		},
		{
			tokenIdentifier: semiFungibleTokenIdentifier2,
			nonce:           2,
			amount:          100,
		},
		{
			tokenIdentifier: semiFungibleTokenIdentifier1,
			nonce:           2,
			amount:          50,
		},
		{
			tokenIdentifier: semiFungibleTokenIdentifier2,
			nonce:           1,
			amount:          200,
		}}
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)

	// transfer all 3 types
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: fungibleTokenIdentifier1,
			nonce:           0,
			amount:          100,
		},
		{
			tokenIdentifier: semiFungibleTokenIdentifier2,
			nonce:           2,
			amount:          100,
		},
		{
			tokenIdentifier: nonFungibleTokenIdentifier1,
			nonce:           7,
			amount:          1,
		}}
	multiTransferToVault(t, net, senderNode,
		vaultScAddress, transfers, nrRoundsToWait,
		expectedIssuerBalance, expectedVaultBalance,
	)
}
