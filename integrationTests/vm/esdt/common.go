package esdt

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/integrationTests"
	testVm "github.com/multiversx/mx-chain-go/integrationTests/vm"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/multiversx/mx-chain-go/process"
	vmFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

// GetESDTTokenData -
func GetESDTTokenData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tickerID []byte,
	nonce uint64,
) *esdt.ESDigitalToken {
	accShardID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != accShardID {
			continue
		}

		esdtData, err := node.BlockchainHook.GetESDTToken(address, tickerID, nonce)
		require.Nil(t, err)
		return esdtData
	}

	return &esdt.ESDigitalToken{Value: big.NewInt(0)}
}

// GetUserAccountWithAddress -
func GetUserAccountWithAddress(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
) state.UserAccountHandler {
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				acc, err := helperNode.AccntState.LoadAccount(address)
				require.Nil(t, err)
				return acc.(state.UserAccountHandler)
			}
		}
	}

	return nil
}

// SetRoles -
func SetRoles(nodes []*integrationTests.TestProcessorNode, addrForRole []byte, tokenIdentifier []byte, roles [][]byte) {
	tokenIssuer := nodes[0]

	txData := "setSpecialRole" +
		"@" + hex.EncodeToString(tokenIdentifier) +
		"@" + hex.EncodeToString(addrForRole)

	for _, role := range roles {
		txData += "@" + hex.EncodeToString(role)
	}

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, big.NewInt(0), vm.ESDTSCAddress, txData, core.MinMetaTxExtraGasCost)
}

// DeployNonPayableSmartContract -
func DeployNonPayableSmartContract(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idxProposers []int,
	nonce *uint64,
	round *uint64,
	fileName string,
) []byte {
	return DeployNonPayableSmartContractFromNode(t, nodes, 0, idxProposers, nonce, round, fileName)
}

// DeployNonPayableSmartContractFromNode -
func DeployNonPayableSmartContractFromNode(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	idDeployer int,
	idxProposers []int,
	nonce *uint64,
	round *uint64,
	fileName string,
) []byte {
	scCode := wasm.GetSCCode(fileName)
	scAddress, _ := nodes[idDeployer].BlockchainHook.NewAddress(nodes[idDeployer].OwnAccount.Address, nodes[idDeployer].OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransaction(
		nodes[idDeployer],
		nodes,
		big.NewInt(0),
		testVm.CreateEmptyAddress(),
		wasm.CreateDeployTxDataNonPayable(scCode),
		integrationTests.AdditionalGasLimit,
	)

	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, 4, *nonce, *round, idxProposers)

	scShardID := nodes[0].ShardCoordinator.ComputeId(scAddress)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != scShardID {
			continue
		}
		_, err := node.AccntState.GetExistingAccount(scAddress)
		require.Nil(t, err)
	}

	return scAddress
}

// CheckAddressHasTokens - Works for both fungible and non-fungible, according to nonce
func CheckAddressHasTokens(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	tickerID []byte,
	nonce int64,
	value int64,
) {
	nonceAsBigInt := big.NewInt(nonce)
	valueAsBigInt := big.NewInt(value)

	esdtData := GetESDTTokenData(t, address, nodes, tickerID, uint64(nonce))

	if esdtData == nil {
		esdtData = &esdt.ESDigitalToken{
			Value: big.NewInt(0),
		}
	}
	if esdtData.Value == nil {
		esdtData.Value = big.NewInt(0)
	}

	if valueAsBigInt.Cmp(esdtData.Value) != 0 {
		require.Fail(t, fmt.Sprintf("esdt NFT balance difference. Token %s, nonce %s, expected %s, but got %s",
			tickerID, nonceAsBigInt.String(), valueAsBigInt.String(), esdtData.Value.String()))
	}
}

// CreateNodesAndPrepareBalances -
func CreateNodesAndPrepareBalances(numOfShards int) ([]*integrationTests.TestProcessorNode, []int) {
	nodesPerShard := 1
	numMetachainNodes := 1

	enableEpochs := config.EnableEpochs{
		OptimizeGasUsedInCrossMiniBlocksEnableEpoch: integrationTests.UnreachableEpoch,
		ScheduledMiniBlocksEnableEpoch:              integrationTests.UnreachableEpoch,
		MiniBlockPartialExecutionEnableEpoch:        integrationTests.UnreachableEpoch,
	}

	nodes := integrationTests.CreateNodesWithEnableEpochs(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		enableEpochs,
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard
	integrationTests.DisplayAndStartNodes(nodes)

	return nodes, idxProposers
}

// IssueNFT -
func IssueNFT(nodes []*integrationTests.TestProcessorNode, esdtType string, ticker string) {
	tokenName := "token"
	issuePrice := big.NewInt(1000)

	tokenIssuer := nodes[0]

	txData := txDataBuilder.NewBuilder()

	issueFunc := "issueNonFungible"
	if esdtType == core.SemiFungibleESDT {
		issueFunc = "issueSemiFungible"
	}
	txData.Clear().Func(issueFunc).Str(tokenName).Str(ticker)
	txData.CanFreeze(false).CanWipe(false).CanPause(false).CanTransferNFTCreateRole(true)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), core.MinMetaTxExtraGasCost)
}

// IssueTestToken -
func IssueTestToken(nodes []*integrationTests.TestProcessorNode, initialSupply int64, ticker string) {
	issueTestToken(nodes, initialSupply, ticker, core.MinMetaTxExtraGasCost)
}

// IssueTestTokenWithCustomGas -
func IssueTestTokenWithCustomGas(nodes []*integrationTests.TestProcessorNode, initialSupply int64, ticker string, gas uint64) {
	issueTestToken(nodes, initialSupply, ticker, gas)
}

// IssueTestTokenWithSpecialRoles -
func IssueTestTokenWithSpecialRoles(nodes []*integrationTests.TestProcessorNode, initialSupply int64, ticker string) {
	issueTestTokenWithSpecialRoles(nodes, initialSupply, ticker, core.MinMetaTxExtraGasCost)
}

func issueTestToken(nodes []*integrationTests.TestProcessorNode, initialSupply int64, ticker string, gas uint64) {
	tokenName := "token"
	issuePrice := big.NewInt(1000)

	tokenIssuer := nodes[0]
	txData := txDataBuilder.NewBuilder()
	txData.Clear().IssueESDT(tokenName, ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), gas)
}

func issueTestTokenWithSpecialRoles(nodes []*integrationTests.TestProcessorNode, initialSupply int64, ticker string, gas uint64) {
	tokenName := "token"
	issuePrice := big.NewInt(1000)

	tokenIssuer := nodes[0]
	txData := txDataBuilder.NewBuilder()
	txData.Clear().IssueESDT(tokenName, ticker, initialSupply, 6)
	txData.CanFreeze(true).CanWipe(true).CanPause(true).CanMint(true).CanBurn(true).CanAddSpecialRoles(true)

	integrationTests.CreateAndSendTransaction(tokenIssuer, nodes, issuePrice, vm.ESDTSCAddress, txData.ToString(), gas)
}

// CheckNumCallBacks -
func CheckNumCallBacks(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	expectedNumCallbacks int,
) {

	contractID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != contractID {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  address,
			FuncName:   "callback_data",
			CallerAddr: address,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		require.Nil(t, err)
		require.NotNil(t, vmOutput)
		require.Equal(t, vmOutput.ReturnCode, vmcommon.Ok)
		require.Equal(t, expectedNumCallbacks, len(vmOutput.ReturnData))
	}
}

// CheckSavedCallBackData -
func CheckSavedCallBackData(
	t *testing.T,
	address []byte,
	nodes []*integrationTests.TestProcessorNode,
	callbackIndex int,
	expectedTokenId string,
	expectedPayment *big.Int,
	expectedResultCode vmcommon.ReturnCode,
	expectedArguments [][]byte) {

	contractID := nodes[0].ShardCoordinator.ComputeId(address)
	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != contractID {
			continue
		}

		scQuery := &process.SCQuery{
			ScAddress:  address,
			FuncName:   "callback_data_at_index",
			CallerAddr: address,
			CallValue:  big.NewInt(0),
			Arguments: [][]byte{
				{byte(callbackIndex)},
			},
		}
		vmOutput, err := node.SCQueryService.ExecuteQuery(scQuery)
		require.Nil(t, err)
		require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
		require.GreaterOrEqual(t, 3, len(vmOutput.ReturnData))
		require.Equal(t, []byte(expectedTokenId), vmOutput.ReturnData[0])
		require.Equal(t, expectedPayment.Bytes(), vmOutput.ReturnData[1])
		if expectedResultCode == vmcommon.Ok {
			require.Equal(t, []byte{0x0}, vmOutput.ReturnData[2])
		} else {
			require.Equal(t, []byte{byte(expectedResultCode)}, vmOutput.ReturnData[2])
		}
		require.Equal(t, expectedArguments, vmOutput.ReturnData[3:])
	}
}

// PrepareFungibleTokensWithLocalBurnAndMint -
func PrepareFungibleTokensWithLocalBurnAndMint(
	t *testing.T,
	nodes []*integrationTests.TestProcessorNode,
	addressWithRoles []byte,
	idxProposers []int,
	round *uint64,
	nonce *uint64,
) string {
	IssueTestToken(nodes, 100, "TKN")

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard := 5
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	tokenIdentifier := string(integrationTests.GetTokenIdentifier(nodes, []byte("TKN")))

	SetRoles(nodes, addressWithRoles, []byte(tokenIdentifier), [][]byte{[]byte(core.ESDTRoleLocalMint), []byte(core.ESDTRoleLocalBurn)})

	time.Sleep(time.Second)
	nrRoundsToPropagateMultiShard = 5
	*nonce, *round = integrationTests.WaitOperationToBeDone(t, nodes, nrRoundsToPropagateMultiShard, *nonce, *round, idxProposers)
	time.Sleep(time.Second)

	return tokenIdentifier
}
