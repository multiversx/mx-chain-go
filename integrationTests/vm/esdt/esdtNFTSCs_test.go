package esdt

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/require"
)

func TestESDTNFTIssueCreateBurnSendViaAsyncViaExecuteOnSC(t *testing.T) {
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

	scAddress := deployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round)

	issuePrice := big.NewInt(1000)
	txData := []byte("nftIssue" + "@" + hex.EncodeToString([]byte("TOKEN")) +
		"@" + hex.EncodeToString([]byte("TKR")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issuePrice,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	txData = []byte("setLocalRoles" + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + "03" + "@" + "05")
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

	txData = []byte("nftCreate" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("name")) +
		"@" + hex.EncodeToString(big.NewInt(10).Bytes()) + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte("abc")) + "@" + hex.EncodeToString([]byte("NFT")))
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

	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 3, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 1, big.NewInt(1))

	txData = []byte("nftBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()))
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

	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 1, big.NewInt(0))

	destinationAddress := nodes[0].OwnAccount.Address
	txData = []byte("transferNftViaAsyncCall" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(2).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("transfer_nft_and_execute" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(3).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
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

	checkAddressHasNft(t, scAddress, destinationAddress, nodes, tokenIdentifier, 2, big.NewInt(1))
	checkAddressHasNft(t, scAddress, destinationAddress, nodes, tokenIdentifier, 3, big.NewInt(1))
	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 2, big.NewInt(0))
	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 3, big.NewInt(0))
}

func TestESDTSemiFTIssueCreateBurnSendViaAsyncViaExecuteOnSC(t *testing.T) {
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

	scAddress := deployNonPayableSmartContract(t, nodes, idxProposers, &nonce, &round)

	issuePrice := big.NewInt(1000)
	txData := []byte("sftIssue" + "@" + hex.EncodeToString([]byte("TOKEN")) +
		"@" + hex.EncodeToString([]byte("TKR")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		issuePrice,
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))
	txData = []byte("setLocalRoles" + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + "03" + "@" + "04" + "@" + "05")
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

	txData = []byte("nftCreate" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString([]byte("name")) +
		"@" + hex.EncodeToString(big.NewInt(10).Bytes()) + "@" + hex.EncodeToString(scAddress) +
		"@" + hex.EncodeToString([]byte("abc")) + "@" + hex.EncodeToString([]byte("NFT")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("nftAddQuantity" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(big.NewInt(10).Bytes()))
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

	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 1, big.NewInt(11))

	txData = []byte("nftBurn" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(big.NewInt(1).Bytes()) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()))
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

	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 1, big.NewInt(9))

	destinationAddress := nodes[0].OwnAccount.Address
	txData = []byte("transferNftViaAsyncCall" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(5).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		scAddress,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	txData = []byte("transfer_nft_and_execute" + "@" + hex.EncodeToString(destinationAddress) +
		"@" + hex.EncodeToString([]byte(tokenIdentifier)) + "@" + hex.EncodeToString(big.NewInt(1).Bytes()) +
		"@" + hex.EncodeToString(big.NewInt(4).Bytes()) + "@" + hex.EncodeToString([]byte("NFT")))
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

	checkAddressHasNft(t, scAddress, destinationAddress, nodes, tokenIdentifier, 1, big.NewInt(9))
	checkAddressHasNft(t, scAddress, scAddress, nodes, tokenIdentifier, 1, big.NewInt(0))
}

func checkAddressHasNft(
	t *testing.T,
	creator []byte,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	nonce uint64,
	quantity *big.Int,
) {
	tokenIdentifierPlusNonce := []byte(tokenName)
	tokenIdentifierPlusNonce = append(tokenIdentifierPlusNonce, big.NewInt(0).SetUint64(nonce).Bytes()...)
	esdtData := getESDTTokenData(t, address, nodes, string(tokenIdentifierPlusNonce))

	if quantity.Cmp(big.NewInt(0)) == 0 {
		require.Nil(t, esdtData.TokenMetaData)
		return
	}

	require.NotNil(t, esdtData.TokenMetaData)
	require.Equal(t, creator, esdtData.TokenMetaData.Creator)
	require.Equal(t, quantity.Bytes(), esdtData.Value.Bytes())
}
