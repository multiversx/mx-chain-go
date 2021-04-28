package nft

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
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
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
	}

	tokenIdentifier, nftMetaData := prepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nodes[1],
		&round,
		&nonce,
		core.NonFungibleESDT,
		1,
		roles,
	)

	// decrease quantity
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
	nrRoundsToPropagateMultiShard := 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// the token data is removed from trie if the quantity is 0, so we should not find it
	nftMetaData.quantity = 0
	checkNftData(
		t,
		nodes[1].OwnAccount.Address,
		nodes[1].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)
}

func TestESDTSemiFungibleTokenCreateAddAndBurn(t *testing.T) {
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
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTAddQuantity),
		[]byte(core.ESDTRoleNFTBurn),
	}

	initialQuantity := int64(5)
	tokenIdentifier, nftMetaData := prepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nodes[1],
		&round,
		&nonce,
		core.SemiFungibleESDT,
		initialQuantity,
		roles,
	)

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
	nrRoundsToPropagateMultiShard := 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData.quantity += quantityToAdd
	checkNftData(
		t,
		nodes[1].OwnAccount.Address,
		nodes[1].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(
		t,
		nodes[1].OwnAccount.Address,
		nodes[1].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

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
	checkNftData(
		t,
		nodes[1].OwnAccount.Address,
		nodes[1].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)
}

func TestESDTNonFungibleTokenTransferSelfShard(t *testing.T) {
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
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTBurn),
	}
	tokenIdentifier, nftMetaData := prepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nodes[1],
		&round,
		&nonce,
		core.NonFungibleESDT,
		1,
		roles,
	)

	// transfer

	// get a node from the shard
	var nodeInSameShard = nodes[0]
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() == nodes[1].ShardCoordinator.SelfId() &&
			!bytes.Equal(node.OwnAccount.Address, node.OwnAccount.Address) {
			nodeInSameShard = node
			break
		}
	}

	nonceArg := hex.EncodeToString(big.NewInt(0).SetUint64(1).Bytes())
	quantityToTransfer := int64(1)
	quantityToTransferArg := hex.EncodeToString(big.NewInt(quantityToTransfer).Bytes())
	txData := []byte(core.BuiltInFunctionESDTNFTTransfer + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToTransferArg + "@" + hex.EncodeToString(nodeInSameShard.OwnAccount.Address))
	integrationTests.CreateAndSendTransaction(
		nodes[1],
		nodes,
		big.NewInt(0),
		nodes[1].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	// check that the new address owns the NFT
	checkNftData(
		t,
		nodes[1].OwnAccount.Address,
		nodeInSameShard.OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	// check that the creator doesn't has the token data in trie anymore
	nftMetaData.quantity = 0
	checkNftData(
		t,
		nodes[1].OwnAccount.Address,
		nodes[1].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)
}

func TestESDTSemiFungibleTokenTransferCrossShard(t *testing.T) {
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
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	// get a node from a different shard
	var nodeInDifferentShard = nodes[0]
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != nodes[0].ShardCoordinator.SelfId() {
			nodeInDifferentShard = node
			break
		}
	}

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTAddQuantity),
		[]byte(core.ESDTRoleNFTBurn),
	}

	initialQuantity := int64(5)
	tokenIdentifier, nftMetaData := prepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nodeInDifferentShard,
		&round,
		&nonce,
		core.SemiFungibleESDT,
		initialQuantity,
		roles,
	)

	// increase quantity
	nonceArg := hex.EncodeToString(big.NewInt(0).SetUint64(1).Bytes())
	quantityToAdd := int64(4)
	quantityToAddArg := hex.EncodeToString(big.NewInt(quantityToAdd).Bytes())
	txData := []byte(core.BuiltInFunctionESDTNFTAddQuantity + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToAddArg)
	integrationTests.CreateAndSendTransaction(
		nodeInDifferentShard,
		nodes,
		big.NewInt(0),
		nodeInDifferentShard.OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData.quantity += quantityToAdd
	checkNftData(
		t,
		nodeInDifferentShard.OwnAccount.Address,
		nodeInDifferentShard.OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(
		t,
		nodeInDifferentShard.OwnAccount.Address,
		nodeInDifferentShard.OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	// transfer
	quantityToTransfer := int64(4)
	quantityToTransferArg := hex.EncodeToString(big.NewInt(quantityToTransfer).Bytes())
	txData = []byte(core.BuiltInFunctionESDTNFTTransfer + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToTransferArg + "@" + hex.EncodeToString(nodes[0].OwnAccount.Address))
	integrationTests.CreateAndSendTransaction(
		nodeInDifferentShard,
		nodes,
		big.NewInt(0),
		nodeInDifferentShard.OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 11
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData.quantity = initialQuantity + quantityToAdd - quantityToTransfer
	checkNftData(
		t,
		nodeInDifferentShard.OwnAccount.Address,
		nodeInDifferentShard.OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	nftMetaData.quantity = quantityToTransfer
	checkNftData(
		t,
		nodeInDifferentShard.OwnAccount.Address,
		nodes[0].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)
}

func TestESDTSemiFungibleTokenTransferToSystemScAddressShouldReceiveBack(t *testing.T) {
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
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
		[]byte(core.ESDTRoleNFTAddQuantity),
		[]byte(core.ESDTRoleNFTBurn),
	}

	initialQuantity := int64(5)
	tokenIdentifier, nftMetaData := prepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nodes[0],
		&round,
		&nonce,
		core.SemiFungibleESDT,
		initialQuantity,
		roles,
	)

	// increase quantity
	nonceArg := hex.EncodeToString(big.NewInt(0).SetUint64(1).Bytes())
	quantityToAdd := int64(4)
	quantityToAddArg := hex.EncodeToString(big.NewInt(quantityToAdd).Bytes())
	txData := []byte(core.BuiltInFunctionESDTNFTAddQuantity + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToAddArg)
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData.quantity += quantityToAdd
	checkNftData(
		t,
		nodes[0].OwnAccount.Address,
		nodes[0].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(
		t,
		nodes[0].OwnAccount.Address,
		nodes[0].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	// transfer
	quantityToTransfer := int64(4)
	quantityToTransferArg := hex.EncodeToString(big.NewInt(quantityToTransfer).Bytes())
	txData = []byte(core.BuiltInFunctionESDTNFTTransfer + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + nonceArg + "@" + quantityToTransferArg + "@" + hex.EncodeToString(vm.ESDTSCAddress))
	integrationTests.CreateAndSendTransaction(
		nodes[0],
		nodes,
		big.NewInt(0),
		nodes[0].OwnAccount.Address,
		string(txData),
		integrationTests.AdditionalGasLimit,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 11
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData.quantity = 0 // make sure that the ESDT SC address didn't receive the token
	checkNftData(
		t,
		nodes[0].OwnAccount.Address,
		vm.ESDTSCAddress,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)

	nftMetaData.quantity = initialQuantity + quantityToAdd // should have the same quantity like before transferring
	checkNftData(
		t,
		nodes[0].OwnAccount.Address,
		nodes[0].OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		1,
	)
}

func testNFTSendCreateRole(t *testing.T, numOfShards int) {
	nodes, idxProposers := esdt.CreateNodesAndPrepareBalances(numOfShards)

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

	roles := [][]byte{
		[]byte(core.ESDTRoleNFTCreate),
	}

	nftCreator := nodes[0]
	initialQuantity := int64(1)
	tokenIdentifier, nftMetaData := prepareNFTWithRoles(
		t,
		nodes,
		idxProposers,
		nftCreator,
		&round,
		&nonce,
		core.NonFungibleESDT,
		initialQuantity,
		roles,
	)

	nftCreatorShId := nftCreator.ShardCoordinator.ComputeId(nftCreator.OwnAccount.Address)
	nextNftCreator := nodes[1]
	for _, node := range nodes {
		if node.ShardCoordinator.ComputeId(node.OwnAccount.Address) != nftCreatorShId {
			nextNftCreator = node
			break
		}
	}

	// transferNFTCreateRole
	txData := []byte("transferNFTCreateRole" + "@" + hex.EncodeToString([]byte(tokenIdentifier)) +
		"@" + hex.EncodeToString(nftCreator.OwnAccount.Address) +
		"@" + hex.EncodeToString(nextNftCreator.OwnAccount.Address))
	integrationTests.CreateAndSendTransaction(
		nftCreator,
		nodes,
		big.NewInt(0),
		vm.ESDTSCAddress,
		string(txData),
		integrationTests.AdditionalGasLimit+core.MinMetaTxExtraGasCost,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 20
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	createNFT(
		[]byte(tokenIdentifier),
		nextNftCreator,
		nodes,
		nftMetaData,
	)

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 2
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(
		t,
		nextNftCreator.OwnAccount.Address,
		nextNftCreator.OwnAccount.Address,
		nodes,
		tokenIdentifier,
		nftMetaData,
		2,
	)
}

func TestESDTNFTSendCreateRoleInShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testNFTSendCreateRole(t, 1)
}

func TestESDTNFTSendCreateRoleInCrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testNFTSendCreateRole(t, 2)
}

func prepareNFTWithRoles(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	nftCreator *integrationTests.TestProcessorNode,
	round *uint64,
	nonce *uint64,
	esdtType string,
	quantity int64,
	roles [][]byte,
) (string, *nftArguments) {
	esdt.IssueNFT(nodes, esdtType, "SFT")

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 10
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("SFT")))

	//// /////// ----- set special roles
	esdt.SetRoles(nodes, nftCreator.OwnAccount.Address, []byte(tokenIdentifier), roles)

	time.Sleep(time.Second)
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	nftMetaData := nftArguments{
		name:       []byte("nft name"),
		quantity:   quantity,
		royalties:  9000,
		hash:       []byte("hash"),
		attributes: []byte("attr"),
		uri:        [][]byte{[]byte("uri")},
	}
	createNFT([]byte(tokenIdentifier), nftCreator, nodes, &nftMetaData)

	time.Sleep(time.Second)
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, 3, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	checkNftData(
		t,
		nftCreator.OwnAccount.Address,
		nftCreator.OwnAccount.Address,
		nodes,
		tokenIdentifier,
		&nftMetaData,
		1,
	)

	return tokenIdentifier, &nftMetaData
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
	creator []byte,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tokenName string,
	args *nftArguments,
	nonce uint64,
) {
	tokenIdentifierPlusNonce := []byte(tokenName)
	tokenIdentifierPlusNonce = append(tokenIdentifierPlusNonce, big.NewInt(0).SetUint64(nonce).Bytes()...)
	esdtData := esdt.GetESDTTokenData(t, address, nodes, string(tokenIdentifierPlusNonce))

	if args.quantity == 0 {
		require.Nil(t, esdtData.TokenMetaData)
		return
	}

	require.NotNil(t, esdtData.TokenMetaData)
	require.Equal(t, creator, esdtData.TokenMetaData.Creator)
	require.Equal(t, args.uri[0], esdtData.TokenMetaData.URIs[0])
	require.Equal(t, args.attributes, esdtData.TokenMetaData.Attributes)
	require.Equal(t, args.name, esdtData.TokenMetaData.Name)
	require.Equal(t, args.hash, esdtData.TokenMetaData.Hash)
	require.Equal(t, uint32(args.royalties), esdtData.TokenMetaData.Royalties)
	require.Equal(t, big.NewInt(args.quantity).Bytes(), esdtData.Value.Bytes())
}
