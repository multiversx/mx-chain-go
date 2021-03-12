package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/require"
)

func TestESDTNonFungibleTokenCreateAndBurn(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	///////////------- send token issue

	issueNFT(nodes, core.NonFungibleESDT)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))

	//// /////// ----- set special role
	setRole(nodes, nodes[1].OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleNFTCreate))
	setRole(nodes, nodes[1].OwnAccount.Address, []byte(tokenIdentifier), []byte(core.ESDTRoleNFTBurn))

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData := nftArguments{
		name:       []byte("nft name"),
		quantity:   1,
		royalties:  9000,
		hash:       []byte("hash"),
		attributes: []byte("attr"),
		uri:        [][]byte{[]byte("uri")},
	}
	createNFT([]byte(tokenIdentifier), nodes[1], nodes, &nftMetaData)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(t, nodes[1].OwnAccount.Address, nodes, tokenIdentifier, &nftMetaData, false)

	// increase quantity
	nonceArg := hex.EncodeToString(big.NewInt(0).SetUint64(1).Bytes())
	quantityToBurn := int64(1)
	quantityToBurnArg := hex.EncodeToString(big.NewInt(quantityToBurn).Bytes())
	txData := []byte(core.BuiltInFunctionESDTNFTBurn + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToBurnArg)
	integrationTests.CreateAndSendTransaction(
		nodes[1],
		nodes,
		big.NewInt(0),
		nodes[1].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// the token data is removed from trie if the quantity is 0, so we should not find it
	checkNftData(t, nodes[1].OwnAccount.Address, nodes, tokenIdentifier, &nftMetaData, true)
}

func TestESDTSemiFungibleTokenCreateAddAndBurn(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 2
	numMetachainNodes := 2

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

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

	///////////------- send token issue

	issueNFT(nodes, core.SemiFungibleESDT)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(getTokenIdentifier(nodes))

	//// /////// ----- set special roles
	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTAddQuantity),
		[]byte(core.ESDTRoleNFTBurn),
	}
	setRoles(nodes, nodes[1].OwnAccount.Address, []byte(tokenIdentifier), roles)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData := nftArguments{
		name:       []byte("nft name"),
		quantity:   5,
		royalties:  9000,
		hash:       []byte("hash"),
		attributes: []byte("attr"),
		uri:        [][]byte{[]byte("uri")},
	}
	createNFT([]byte(tokenIdentifier), nodes[1], nodes, &nftMetaData)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(t, nodes[1].OwnAccount.Address, nodes, tokenIdentifier, &nftMetaData, false)

	// increase quantity
	nonceArg := hex.EncodeToString(big.NewInt(0).SetUint64(1).Bytes())
	quantityToAdd := int64(4)
	quantityToAddArg := hex.EncodeToString(big.NewInt(quantityToAdd).Bytes())
	txData := []byte(core.BuiltInFunctionESDTNFTAddQuantity + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToAddArg)
	integrationTests.CreateAndSendTransaction(
		nodes[1],
		nodes,
		big.NewInt(0),
		nodes[1].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData.quantity += quantityToAdd
	checkNftData(t, nodes[1].OwnAccount.Address, nodes, tokenIdentifier, &nftMetaData, false)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(t, nodes[1].OwnAccount.Address, nodes, tokenIdentifier, &nftMetaData, false)

	// burn quantity
	quantityToBurn := int64(4)
	quantityToBurnArg := hex.EncodeToString(big.NewInt(quantityToBurn).Bytes())
	txData = []byte(core.BuiltInFunctionESDTNFTBurn + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToBurnArg)
	integrationTests.CreateAndSendTransaction(
		nodes[1],
		nodes,
		big.NewInt(0),
		nodes[1].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData.quantity -= quantityToBurn
	checkNftData(t, nodes[1].OwnAccount.Address, nodes, tokenIdentifier, &nftMetaData, false)
}

func issueNFT(nodes []*integrationTests.TestProcessorNode, esdtType string) {
	ticker := "SFT"
	tokenName := "token"
	issuePrice := big.NewInt(1000)

	tokenIssuer := nodes[0]

	txData := txDataBuilder.NewBuilder()

	issueFunc := "issueNonFungible"
	if esdtType == core.SemiFungibleESDT {
		issueFunc = "issueSemiFungible"
	}
	txData.Clear().Func(issueFunc).Str(tokenName).Str(ticker)
	txData.CanFreeze(false).CanWipe(false).CanPause(false)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)
}

type nftArguments struct {
	name       []byte
	quantity   int64
	royalties  int64
	hash       []byte
	attributes []byte
	uri        [][]byte
}

func createNFT(tokenIdentifier []byte, issuer *integrationTests.TestProcessorNode, nodes []*integrationTests.TestProcessorNode, args *nftArguments) {
	txData := fmt.Sprintf("%s@%s@%s@%s@%s@%s@%s@%s@",
		core.BuiltInFunctionESDTNFTCreate,
		hex.EncodeToString(tokenIdentifier),
		hex.EncodeToString(big.NewInt(args.quantity).Bytes()),
		hex.EncodeToString(args.name),
		hex.EncodeToString(big.NewInt(args.royalties).Bytes()),
		hex.EncodeToString(args.hash),
		hex.EncodeToString(args.attributes),
		hex.EncodeToString(args.uri[0]),
	)

	integrationTests.CreateAndSendTransaction(issuer, nodes, big.NewInt(0), issuer.OwnAccount.Address, txData, integrationTests.AdditionalGasLimit)
}

func checkNftData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	args *nftArguments,
	shouldNotFound bool,
) {
	tokenIdentifierPlusNonce := []byte(tokenName)
	tokenIdentifierPlusNonce = append(tokenIdentifierPlusNonce, big.NewInt(0).SetUint64(1).Bytes()...)
	esdtData := getESDTTokenData(t, address, nodes, string(tokenIdentifierPlusNonce))

	if shouldNotFound {
		require.Nil(t, esdtData.TokenMetaData)
		return
	}

	require.NotNil(t, esdtData.TokenMetaData)
	require.Equal(t, address, esdtData.TokenMetaData.Creator)
	require.Equal(t, args.uri[0], esdtData.TokenMetaData.URIs[0])
	require.Equal(t, args.attributes, esdtData.TokenMetaData.Attributes)
	require.Equal(t, args.name, esdtData.TokenMetaData.Name)
	require.Equal(t, args.hash, esdtData.TokenMetaData.Hash)
	require.Equal(t, uint32(args.royalties), esdtData.TokenMetaData.Royalties)
	require.Equal(t, big.NewInt(args.quantity).Bytes(), esdtData.Value.Bytes())
}
