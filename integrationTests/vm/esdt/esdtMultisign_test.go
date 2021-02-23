package esdt

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"strings"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var vmType = []byte{5, 0}
var emptyAddress = make([]byte, 32)
var log = logger.GetOrCreate("integrationtests/vm/esdt")

func TestESDTTransferWithMultisig(t *testing.T) {
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

	initialVal := big.NewInt(10000000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	multisignContractAddress := deployMultisig(t, nodes, 0, 1, 2)

	time.Sleep(time.Second)
	numRoundsToPropagateIntraShard := 2
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, numRoundsToPropagateIntraShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	//----- issue ESDT token
	initalSupply := big.NewInt(10000000000)
	proposeIssueTokenAndTransferFunds(nodes, multisignContractAddress, initalSupply, 0)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, numRoundsToPropagateIntraShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	actionID := getActionID(t, nodes, multisignContractAddress)
	log.Info("got action ID", "action ID", actionID)

	boardMembersSignActionID(nodes, multisignContractAddress, actionID, 1, 2)

	time.Sleep(time.Second)
	numRoundsToPropagateCrossShard := 10
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, numRoundsToPropagateCrossShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	performActionID(nodes, multisignContractAddress, actionID, 0)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, numRoundsToPropagateCrossShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := getTokenIdentifier(nodes)
	checkAddressHasESDTTokens(t, multisignContractAddress, nodes, string(tokenIdentifier), initalSupply)

	checkCallBackWasSaved(t, nodes, multisignContractAddress)

	//----- transfer ESDT token
	destinationAddress, _ := integrationTests.TestAddressPubkeyConverter.Decode("erd1j25xk97yf820rgdp3mj5scavhjkn6tjyn0t63pmv5qyjj7wxlcfqqe2rw5")
	transferValue := big.NewInt(10)
	proposeTransferToken(nodes, multisignContractAddress, transferValue, 0, destinationAddress)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, numRoundsToPropagateIntraShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	actionID = getActionID(t, nodes, multisignContractAddress)
	log.Info("got action ID", "action ID", actionID)

	boardMembersSignActionID(nodes, multisignContractAddress, actionID, 1, 2)

	time.Sleep(time.Second)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, numRoundsToPropagateCrossShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	performActionID(nodes, multisignContractAddress, actionID, 0)

	time.Sleep(time.Second)
	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, numRoundsToPropagateCrossShard, nonce, round, idxProposers)
	time.Sleep(time.Second)

	expectedBalance := big.NewInt(0).Set(initalSupply)
	expectedBalance.Sub(expectedBalance, transferValue)
	checkAddressHasESDTTokens(t, multisignContractAddress, nodes, string(tokenIdentifier), expectedBalance)
	checkAddressHasESDTTokens(t, destinationAddress, nodes, string(tokenIdentifier), transferValue)
}

func checkCallBackWasSaved(t *testing.T, nodes []*integrationTests.TestProcessorNode, contract []byte) {
	contractID := nodes[0].ShardCoordinator.ComputeId(contract)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != contractID {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  contract,
			FuncName:   "callback_log",
			CallerAddr: contract,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		assert.Nil(t, err)
		assert.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
		assert.Equal(t, 1, len(vmOutput.ReturnData))
	}
}

func deployMultisig(t *testing.T, nodes []*integrationTests.TestProcessorNode, ownerIdx int, proposersIndexes ...int) []byte {
	codeMetaData := &vmcommon.CodeMetadata{
		Payable:     true,
		Upgradeable: false,
		Readable:    true,
	}

	contractBytes, err := ioutil.ReadFile("./testdata/multisig-callback.wasm")
	require.Nil(t, err)
	proposers := make([]string, 0, len(proposersIndexes)+1)
	proposers = append(proposers, hex.EncodeToString(nodes[ownerIdx].OwnAccount.Address))
	for _, proposerIdx := range proposersIndexes {
		walletAddressAsHex := hex.EncodeToString(nodes[proposerIdx].OwnAccount.Address)
		proposers = append(proposers, walletAddressAsHex)
	}

	parameters := []string{
		hex.EncodeToString(contractBytes),
		hex.EncodeToString(vmType),
		hex.EncodeToString(codeMetaData.ToBytes()),
		hex.EncodeToString(big.NewInt(int64(len(proposersIndexes))).Bytes()),
	}
	parameters = append(parameters, proposers...)

	txData := strings.Join(
		parameters,
		"@",
	)

	multisigContractAddress, err := nodes[ownerIdx].BlockchainHook.NewAddress(
		nodes[ownerIdx].OwnAccount.Address,
		nodes[ownerIdx].OwnAccount.Nonce,
		vmType,
	)
	require.Nil(t, err)

	log.Info("multisign contract", "address", integrationTests.TestAddressPubkeyConverter.Encode(multisigContractAddress))
	integrationTests.CreateAndSendTransaction(nodes[ownerIdx], nodes, big.NewInt(0), emptyAddress, txData, 100000)

	return multisigContractAddress
}

func proposeIssueTokenAndTransferFunds(
	nodes []*integrationTests.TestProcessorNode,
	multisignContractAddress []byte,
	initalSupply *big.Int,
	ownerIdx int,
) {
	tokenName := []byte("token")
	issuePrice := big.NewInt(1000)
	ticker := "TKN"
	multisigParams := []string{
		"proposeSCCall",
		hex.EncodeToString(vm.ESDTSCAddress),
		hex.EncodeToString(issuePrice.Bytes()),
	}

	esdtParams := []string{
		hex.EncodeToString([]byte("issue")),
		hex.EncodeToString(tokenName),
		hex.EncodeToString([]byte(ticker)),
		hex.EncodeToString(initalSupply.Bytes()),
		hex.EncodeToString([]byte{6}),
	}

	hexEncodedTrue := hex.EncodeToString([]byte("true"))
	tokenPropertiesParams := []string{
		hex.EncodeToString([]byte("canFreeze")),
		hexEncodedTrue,
		hex.EncodeToString([]byte("canWipe")),
		hexEncodedTrue,
		hex.EncodeToString([]byte("canPause")),
		hexEncodedTrue,
		hex.EncodeToString([]byte("canMint")),
		hexEncodedTrue,
		hex.EncodeToString([]byte("canBurn")),
		hexEncodedTrue,
	}

	params := append(multisigParams, esdtParams...)
	params = append(params, tokenPropertiesParams...)
	txData := strings.Join(params, "@")

	integrationTests.CreateAndSendTransaction(nodes[ownerIdx], nodes, big.NewInt(1000000), multisignContractAddress, "deposit", 100000)
	integrationTests.CreateAndSendTransaction(nodes[ownerIdx], nodes, big.NewInt(0), multisignContractAddress, txData, 100000)
}

func getActionID(t *testing.T, nodes []*integrationTests.TestProcessorNode, multisignContractAddress []byte) []byte {
	node := getSameShardNode(nodes, multisignContractAddress)
	accnt, _ := node.AccntState.LoadAccount(multisignContractAddress)
	_ = accnt

	query := &process.SCQuery{
		ScAddress:  multisignContractAddress,
		FuncName:   "getPendingActionFullInfo",
		CallerAddr: make([]byte, 0),
		CallValue:  big.NewInt(0),
		Arguments:  make([][]byte, 0),
	}

	vmOutput, err := node.SCQueryService.ExecuteQuery(query)
	require.Nil(t, err)
	require.Equal(t, 1, len(vmOutput.ReturnData))

	actionFullInfo := vmOutput.ReturnData[0]

	return actionFullInfo[:4]
}

func getSameShardNode(nodes []*integrationTests.TestProcessorNode, address []byte) *integrationTests.TestProcessorNode {
	shId := nodes[0].ShardCoordinator.ComputeId(address)
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == shId {
			return n
		}
	}

	return nil
}

func boardMembersSignActionID(
	nodes []*integrationTests.TestProcessorNode,
	multisignContractAddress []byte,
	actionID []byte,
	signersIndexes ...int,
) {
	for _, index := range signersIndexes {
		node := nodes[index]
		params := []string{
			"sign",
			hex.EncodeToString(actionID),
		}

		txData := strings.Join(params, "@")
		integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), multisignContractAddress, txData, 100000)
	}
}

func performActionID(
	nodes []*integrationTests.TestProcessorNode,
	multisignContractAddress []byte,
	actionID []byte,
	nodeIndex int,
) {
	node := nodes[nodeIndex]
	params := []string{
		"performAction",
		hex.EncodeToString(actionID),
	}

	txData := strings.Join(params, "@")
	integrationTests.CreateAndSendTransaction(node, nodes, big.NewInt(0), multisignContractAddress, txData, 1500000)
}

func proposeTransferToken(
	nodes []*integrationTests.TestProcessorNode,
	multisignContractAddress []byte,
	transferValue *big.Int,
	ownerIdx int,
	destinationAddress []byte,
) {
	tokenID := getTokenIdentifier(nodes)
	multisigParams := []string{
		"proposeSCCall",
		hex.EncodeToString(destinationAddress),
		"00",
	}

	esdtParams := []string{
		hex.EncodeToString([]byte("ESDTTransfer")),
		hex.EncodeToString(tokenID),
		hex.EncodeToString(transferValue.Bytes()),
	}

	params := append(multisigParams, esdtParams...)
	txData := strings.Join(params, "@")

	integrationTests.CreateAndSendTransaction(nodes[ownerIdx], nodes, big.NewInt(0), multisignContractAddress, txData, 100000)
}
